package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/envfile"
)

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
	Version: Version,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate("mur version {{.Version}}\n")

	// Load ~/.mur/.env before any subcommand runs
	_ = envfile.Load()

	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "V", false, "verbose output")
}
