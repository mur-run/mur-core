# 04: Learn Package

**Status:** Done  
**Priority:** Medium  
**Effort:** Medium (2-3 hours)

## Problem

目前 `mur learn` 只有 stub 實作：
- `mur learn list` 只掃描目錄，沒有統一的 pattern 結構
- `mur learn add` 印出 template 但無法實際新增
- `mur learn sync` 印出 "not implemented"

需要一個 `internal/learn` package 統一管理 patterns，並能同步到各 AI CLI 工具。

## Solution

建立 `internal/learn/` package 提供：
1. **Pattern struct** — 統一的 pattern 資料結構
2. **Storage** — 以 YAML 格式儲存在 `~/.murmur/patterns/`
3. **CRUD 操作** — List, Add, Get, Delete
4. **SyncPatterns()** — 同步到 Claude Code / Gemini CLI

## Implementation

### Files to Create

```
internal/
  learn/
    pattern.go      # Pattern struct + CRUD
    sync.go         # SyncPatterns to CLI tools
```

### Pattern Struct

```go
// internal/learn/pattern.go
package learn

type Pattern struct {
    Name        string  `yaml:"name"`
    Description string  `yaml:"description"`
    Content     string  `yaml:"content"`
    Domain      string  `yaml:"domain"`       // dev, devops, business
    Category    string  `yaml:"category"`     // pattern, decision, lesson
    Confidence  float64 `yaml:"confidence"`   // 0.0 - 1.0
    CreatedAt   string  `yaml:"created_at"`
    UpdatedAt   string  `yaml:"updated_at"`
}
```

### Storage Format

每個 pattern 儲存為獨立 YAML 檔案：
```
~/.murmur/patterns/
  git-commit-style.yaml
  error-handling.yaml
  api-design.yaml
```

### Core Functions

```go
// PatternsDir returns ~/.murmur/patterns/
func PatternsDir() (string, error)

// List returns all patterns
func List() ([]Pattern, error)

// Get returns a pattern by name
func Get(name string) (*Pattern, error)

// Add creates or updates a pattern
func Add(p Pattern) error

// Delete removes a pattern
func Delete(name string) error
```

### Sync Functions

```go
// internal/learn/sync.go

// SyncPatterns syncs all patterns to CLI tools
func SyncPatterns() ([]SyncResult, error)

// syncToClaudeCode syncs patterns to ~/.claude/skills/learned-{name}/SKILL.md
func syncToClaudeCode(patterns []Pattern) error

// syncToGeminiCLI syncs patterns to ~/.gemini/skills/learned-{name}.md
func syncToGeminiCLI(patterns []Pattern) error

// patternToSkill converts a Pattern to SKILL.md format
func patternToSkill(p Pattern) string
```

### SKILL.md Format

```markdown
# {Pattern Name}

{Description}

## Domain
{domain} / {category}

## Content

{content}

---
Confidence: {confidence}
Synced from murmur-ai
```

### Update Commands

修改 `cmd/mur/cmd/learn.go`：

1. **learn list**
   - 使用 `learn.List()` 取得所有 patterns
   - 顯示 name, domain, category, confidence

2. **learn add <name>**
   - 互動式輸入 description, content, domain, category
   - 支援從 stdin 讀取（`cat pattern.yaml | mur learn add mypattern`）
   - 使用 `learn.Add()` 儲存

3. **learn sync**
   - 呼叫 `learn.SyncPatterns()`
   - 顯示同步結果

4. **learn get <name>** (新增)
   - 顯示單一 pattern 內容

5. **learn delete <name>** (新增)
   - 刪除指定 pattern

### Update sync.go

修改 `internal/sync/sync.go`：

```go
func SyncAll() (map[string][]SyncResult, error) {
    results := make(map[string][]SyncResult)
    
    // Sync MCP
    mcpResults, err := SyncMCP()
    // ...
    
    // Sync Hooks
    hooksResults, err := SyncHooks()
    // ...
    
    // Sync Patterns (new)
    patternResults, err := learn.SyncPatterns()
    if err != nil {
        // Log warning but don't fail
    }
    results["patterns"] = patternResults
    
    return results, nil
}
```

## Tests

```go
// internal/learn/pattern_test.go
func TestPatternCRUD(t *testing.T)    // 測試 Add/Get/List/Delete
func TestPatternsDir(t *testing.T)    // 測試目錄建立
func TestPatternValidation(t *testing.T) // 測試 name 驗證

// internal/learn/sync_test.go
func TestPatternToSkill(t *testing.T) // 測試格式轉換
func TestSyncPatterns(t *testing.T)   // 測試同步（mock fs）
```

## Acceptance Criteria

- [x] `go build ./...` 無 warning
- [x] `mur learn add test-pattern` 建立 `~/.murmur/patterns/test-pattern.yaml`
- [x] `mur learn list` 列出所有 patterns
- [x] `mur learn get test-pattern` 顯示 pattern 內容
- [x] `mur learn delete test-pattern` 刪除 pattern
- [x] `mur learn sync` 同步到 Claude Code 和 Gemini CLI
- [x] `mur sync all` 包含 patterns 同步

## Dependencies

- `gopkg.in/yaml.v3` (已在 go.mod)

## Related

- `internal/sync/sync.go` — 需要 import learn package
- `~/.claude/skills/` — Claude Code skills 目錄
- `~/.gemini/skills/` — Gemini CLI skills 目錄
