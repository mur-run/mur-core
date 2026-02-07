package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize murmur configuration",
	Long:  `Initialize murmur by creating the config directory and default settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		murDir := filepath.Join(home, ".murmur")

		// Create directories
		dirs := []string{
			murDir,
			filepath.Join(murDir, "learned"),
			filepath.Join(murDir, "learned", "_global"),
			filepath.Join(murDir, "learned", "dev"),
			filepath.Join(murDir, "learned", "devops"),
			filepath.Join(murDir, "learned", "infra"),
			filepath.Join(murDir, "learned", "security"),
			filepath.Join(murDir, "learned", "product"),
			filepath.Join(murDir, "learned", "business"),
			filepath.Join(murDir, "learned", "ops"),
		}

		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}

		// Create default config
		configPath := filepath.Join(murDir, "config.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			defaultConfig := `# Murmur Configuration
# https://github.com/mur-run/mur-core

# Default AI tool to use
default_tool: claude

# Available tools
tools:
  claude:
    enabled: true
    binary: claude
    flags: ["-p"]
  gemini:
    enabled: true
    binary: gemini
    flags: ["-p"]
  auggie:
    enabled: false
    binary: auggie
    flags: []

# Learning settings
learning:
  auto_extract: true
  sync_to_tools: true
  pattern_limit: 5  # Free tier limit

# MCP settings
mcp:
  sync_enabled: true
  servers: {}
`
			if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
		}

		fmt.Println("✓ Murmur initialized at ~/.murmur")
		fmt.Println("")
		fmt.Println("Next steps:")
		fmt.Println("  mur config     # Configure your AI tools")
		fmt.Println("  mur health     # Check tool availability")
		fmt.Println("  mur run -p \"your prompt\"  # Run a task")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
