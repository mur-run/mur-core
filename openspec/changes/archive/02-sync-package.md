# Change Spec: Sync Package

## Goal
Implement `internal/sync` package to sync ~/.murmur/config.yaml MCP settings to AI CLI tools.

## Scope
- `internal/sync/sync.go` — SyncMCP(), SyncAll()
- Update `cmd/mur/cmd/sync.go` to use the new package

## Target CLIs
- Claude Code: `~/.claude/settings.json`
- Gemini CLI: `~/.gemini/settings.json`

## Implementation

### SyncMCP()
1. Load murmur config
2. Read each CLI's settings.json (or create if missing)
3. Merge MCP servers into `mcpServers` key
4. Write back settings.json

### SyncAll()
- Currently: just calls SyncMCP()
- Future: add hooks sync

## Commands
- `mur sync mcp` — sync MCP servers only
- `mur sync all` — sync everything

## Out of Scope
- Hooks sync (next spec)
- Learn/skills sync (future)
