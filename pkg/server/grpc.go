package server

import (
	"context"
	"fmt"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator"
	"github.com/smallbiznis/go-lib/pkg/env"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	GrpcServerProvider = fx.Module("grpc.server", fx.Options(
		fx.Provide(
			NewServerOption,
			NewGrpcServer,
		),
	))
	GrpcServerInvoke = fx.Module("grpc.invoke", fx.Options(
		fx.Invoke(func(lc fx.Lifecycle, server *grpc.Server) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					lis, err := net.Listen("tcp", env.Lookup("GRPC_PORT", ":4317"))
					if err != nil {
						return err
					}
					go server.Serve(lis)
					return nil
				},
				OnStop: func(ctx context.Context) error {
					server.GracefulStop()
					return nil
				},
			})
		}),
	))
)

func InterceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

// Unary Server Interceptor - Custom Middleware
func TraceInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Ambil metadata (header) dari konteks
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("metadata is not present in context")
	}

	zap.L().With(zap.Any("metadata", md)).Info("TraceInterceptor")

	// Cari trace_id dalam header (gunakan X-B3-TraceId atau yang sesuai)
	traceID := ""
	if len(md["X-B3-TraceId"]) > 0 {
		traceID = md["X-B3-TraceId"][0]
	} else if len(md["X-Request-Id"]) > 0 {
		traceID = md["X-Request-Id"][0]
	}

	// Log trace_id atau simpan dalam context
	if traceID != "" {
		// Anda dapat menggunakan trace_id di sini atau melanjutkan dengan span tracing
		fmt.Println("Received trace_id:", traceID)

		// Buat span dengan trace_id (opsional)
		tracer := otel.Tracer("server")
		_, span := tracer.Start(ctx, "Processing request", trace.WithAttributes(attribute.String("trace_id", traceID)))
		defer span.End()

		// Tambahkan trace_id ke dalam konteks untuk diteruskan ke proses berikutnya
		ctx = context.WithValue(ctx, "trace_id", traceID)
	}

	// Lanjutkan dengan eksekusi handler gRPC
	return handler(ctx, req)
}

func NewServerOption(
	trace *sdktrace.TracerProvider,
	metric *metric.MeterProvider,
) (options []grpc.ServerOption) {

	options = []grpc.ServerOption{
		grpc.UnaryInterceptor(TraceInterceptor),
		grpc.ChainUnaryInterceptor(
			validator.UnaryServerInterceptor(validator.WithFailFast()),
			logging.UnaryServerInterceptor(InterceptorLogger(zap.L())),
		),
		grpc.ChainStreamInterceptor(
			validator.StreamServerInterceptor(validator.WithFailFast()),
			logging.StreamServerInterceptor(InterceptorLogger(zap.L())),
		),
		grpc.StatsHandler(
			otelgrpc.NewServerHandler(
				otelgrpc.WithTracerProvider(trace),
				otelgrpc.WithMeterProvider(metric),
			),
		),
	}

	return
}

func NewGrpcServer(trace *sdktrace.TracerProvider, opts ...grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(opts...)
}
