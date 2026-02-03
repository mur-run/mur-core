# 09: Team Sharing

**Status:** Done  
**Priority:** High  
**Effort:** Medium (2-3 hours)

## Problem

目前 murmur patterns、hooks、skills 只能在本機使用。團隊成員無法共享：
- 學到的 patterns
- 自訂的 hooks
- MCP 設定

需要一個 Git-based 的團隊共享機制，讓知識可以跨團隊同步。

## Solution

建立 `internal/team` package 提供 Git-based 的團隊共享：
1. **Team Repo** — 一個 git repo 存放共享的 patterns/hooks/skills
2. **Config** — 在 config 加入 team 設定
3. **Commands** — `mur team init/pull/push/sync/status`
4. **Integration** — patterns 可標記為 team-shared

## Team Repo Structure

```
team-repo/
├── patterns/     # 共享的 patterns (*.yaml)
├── hooks/        # 共享的 hooks (hooks.yaml)
├── skills/       # 共享的 skills (*/SKILL.md)
└── mcp/          # 共享的 MCP 設定 (servers.yaml)
```

## Implementation

### Config Changes

```go
// internal/config/config.go

type Config struct {
    // ... existing fields
    Team TeamConfig `yaml:"team"`
}

type TeamConfig struct {
    Repo     string `yaml:"repo"`      // git repo URL
    Branch   string `yaml:"branch"`    // default "main"
    AutoSync bool   `yaml:"auto_sync"` // auto sync on pull
}
```

### Files to Create

```
internal/
  team/
    team.go         # Core team operations
```

### Core Functions

```go
// internal/team/team.go
package team

// TeamDir returns ~/.murmur/team/
func TeamDir() (string, error)

// IsInitialized checks if team repo is configured and cloned
func IsInitialized() bool

// Clone clones the team repo to ~/.murmur/team/
func Clone(repoURL string) error

// Pull pulls latest changes from remote
func Pull() error

// Push pushes local changes to remote
func Push(message string) error

// Sync performs bidirectional sync
func Sync() error

// Status returns current team repo status
func Status() (*TeamStatus, error)

type TeamStatus struct {
    Initialized bool
    RepoURL     string
    Branch      string
    LocalPath   string
    Ahead       int
    Behind      int
    Modified    []string
}
```

### Commands

```go
// cmd/mur/cmd/team.go

var teamCmd = &cobra.Command{
    Use:   "team",
    Short: "Manage team knowledge sharing",
}

// mur team init <repo-url>
var teamInitCmd = &cobra.Command{
    Use:   "init <repo-url>",
    Short: "Initialize team repo connection",
}

// mur team pull
var teamPullCmd = &cobra.Command{
    Use:   "pull",
    Short: "Pull latest team changes",
}

// mur team push
var teamPushCmd = &cobra.Command{
    Use:   "push",
    Short: "Push local changes to team",
}

// mur team sync
var teamSyncCmd = &cobra.Command{
    Use:   "sync",
    Short: "Bidirectional sync with team",
}

// mur team status
var teamStatusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show team repo status",
}
```

### Integration with Learn

修改 `internal/learn/pattern.go`：

```go
type Pattern struct {
    // ... existing fields
    TeamShared bool `yaml:"team_shared"` // 是否共享到 team
}
```

修改 `internal/learn/sync.go`：

```go
// SyncToTeam syncs team-shared patterns to team repo
func SyncToTeam() error {
    patterns, _ := List()
    for _, p := range patterns {
        if p.TeamShared {
            // Copy to ~/.murmur/team/patterns/
        }
    }
}
```

## Git Operations

使用 `os/exec` 執行 git 命令：

```go
func runGit(args ...string) (string, error) {
    teamDir, _ := TeamDir()
    cmd := exec.Command("git", args...)
    cmd.Dir = teamDir
    output, err := cmd.CombinedOutput()
    return string(output), err
}
```

## Acceptance Criteria

- [x] `go build ./...` 無 warning
- [x] `mur team init <url>` clone repo 到 `~/.murmur/team/`
- [x] `mur team status` 顯示 repo 狀態
- [x] `mur team pull` 拉取最新變更
- [x] `mur team push` 推送本地變更
- [x] `mur team sync` 雙向同步
- [x] Pattern 可標記 `team_shared: true`
- [x] `mur learn sync` 自動同步 team-shared patterns

## Dependencies

- Git CLI (系統已安裝)

## Related

- `internal/learn/` — Pattern 管理
- `internal/config/` — 設定管理
- `~/.murmur/team/` — Team repo 本地路徑
