package fan

import (
	"fmt"
	"sync"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

// PWMController handles software PWM on a GPIO pin
type PWMController struct {
	chipName  string
	pin       int
	frequency float64 // Hz
	dutyCycle float64 // 0.0 to 100.0

	running bool
	mu      sync.Mutex
	stop    chan struct{}
}

// NewPWMController creates a new software PWM controller
func NewPWMController(chipName string, pin int, frequency float64) *PWMController {
	return &PWMController{
		chipName:  chipName,
		pin:       pin,
		frequency: frequency,
		dutyCycle: 0.0,
		stop:      make(chan struct{}),
	}
}

// Start begins the PWM loop
func (pwm *PWMController) Start() error {
	pwm.mu.Lock()
	if pwm.running {
		pwm.mu.Unlock()
		return nil
	}
	pwm.running = true
	pwm.mu.Unlock()

	// Request the line
	line, err := gpiocdev.RequestLine(pwm.chipName, pwm.pin, gpiocdev.AsOutput(0))
	if err != nil {
		return fmt.Errorf("failed to request GPIO %d on %s: %w", pwm.pin, pwm.chipName, err)
	}

	go func() {
		defer line.Close()

		// Calculate period
		period := time.Duration(float64(time.Second) / pwm.frequency)

		ticker := time.NewTicker(period)
		defer ticker.Stop()

		for {
			select {
			case <-pwm.stop:
				// Turn off before stopping
				_ = line.SetValue(0)
				return
			default:
				pwm.mu.Lock()
				dc := pwm.dutyCycle
				pwm.mu.Unlock()

				if dc <= 0.0 {
					_ = line.SetValue(0)
					time.Sleep(period)
				} else if dc >= 100.0 {
					_ = line.SetValue(1)
					time.Sleep(period)
				} else {
					// On time
					onDuration := time.Duration(float64(period) * dc / 100.0)
					_ = line.SetValue(1)
					time.Sleep(onDuration)

					// Off time
					_ = line.SetValue(0)
					time.Sleep(period - onDuration)
				}
			}
		}
	}()

	return nil
}

// Stop stops the PWM loop
func (pwm *PWMController) Stop() {
	pwm.mu.Lock()
	defer pwm.mu.Unlock()
	if pwm.running {
		close(pwm.stop)
		pwm.running = false
	}
}

// SetDutyCycle updates the duty cycle (0-100)
func (pwm *PWMController) SetDutyCycle(dc float64) {
	pwm.mu.Lock()
	defer pwm.mu.Unlock()
	if dc < 0 {
		dc = 0
	}
	if dc > 100 {
		dc = 100
	}
	pwm.dutyCycle = dc
}
