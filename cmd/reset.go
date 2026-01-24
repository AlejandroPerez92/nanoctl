package cmd

import (
	"NanoCtl/pkg/gpio"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset [slot]",
	Short: "Reset a node in the specified slot",
	Long: `Reset a CM5 node in the specified slot.
This simulates a single short press to perform a power cycle/reset.`,
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

		if err := controller.Reset(slotNum, gpio.BoardType(boardType)); err != nil {
			fmt.Fprintf(os.Stderr, "Error resetting slot %s: %v\n", slot, err)
			os.Exit(1)
		}

		fmt.Printf("Successfully reset slot %s\n", slot)
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
	resetCmd.Flags().StringP("board", "b", "cm5", "Board type (only 'cm5' is supported)")
}
