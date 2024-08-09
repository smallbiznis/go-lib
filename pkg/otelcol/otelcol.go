package otelcol

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

func InitTraceProvider(resource *resource.Resource) (func(context.Context) error, error) {
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
		zap.Error(err)
		return nil, err
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(tracerExp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

func InitMetricProvider(resource *resource.Resource) (func(context.Context) error, error) {
	ctx := context.Background()
	// Set up a metrics exporter
	metricClient, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
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

	return mp.Shutdown, nil
}
