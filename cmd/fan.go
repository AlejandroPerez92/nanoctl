package cmd

import (
	"NanoCtl/pkg/config"
	"NanoCtl/pkg/fan"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
		// Load configuration from file
		cfg, err := config.LoadFanConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
			os.Exit(1)
		}

		// Parse check interval duration
		checkInterval, err := cfg.GetCheckIntervalDuration()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing check interval: %v\n", err)
			os.Exit(1)
		}

		// Convert to fan.MonitorConfig
		monitorConfig := fan.MonitorConfig{
			ChipName:      cfg.GPIO.ChipName,
			Pin:           cfg.GPIO.Pin,
			TargetTemp:    cfg.Temperature.Target,
			Kp:            cfg.PID.Kp,
			Ki:            cfg.PID.Ki,
			Kd:            cfg.PID.Kd,
			CheckInterval: checkInterval,
		}

		// Handle graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigChan
			fmt.Println("\nReceived interrupt, shutting down...")
			cancel()
		}()

		if err := fan.RunMonitor(ctx, monitorConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Fan monitor error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(fanCmd)

	fanCmd.Flags().StringVar(&configPath, "config", config.DefaultConfigPath, "Path to configuration file")
}
