package fan

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/felixge/pidctrl"
)

const (
	ThermalZonePath = "/sys/class/thermal/thermal_zone0/temp"
)

// GetCPUTemp reads the CPU temperature from the thermal zone
func GetCPUTemp() (float64, error) {
	data, err := os.ReadFile(ThermalZonePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read thermal zone: %w", err)
	}

	tempStr := strings.TrimSpace(string(data))
	tempMilli, err := strconv.ParseInt(tempStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse temperature: %w", err)
	}

	return float64(tempMilli) / 1000.0, nil
}

// MonitorConfig holds configuration for the fan monitor
type MonitorConfig struct {
	ChipName      string
	Pin           int
	TargetTemp    float64
	Kp, Ki, Kd    float64
	CheckInterval time.Duration
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
			temp, err := GetCPUTemp()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading temp: %v\n", err)
				continue
			}

			pid.SetPID(-config.Kp, -config.Ki, -config.Kd)

			output := pid.Update(temp)

			pwm.SetDutyCycle(output)
		}
	}
}
