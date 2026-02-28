package fan

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// HardwarePWMController controls PWM via sysfs.
type HardwarePWMController struct {
	chip     string
	channel  int
	periodNs int64
	inverted bool

	dutyCycle float64
	running   bool
	mu        sync.Mutex
}

// NewHardwarePWMController creates a new hardware PWM controller.
func NewHardwarePWMController(chip string, channel int, periodNs int64, inverted bool) *HardwarePWMController {
	return &HardwarePWMController{
		chip:      chip,
		channel:   channel,
		periodNs:  periodNs,
		inverted:  inverted,
		dutyCycle: 0.0,
	}
}

// Start configures and enables the PWM channel.
func (pwm *HardwarePWMController) Start() error {
	pwm.mu.Lock()
	defer pwm.mu.Unlock()
	if pwm.running {
		return nil
	}

	if err := pwm.exportChannel(); err != nil {
		return err
	}

	if err := pwm.writeSysfs("enable", "0"); err != nil {
		return fmt.Errorf("failed to disable PWM: %w", err)
	}
	if err := pwm.writeSysfs("period", strconv.FormatInt(pwm.periodNs, 10)); err != nil {
		return fmt.Errorf("failed to set PWM period: %w", err)
	}
	if err := pwm.writeDutyCycleLocked(); err != nil {
		return err
	}
	if err := pwm.writeSysfs("enable", "1"); err != nil {
		return fmt.Errorf("failed to enable PWM: %w", err)
	}

	pwm.running = true
	return nil
}

// Stop disables the PWM channel.
func (pwm *HardwarePWMController) Stop() {
	pwm.mu.Lock()
	defer pwm.mu.Unlock()
	if !pwm.running {
		return
	}
	_ = pwm.writeSysfs("enable", "0")
	pwm.running = false
}

// SetDutyCycle updates the duty cycle (0-100).
func (pwm *HardwarePWMController) SetDutyCycle(dc float64) {
	pwm.mu.Lock()
	defer pwm.mu.Unlock()
	if dc < 0 {
		dc = 0
	}
	if dc > 100 {
		dc = 100
	}
	pwm.dutyCycle = dc
	if pwm.running {
		_ = pwm.writeDutyCycleLocked()
	}
}

func (pwm *HardwarePWMController) exportChannel() error {
	if _, err := os.Stat(pwm.channelPath()); err == nil {
		return nil
	}
	if err := os.WriteFile(pwm.exportPath(), []byte(strconv.Itoa(pwm.channel)), 0644); err != nil {
		return fmt.Errorf("failed to export PWM channel %d: %w", pwm.channel, err)
	}

	for i := 0; i < 20; i++ {
		if _, err := os.Stat(pwm.channelPath()); err == nil {
			return nil
		}
		time.Sleep(25 * time.Millisecond)
	}

	return fmt.Errorf("PWM channel %d did not appear after export", pwm.channel)
}

func (pwm *HardwarePWMController) writeDutyCycleLocked() error {
	dutyNs := pwm.dutyNsLocked()
	if err := pwm.writeSysfs("duty_cycle", strconv.FormatInt(dutyNs, 10)); err != nil {
		return fmt.Errorf("failed to set PWM duty cycle: %w", err)
	}
	return nil
}

func (pwm *HardwarePWMController) dutyNsLocked() int64 {
	dutyNs := int64(math.Round(float64(pwm.periodNs) * pwm.dutyCycle / 100.0))
	if dutyNs < 0 {
		dutyNs = 0
	}
	if dutyNs > pwm.periodNs {
		dutyNs = pwm.periodNs
	}
	if pwm.inverted {
		return pwm.periodNs - dutyNs
	}
	return dutyNs
}

func (pwm *HardwarePWMController) writeSysfs(name, value string) error {
	return os.WriteFile(filepath.Join(pwm.channelPath(), name), []byte(value), 0644)
}

func (pwm *HardwarePWMController) exportPath() string {
	return filepath.Join("/sys/class/pwm", pwm.chip, "export")
}

func (pwm *HardwarePWMController) channelPath() string {
	return filepath.Join("/sys/class/pwm", pwm.chip, fmt.Sprintf("pwm%d", pwm.channel))
}
