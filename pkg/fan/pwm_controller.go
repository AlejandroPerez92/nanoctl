package fan

import "fmt"

// PWMConfig defines PWM control configuration.
type PWMConfig struct {
	Mode         string
	FrequencyKHz float64
	Hardware     HardwarePWMConfig
}

// HardwarePWMConfig defines sysfs hardware PWM settings.
type HardwarePWMConfig struct {
	Chip     string
	Channel  int
	Inverted bool
}

type pwmController interface {
	Start() error
	Stop()
	SetDutyCycle(dc float64)
}

func newPWMController(config MonitorConfig) (pwmController, error) {
	switch config.PWM.Mode {
	case "software":
		return NewPWMController(config.ChipName, config.Pin, frequencyHz(config.PWM.FrequencyKHz)), nil
	case "hardware":
		periodNs, err := periodNsFromFrequency(config.PWM.FrequencyKHz)
		if err != nil {
			return nil, err
		}
		return NewHardwarePWMController(
			config.PWM.Hardware.Chip,
			config.PWM.Hardware.Channel,
			periodNs,
			config.PWM.Hardware.Inverted,
		), nil
	default:
		return nil, fmt.Errorf("unsupported PWM mode: %s", config.PWM.Mode)
	}
}

func printPWMConfig(config MonitorConfig) {
	switch config.PWM.Mode {
	case "software":
		fmt.Printf("PWM: software %.2fkHz on %s pin %d\n", config.PWM.FrequencyKHz, config.ChipName, config.Pin)
	case "hardware":
		periodNs, err := periodNsFromFrequency(config.PWM.FrequencyKHz)
		if err != nil {
			fmt.Printf("PWM: hardware %s pwm%d inverted=%t (invalid frequency: %v)\n", config.PWM.Hardware.Chip, config.PWM.Hardware.Channel, config.PWM.Hardware.Inverted, err)
			return
		}
		fmt.Printf(
			"PWM: hardware %s pwm%d period %dns (%.2fkHz) inverted=%t\n",
			config.PWM.Hardware.Chip,
			config.PWM.Hardware.Channel,
			periodNs,
			config.PWM.FrequencyKHz,
			config.PWM.Hardware.Inverted,
		)
	}
}
