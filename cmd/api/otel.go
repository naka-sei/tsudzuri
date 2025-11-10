package main

import (
	"context"
	"log"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"

	"github.com/naka-sei/tsudzuri/config"
)

// InitializeTracer initializes and returns a TracerProvider that exports traces to Google
func InitializeTracer(conf *config.Config) *sdktrace.TracerProvider {
	ctx := context.Background()

	// Try to create GCP Cloud Trace exporter. If it fails (e.g., missing ADC),
	// fall back to a provider without an exporter to avoid panics.
	exporter, err := texporter.New(texporter.WithProjectID(conf.GoogleCloudProject))
	if err != nil {
		// Log the error but continue with a noop exporter (no WithBatcher)
		// so local/dev environments without credentials won't crash.
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

	// Build tracer provider options, adding the batcher only when exporter is available.
	opts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	if exporter != nil {
		opts = append(opts, sdktrace.WithBatcher(exporter))
	} else {
		// Avoid sampling work locally if we're not exporting anywhere.
		opts = append(opts, sdktrace.WithSampler(sdktrace.NeverSample()))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return tp
}
