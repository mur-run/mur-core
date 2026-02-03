package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/karajanchang/murmur-ai/internal/config"
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
	Long: `Set a configuration value.

Supported keys:
  default_tool    - The default AI tool to use

Examples:
  mur config set default_tool gemini`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		switch key {
		case "default_tool":
			if err := cfg.SetDefaultTool(value); err != nil {
				// List available tools
				var tools []string
				for name := range cfg.Tools {
					tools = append(tools, name)
				}
				return fmt.Errorf("%w. Available: %s", err, strings.Join(tools, ", "))
			}
		default:
			return fmt.Errorf("unknown key: %s. Supported: default_tool", key)
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("✓ Set %s = %s\n", key, value)
		return nil
	},
}

var configDefaultCmd = &cobra.Command{
	Use:   "default [tool]",
	Short: "Set the default AI tool",
	Long: `Set the default AI tool to use with 'mur run'.

Available tools depend on your config. Common options:
  claude, gemini, auggie, codex

Examples:
  mur config default gemini
  mur config default claude`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tool := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := cfg.SetDefaultTool(tool); err != nil {
			// List available tools
			var tools []string
			for name := range cfg.Tools {
				tools = append(tools, name)
			}
			return fmt.Errorf("%w. Available: %s", err, strings.Join(tools, ", "))
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("✓ Default tool set to: %s\n", tool)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configDefaultCmd)
}
