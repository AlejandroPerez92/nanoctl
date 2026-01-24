package config

import (
	_ "embed"
	"fmt"
	"os"
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

	Temperature struct {
		Target float64 `yaml:"target"`
	} `yaml:"temperature"`

	PID struct {
		Kp float64 `yaml:"kp"`
		Ki float64 `yaml:"ki"`
		Kd float64 `yaml:"kd"`
	} `yaml:"pid"`

	Monitor struct {
		CheckInterval string `yaml:"check_interval"`
	} `yaml:"monitor"`
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
	if config.Temperature.Target == 0 {
		config.Temperature.Target = 55.0
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

// GetDefaultConfigYAML returns the embedded default configuration as a string
func GetDefaultConfigYAML() string {
	return string(defaultConfigYAML)
}
