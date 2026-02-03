# 08: Skills Sync

**Status:** Done  
**Priority:** Medium  
**Effort:** Medium (2-3 hours)

## Problem

AI CLI 工具（Claude Code, Gemini CLI, Auggie）各有自己的 skills/方法論系統，但：
- 無法統一管理 skills
- 換工具時需要重新設定
- Superpowers 等 plugin 的 skills 無法跨工具共用

需要一個統一的 skills 管理和同步機制。

## Solution

建立 `~/.murmur/skills/` 作為統一 skills 來源，同步到各 CLI：
- Claude Code: `~/.claude/skills/`
- Gemini CLI: `~/.gemini/skills/`
- Auggie: `~/.augment/skills/`

支援從 Superpowers plugin 匯入 skills。

## Implementation

### Files to Create/Modify

```
internal/
  sync/
    sync.go          # 加入 SyncSkills()
    skills.go        # Skills 同步邏輯（新增）

cmd/mur/cmd/
    sync.go          # 加入 `mur sync skills`
    skills.go        # 新增 `mur skills` 命令群組
```

### Skills Format

統一使用 SKILL.md 格式：

```markdown
# Skill Name

## Description

簡短描述這個 skill 的用途。

## Instructions

詳細的指示，AI 會遵循這些指示來執行任務。

- 步驟 1
- 步驟 2
- 步驟 3
```

### Core Types & Functions

```go
// internal/sync/skills.go
package sync

// Skill represents a skill/methodology that can be synced.
type Skill struct {
    Name        string
    Description string
    Instructions string
    SourcePath  string
}

// SkillsTarget represents where skills are synced to.
type SkillsTarget struct {
    Name       string
    SkillsDir  string // relative to home
}

// DefaultSkillsTargets returns supported CLI skill directories.
func DefaultSkillsTargets() []SkillsTarget

// ListSkills returns all available skills from ~/.murmur/skills/
func ListSkills() ([]Skill, error)

// SyncSkills syncs skills to all CLI tools.
func SyncSkills() ([]SyncResult, error)

// ImportSkill imports a skill from a file path.
func ImportSkill(path string) error

// ImportFromSuperpowers imports skills from Superpowers plugin.
func ImportFromSuperpowers() (int, error)
```

### Commands

1. **`mur sync skills`**
   ```bash
   $ mur sync skills
   Syncing skills...
   
     ✓ Claude Code: synced 5 skills
     ✓ Gemini CLI: synced 5 skills
     ✓ Auggie: synced 5 skills
   
   Done.
   ```

2. **`mur skills list`**
   ```bash
   $ mur skills list
   Available skills:
   
     • code-review — Code review methodology
     • tdd-workflow — Test-driven development workflow
     • debugging — Systematic debugging approach
   ```

3. **`mur skills import <path>`**
   ```bash
   $ mur skills import ~/my-skill.md
   Imported skill: my-skill
   ```

4. **`mur skills import --superpowers`**
   ```bash
   $ mur skills import --superpowers
   Importing from Superpowers plugin...
   Imported 3 skills: code-review, tdd-workflow, debugging
   ```

### Update SyncAll

修改 `SyncAll()` 包含 skills 同步：

```go
func SyncAll() (map[string][]SyncResult, error) {
    results := make(map[string][]SyncResult)

    // ... existing MCP and hooks sync ...

    // Sync Skills
    skillsResults, err := SyncSkills()
    if err != nil {
        // Log warning but don't fail - skills might not exist yet
        fmt.Printf("Skills sync warning: %v\n", err)
    } else {
        results["skills"] = skillsResults
    }

    return results, nil
}
```

## Acceptance Criteria

- [x] `go build ./...` 無 warning
- [x] `mur sync skills` 同步 skills 到所有 CLI
- [x] `mur skills list` 列出可用 skills
- [x] `mur skills import <path>` 匯入 skill
- [x] `mur skills import --superpowers` 從 Superpowers 匯入
- [x] `mur sync all` 包含 skills 同步
- [x] Skills 正確複製到各 CLI 目錄

## Directory Structure

```
~/.murmur/
  skills/
    code-review.md
    tdd-workflow.md
    debugging.md

~/.claude/
  skills/           # synced from murmur
    code-review.md
    tdd-workflow.md
    debugging.md

~/.gemini/
  skills/           # synced from murmur
    code-review.md
    ...

~/.augment/
  skills/           # synced from murmur
    code-review.md
    ...
```

## Dependencies

- None (uses standard library)

## Related

- Superpowers plugin: `~/.claude/plugins/using-superpowers/skills/`
- 02-sync-package.md (sync 基礎架構)
