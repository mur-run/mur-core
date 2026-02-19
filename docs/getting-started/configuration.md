# Configuration

MUR Core's configuration lives at `~/.mur/config.yaml`.

## Full Configuration Reference

```yaml
# Default tool when routing is disabled
default_tool: claude

# Smart routing settings
routing:
  mode: auto  # auto | manual | cost-first | quality-first
  complexity_threshold: 0.5  # 0.0-1.0

# AI tool definitions
tools:
  claude:
    enabled: true
    binary: claude
    tier: paid
    capabilities: [coding, analysis, complex]
    flags: []  # Additional flags to pass
  gemini:
    enabled: true
    binary: gemini
    tier: free
    capabilities: [coding, simple-qa]
    flags: []
  auggie:
    enabled: false
    binary: auggie
    tier: free
    capabilities: [coding]
    flags: []

# MCP server configuration (synced to all tools)
mcp:
  servers:
    filesystem:
      command: npx
      args: ["-y", "@modelcontextprotocol/server-filesystem", "/home/user"]
    github:
      command: npx
      args: ["-y", "@modelcontextprotocol/server-github"]
      env:
        GITHUB_TOKEN: ${GITHUB_TOKEN}

# Hook configuration (synced to all tools)
hooks:
  UserPromptSubmit:
    - matcher: ""  # Match all
      hooks:
        - type: command
          command: echo "Starting prompt..."
  Stop:
    - matcher: ""
      hooks:
        - type: command
          command: echo "Session ended"

# Learning settings
learning:
  auto_extract: true      # Auto-extract patterns from sessions
  sync_to_tools: true     # Sync patterns to AI tool instructions
  repo_url: ""            # Git repo for pattern sync
  auto_push: false        # Auto-push patterns after extraction
```

## Routing Modes

### `auto` (default)

Analyzes prompt complexity and routes accordingly:

- **Complexity < threshold** → Free tool (Gemini)
- **Complexity ≥ threshold** → Paid tool (Claude)
- **Needs tool use** → Paid tool (Claude)

### `manual`

Always uses `default_tool`. No automatic routing.

### `cost-first`

Aggressively prefers free tools:

- Only uses paid tools when complexity > 0.8
- Or when tool use is required

### `quality-first`

Prefers paid tools except for trivial tasks:

- Only uses free tools when complexity < 0.2
- And no tool use is required

## Environment Variables

You can use environment variables in config:

```yaml
mcp:
  servers:
    github:
      env:
        GITHUB_TOKEN: ${GITHUB_TOKEN}
```

## View Current Configuration

```bash
mur config show
```

## Change Settings

```bash
# Set default tool
mur config default gemini

# Set routing mode
mur config routing cost-first
```

## Configuration Locations

| File | Purpose |
|------|---------|
| `~/.mur/config.yaml` | Main configuration |
| `~/.mur/patterns/` | Learned patterns |
| `~/.mur/stats.json` | Usage statistics |
