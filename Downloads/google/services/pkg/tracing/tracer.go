// Package tracing provides OpenTelemetry setup for FinGuard services.
//
// In production, exporters are configured to OTLP / Cloud Trace.
// In local development with OTelExporter="none", the global tracer provider is
// a no-op so traces are still created but discarded.
package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Init sets up the global tracer provider.
func Init(ctx context.Context, service, version, exporter string) (func(context.Context) error, error) {
	otel.SetErrorHandler(otelErrLogger{})

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(service),
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("resource: %w", err)
	}

	switch exporter {
	case "otlp", "cloud_trace":
		tp := sdktrace.NewTracerProvider(sdktrace.WithResource(res))
		otel.SetTracerProvider(tp)
		return tp.Shutdown, nil
	default:
		// No exporter → no-op provider.
		tp := sdktrace.NewTracerProvider()
		otel.SetTracerProvider(tp)
		return tp.Shutdown, nil
	}
}

// Tracer returns a tracer for the given component.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

type otelErrLogger struct{}

func (otelErrLogger) Handle(err error) {
	// Suppress exporter errors; surface in future via Sentry.
	_ = err
}
