package main

import (
	"context"
	"log"
	"os"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"

	"github.com/naka-sei/tsudzuri/config"
)

// InitializeTracer initializes and returns a TracerProvider that exports traces to Google
func InitializeTracer(conf *config.Config) *sdktrace.TracerProvider {
	ctx := context.Background()

	// Select exporter based on configuration. When stdout exporter is enabled,
	// we prefer immediate feedback over remote delivery to Cloud Trace.
	var (
		exporter sdktrace.SpanExporter
		err      error
	)
	switch {
	case conf.EnableStdoutTraceExporter:
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
			stdouttrace.WithWriter(os.Stdout),
		)
	default:
		// Try to create GCP Cloud Trace exporter. If it fails (e.g., missing ADC),
		// fall back to a provider without an exporter to avoid panics.
		exporter, err = texporter.New(texporter.WithProjectID(conf.GoogleCloudProject))
	}
	if err != nil {
		// Log the error but continue with a noop exporter so the app keeps running locally.
		log.Println(err)
		exporter = nil
	}

	res, err := resource.New(ctx,
		resource.WithDetectors(gcp.NewDetector()),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("tsudzuri"),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	// Build tracer provider options, adding the exporter only when available.
	opts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	switch {
	case exporter != nil && conf.EnableStdoutTraceExporter:
		opts = append(opts,
			sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
	case exporter != nil:
		opts = append(opts, sdktrace.WithBatcher(exporter))
	default:
		// Avoid sampling work locally if we're not exporting anywhere.
		opts = append(opts, sdktrace.WithSampler(sdktrace.NeverSample()))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return tp
}
