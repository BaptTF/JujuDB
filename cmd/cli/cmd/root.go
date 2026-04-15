package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "jujudb",
	Short:         "JujuDB CLI - Manage your family inventory",
	Long:          "CLI tool to interact with JujuDB, a family inventory manager (freezer, fridge, pantry, etc.).",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
