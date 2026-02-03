# 20: Slack/Discord Notifications

**Status:** In Progress  
**Priority:** Medium  
**Effort:** Small (2-3 hours)

## Problem

當 murmur-ai 學到新 patterns 時，沒有通知機制讓使用者或團隊知道：
- 新 pattern 被加入
- `mur learn extract --auto` 發現新 patterns
- `mur learn auto-merge` 建立 PR

團隊需要即時知道學習進度，並且能快速檢視新發現的 patterns。

## Solution

加入 Slack/Discord webhook 通知功能：
1. **Config** — 在 `~/.murmur/config.yaml` 設定 webhook URLs
2. **Notify Package** — `internal/notify/` 處理通知發送
3. **Learn 整合** — 在適當時機觸發通知
4. **Commands** — `mur notify test` 測試設定、`mur config notifications` 設定 webhook

## Implementation

### Config Changes

```yaml
# ~/.murmur/config.yaml
notifications:
  enabled: true
  slack:
    webhook_url: "https://hooks.slack.com/services/..."
    channel: "#murmur-ai"  # optional, webhook 通常已綁定 channel
  discord:
    webhook_url: "https://discord.com/api/webhooks/..."
```

```go
// internal/config/config.go additions

type NotificationsConfig struct {
    Enabled bool          `yaml:"enabled"`
    Slack   SlackConfig   `yaml:"slack"`
    Discord DiscordConfig `yaml:"discord"`
}

type SlackConfig struct {
    WebhookURL string `yaml:"webhook_url"`
    Channel    string `yaml:"channel"`
}

type DiscordConfig struct {
    WebhookURL string `yaml:"webhook_url"`
}
```

### Notify Package

```go
// internal/notify/notify.go

package notify

type Options struct {
    Title       string
    PatternName string
    Confidence  float64
    Preview     string // content preview
    Source      string // session ID, etc.
}

// Notify sends a notification to all configured channels.
func Notify(message string, opts Options) error

// NotifySlack sends to Slack webhook.
func NotifySlack(webhookURL string, message string, opts Options) error

// NotifyDiscord sends to Discord webhook.
func NotifyDiscord(webhookURL string, message string, opts Options) error
```

### Slack Format (Rich Message)

```json
{
  "blocks": [
    {
      "type": "header",
      "text": {"type": "plain_text", "text": "🧠 New Pattern Learned"}
    },
    {
      "type": "section",
      "fields": [
        {"type": "mrkdwn", "text": "*Pattern:* error-handling-go"},
        {"type": "mrkdwn", "text": "*Confidence:* 85%"}
      ]
    },
    {
      "type": "section",
      "text": {"type": "mrkdwn", "text": "Always handle errors explicitly in Go..."}
    }
  ]
}
```

### Discord Format (Embed)

```json
{
  "embeds": [{
    "title": "🧠 New Pattern Learned",
    "color": 5814783,
    "fields": [
      {"name": "Pattern", "value": "error-handling-go", "inline": true},
      {"name": "Confidence", "value": "85%", "inline": true}
    ],
    "description": "Always handle errors explicitly in Go..."
  }]
}
```

### Learn Integration Points

1. **`mur learn add`** — 新增 pattern 後通知
2. **`mur learn extract --auto`** — 發現並儲存新 patterns 後通知
3. **`mur learn auto-merge`** — PR 建立後通知

### New Commands

```bash
# Test notifications
mur notify test                    # Send test to all configured
mur notify test --slack           # Test Slack only
mur notify test --discord         # Test Discord only

# Configure webhooks
mur config notifications           # Show current settings
mur config notifications --enable  # Enable notifications
mur config notifications --disable # Disable notifications
mur config notifications slack <webhook-url>
mur config notifications discord <webhook-url>
```

## Files to Create/Modify

```
internal/
  notify/
    notify.go       # Main notify logic
    slack.go        # Slack webhook
    discord.go      # Discord webhook
  config/
    config.go       # Add NotificationsConfig
cmd/
  mur/
    cmd/
      notify.go     # notify test command
      config.go     # Add notifications subcommand
      learn.go      # Integrate notifications
```

## Acceptance Criteria

- [ ] `go build ./...` 無 warning
- [ ] `mur notify test` 發送測試通知
- [ ] `mur config notifications` 顯示/設定 webhook
- [ ] `mur learn add` 後發送通知 (if enabled)
- [ ] `mur learn extract --auto` 發現新 patterns 後發送通知
- [ ] `mur learn auto-merge` PR 建立後發送通知
- [ ] Slack/Discord rich format 顯示 pattern 資訊

## Dependencies

- 無新增依賴（使用標準庫 net/http）

## Related

- `internal/learn/pattern.go` — Pattern struct
- `internal/learn/extract.go` — Auto-extract
- `internal/learning/repo.go` — Auto-merge
