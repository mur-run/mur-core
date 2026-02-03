package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage murmur configuration",
	Long:  `View and edit murmur configuration.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		configPath := filepath.Join(home, ".murmur", "config.yaml")

		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("config not found. Run 'mur init' first")
		}

		fmt.Println(string(data))
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		// TODO: implement proper YAML editing
		fmt.Printf("Setting %s = %s\n", key, value)
		fmt.Println("⚠️  Not implemented yet. Edit ~/.murmur/config.yaml directly.")
		return nil
	},
}

var configDefaultCmd = &cobra.Command{
	Use:   "default [tool]",
	Short: "Set the default AI tool",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tool := args[0]

		valid := map[string]bool{
			"claude": true,
			"gemini": true,
			"auggie": true,
			"codex":  true,
		}

		if !valid[tool] {
			return fmt.Errorf("unknown tool: %s. Available: claude, gemini, auggie, codex", tool)
		}

		// TODO: implement proper YAML editing
		fmt.Printf("✓ Default tool set to: %s\n", tool)
		fmt.Println("⚠️  Not fully implemented yet. Edit ~/.murmur/config.yaml directly.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configDefaultCmd)
}
