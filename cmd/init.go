package cmd

import (
	"fmt"
	"github.com/AlejandroPerez92/nanoctl/pkg/gpio"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the Nano Cluster hardware",
	Long: `Performs initialization tasks for the Nano Cluster board.
Currently, this resets the switch chip (GPIO 0).
This should be run once after booting the carrier board.`,
	Run: func(cmd *cobra.Command, args []string) {
		controller := gpio.NewController()

		if err := controller.ResetSwitch(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing hardware: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Initialization complete.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
