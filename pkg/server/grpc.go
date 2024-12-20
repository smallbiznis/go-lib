package server

import (
	"context"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator"
	"github.com/smallbiznis/go-lib/pkg/env"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var (
	GrpcServerProvider = fx.Module("grpc.server", fx.Options(
		fx.Provide(
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

func NewGrpcServer(trace *sdktrace.TracerProvider, opts ...grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			validator.UnaryServerInterceptor(validator.WithFailFast()),
		),
		grpc.ChainStreamInterceptor(
			validator.StreamServerInterceptor(validator.WithFailFast()),
		),
		grpc.StatsHandler(
			otelgrpc.NewServerHandler(
				otelgrpc.WithTracerProvider(trace),
			),
		),
	)
}
