# 12: Codex & OpenCode CLI Support

**Status:** Done  
**Priority:** Medium  
**Effort:** Medium (2-3 hours)

## Problem

murmur currently supports Claude Code, Gemini CLI, and Auggie. Two more AI coding CLIs should be added:

1. **Codex CLI** (OpenAI's coding CLI) - paid tier, full capabilities
2. **OpenCode** (open-source LLM coding CLI) - free tier

## Solution

Add both tools to murmur's config, sync, and routing systems.

### Codex CLI
- Binary: `codex`
- Tier: `paid`
- Config method: `~/.codex/instructions.md` (no settings.json)
- Capabilities: coding, analysis, tool-use, architecture

### OpenCode
- Binary: `opencode`
- Tier: `free`
- Config method: `~/.opencode/settings.json`
- Capabilities: coding, simple-qa, analysis

## Implementation

### 1. Update internal/config/config.go

Add to `defaultConfig()`:

```go
"codex": {
    Enabled:      true,
    Binary:       "codex",
    Flags:        []string{},  // uses stdin or -q for quiet
    Tier:         "paid",
    Capabilities: []string{"coding", "analysis", "tool-use", "architecture"},
},
"opencode": {
    Enabled:      true,
    Binary:       "opencode",
    Flags:        []string{},
    Tier:         "free",
    Capabilities: []string{"coding", "simple-qa", "analysis"},
},
```

### 2. Update internal/sync/sync.go

Add to `DefaultTargets()`:

```go
{Name: "OpenCode", ConfigPath: ".opencode/settings.json"},
```

For Codex, create special handling since it uses instructions.md instead of settings.json.

Add new function `SyncPatternsToCodex()` that writes to `~/.codex/instructions.md`.

### 3. Health Check

Already includes codex and opencode - no changes needed.

### 4. Router

No changes needed - router uses tier from config automatically.

## Files Changed

- `internal/config/config.go` - add codex/opencode to defaults
- `internal/sync/sync.go` - add OpenCode target, Codex pattern sync

## Testing

```bash
# Build
go build ./...

# Manual test
mur health  # should show codex and opencode status
mur run -p "test" -t codex
mur run -p "test" -t opencode
mur run -p "simple question" --explain  # should route to free
```

## Acceptance Criteria

- [x] `go build ./...` compiles without errors
- [ ] codex appears in default config with tier=paid
- [ ] opencode appears in default config with tier=free
- [ ] `mur sync` syncs MCP to opencode's settings.json
- [ ] Codex patterns sync to ~/.codex/instructions.md
- [ ] Router selects codex for complex tasks (paid tier)
- [ ] Router selects opencode for simple tasks (free tier)
