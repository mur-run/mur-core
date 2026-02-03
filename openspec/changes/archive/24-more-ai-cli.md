# Change Spec: More AI CLI Support

## Metadata
- **ID**: 24
- **Status**: In Progress
- **Priority**: Medium
- **Author**: murmur-ai team
- **Created**: 2025-02-03

## Summary
Add support for additional AI CLI tools: Aider, Continue, and Cursor.

## Motivation
Expand murmur-ai's compatibility with popular AI coding tools beyond Claude Code, Gemini CLI, Auggie, Codex, and OpenCode.

## Tools Research

### Aider
- **Type**: Terminal-based AI pair programming
- **Config**: `~/.aider.conf.yml` (global) or `.aider.conf.yml` (project-level)
- **Tier**: Free (can use free models) or Paid (with API keys)
- **Install**: `pip install aider-chat`
- **Binary**: `aider`
- **Skills/Prompts**: Uses conventions files, `.aider/` directory

### Continue
- **Type**: IDE extension (VS Code, JetBrains) with CLI capabilities
- **Config**: `~/.continue/config.json`
- **Tier**: Free (open source, bring your own API keys)
- **Binary**: `continue` (if CLI installed)
- **MCP Support**: Yes, via config.json `mcpServers`
- **Skills/Prompts**: Custom instructions in config.json

### Cursor
- **Type**: AI-first code editor (fork of VS Code)
- **Config**: 
  - `~/.cursor/` directory
  - `.cursorrules` (project-level instructions)
  - Cursor Settings (GUI-based)
- **Tier**: Paid (subscription model)
- **Binary**: `cursor` (CLI)
- **Skills/Prompts**: `.cursorrules` files

## Implementation

### 1. Update internal/config/config.go
- Add `aider`, `continue`, `cursor` to default tools
- Set appropriate tiers and capabilities

### 2. Update internal/sync/sync.go
- Add Continue to DefaultTargets (for MCP sync)
- Add Aider config sync support
- Add Cursor sync support (via .cursorrules)

### 3. Update internal/learn/sync.go
- Add pattern sync for Aider (conventions file)
- Add pattern sync for Continue (custom instructions)
- Add pattern sync for Cursor (.cursorrules)

### 4. Update cmd/mur/cmd/health.go
- Add health checks for aider, continue, cursor

## Config Format Examples

### Aider (~/.aider.conf.yml)
```yaml
# Global aider configuration
model: claude-3-sonnet
# Custom instructions can be added via conventions
```

### Continue (~/.continue/config.json)
```json
{
  "models": [...],
  "mcpServers": {...},
  "customInstructions": "..."
}
```

### Cursor (.cursorrules)
```markdown
# Project-specific Cursor instructions
Plain text/markdown format
```

## Rollout
1. Implement config changes
2. Implement sync support
3. Implement health checks
4. Test with installed tools
5. Update documentation
