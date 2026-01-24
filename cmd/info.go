package cmd

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"github.com/warthog618/go-gpiocdev"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "List available GPIO chips and their info",
	Long:  `Scans the system for available GPIO character devices and prints their details.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Scanning for GPIO chips...")

		// gpiocdev.Chips() returns a list of paths
		chips := gpiocdev.Chips()

		if len(chips) == 0 {
			fmt.Println("No GPIO chips found.")
			return
		}

		sort.Strings(chips)

		for _, path := range chips {
			printChipInfo(path)
		}
	},
}

func printChipInfo(path string) {
	c, err := gpiocdev.NewChip(path)
	if err != nil {
		fmt.Printf("  %s: [Error opening] %v\n", path, err)
		return
	}
	defer c.Close()

	name := c.Name
	label := c.Label
	lines := c.Lines

	fmt.Printf("  %s: %s (%s) - %d lines\n", filepath.Base(path), name, label, lines)
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
