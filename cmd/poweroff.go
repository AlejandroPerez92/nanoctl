package cmd

import (
	"NanoCtl/pkg/gpio"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var powerOffCmd = &cobra.Command{
	Use:   "poweroff [slot]",
	Short: "Power off a node in the specified slot",
	Long: `Power off a CM5 node in the specified slot.
This simulates two short presses of the power button for graceful shutdown
(compatible with both Desktop and headless systems).

Use the --force flag for a hard power off (8 second hold).`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slot := args[0]
		boardType, _ := cmd.Flags().GetString("board")
		force, _ := cmd.Flags().GetBool("force")

		// Validate board type
		if boardType != "cm5" {
			fmt.Fprintf(os.Stderr, "Error: unsupported board type '%s'. Only 'cm5' is supported.\n", boardType)
			os.Exit(1)
		}

		controller := gpio.NewController()

		// Parse the slot argument
		slotNum, err := strconv.Atoi(slot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: slot must be a number, got '%s'\n", slot)
			os.Exit(1)
		}

		if force {
			err = controller.ForceOff(slotNum, gpio.BoardType(boardType))
		} else {
			err = controller.PowerOff(slotNum, gpio.BoardType(boardType))
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error powering off slot %s: %v\n", slot, err)
			os.Exit(1)
		}

		if force {
			fmt.Printf("Successfully force powered off slot %s\n", slot)
		} else {
			fmt.Printf("Successfully powered off slot %s\n", slot)
		}
	},
}

func init() {
	rootCmd.AddCommand(powerOffCmd)
	powerOffCmd.Flags().StringP("board", "b", "cm5", "Board type (only 'cm5' is supported)")
	powerOffCmd.Flags().BoolP("force", "f", false, "Force power off (8 second hold)")
}
