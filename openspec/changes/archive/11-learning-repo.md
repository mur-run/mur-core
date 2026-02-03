# Change Spec 11: Learning Repo Sync

## Goal

Enable learning patterns to sync to a dedicated git repository, with each machine using its own branch. This allows:
- Personal patterns per machine (via hostname-based branches)
- Shared patterns across machines (via main branch)
- Automatic push after pattern extraction

## Changes

### 1. Config Update (`internal/config/config.go`)

Extend `LearningConfig`:

```go
type LearningConfig struct {
    AutoExtract   bool   `yaml:"auto_extract"`
    SyncToTools   bool   `yaml:"sync_to_tools"`
    PatternLimit  int    `yaml:"pattern_limit"`
    // New fields for learning repo
    Repo          string `yaml:"repo"`           // git repo URL
    Branch        string `yaml:"branch"`         // branch name (default: hostname)
    AutoPush      bool   `yaml:"auto_push"`      // push after extract
    PullFromMain  bool   `yaml:"pull_from_main"` // pull shared patterns from main
}
```

### 2. New Package: `internal/learning/repo.go`

Functions:
- `RepoDir() string` — returns `~/.murmur/learning-repo/`
- `IsInitialized() bool` — check if repo exists
- `InitRepo(repoURL string) error` — clone repo, create branch
- `Push() error` — commit and push to own branch
- `Pull() error` — pull from main branch
- `Sync() error` — push own + pull main
- `DefaultBranch() string` — returns hostname

### 3. Command Updates (`cmd/mur/cmd/learn.go`)

New subcommands:
- `mur learn init <repo-url>` — initialize learning repo
- `mur learn push` — push patterns to own branch
- `mur learn pull` — pull shared patterns from main
- `mur learn sync` — push + pull

### 4. Extract Integration

Update `runExtractAuto()` to call `learning.Push()` if `auto_push` is enabled.

## File Changes

| File | Change |
|------|--------|
| `internal/config/config.go` | Add new LearningConfig fields |
| `internal/learning/repo.go` | New file: repo management |
| `cmd/mur/cmd/learn.go` | Add init/push/pull/sync subcommands |

## Testing

1. `mur learn init git@github.com:user/learning.git`
2. `mur learn extract --auto` → should auto-push if enabled
3. `mur learn push` → manual push
4. `mur learn pull` → pull from main
5. `mur learn sync` → both directions
