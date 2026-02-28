package config

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed default.yaml
var defaultConfigYAML []byte

// DefaultConfigPath is the default location for the fan configuration file
const DefaultConfigPath = "/etc/nanoctl/fan.yaml"

// FanConfig represents the fan controller configuration
type FanConfig struct {
	GPIO struct {
		ChipName string `yaml:"chip_name"`
		Pin      int    `yaml:"pin"`
	} `yaml:"gpio"`

	PWM struct {
		Mode         string  `yaml:"mode"`
		FrequencyKHz float64 `yaml:"frequency_khz"`
		Hardware     struct {
			Chip     string `yaml:"chip"`
			Channel  int    `yaml:"channel"`
			Inverted bool   `yaml:"inverted"`
		} `yaml:"hardware"`
	} `yaml:"pwm"`

	Temperature struct {
		Target float64      `yaml:"target"`
		Source SourceConfig `yaml:"source"`
	} `yaml:"temperature"`

	PID struct {
		Kp float64 `yaml:"kp"`
		Ki float64 `yaml:"ki"`
		Kd float64 `yaml:"kd"`
	} `yaml:"pid"`

	Metrics struct {
		Enabled  bool        `yaml:"enabled"`
		Endpoint string      `yaml:"endpoint"` // e.g. "localhost:4317"
		Insecure bool        `yaml:"insecure"`
		Interval string      `yaml:"interval"` // e.g. "5s"
		Auth     *AuthConfig `yaml:"auth,omitempty"`
	} `yaml:"metrics"`

	Monitor struct {
		CheckInterval string `yaml:"check_interval"`
	} `yaml:"monitor"`
}

// SourceConfig holds configuration for temperature sources
type SourceConfig struct {
	Primary    string            `yaml:"primary"`              // "prometheus" or "file"
	Fallback   string            `yaml:"fallback"`             // "file"
	Prometheus *PrometheusConfig `yaml:"prometheus,omitempty"` // Optional
	File       FileSourceConfig  `yaml:"file"`
}

// PrometheusConfig holds configuration specific to the Prometheus source.
type PrometheusConfig struct {
	Host    string      `yaml:"host"`              // Required: http://host:port or https://host:port
	Query   string      `yaml:"query,omitempty"`   // Optional: defaults to max(node_hwmon_temp_celsius{sensor="temp0"})
	Timeout string      `yaml:"timeout,omitempty"` // Optional: defaults to "5s"
	Auth    *AuthConfig `yaml:"auth,omitempty"`    // Optional: Basic auth
}

// AuthConfig holds authentication details.
type AuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password,omitempty"`
}

// FileSourceConfig holds configuration specific to the file source.
type FileSourceConfig struct {
	Path string `yaml:"path"`
}

