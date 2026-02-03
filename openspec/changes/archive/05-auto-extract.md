# 05: Auto-Extract Patterns from Sessions

**Status:** Done  
**Priority:** Medium  
**Effort:** Medium (3-4 hours)

## Problem

目前 patterns 必須手動建立。開發者在 coding sessions 中經常會：
- 重複解決相同問題
- 發現有用的 patterns 但忘記記錄
- 錯過從 AI 對話中學習的機會

需要自動從 Claude Code session transcripts 擷取有價值的 patterns。

## Solution

在 `internal/learn/` 加入 extraction 功能：
1. **ExtractFromSession()** — 分析 session transcript 並辨識 patterns
2. **Session 整合** — 讀取 `~/.claude/projects/*/*.jsonl`
3. **Command** — `mur learn extract` 手動或自動擷取
4. **Confirmation** — 使用者確認後才儲存 pattern

## Implementation

### Files to Create/Modify

```
internal/
  learn/
    extract.go      # Extraction logic
    session.go      # Session reading & parsing
cmd/
  mur/
    cmd/
      learn.go      # Add extract subcommand
```

### Session Struct

```go
// internal/learn/session.go

// Session represents a Claude Code session.
type Session struct {
    ID        string
    Project   string
    Path      string
    Messages  []SessionMessage
    CreatedAt time.Time
}

// SessionMessage represents a message in a session.
type SessionMessage struct {
    Type      string   // "user", "assistant", "progress", etc.
    Role      string   // "user", "assistant"
    Content   string   // Text content
    Timestamp time.Time
}

// ListSessions returns available sessions from ~/.claude/projects/
func ListSessions() ([]Session, error)

// LoadSession loads a session by ID or path.
func LoadSession(idOrPath string) (*Session, error)

// RecentSessions returns sessions from the last N days.
func RecentSessions(days int) ([]Session, error)
```

### Extraction Logic

```go
// internal/learn/extract.go

// ExtractedPattern represents a potential pattern found in a session.
type ExtractedPattern struct {
    Pattern     Pattern  // The pattern to potentially save
    Source      string   // Session ID
    Evidence    []string // Relevant snippets that support this pattern
    Confidence  float64  // Extraction confidence
}

// ExtractFromSession analyzes a session and extracts patterns.
// Uses keyword/heuristic matching (no LLM required).
func ExtractFromSession(sessionPath string) ([]ExtractedPattern, error)

// extractFromMessages performs the actual extraction.
func extractFromMessages(messages []SessionMessage) []ExtractedPattern

// PatternMatchers contains keyword patterns to detect.
var PatternMatchers = []PatternMatcher{
    // Best practices
    {Keywords: []string{"best practice", "recommended", "should always"},
     Category: "pattern", Domain: "dev"},
    
    // Error handling
    {Keywords: []string{"error handling", "handle error", "catch", "recover"},
     Category: "pattern", Domain: "dev"},
    
    // Decisions
    {Keywords: []string{"decided to", "chose", "trade-off", "reason"},
     Category: "decision", Domain: "dev"},
    
    // Lessons learned
    {Keywords: []string{"learned", "realized", "mistake", "gotcha", "pitfall"},
     Category: "lesson", Domain: "dev"},
    
    // Templates/snippets
    {Keywords: []string{"template", "boilerplate", "snippet", "example"},
     Category: "template", Domain: "dev"},
}

// PatternMatcher defines how to detect a pattern type.
type PatternMatcher struct {
    Keywords []string
    Category string
    Domain   string
}
```

### Extract Command

```go
// cmd/mur/cmd/learn.go (additions)

var learnExtractCmd = &cobra.Command{
    Use:   "extract [--session <id>] [--auto] [--dry-run]",
    Short: "Extract patterns from coding sessions",
    Long: `Extract patterns from Claude Code session transcripts.

Examples:
  mur learn extract                      # Interactive: choose session
  mur learn extract --session abc123     # From specific session
  mur learn extract --auto               # Scan recent sessions
  mur learn extract --auto --dry-run     # Preview without saving`,
    RunE: runExtract,
}

func runExtract(cmd *cobra.Command, args []string) error {
    sessionID, _ := cmd.Flags().GetString("session")
    auto, _ := cmd.Flags().GetBool("auto")
    dryRun, _ := cmd.Flags().GetBool("dry-run")
    
    if auto {
        return extractAuto(dryRun)
    }
    
    if sessionID != "" {
        return extractFromSession(sessionID, dryRun)
    }
    
    // Interactive: list sessions and let user choose
    return extractInteractive(dryRun)
}

