package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	// These variables are set during build time via -ldflags
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of nanoctl",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("nanoctl version %s (commit: %s, built: %s)\n", Version, Commit, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
