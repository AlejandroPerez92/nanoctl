package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/AlejandroPerez92/nanoctl/pkg/config"
	"github.com/AlejandroPerez92/nanoctl/pkg/fan"
	"github.com/AlejandroPerez92/nanoctl/pkg/metrics"
	"github.com/AlejandroPerez92/nanoctl/pkg/temperature"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	configPath string
)

var fanCmd = &cobra.Command{
	Use:   "fan",
	Short: "Control cooling fan based on CPU temperature",
	Long: `Starts a daemon-like process that monitors CPU temperature and 
controls a fan using PWM (Pulse Width Modulation) on a GPIO pin.
Implements a PID controller to maintain the target temperature.

Configuration is loaded from /etc/nanoctl/fan.yaml by default.
Run 'sudo nanoctl install-service' to create a default configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runFan(cmd, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runFan(cmd *cobra.Command, args []string) error {
	// Load configuration from file
	cfg, err := config.LoadFanConfig(configPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Create temperature source with fallback
	tempSource, err := createTemperatureSource(cfg)
	if err != nil {
		return fmt.Errorf("error creating temperature source: %w", err)
	}
	defer tempSource.Close()

	// Parse check interval duration
	checkInterval, err := cfg.GetCheckIntervalDuration()
	if err != nil {
		return fmt.Errorf("error parsing check interval: %w", err)
	}

	// Convert to fan.MonitorConfig
	monitorConfig := fan.MonitorConfig{
		ChipName: cfg.GPIO.ChipName,
		Pin:      cfg.GPIO.Pin,
		PWM: fan.PWMConfig{
			Mode:         cfg.PWM.Mode,
			FrequencyKHz: cfg.PWM.FrequencyKHz,
			Hardware: fan.HardwarePWMConfig{
				Chip:     cfg.PWM.Hardware.Chip,
				Channel:  cfg.PWM.Hardware.Channel,
				Inverted: cfg.PWM.Hardware.Inverted,
			},
		},
		TargetTemp:    cfg.Temperature.Target,
		Kp:            cfg.PID.Kp,
		Ki:            cfg.PID.Ki,
		Kd:            cfg.PID.Kd,
		CheckInterval: checkInterval,
		TempSource:    tempSource,
	}

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt, shutting down...")
		cancel()
	}()

	// Initialize OTel Metrics if enabled
	if cfg.Metrics.Enabled {
		interval, err := time.ParseDuration(cfg.Metrics.Interval)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid metrics interval '%s', defaulting to 10s: %v\n", cfg.Metrics.Interval, err)
			interval = 10 * time.Second
		}

		headers := make(map[string]string)
		if cfg.Metrics.Auth != nil && cfg.Metrics.Auth.Username != "" {
			auth := fmt.Sprintf("%s:%s", cfg.Metrics.Auth.Username, cfg.Metrics.Auth.Password)
			encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
			headers["Authorization"] = "Basic " + encodedAuth
		}

		shutdown, err := metrics.InitOTLP(ctx, cfg.Metrics.Endpoint, cfg.Metrics.Insecure, interval, headers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize metrics: %v\n", err)
		} else {
			fmt.Println("Metrics pushing enabled to", cfg.Metrics.Endpoint)
			defer func() {
				if err := shutdown(context.Background()); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to shutdown metrics: %v\n", err)
				}
			}()
		}
	}

	if err := fan.RunMonitor(ctx, monitorConfig); err != nil {
		return fmt.Errorf("fan monitor error: %w", err)
	}

	return nil
}

func createTemperatureSource(cfg *config.FanConfig) (temperature.Source, error) {
	primary := cfg.Temperature.Source.Primary
	fallback := cfg.Temperature.Source.Fallback

	// Try primary source first
	if primary == "prometheus" {
		if cfg.Temperature.Source.Prometheus == nil {
			return nil, fmt.Errorf("prometheus configuration is required when primary source is prometheus")
		}

		// Map config to temperature package config struct
		promConfig := temperature.PrometheusConfig{
			Host:    cfg.Temperature.Source.Prometheus.Host,
			Query:   cfg.Temperature.Source.Prometheus.Query,
			Timeout: cfg.Temperature.Source.Prometheus.Timeout,
		}

		if cfg.Temperature.Source.Prometheus.Auth != nil {
			promConfig.Auth = temperature.AuthConfig{
				Username: cfg.Temperature.Source.Prometheus.Auth.Username,
				Password: cfg.Temperature.Source.Prometheus.Auth.Password,
			}
		}

		promSource, err := temperature.NewPrometheusSource(promConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create Prometheus source: %v\n", err)
			return createFallbackSource(cfg, fallback)
		}

		// Test connection
		_, err = promSource.GetTemperature()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prometheus connection test failed: %v\n", err)
			fmt.Fprintf(os.Stderr, "Falling back to %s source\n", fallback)
			return createFallbackSource(cfg, fallback)
		}

		fmt.Println("Using Prometheus as temperature source")
		return promSource, nil
	}

	// Use file source
	if primary == "file" {
		fileSource := temperature.NewFileSource(cfg.Temperature.Source.File.Path)
		fmt.Println("Using file as temperature source")
		return fileSource, nil
	}

	return nil, fmt.Errorf("unknown primary source type: %s", primary)
}

func createFallbackSource(cfg *config.FanConfig, fallback string) (temperature.Source, error) {
	if fallback != "file" {
		return nil, fmt.Errorf("unsupported fallback source: %s", fallback)
	}

	fileSource := temperature.NewFileSource(cfg.Temperature.Source.File.Path)
	fmt.Println("Using file as fallback temperature source")
	return fileSource, nil
}

func init() {
	rootCmd.AddCommand(fanCmd)

	fanCmd.Flags().StringVar(&configPath, "config", config.DefaultConfigPath, "Path to configuration file")
}
