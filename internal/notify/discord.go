package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// discordMessage represents a Discord webhook message.
type discordMessage struct {
	Embeds []discordEmbed `json:"embeds"`
}

type discordEmbed struct {
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Color       int            `json:"color"`
	Fields      []discordField `json:"fields,omitempty"`
	Footer      *discordFooter `json:"footer,omitempty"`
	Timestamp   string         `json:"timestamp,omitempty"`
}

type discordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type discordFooter struct {
	Text string `json:"text"`
}

// Discord embed colors
const (
	colorGreen  = 0x58D68D // Success/pattern added
	colorBlue   = 0x5DADE2 // Info/extracted
	colorPurple = 0xAF7AC5 // PR created
	colorGray   = 0x95A5A6 // Test
)

// NotifyDiscord sends a notification to a Discord webhook.
func NotifyDiscord(webhookURL string, event string, opts Options) error {
	msg := buildDiscordMessage(event, opts)

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal discord message: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send discord notification: %w", err)
	}
	defer resp.Body.Close()

	// Discord returns 204 No Content on success
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("discord returned status %d", resp.StatusCode)
	}

	return nil
}

func buildDiscordMessage(event string, opts Options) discordMessage {
	embed := discordEmbed{
		Title:     formatTitle(event),
		Color:     getColorForEvent(event),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	switch event {
	case EventPatternAdded, EventTest:
		if opts.PatternName != "" {
			embed.Fields = append(embed.Fields, discordField{
				Name:   "Pattern",
				Value:  fmt.Sprintf("`%s`", opts.PatternName),
				Inline: true,
			})
		}
		if opts.Confidence > 0 {
			embed.Fields = append(embed.Fields, discordField{
				Name:   "Confidence",
				Value:  fmt.Sprintf("%.0f%%", opts.Confidence*100),
				Inline: true,
			})
		}
		if opts.Preview != "" {
			embed.Description = truncate(opts.Preview, 500)
		}

	case EventPatternsExtracted:
		embed.Description = fmt.Sprintf("Extracted **%d** new pattern(s)", opts.Count)
		if opts.Source != "" {
			embed.Fields = append(embed.Fields, discordField{
				Name:   "Source",
				Value:  fmt.Sprintf("`%s`", opts.Source),
				Inline: true,
			})
		}

	case EventPRCreated:
		if opts.PatternName != "" {
			embed.Fields = append(embed.Fields, discordField{
				Name:   "Pattern",
				Value:  fmt.Sprintf("`%s`", opts.PatternName),
				Inline: true,
			})
		}
		if opts.Confidence > 0 {
			embed.Fields = append(embed.Fields, discordField{
				Name:   "Confidence",
				Value:  fmt.Sprintf("%.0f%%", opts.Confidence*100),
				Inline: true,
			})
		}
		if opts.PRURL != "" {
			embed.Description = fmt.Sprintf("[View PR](%s)", opts.PRURL)
		}
	}

	embed.Footer = &discordFooter{Text: "murmur-ai"}

	return discordMessage{Embeds: []discordEmbed{embed}}
}

func getColorForEvent(event string) int {
	switch event {
	case EventPatternAdded:
		return colorGreen
	case EventPatternsExtracted:
		return colorBlue
	case EventPRCreated:
		return colorPurple
	default:
		return colorGray
	}
}