// LoadFanConfig loads the fan configuration from a YAML file
func LoadFanConfig(path string) (*FanConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s. Run 'sudo nanoctl install-service' to create it", path)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config FanConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Apply defaults
	applyDefaults(&config)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// applyDefaults applies default values to empty fields
func applyDefaults(config *FanConfig) {
	if config.GPIO.ChipName == "" {
		config.GPIO.ChipName = "gpiochip0"
	}
	if config.GPIO.Pin == 0 {
		config.GPIO.Pin = 13
	}
	if config.PWM.Mode == "" {
		config.PWM.Mode = "software"
	}
	if config.PWM.FrequencyKHz == 0 {
		config.PWM.FrequencyKHz = 25.0
	}
	if config.PWM.Hardware.Chip == "" {
		config.PWM.Hardware.Chip = "pwmchip0"
	}
	if config.PWM.Hardware.Channel == 0 {
		config.PWM.Hardware.Channel = 1
	}
	if config.Temperature.Target == 0 {
		config.Temperature.Target = 55.0
	}

	// Default temperature source settings
	if config.Temperature.Source.Primary == "" {
		config.Temperature.Source.Primary = "file"
	}
	if config.Temperature.Source.Fallback == "" {
		config.Temperature.Source.Fallback = "file"
	}
	if config.Temperature.Source.File.Path == "" {
		config.Temperature.Source.File.Path = "/sys/class/thermal/thermal_zone0/temp"
	}

	if config.PID.Kp == 0 {
		config.PID.Kp = 5.0
	}
	if config.PID.Ki == 0 {
		config.PID.Ki = 0.1
	}
	if config.PID.Kd == 0 {
		config.PID.Kd = 0.5
	}

	// Default metrics settings
	if config.Metrics.Endpoint == "" {
		config.Metrics.Endpoint = "localhost:4317"
	}
	if config.Metrics.Interval == "" {
		config.Metrics.Interval = "10s"
	}

	if config.Monitor.CheckInterval == "" {
		config.Monitor.CheckInterval = "1s"
	}
}

// Validate validates the configuration values
func (c *FanConfig) Validate() error {
	// Validate GPIO pin (valid BCM pins for RPi are 0-27)
	if c.GPIO.Pin < 0 || c.GPIO.Pin > 27 {
		return fmt.Errorf("gpio.pin must be between 0 and 27, got %d", c.GPIO.Pin)
	}

	// Validate PWM configuration
	switch c.PWM.Mode {
	case "software":
		if c.PWM.FrequencyKHz <= 0 {
			return fmt.Errorf("pwm.frequency_khz must be positive, got %.2f", c.PWM.FrequencyKHz)
		}
	case "hardware":
		if c.PWM.Hardware.Chip == "" {
			return fmt.Errorf("pwm.hardware.chip is required when pwm.mode is 'hardware'")
		}
		if c.PWM.Hardware.Channel < 0 {
			return fmt.Errorf("pwm.hardware.channel must be >= 0, got %d", c.PWM.Hardware.Channel)
		}
	default:
		return fmt.Errorf("pwm.mode must be 'software' or 'hardware', got '%s'", c.PWM.Mode)
	}

	// Validate target temperature (reasonable range)
	if c.Temperature.Target < 20.0 || c.Temperature.Target > 90.0 {
		return fmt.Errorf("temperature.target must be between 20 and 90°C, got %.1f", c.Temperature.Target)
	}

	// Validate PID values (must be positive)
	if c.PID.Kp <= 0 {
		return fmt.Errorf("pid.kp must be positive, got %.2f", c.PID.Kp)
	}
	if c.PID.Ki <= 0 {
		return fmt.Errorf("pid.ki must be positive, got %.2f", c.PID.Ki)
	}
	if c.PID.Kd <= 0 {
		return fmt.Errorf("pid.kd must be positive, got %.2f", c.PID.Kd)
	}

	// Validate check interval (must be parseable as duration)
	if _, err := time.ParseDuration(c.Monitor.CheckInterval); err != nil {
		return fmt.Errorf("monitor.check_interval must be a valid duration (e.g., '1s', '500ms'): %w", err)
	}

	// Validate temperature source configuration
	if err := c.validateTemperatureSource(); err != nil {
		return err
	}

	return nil
}

func (c *FanConfig) validateTemperatureSource() error {
	// Validate primary source type
	if c.Temperature.Source.Primary != "file" && c.Temperature.Source.Primary != "prometheus" {
		return fmt.Errorf("temperature.source.primary must be 'file' or 'prometheus', got '%s'", c.Temperature.Source.Primary)
	}

	// If prometheus is primary, validate prometheus config
	if c.Temperature.Source.Primary == "prometheus" {
		if c.Temperature.Source.Prometheus == nil {
			return fmt.Errorf("temperature.source.prometheus configuration is required when primary is 'prometheus'")
		}

		if c.Temperature.Source.Prometheus.Host == "" {
			return fmt.Errorf("temperature.source.prometheus.host is required")
		}

		// Validate URL format
		if !strings.HasPrefix(c.Temperature.Source.Prometheus.Host, "http://") &&
			!strings.HasPrefix(c.Temperature.Source.Prometheus.Host, "https://") {
			return fmt.Errorf("temperature.source.prometheus.host must start with http:// or https://")
		}

		// Validate timeout if provided
		if c.Temperature.Source.Prometheus.Timeout != "" {
			if _, err := time.ParseDuration(c.Temperature.Source.Prometheus.Timeout); err != nil {
				return fmt.Errorf("temperature.source.prometheus.timeout must be a valid duration: %w", err)
			}
		}
	}

	// Validate file source path
	if c.Temperature.Source.File.Path == "" {
		return fmt.Errorf("temperature.source.file.path is required")
	}

	return nil
}

// GetCheckIntervalDuration parses and returns the check interval as time.Duration
func (c *FanConfig) GetCheckIntervalDuration() (time.Duration, error) {
	return time.ParseDuration(c.Monitor.CheckInterval)
}

// CreateDefaultConfig creates a default configuration file at the specified path
func CreateDefaultConfig(path string) error {
	// Ensure directory exists
	dir := "/etc/nanoctl"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write default config
	if err := os.WriteFile(path, defaultConfigYAML, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
