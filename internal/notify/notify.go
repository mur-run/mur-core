// Package notify provides notification functionality for Slack and Discord.
package notify

import (
	"fmt"

	"github.com/mur-run/mur-core/internal/config"
)

// Options contains notification options.
type Options struct {
	Title       string  // Notification title (e.g., "New Pattern Learned")
	PatternName string  // Pattern name
	Confidence  float64 // Confidence score (0.0-1.0)
	Preview     string  // Content preview
	Source      string  // Source (e.g., session ID)
	PRURL       string  // PR URL (for auto-merge notifications)
	Count       int     // Count (for batch notifications)
}

// Event types for notifications.
const (
	EventPatternAdded     = "pattern_added"
	EventPatternsExtracted = "patterns_extracted"
	EventPRCreated        = "pr_created"
	EventTest             = "test"
)

// Notify sends a notification to all configured channels.
func Notify(event string, opts Options) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.Notifications.Enabled {
		return nil
	}

	var errs []error

	// Send to Slack if configured
	if cfg.Notifications.Slack.WebhookURL != "" {
		if err := NotifySlack(cfg.Notifications.Slack.WebhookURL, event, opts); err != nil {
			errs = append(errs, fmt.Errorf("slack: %w", err))
		}
	}

	// Send to Discord if configured
	if cfg.Notifications.Discord.WebhookURL != "" {
		if err := NotifyDiscord(cfg.Notifications.Discord.WebhookURL, event, opts); err != nil {
			errs = append(errs, fmt.Errorf("discord: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("notification errors: %v", errs)
	}

	return nil
}

// NotifySlackOnly sends only to Slack.
func NotifySlackOnly(event string, opts Options) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Notifications.Slack.WebhookURL == "" {
		return fmt.Errorf("slack webhook not configured")
	}

	return NotifySlack(cfg.Notifications.Slack.WebhookURL, event, opts)
}

// NotifyDiscordOnly sends only to Discord.
func NotifyDiscordOnly(event string, opts Options) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Notifications.Discord.WebhookURL == "" {
		return fmt.Errorf("discord webhook not configured")
	}

	return NotifyDiscord(cfg.Notifications.Discord.WebhookURL, event, opts)
}

// IsConfigured returns true if at least one notification channel is configured.
func IsConfigured() bool {
	cfg, err := config.Load()
	if err != nil {
		return false
	}

	return cfg.Notifications.Enabled &&
		(cfg.Notifications.Slack.WebhookURL != "" ||
			cfg.Notifications.Discord.WebhookURL != "")
}

// formatTitle returns the title for a given event type.
func formatTitle(event string) string {
	switch event {
	case EventPatternAdded:
		return "🧠 New Pattern Learned"
	case EventPatternsExtracted:
		return "🔍 Patterns Extracted"
	case EventPRCreated:
		return "🔀 Auto-Merge PR Created"
	case EventTest:
		return "🧪 Test Notification"
	default:
		return "📢 Murmur Notification"
	}
}

// truncate shortens a string to max length with ellipsis.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
