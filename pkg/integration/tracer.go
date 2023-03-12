package integration

import (
	"context"
	"fmt"
	"github.com/go-logr/zapr"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

func ConfigureTracer(context context.Context, useStdoutExporter bool) (*trace.TracerProvider, error) {
	var exporter trace.SpanExporter
	if useStdoutExporter {
		writer := &zapio.Writer{Log: zap.L(), Level: zap.DebugLevel}
		var err error
		if exporter, err = stdouttrace.New(stdouttrace.WithWriter(writer)); err != nil {
			return nil, err
		}
	} else {
		var err error
		if exporter, err = otlptracegrpc.New(context); err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}
	}
	staticRes, err := resource.Merge(resource.Default(), resource.NewSchemaless(semconv.ServiceName("app")))
	if err != nil {
		return nil, fmt.Errorf("to merge static resources: %w", err)
	}
	otelRes, err := resource.Merge(staticRes, resource.Environment())
	if err != nil {
		return nil, fmt.Errorf("to merge env resources: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(otelRes),
	)

	otel.SetTracerProvider(tp)
	otel.SetLogger(zapr.NewLogger(zap.L()))
	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())
	return tp, nil
}
