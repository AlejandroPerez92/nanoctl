package temperature

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// PrometheusSource implements the Source interface for querying Prometheus.
type PrometheusSource struct {
	client  api.Client
	api     promv1.API
	query   string
	timeout time.Duration
}

// NewPrometheusSource creates a new Prometheus-based temperature source.
func NewPrometheusSource(config PrometheusConfig) (*PrometheusSource, error) {
	// Validate required fields
	if config.Host == "" {
		return nil, fmt.Errorf("prometheus host is required")
	}

	// Apply defaults
	if config.Query == "" {
		config.Query = `max(node_hwmon_temp_celsius{sensor="temp0"})`
	}
	if config.Timeout == "" {
		config.Timeout = "5s"
	}

	timeout, err := time.ParseDuration(config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout: %w", err)
	}

	// Build client config
	clientConfig := api.Config{
		Address: config.Host,
	}

	// Add basic auth if provided
	if config.Auth.Username != "" {
		clientConfig.RoundTripper = &basicAuthTransport{
			Username:  config.Auth.Username,
			Password:  config.Auth.Password,
			Transport: api.DefaultRoundTripper,
		}
	}

	client, err := api.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus client: %w", err)
	}

	return &PrometheusSource{
		client:  client,
		api:     promv1.NewAPI(client),
		query:   config.Query,
		timeout: timeout,
	}, nil
}

// GetTemperature executes the PromQL query to fetch the temperature.
func (p *PrometheusSource) GetTemperature() (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	result, warnings, err := p.api.Query(ctx, p.query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("prometheus query failed: %w", err)
	}

	if len(warnings) > 0 {
		// Log warnings but don't fail, use stderr to avoid polluting stdout if used for CLI output
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "Prometheus warning: %s\n", w)
		}
	}

	// Early return if no result
	if result == nil {
		return 0, fmt.Errorf("no result from prometheus query")
	}

	// Parse result based on type. We expect a Vector (instant query) or Scalar.
	// Usually Query returns a Vector or Scalar.
	switch v := result.(type) {
	case model.Vector:
		if len(v) == 0 {
			return 0, fmt.Errorf("no temperature data returned from query")
		}
		// We take the first element. The query should ideally return a single scalar or vector.
		// If it returns multiple, taking the first is a reasonable default if aggregation was expected but not fully done.
		return float64(v[0].Value), nil
	case *model.Scalar:
		return float64(v.Value), nil
	default:
		return 0, fmt.Errorf("unexpected result type: %T", result)
	}
}

// Close implements the Source interface.
func (p *PrometheusSource) Close() error {
	// Prometheus client doesn't need explicit cleanup
	return nil
}

// basicAuthTransport implements HTTP basic authentication
type basicAuthTransport struct {
	Username  string
	Password  string
	Transport http.RoundTripper
}

func (t *basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Username != "" {
		req.SetBasicAuth(t.Username, t.Password)
	}
	// Use default transport if t.Transport is nil
	if t.Transport == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return t.Transport.RoundTrip(req)
}
