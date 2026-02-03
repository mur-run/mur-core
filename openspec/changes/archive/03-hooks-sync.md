# Change Spec: Hooks Sync

## Goal
Implement `mur sync hooks` to sync hooks configuration from murmur to Claude Code and Gemini CLI.

## Scope
- Add `HooksConfig` to `internal/config/config.go`
- Add `SyncHooks()` to `internal/sync/sync.go`
- Update `cmd/mur/cmd/sync.go` to call `SyncHooks()`

## Target CLIs
- Claude Code: `~/.claude/settings.json`
- Gemini CLI: `~/.gemini/settings.json`

## Event Mapping

| Murmur Event | Claude Code Event | Gemini CLI Event |
|--------------|-------------------|------------------|
| UserPromptSubmit | UserPromptSubmit | BeforeAgent |
| Stop | Stop | AfterAgent |
| BeforeTool | PreToolUse | BeforeTool |
| AfterTool | PostToolUse | AfterTool |

## Configuration Format

### Source: `~/.murmur/config.yaml`
```yaml
hooks:
  UserPromptSubmit:
    - matcher: ""
      hooks:
        - type: command
          command: "echo 'before prompt'"
  Stop:
    - matcher: ""
      hooks:
        - type: command
          command: "bash ~/scripts/on_stop.sh"
```

### Output: Claude Code (`~/.claude/settings.json`)
```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "matcher": "",
        "hooks": [
          {"type": "command", "command": "echo 'before prompt'"}
        ]
      }
    ],
    "Stop": [...]
  }
}
```

### Output: Gemini CLI (`~/.gemini/settings.json`)
```json
{
  "hooks": {
    "BeforeAgent": [...],
    "AfterAgent": [...]
  }
}
```

## Implementation

### Config Types (internal/config/config.go)
```go
type HooksConfig struct {
    UserPromptSubmit []HookGroup `yaml:"UserPromptSubmit"`
    Stop             []HookGroup `yaml:"Stop"`
    BeforeTool       []HookGroup `yaml:"BeforeTool"`
    AfterTool        []HookGroup `yaml:"AfterTool"`
}

type HookGroup struct {
    Matcher string `yaml:"matcher"`
    Hooks   []Hook `yaml:"hooks"`
}

type Hook struct {
    Type    string `yaml:"type"`
    Command string `yaml:"command"`
}
```

### SyncHooks() (internal/sync/sync.go)
1. Load murmur config
2. For each CLI target:
   - Read existing settings.json
   - Map murmur events to CLI-specific events
   - Merge hooks into settings
   - Write back settings.json

### Commands
- `mur sync hooks` — sync hooks only
- `mur sync all` — now includes hooks

## Out of Scope
- Hook validation/testing
- Per-project hooks
- Custom event types beyond the four defined
