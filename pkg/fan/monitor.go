package fan

import (
	"context"
	"fmt"
	"github.com/AlejandroPerez92/nanoctl/pkg/temperature"
	"os"
	"time"

	"github.com/felixge/pidctrl"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// MonitorConfig holds configuration for the fan monitor
type MonitorConfig struct {
	ChipName      string
	Pin           int
	PWM           PWMConfig
	TargetTemp    float64
	Kp, Ki, Kd    float64
	CheckInterval time.Duration
	TempSource    temperature.Source
}

func periodNsFromFrequency(frequencyKHz float64) (int64, error) {
	if frequencyKHz <= 0 {
		return 0, fmt.Errorf("pwm.frequency_khz must be positive, got %.2f", frequencyKHz)
	}
	return int64(1e6 / frequencyKHz), nil
}

func frequencyHz(frequencyKHz float64) float64 {
	return frequencyKHz * 1000.0
}

// RunMonitor starts the fan control monitor
func RunMonitor(ctx context.Context, config MonitorConfig) error {
	// Initialize PWM Controller
	controller, err := newPWMController(config)
	if err != nil {
		return fmt.Errorf("failed to create PWM controller: %w", err)
	}
	if err := controller.Start(); err != nil {
		return fmt.Errorf("failed to start PWM controller: %w", err)
	}
	defer controller.Stop()

	// Initialize PID Controller
	pid := pidctrl.NewPIDController(config.Kp, config.Ki, config.Kd)
	pid.SetOutputLimits(0.0, 100.0)
	pid.Set(config.TargetTemp)

	// Initialize Metrics
	meter := otel.Meter("nanoctl")
	tempGauge, err := meter.Float64Gauge("nanoctl.temperature.celsius",
		metric.WithDescription("Current CPU temperature"),
		metric.WithUnit("Ce"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temperature gauge: %v\n", err)
	}

	fanGauge, err := meter.Float64Gauge("nanoctl.fan.duty_cycle.percent",
		metric.WithDescription("Current Fan PWM duty cycle"),
		metric.WithUnit("%"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create fan gauge: %v\n", err)
	}

	fmt.Printf("Starting Fan Monitor...\n")
	fmt.Printf("Target Temp: %.1f°C\n", config.TargetTemp)
	printPWMConfig(config)

	ticker := time.NewTicker(config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Monitor stopping...")
			return nil
		case <-ticker.C:
			// Use the configured temperature source
			temp, err := config.TempSource.GetTemperature()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading temp: %v\n", err)
				continue
			}

			pid.SetPID(-config.Kp, -config.Ki, -config.Kd)

			output := pid.Update(temp)

			controller.SetDutyCycle(output)

			// Record metrics
			if tempGauge != nil {
				tempGauge.Record(ctx, temp)
			}
			if fanGauge != nil {
				fanGauge.Record(ctx, output)
			}
		}
	}
}
