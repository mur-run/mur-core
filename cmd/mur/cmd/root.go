package cmd

import (
	"github.com/spf13/cobra"
)

var version = "0.4.0"

var rootCmd = &cobra.Command{
	Use:   "mur",
	Short: "Murmur — Multi-AI CLI 統一管理層 + 跨工具學習系統",
	Long: `Murmur (mur) is a unified management layer for AI CLI tools.

Features:
  • Multi-tool runner — run any AI with one command
  • MCP sync — configure once, sync everywhere
  • Cross-tool learning — what Claude learns, Gemini knows
  • Team knowledge base — share patterns across your team
  • Smart routing — auto-select the cheapest tool for the task

Learn more: https://github.com/mur-run/mur-cli`,
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
