package cmd

import (
	"NanoCtl/pkg/config"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

const serviceTemplate = `[Unit]
Description=NanoCtl Fan Controller
After=multi-user.target

[Service]
ExecStart={{.BinaryPath}} fan
Restart=always
User=root
Type=simple

[Install]
WantedBy=multi-user.target
`

type ServiceConfig struct {
	BinaryPath string
}

var installServiceCmd = &cobra.Command{
	Use:   "install-service",
	Short: "Install nanoctl-fan as a systemd service",
	Long: `Creates and enables a systemd service for the fan controller.

This command will:
1. Create a default configuration file at /etc/nanoctl/fan.yaml (if it doesn't exist)
2. Create a systemd service file at /etc/systemd/system/nanoctl-fan.service
3. Enable and start the service

After installation, you can edit /etc/nanoctl/fan.yaml to customize the fan controller settings.

Example:
  sudo nanoctl install-service`,
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

		// Check if config file exists, if not create default
		configPath := config.DefaultConfigPath
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Printf("Creating default configuration at %s...\n", configPath)
			if err := config.CreateDefaultConfig(configPath); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create config file: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Configuration file created.")
		} else {
			fmt.Printf("Configuration file already exists at %s\n", configPath)
		}

		// Create systemd service file
		serviceConfig := ServiceConfig{
			BinaryPath: binaryPath,
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

		if err := tmpl.Execute(f, serviceConfig); err != nil {
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

		fmt.Println("\nSuccess! Fan controller is now running as a service.")
		fmt.Printf("Configuration file: %s\n", configPath)
		fmt.Println("Check status with: systemctl status nanoctl-fan")
		fmt.Println("View logs with: journalctl -u nanoctl-fan -f")
		fmt.Printf("\nTo customize settings, edit %s and restart the service:\n", configPath)
		fmt.Println("  sudo systemctl restart nanoctl-fan")
	},
}

func init() {
	rootCmd.AddCommand(installServiceCmd)
}
