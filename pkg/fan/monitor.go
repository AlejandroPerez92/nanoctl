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
	TargetTemp    float64
	Kp, Ki, Kd    float64
	CheckInterval time.Duration
	TempSource    temperature.Source
}

// RunMonitor starts the fan control monitor
func RunMonitor(ctx context.Context, config MonitorConfig) error {
	// Initialize PWM Controller (50Hz)
	pwm := NewPWMController(config.ChipName, config.Pin, 50.0)
	if err := pwm.Start(); err != nil {
		return fmt.Errorf("failed to start PWM controller: %w", err)
	}
	defer pwm.Stop()

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
	fmt.Printf("GPIO: %s pin %d\n", config.ChipName, config.Pin)

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

			pwm.SetDutyCycle(output)

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
