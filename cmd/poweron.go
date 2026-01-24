package cmd

import (
	"NanoCtl/pkg/gpio"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var powerOnCmd = &cobra.Command{
	Use:   "poweron [slot]",
	Short: "Power on a node in the specified slot",
	Long: `Power on a CM5 node in the specified slot.
This simulates a single short press of the power button.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slot := args[0]
		boardType, _ := cmd.Flags().GetString("board")

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

		if err := controller.PowerOn(slotNum, gpio.BoardType(boardType)); err != nil {
			fmt.Fprintf(os.Stderr, "Error powering on slot %s: %v\n", slot, err)
			os.Exit(1)
		}

		fmt.Printf("Successfully powered on slot %s\n", slot)
	},
}

func init() {
	rootCmd.AddCommand(powerOnCmd)
	powerOnCmd.Flags().StringP("board", "b", "cm5", "Board type (only 'cm5' is supported)")
}
