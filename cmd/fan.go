package cmd

import (
	"NanoCtl/pkg/fan"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	fanChip       string
	fanPin        int
	targetTemp    float64
	kp, ki, kd    float64
	checkInterval time.Duration
)

var fanCmd = &cobra.Command{
	Use:   "fan",
	Short: "Control cooling fan based on CPU temperature",
	Long: `Starts a daemon-like process that monitors CPU temperature and 
controls a fan using PWM (Pulse Width Modulation) on a GPIO pin.
Implements a PID controller to maintain the target temperature.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := fan.MonitorConfig{
			ChipName:      fanChip,
			Pin:           fanPin,
			TargetTemp:    targetTemp,
			Kp:            kp,
			Ki:            ki,
			Kd:            kd,
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

		if err := fan.RunMonitor(ctx, config); err != nil {
			fmt.Fprintf(os.Stderr, "Fan monitor error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(fanCmd)

	fanCmd.Flags().StringVar(&fanChip, "chip", "gpiochip0", "GPIO chip name (e.g., gpiochip0 for RPi5)")
	fanCmd.Flags().IntVar(&fanPin, "pin", 13, "GPIO pin number (BCM)")
	fanCmd.Flags().Float64Var(&targetTemp, "target", 55.0, "Target CPU temperature in Celsius")

	// Default PID values - these might need tuning
	fanCmd.Flags().Float64Var(&kp, "kp", 5.0, "Proportional gain")
	fanCmd.Flags().Float64Var(&ki, "ki", 0.1, "Integral gain")
	fanCmd.Flags().Float64Var(&kd, "kd", 0.5, "Derivative gain")

	fanCmd.Flags().DurationVar(&checkInterval, "interval", 1*time.Second, "Temperature check interval")
}
