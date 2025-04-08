package tracing

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"time"
)

var service string

func InitTracer(serviceName, otelEndpoint string) (*trace.TracerProvider, error) {

	service = serviceName
	ctx := context.Background()

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(otelEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create otel exporter: %w", err)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter, trace.WithBatchTimeout(time.Second)),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	otel.SetTracerProvider(traceProvider)
	return traceProvider, nil
}

func GetTracer() oteltrace.Tracer {
	return otel.Tracer(service)
}

type TraceInfo struct {
	TraceID string
	SpanID  string
}

func GetTraceInfo(ctx context.Context) (TraceInfo, error) {
	spanContext := oteltrace.SpanFromContext(ctx).SpanContext()

	if !spanContext.IsValid() {
		return TraceInfo{}, fmt.Errorf("invalid span context")
	}

	return TraceInfo{
		TraceID: spanContext.TraceID().String(),
		SpanID:  spanContext.SpanID().String(),
	}, nil
}
