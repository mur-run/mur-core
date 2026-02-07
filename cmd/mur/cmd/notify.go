package cmd

import (
	"fmt"

	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/notify"
	"github.com/spf13/cobra"
)

var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Manage notifications",
	Long:  `Send test notifications and manage notification settings.`,
}

var notifyTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Send a test notification",
	Long: `Send a test notification to verify webhook configuration.

Examples:
  mur notify test             # Test all configured webhooks
  mur notify test --slack     # Test Slack only
  mur notify test --discord   # Test Discord only`,
	RunE: func(cmd *cobra.Command, args []string) error {
		slackOnly, _ := cmd.Flags().GetBool("slack")
		discordOnly, _ := cmd.Flags().GetBool("discord")

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		opts := notify.Options{
			PatternName: "test-pattern",
			Confidence:  0.85,
			Preview:     "This is a test notification from murmur-ai to verify your webhook configuration is working correctly.",
		}

		if slackOnly {
			if cfg.Notifications.Slack.WebhookURL == "" {
				return fmt.Errorf("slack webhook not configured. Run: mur config notifications slack <webhook-url>")
			}
			fmt.Println("Sending test notification to Slack...")
			if err := notify.NotifySlackOnly(notify.EventTest, opts); err != nil {
				return fmt.Errorf("failed to send: %w", err)
			}
			fmt.Println("✓ Slack notification sent")
			return nil
		}

		if discordOnly {
			if cfg.Notifications.Discord.WebhookURL == "" {
				return fmt.Errorf("discord webhook not configured. Run: mur config notifications discord <webhook-url>")
			}
			fmt.Println("Sending test notification to Discord...")
			if err := notify.NotifyDiscordOnly(notify.EventTest, opts); err != nil {
				return fmt.Errorf("failed to send: %w", err)
			}
			fmt.Println("✓ Discord notification sent")
			return nil
		}

		// Send to all configured
		if !notify.IsConfigured() {
			return fmt.Errorf("no notifications configured. Run: mur config notifications --help")
		}

		// Temporarily enable for test
		wasEnabled := cfg.Notifications.Enabled
		cfg.Notifications.Enabled = true

		fmt.Println("Sending test notifications...")

		var sent []string
		if cfg.Notifications.Slack.WebhookURL != "" {
			if err := notify.NotifySlack(cfg.Notifications.Slack.WebhookURL, notify.EventTest, opts); err != nil {
				fmt.Printf("  ✗ Slack: %v\n", err)
			} else {
				fmt.Println("  ✓ Slack")
				sent = append(sent, "Slack")
			}
		}

		if cfg.Notifications.Discord.WebhookURL != "" {
			if err := notify.NotifyDiscord(cfg.Notifications.Discord.WebhookURL, notify.EventTest, opts); err != nil {
				fmt.Printf("  ✗ Discord: %v\n", err)
			} else {
				fmt.Println("  ✓ Discord")
				sent = append(sent, "Discord")
			}
		}

		if len(sent) == 0 {
			return fmt.Errorf("no notifications were sent successfully")
		}

		if !wasEnabled {
			fmt.Println("")
			fmt.Println("Note: Notifications are currently disabled.")
			fmt.Println("Run 'mur config notifications --enable' to enable them.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(notifyCmd)
	notifyCmd.AddCommand(notifyTestCmd)

	notifyTestCmd.Flags().Bool("slack", false, "Test Slack webhook only")
	notifyTestCmd.Flags().Bool("discord", false, "Test Discord webhook only")
}
