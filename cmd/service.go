package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

const serviceTemplate = `[Unit]
Description=NanoCtl Fan Controller
After=multi-user.target

[Service]
ExecStart={{.BinaryPath}} fan --chip {{.Chip}} --pin {{.Pin}} --target {{.Target}} --kp {{.Kp}} --ki {{.Ki}} --kd {{.Kd}} --interval {{.Interval}}
Restart=always
User=root
Type=simple

[Install]
WantedBy=multi-user.target
`

type ServiceConfig struct {
	BinaryPath string
	Chip       string
	Pin        int
	Target     float64
	Kp, Ki, Kd float64
	Interval   string
}

var installServiceCmd = &cobra.Command{
	Use:   "install-service",
	Short: "Install nanoctl-fan as a systemd service",
	Long: `Creates and enables a systemd service for the fan controller.
It uses the flags provided to this command to configure the service.

Example:
  sudo nanoctl install-service --target 60 --pin 13`,
	Run: func(cmd *cobra.Command, args []string) {
		if os.Geteuid() != 0 {
			fmt.Println("Error: This command must be run as root (sudo).")
			os.Exit(1)
		}

		// Get absolute path to the current binary
		binaryPath, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get binary path: %v\n", err)
			os.Exit(1)
		}

		// Resolve symlinks just in case
		binaryPath, err = filepath.EvalSymlinks(binaryPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to resolve binary path: %v\n", err)
			os.Exit(1)
		}

		config := ServiceConfig{
			BinaryPath: binaryPath,
			Chip:       fanChip,
			Pin:        fanPin,
			Target:     targetTemp,
			Kp:         kp,
			Ki:         ki,
			Kd:         kd,
			Interval:   checkInterval.String(),
		}

		servicePath := "/etc/systemd/system/nanoctl-fan.service"
		fmt.Printf("Creating service file at %s...\n", servicePath)

		f, err := os.Create(servicePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create service file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		tmpl, err := template.New("service").Parse(serviceTemplate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse template: %v\n", err)
			os.Exit(1)
		}

		if err := tmpl.Execute(f, config); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write service file: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Service file created.")
		fmt.Println("Reloading systemd daemon...")
		if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to reload daemon: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Enabling and starting nanoctl-fan service...")
		if err := exec.Command("systemctl", "enable", "--now", "nanoctl-fan").Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to enable/start service: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Success! Fan controller is now running as a service.")
		fmt.Println("Check status with: systemctl status nanoctl-fan")
	},
}

func init() {
	rootCmd.AddCommand(installServiceCmd)

	// Reuse flags from fan command logic
	// We need to redefine them here or access the variables.
	// Since variables are package-level in cmd/fan.go, we can reuse them if they are exported or in same package.
	// They are in 'cmd' package, but lowercase. So they are accessible in this file (same package).

	installServiceCmd.Flags().StringVar(&fanChip, "chip", "gpiochip0", "GPIO chip name")
	installServiceCmd.Flags().IntVar(&fanPin, "pin", 13, "GPIO pin number")
	installServiceCmd.Flags().Float64Var(&targetTemp, "target", 55.0, "Target CPU temperature")
	installServiceCmd.Flags().Float64Var(&kp, "kp", 5.0, "Proportional gain")
	installServiceCmd.Flags().Float64Var(&ki, "ki", 0.1, "Integral gain")
	installServiceCmd.Flags().Float64Var(&kd, "kd", 0.5, "Derivative gain")
	installServiceCmd.Flags().DurationVar(&checkInterval, "interval", 1*time.Second, "Temperature check interval")
}
