# 07: Auggie Support

**Status:** Done  
**Priority:** Medium  
**Effort:** Medium (2-3 hours)

## Problem

Auggie (Augment CLI) is a free AI coding assistant that should be supported by murmur for cost-first routing. Currently murmur only supports Claude Code and Gemini CLI.

Auggie features:
- Free tier (no cost)
- Coding and simple-qa capabilities
- MCP server support via `~/.augment/settings.json`
- Instruction via `-i` flag or `--print` for non-interactive mode
- Skills/rules support via `--rules` flag

## Solution

Add Auggie CLI support to murmur:

1. **Config:** Add auggie to default tools with tier=free
2. **Run:** Handle auggie's CLI format (`-i` for instruction, `-p` for print mode)
3. **Sync:** Support `~/.augment/settings.json` for MCP sync
4. **Router:** Include auggie as free tier option
5. **Health:** Already exists, just verify

## Implementation

### 1. Update internal/config/config.go

Update `defaultConfig()` to enable auggie with proper capabilities:

```go
"auggie": {
    Enabled:      true,
    Binary:       "auggie",
    Flags:        []string{"-p", "-i"},  // print mode + instruction flag
    Tier:         "free",
    Capabilities: []string{"coding", "simple-qa"},
},
```

### 2. Update internal/run (cmd/mur/cmd/run.go)

Auggie uses different CLI format:
- Claude/Gemini: `claude -p "prompt"` (prompt as argument)
- Auggie: `auggie -p -i "prompt"` (prompt follows `-i` flag)

Update `runExecute` to handle this format difference.

### 3. Update internal/sync/sync.go

Add Auggie to `DefaultTargets()`:

```go
{Name: "Auggie", ConfigPath: ".augment/settings.json"},
```

Add event mapping for auggie hooks (if supported).

### 4. Health Check

Already checks for `auggie` binary - no changes needed.

## CLI Differences

| Tool | Print Mode | Instruction |
|------|-----------|-------------|
| Claude | `-p "prompt"` | prompt is positional |
| Gemini | `-p "prompt"` | prompt is positional |
| Auggie | `-p -i "prompt"` | `-i` flag required |

## Files Changed

- `internal/config/config.go` - enable auggie in defaults
- `cmd/mur/cmd/run.go` - handle auggie CLI format
- `internal/sync/sync.go` - add auggie sync target + event mapping

## Testing

```bash
# Build
go build ./...

# Test
go test ./...

# Manual test
mur health
mur run -p "what is 2+2" -t auggie
mur run -p "simple question" --explain  # should route to free (auggie/gemini)
```

## Out of Scope

- Auggie session/resume support
- Auggie persona support
- Auggie `--rules` file sync (future enhancement)