func extractAuto(dryRun bool) error {
    // Get recent sessions (last 7 days)
    sessions, err := learn.RecentSessions(7)
    // ...
    
    for _, session := range sessions {
        patterns, err := learn.ExtractFromSession(session.Path)
        // Display and optionally save
    }
}

func extractFromSession(sessionID string, dryRun bool) error {
    session, err := learn.LoadSession(sessionID)
    patterns, err := learn.ExtractFromSession(session.Path)
    
    for _, ep := range patterns {
        displayExtractedPattern(ep)
        
        if !dryRun {
            // Confirm with user
            if confirmSave(ep.Pattern.Name) {
                learn.Add(ep.Pattern)
            }
        }
    }
}
```

### Flags

```go
func init() {
    learnCmd.AddCommand(learnExtractCmd)
    
    learnExtractCmd.Flags().StringP("session", "s", "", "Session ID to extract from")
    learnExtractCmd.Flags().Bool("auto", false, "Automatically scan recent sessions")
    learnExtractCmd.Flags().Bool("dry-run", false, "Show what would be extracted without saving")
}
```

## Heuristic Extraction Rules

不使用 LLM，而是用 keyword/heuristic 方式：

### Detection Patterns

1. **Best Practices**
   - Keywords: "best practice", "recommended", "should always", "prefer"
   - Context: Look for sentences with actionable advice
   
2. **Error Handling**
   - Keywords: "error handling", "handle error", "catch", "recover", "panic"
   - Context: Code blocks with error checks
   
3. **Architecture Decisions**
   - Keywords: "decided to", "chose", "because", "trade-off", "instead of"
   - Context: Explanations of why a certain approach was taken
   
4. **Gotchas/Lessons**
   - Keywords: "gotcha", "pitfall", "careful", "watch out", "important"
   - Context: Warnings or lessons learned

5. **Templates**
   - Keywords: "template", "boilerplate", "scaffold", "starter"
   - Context: Code blocks meant for reuse

### Confidence Scoring

```go
func calculateConfidence(text string, matcher PatternMatcher) float64 {
    score := 0.0
    
    // Base score for keyword matches
    for _, kw := range matcher.Keywords {
        if strings.Contains(strings.ToLower(text), kw) {
            score += 0.2
        }
    }
    
    // Bonus for code blocks
    if strings.Contains(text, "```") {
        score += 0.1
    }
    
    // Bonus for longer, structured content
    if len(text) > 200 {
        score += 0.1
    }
    
    return min(score, 1.0)
}
```

## Output Format

```
Extracted Patterns
==================

1. [pattern] error-handling-go (confidence: 0.7)
   Source: session 93e7f773
   
   Always handle errors explicitly in Go:
   ```go
   if err != nil {
       return fmt.Errorf("context: %w", err)
   }
   ```
   
   Save this pattern? [y/N/e(dit)]

2. [decision] use-cobra-cli (confidence: 0.6)
   Source: session 93e7f773
   
   Chose Cobra for CLI framework because...
   
   Save this pattern? [y/N/e(dit)]
```

## Tests

```go
// internal/learn/extract_test.go
func TestExtractFromMessages(t *testing.T)
func TestCalculateConfidence(t *testing.T)
func TestPatternMatchers(t *testing.T)

// internal/learn/session_test.go
func TestListSessions(t *testing.T)
func TestLoadSession(t *testing.T)
func TestRecentSessions(t *testing.T)
```

## Acceptance Criteria

- [x] `go build ./...` 無 warning
- [x] `mur learn extract --help` 顯示用法
- [x] `mur learn extract --session <id>` 從指定 session 擷取
- [x] `mur learn extract --auto` 掃描最近 sessions
- [x] `mur learn extract --dry-run` 顯示但不儲存
- [x] 使用者確認後才儲存 pattern
- [ ] 支援編輯 pattern 再儲存 (future enhancement)

## Future Enhancements

- Hook 整合：在 Stop event 時自動背景擷取
- LLM 輔助：使用 AI 提高擷取品質
- 跨 session 分析：辨識多個 sessions 中重複出現的 patterns
- 自動分類：更精確的 domain/category 判斷

## Dependencies

- 無新增依賴

## Related

- `internal/learn/pattern.go` — Pattern struct
- `internal/learn/sync.go` — 擷取後可同步到 AI tools
- `~/.claude/projects/` — Claude Code sessions 儲存位置
