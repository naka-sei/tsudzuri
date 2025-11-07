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
	exporter, err := texporter.New(texporter.WithProjectID(conf.GoogleCloudProject))
	if err != nil {
		log.Println(err)
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
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return tp
}
