package gpio

import (
	"fmt"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

const (
	GPIOChip = "gpiochip14"
	SlotGPIO = 2
)

// BoardType represents the type of board in the slot
type BoardType string

const (
	BoardCM5 BoardType = "cm5"
)

// Controller handles GPIO operations for node control
type Controller struct {
	chipName string
}

// NewController creates a new GPIO controller
func NewController() *Controller {
	return &Controller{
		chipName: GPIOChip,
	}
}

// pulseGPIO sends a pulse to a GPIO pin
// The pulse consists of setting the pin low, waiting for duration, then setting it high
func (c *Controller) pulseGPIO(pin int, duration time.Duration) error {
	// Request the line as output with initial value low (0)
	line, err := gpiocdev.RequestLine(c.chipName, pin, gpiocdev.AsOutput(0))
	if err != nil {
		return fmt.Errorf("failed to request GPIO %d: %w", pin, err)
	}
	defer line.Close()

	// Wait for specified duration while GPIO is low
	time.Sleep(duration)

	// Set GPIO high to complete the pulse
	if err := line.SetValue(1); err != nil {
		return fmt.Errorf("failed to set GPIO %d high: %w", pin, err)
	}

	return nil
}

// PowerOn powers on a CM5 node
// Simulates a single short press (power on)
func (c *Controller) PowerOn(slot int, boardType BoardType) error {
	if boardType != BoardCM5 {
		return fmt.Errorf("unsupported board type: %s (only cm5 is supported)", boardType)
	}

	fmt.Printf("Powering on slot %d (CM5)...\n", slot)

	// Single short press to power on
	if err := c.pulseGPIO(slot, 1*time.Second); err != nil {
		return fmt.Errorf("failed to power on: %w", err)
	}

	fmt.Printf("Power on signal sent to slot %d\n", slot)
	return nil
}

// PowerOff powers off a CM5 node
// Simulates two short presses (shutdown for Desktop systems)
// For headless systems, one press is sufficient, but two presses work for both
func (c *Controller) PowerOff(slot int, boardType BoardType) error {
	if boardType != BoardCM5 {
		return fmt.Errorf("unsupported board type: %s (only cm5 is supported)", boardType)
	}

	fmt.Printf("Powering off slot %d (CM5)...\n", slot)

	// First short press
	if err := c.pulseGPIO(slot, 1*time.Second); err != nil {
		return fmt.Errorf("failed to send first power off signal: %w", err)
	}

	fmt.Printf("Power off signal sent to slot %d\n", slot)
	return nil
}

// ForceOff forces off a node by holding the GPIO low for 8 seconds
// This is a hard power off, similar to holding a physical power button
func (c *Controller) ForceOff(slot int, boardType BoardType) error {
	if boardType != BoardCM5 {
		return fmt.Errorf("unsupported board type: %s (only cm5 is supported)", boardType)
	}

	fmt.Printf("Force powering off slot %d (CM5)...\n", slot)

	// Hold GPIO low for 8 seconds
	if err := c.pulseGPIO(slot, 8*time.Second); err != nil {
		return fmt.Errorf("failed to force off: %w", err)
	}

	fmt.Printf("Force power off signal sent to slot %d\n", slot)
	return nil
}

// Reset performs a reset on a CM5 node
// This is equivalent to a power cycle
func (c *Controller) Reset(slot int, boardType BoardType) error {
	if boardType != BoardCM5 {
		return fmt.Errorf("unsupported board type: %s (only cm5 is supported)", boardType)
	}

	fmt.Printf("Resetting slot %d (CM5)...\n", slot)

	// Single short press for reset
	if err := c.pulseGPIO(slot, 1*time.Second); err != nil {
		return fmt.Errorf("failed to reset: %w", err)
	}

	fmt.Printf("Reset signal sent to slot %d\n", slot)
	return nil
}

// ResetSwitch performs a reset on the switch chip (GPIO 0 on gpiochip2)
// This toggles the GPIO 0 low then high to reset the switch
func (c *Controller) ResetSwitch() error {
	fmt.Printf("Resetting switch chip (GPIO 0)...\n")

	// Use a short pulse (100ms) to reset
	// The original script was: 0=0 && 0=1
	// pulseGPIO sets 0, waits, sets 1.
	if err := c.pulseGPIO(0, 100*time.Millisecond); err != nil {
		return fmt.Errorf("failed to reset switch chip: %w", err)
	}

	fmt.Printf("Switch chip reset signal sent\n")
	return nil
}
