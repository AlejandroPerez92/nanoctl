package metrics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitOTLP initializes the OpenTelemetry OTLP metric exporter.
// It returns a shutdown function that should be called when the application exits.
func InitOTLP(ctx context.Context, endpoint string, insecure bool, interval time.Duration, headers map[string]string) (func(context.Context) error, error) {
	// Determine protocol based on endpoint prefix or explicit configuration
	// If endpoint starts with http:// or https://, use HTTP exporter
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return initOTLPHTTP(ctx, endpoint, insecure, interval, headers)
	}
	return initOTLPGRPC(ctx, endpoint, insecure, interval, headers)
}

func initOTLPHTTP(ctx context.Context, endpoint string, insecure bool, interval time.Duration, headers map[string]string) (func(context.Context) error, error) {
	var opts []otlpmetrichttp.Option
	opts = append(opts, otlpmetrichttp.WithEndpointURL(endpoint))

	if insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	if len(headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(headers))
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP HTTP metric exporter: %w", err)
	}

	return setupProvider(ctx, exporter, interval)
}

func initOTLPGRPC(ctx context.Context, endpoint string, insecure bool, interval time.Duration, headers map[string]string) (func(context.Context) error, error) {
	var opts []otlpmetricgrpc.Option
	opts = append(opts, otlpmetricgrpc.WithEndpoint(endpoint))

	if insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	if len(headers) > 0 {
		opts = append(opts, otlpmetricgrpc.WithHeaders(headers))
	}

	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP gRPC metric exporter: %w", err)
	}

	return setupProvider(ctx, exporter, interval)
}

func setupProvider(ctx context.Context, exporter metric.Exporter, interval time.Duration) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("nanoctl"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	reader := metric.NewPeriodicReader(exporter, metric.WithInterval(interval))
	provider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(reader),
	)

	otel.SetMeterProvider(provider)

	return provider.Shutdown, nil
}
