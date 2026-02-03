package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// slackMessage represents a Slack message with blocks.
type slackMessage struct {
	Blocks []slackBlock `json:"blocks"`
}

type slackBlock struct {
	Type   string           `json:"type"`
	Text   *slackText       `json:"text,omitempty"`
	Fields []slackTextField `json:"fields,omitempty"`
}

type slackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type slackTextField struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NotifySlack sends a notification to a Slack webhook.
func NotifySlack(webhookURL string, event string, opts Options) error {
	msg := buildSlackMessage(event, opts)

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	return nil
}

func buildSlackMessage(event string, opts Options) slackMessage {
	blocks := []slackBlock{
		{
			Type: "header",
			Text: &slackText{
				Type: "plain_text",
				Text: formatTitle(event),
			},
		},
	}

	switch event {
	case EventPatternAdded, EventTest:
		if opts.PatternName != "" {
			fields := []slackTextField{
				{Type: "mrkdwn", Text: fmt.Sprintf("*Pattern:* `%s`", opts.PatternName)},
			}
			if opts.Confidence > 0 {
				fields = append(fields, slackTextField{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Confidence:* %.0f%%", opts.Confidence*100),
				})
			}
			blocks = append(blocks, slackBlock{
				Type:   "section",
				Fields: fields,
			})
		}
		if opts.Preview != "" {
			blocks = append(blocks, slackBlock{
				Type: "section",
				Text: &slackText{
					Type: "mrkdwn",
					Text: truncate(opts.Preview, 500),
				},
			})
		}

	case EventPatternsExtracted:
		text := fmt.Sprintf("Extracted *%d* new pattern(s)", opts.Count)
		if opts.Source != "" {
			text += fmt.Sprintf(" from session `%s`", opts.Source)
		}
		blocks = append(blocks, slackBlock{
			Type: "section",
			Text: &slackText{
				Type: "mrkdwn",
				Text: text,
			},
		})

	case EventPRCreated:
		fields := []slackTextField{
			{Type: "mrkdwn", Text: fmt.Sprintf("*Pattern:* `%s`", opts.PatternName)},
		}
		if opts.Confidence > 0 {
			fields = append(fields, slackTextField{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*Confidence:* %.0f%%", opts.Confidence*100),
			})
		}
		blocks = append(blocks, slackBlock{
			Type:   "section",
			Fields: fields,
		})
		if opts.PRURL != "" {
			blocks = append(blocks, slackBlock{
				Type: "section",
				Text: &slackText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("<%s|View PR>", opts.PRURL),
				},
			})
		}
	}

	// Add divider at the end
	blocks = append(blocks, slackBlock{Type: "divider"})

	return slackMessage{Blocks: blocks}
}
