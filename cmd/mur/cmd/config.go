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

var configRoutingCmd = &cobra.Command{
	Use:   "routing [mode]",
	Short: "Set routing mode",
	Long: `Set the automatic routing mode for tool selection.

Modes:
  auto          Smart routing based on complexity (default)
  manual        Always use default_tool
  cost-first    Prefer free tools unless very complex
  quality-first Prefer paid tools unless very simple

Examples:
  mur config routing auto
  mur config routing cost-first
  mur config routing manual`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mode := args[0]
		validModes := []string{"auto", "manual", "cost-first", "quality-first"}

		valid := false
		for _, m := range validModes {
			if mode == m {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid mode: %s. Valid modes: %v", mode, validModes)
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		cfg.Routing.Mode = mode
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("✓ Routing mode set to: %s\n", mode)
		return nil
	},
}

var configNotificationsCmd = &cobra.Command{
	Use:   "notifications [slack|discord] [webhook-url]",
	Short: "Configure notifications",
	Long: `Configure Slack and Discord webhook notifications.

Examples:
  mur config notifications                      # Show current settings
  mur config notifications --enable             # Enable notifications
  mur config notifications --disable            # Disable notifications
  mur config notifications slack <webhook-url>  # Set Slack webhook
  mur config notifications discord <webhook-url> # Set Discord webhook`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		enable, _ := cmd.Flags().GetBool("enable")
		disable, _ := cmd.Flags().GetBool("disable")

		// Handle enable/disable flags
		if enable {
			cfg.Notifications.Enabled = true
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Println("✓ Notifications enabled")
			return nil
		}

		if disable {
			cfg.Notifications.Enabled = false
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Println("✓ Notifications disabled")
			return nil
		}

		// Handle webhook configuration
		if len(args) >= 2 {
			service := args[0]
			webhookURL := args[1]

			switch service {
			case "slack":
				cfg.Notifications.Slack.WebhookURL = webhookURL
				if err := cfg.Save(); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}
				fmt.Println("✓ Slack webhook configured")
				fmt.Println("  Run 'mur notify test --slack' to test")
				return nil
			case "discord":
				cfg.Notifications.Discord.WebhookURL = webhookURL
				if err := cfg.Save(); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}
				fmt.Println("✓ Discord webhook configured")
				fmt.Println("  Run 'mur notify test --discord' to test")
				return nil
			default:
				return fmt.Errorf("unknown service: %s. Use 'slack' or 'discord'", service)
			}
		}

		// Show current settings
		fmt.Println("Notification Settings")
		fmt.Println("=====================")
		fmt.Println("")

		status := "disabled"
		if cfg.Notifications.Enabled {
			status = "enabled"
		}
		fmt.Printf("  Status: %s\n", status)
		fmt.Println("")

		fmt.Println("  Slack:")
		if cfg.Notifications.Slack.WebhookURL != "" {
			// Mask webhook URL for security
			masked := maskWebhook(cfg.Notifications.Slack.WebhookURL)
			fmt.Printf("    Webhook: %s\n", masked)
		} else {
			fmt.Println("    Webhook: (not configured)")
		}
		if cfg.Notifications.Slack.Channel != "" {
			fmt.Printf("    Channel: %s\n", cfg.Notifications.Slack.Channel)
		}
		fmt.Println("")

		fmt.Println("  Discord:")
		if cfg.Notifications.Discord.WebhookURL != "" {
			masked := maskWebhook(cfg.Notifications.Discord.WebhookURL)
			fmt.Printf("    Webhook: %s\n", masked)
		} else {
			fmt.Println("    Webhook: (not configured)")
		}
		fmt.Println("")

		fmt.Println("Commands:")
		fmt.Println("  mur config notifications --enable")
		fmt.Println("  mur config notifications --disable")
		fmt.Println("  mur config notifications slack <webhook-url>")
		fmt.Println("  mur config notifications discord <webhook-url>")
		fmt.Println("  mur notify test")

		return nil
	},
}

// maskWebhook hides most of a webhook URL for security.
func maskWebhook(url string) string {
	if len(url) < 20 {
		return "***"
	}
	return url[:20] + "..." + url[len(url)-8:]
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configDefaultCmd)
	configCmd.AddCommand(configRoutingCmd)
	configCmd.AddCommand(configNotificationsCmd)

	configNotificationsCmd.Flags().Bool("enable", false, "Enable notifications")
	configNotificationsCmd.Flags().Bool("disable", false, "Disable notifications")
}
