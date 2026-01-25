package cmd

import (
	"fmt"
	"github.com/AlejandroPerez92/nanoctl/pkg/config"
	"github.com/AlejandroPerez92/nanoctl/pkg/temperature"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var checkPrometheusCmd = &cobra.Command{
	Use:   "check-prometheus",
	Short: "Test Prometheus connection and query",
	Long: `Verifies that the Prometheus configuration is correct by:
1. Checking connectivity to the Prometheus server
2. Executing the configured temperature query
3. Displaying the returned temperature value

This command is useful for troubleshooting Prometheus integration.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.LoadFanConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
			os.Exit(1)
		}

		// Check if Prometheus is configured
		if cfg.Temperature.Source.Prometheus == nil {
			fmt.Println("Prometheus is not configured in the config file.")
			fmt.Printf("Edit %s to add Prometheus configuration.\n", configPath)
			os.Exit(1)
		}

		promConfig := cfg.Temperature.Source.Prometheus

		fmt.Println("Prometheus Configuration:")
		fmt.Printf("  Host:    %s\n", promConfig.Host)
		fmt.Printf("  Query:   %s\n", getQueryOrDefault(promConfig.Query))
		fmt.Printf("  Timeout: %s\n", getTimeoutOrDefault(promConfig.Timeout))
		if promConfig.Auth != nil && promConfig.Auth.Username != "" {
			fmt.Printf("  Auth:    Basic (username: %s)\n", promConfig.Auth.Username)
		} else {
			fmt.Printf("  Auth:    None\n")
		}
		fmt.Println()

		fmt.Println("Testing connection...")

		// Map to temperature package config
		tempPromConfig := temperature.PrometheusConfig{
			Host:    promConfig.Host,
			Query:   promConfig.Query,
			Timeout: promConfig.Timeout,
		}
		if promConfig.Auth != nil {
			tempPromConfig.Auth = temperature.AuthConfig{
				Username: promConfig.Auth.Username,
				Password: promConfig.Auth.Password,
			}
		}

		// Create Prometheus source
		start := time.Now()
		promSource, err := temperature.NewPrometheusSource(tempPromConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Failed to create Prometheus client: %v\n", err)
			os.Exit(1)
		}
		defer promSource.Close()

		fmt.Println("✓ Prometheus client created successfully")

		// Execute query
		fmt.Println("\nExecuting query...")
		temp, err := promSource.GetTemperature()
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Query failed: %v\n", err)
			os.Exit(1)
		}

		elapsed := time.Since(start)

		fmt.Printf("✓ Query successful (took %v)\n", elapsed)
		fmt.Printf("\nTemperature: %.2f°C\n", temp)
		fmt.Println("\n✓ Prometheus connection is working correctly!")
	},
}

func getQueryOrDefault(query string) string {
	if query == "" {
		return `max(node_hwmon_temp_celsius{sensor="temp0"})` + " (default)"
	}
	return query
}

func getTimeoutOrDefault(timeout string) string {
	if timeout == "" {
		return "5s (default)"
	}
	return timeout
}

func init() {
	rootCmd.AddCommand(checkPrometheusCmd)
	checkPrometheusCmd.Flags().StringVar(&configPath, "config", config.DefaultConfigPath, "Path to configuration file")
}
