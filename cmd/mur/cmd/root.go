package cmd

import (
	"github.com/spf13/cobra"
)

var version = "0.5.1"

var rootCmd = &cobra.Command{
	Use:   "mur",
	Short: "Continuous learning for AI assistants",
	Long: `mur — Continuous learning for AI assistants.

Learn once, remember forever. mur syncs your patterns to all AI CLIs.

Quick start:
  mur init              # Interactive setup
  mur sync              # Sync patterns to CLIs
  mur learn             # Manage patterns
  mur stats             # View statistics

Learn more: https://github.com/mur-run/mur-core`,
	Version: version,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate("mur version {{.Version}}\n")

	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}
