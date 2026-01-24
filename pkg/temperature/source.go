// Package temperature provides temperature reading sources for the fan controller.
// It supports both local file-based reading and remote Prometheus queries.
package temperature

import (
	"fmt"
)

// Source defines the interface for fetching temperature data.
type Source interface {
	// GetTemperature returns the current temperature in degrees Celsius.
	GetTemperature() (float64, error)
	// Close releases any resources associated with the source.
	Close() error
}

// SourceType represents the type of temperature source.
type SourceType string

const (
	// SourceFile represents a local file-based temperature source.
	SourceFile SourceType = "file"
	// SourcePrometheus represents a Prometheus-based temperature source.
	SourcePrometheus SourceType = "prometheus"
)

// SourceConfig holds the configuration for creating a new Source.
// This is a simplified config used by the factory.
type SourceConfig struct {
	Type       SourceType
	FilePath   string
	Prometheus PrometheusConfig
}

// PrometheusConfig holds configuration specific to the Prometheus source.
type PrometheusConfig struct {
	Host    string
	Query   string
	Timeout string
	Auth    AuthConfig
}

// AuthConfig holds authentication details.
type AuthConfig struct {
	Username string
	Password string
}

// NewSource creates a new temperature source based on the provided configuration.
func NewSource(config SourceConfig) (Source, error) {
	if config.Type == SourcePrometheus {
		return NewPrometheusSource(config.Prometheus)
	}

	if config.Type == SourceFile {
		return NewFileSource(config.FilePath), nil
	}

	return nil, fmt.Errorf("unknown source type: %s", config.Type)
}
