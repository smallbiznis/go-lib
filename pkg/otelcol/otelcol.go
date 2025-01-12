package otelcol

import (
	"context"

	otelpyroscope "github.com/grafana/otel-profiling-go"
	"github.com/smallbiznis/go-lib/pkg/env"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	Resource = fx.Module("otelcol.resource", fx.Options(
		fx.Provide(
			InitResource,
		),
	))

	TraceProvider = fx.Module("otelcol.trace", fx.Options(
		fx.Provide(
			InitTraceProvider,
		),
	))

	MetricProvider = fx.Module("otelcol.metric", fx.Options(
		fx.Provide(
			InitMetricProvider,
		),
	))
)

func InitResource() (res *resource.Resource, err error) {
	ctx := context.Background()
	extra, err := resource.New(ctx,
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithContainer(),
		resource.WithAttributes(
			semconv.ServiceName(env.Lookup("SERVICE_NAME", "example")),
			semconv.ServiceVersion(env.Lookup("SERVICE_VERSION", "1.0.0")),
			semconv.ServiceNamespace(env.Lookup("SERVICE_NAME", "smallbiznis")),
		),
	)
	if err != nil {
		return nil, err
	}

	res, err = resource.Merge(
		resource.Default(),
		extra,
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func InitTraceProvider(resource *resource.Resource) (*sdktrace.TracerProvider, error) {
	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns.
	// ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()

	// Set up a trace exporter
	traceClient := otlptracehttp.NewClient(
		otlptracehttp.WithInsecure(),
	)

	tracerExp, err := otlptrace.New(context.Background(), traceClient)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("Failed to connecting otel traceprovider")
		return nil, err
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(tracerExp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(tracerExp),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(resource),
	)

	otel.SetTracerProvider(otelpyroscope.NewTracerProvider(tracerProvider))

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider, nil
}

func InitMetricProvider(resource *resource.Resource) (*metric.MeterProvider, error) {
	ctx := context.Background()
	// Set up a metrics exporter
	metricClient, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("Failed to connecting otel meterprovider")
		return nil, err
	}

	mp := metric.NewMeterProvider(
		metric.WithResource(resource),
		metric.WithReader(
			metric.NewPeriodicReader(metricClient),
		),
	)
	defer func() {
		if err := mp.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()
	otel.SetMeterProvider(mp)

	return mp, nil
}
