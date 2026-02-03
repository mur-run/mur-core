# 13: Pattern Auto-Merge

**Status:** Done  
**Priority:** Medium  
**Effort:** Medium (2-3 hours)

## Problem

目前 patterns 需要手動發 PR 到 main branch 才能分享給團隊。高 confidence 的 patterns 應該能自動 merge，減少人工操作。

## Solution

建立 auto-merge 功能：
1. **AutoMerge()** — 檢查當前分支的 patterns，將 confidence >= threshold 的自動發 PR 到 main
2. **GitHub CLI 整合** — 使用 `gh` CLI 發 PR
3. **Config 擴展** — 新增 `auto_merge` 和 `merge_threshold` 設定

## Implementation

### Config Changes

```yaml
learning:
  auto_merge: true           # 是否啟用自動 merge
  merge_threshold: 0.8       # confidence 門檻（0.0-1.0）
```

### Files to Modify

```
internal/
  config/config.go           # 新增 AutoMerge, MergeThreshold
  learning/repo.go           # 新增 AutoMerge(), GetHighConfidencePatterns()
cmd/
  mur/cmd/learn.go           # 新增 auto-merge 子命令，push --auto-merge flag
```

### Core Functions

```go
// internal/learning/repo.go

// AutoMerge checks patterns with confidence >= threshold and creates PRs.
func AutoMerge() (*AutoMergeResult, error)

// GetHighConfidencePatterns returns patterns meeting the threshold.
func GetHighConfidencePatterns(threshold float64) ([]learn.Pattern, error)

// CreatePatternPR creates a PR for a pattern using gh CLI.
func CreatePatternPR(pattern learn.Pattern) error
```

### New Commands

1. **`mur learn auto-merge`** — 手動觸發 auto-merge
2. **`mur learn push --auto-merge`** — push 後自動檢查並 merge

### PR Format

- **Title:** `Add pattern: {pattern-name}`
- **Body:** Pattern 內容摘要（description + domain + category + confidence）
- **Labels:** `auto-merge`, `pattern`

## Tests

- `go build ./...` 無 warning
- `mur learn auto-merge --dry-run` 顯示會發的 PR
- 實際發 PR 到 GitHub（需要 gh CLI 已認證）

## Acceptance Criteria

- [x] Config 支援 `auto_merge` 和 `merge_threshold`
- [x] `mur learn auto-merge` 命令可用
- [x] `mur learn push --auto-merge` flag 可用
- [x] 使用 `gh` CLI 發 PR
- [x] `go build ./...` 無 warning

## Dependencies

- GitHub CLI (`gh`) 需要預先安裝並認證
- 需要 repo 有 push 權限

## Related

- 12-learning-repo-sync.md（learning repo 基礎功能）
