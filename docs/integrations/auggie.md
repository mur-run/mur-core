# Auggie Integration

Auggie is a free AI coding assistant from Augment Code.

## Overview

| Property | Value |
|----------|-------|
| **Tool Name** | `auggie` |
| **Binary** | `auggie` |
| **Tier** | Free |
| **Capabilities** | coding |
| **Config Path** | `~/.augment/settings.json` |
| **Hooks Support** | ✅ Yes (v0.15.0+) |

## Status

✅ **Fully Supported** - Hooks, patterns, and MCP sync all work.

Current support:

- ✅ Hook sync (SessionStart, Stop events)
- ✅ MCP server sync
- ✅ Pattern sync
- ✅ Skill sync

## Installation

```bash
npm install -g @augment-code/auggie
```

Verify:

```bash
auggie --version
```

## Quick Setup

```bash
# Install mur hooks for Auggie
mur init --hooks
```

This configures `~/.augment/settings.json` with:

- **SessionStart** hook — Injects context-aware patterns at session start
- **Stop** hook — Extracts patterns and syncs when agent finishes

## Configuration in Murmur

```yaml
# ~/.mur/config.yaml
tools:
  auggie:
    enabled: true
    binary: auggie
    tier: free
    capabilities: [coding]
    flags: []
```

## Sync Features

### Hooks ✅

```bash
mur init --hooks
```

Auggie supports hooks via `~/.augment/settings.json`. Supported events:

| Event | Description |
|-------|-------------|
| `SessionStart` | Runs when Auggie starts a new session |
| `SessionEnd` | Runs when a session ends |
| `PreToolUse` | Runs before a tool executes |
| `PostToolUse` | Runs after a tool completes |
| `Stop` | Runs when agent finishes responding |

Mur uses `SessionStart` for pattern injection and `Stop` for pattern extraction.

### MCP Servers ✅

```bash
mur sync mcp
```

Auggie supports MCP servers via `~/.augment/settings.json`.

### Patterns ✅

Patterns sync to Auggie's instruction context.

### Skills ✅

Skills sync to Auggie's configuration.

## Hooks Configuration

After running `mur init --hooks`, your `~/.augment/settings.json` will include:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "bash /Users/YOU/.mur/hooks/on-prompt.sh"
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "bash /Users/YOU/.mur/hooks/on-stop.sh"
          }
        ]
      }
    ]
  }
}
```

## Verify Hooks

To verify hooks are working:

```bash
# Add debug logging
echo 'echo "[mur] hook @ $(date)" >> /tmp/mur-hook.log' >> ~/.mur/hooks/on-prompt.sh

# Run auggie
auggie "test"

# Check log
cat /tmp/mur-hook.log
```

## When to Use Auggie

Auggie is good for:

- Free coding assistance (uses GPT-5.2)
- Quick code generation
- Project-aware assistance (auto-indexes your codebase)

With murmur routing:

```yaml
routing:
  prefer_free: [gemini, auggie]  # Try Gemini first, then Auggie
```

## Direct Usage

```bash
auggie "write a hello world in Rust"
```

Through murmur:

```bash
mur run -t auggie -p "write a hello world in Rust"
```

## See Also

- [Augment Code](https://www.augmentcode.com/)
- [Auggie CLI Docs](https://docs.augmentcode.com/cli/)
- [Auggie Hooks Docs](https://docs.augmentcode.com/cli/hooks)
- [Smart Routing](../concepts/routing.md)
