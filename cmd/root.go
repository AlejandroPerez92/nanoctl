package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nanoctl",
	Short: "NanoCtl - Manage your Nano Cluster",
	Long: `NanoCtl is a CLI tool for managing Nano Cluster nodes.
It provides commands to power on, power off, and reset CM5 nodes
in your cluster using GPIO controls.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here if needed
}
