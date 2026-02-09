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

## Status

🔜 **Coming Soon** - Full integration is in development.

Current support:

- ✅ MCP server sync
- ❌ Hook sync (not supported by Auggie)
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

## Configuration in Murmur

```yaml
# ~/.mur/config.yaml
tools:
  auggie:
    enabled: false  # Enable when ready
    binary: auggie
    tier: free
    capabilities: [coding]
    flags: []
```

## Sync Features

### MCP Servers ✅

```bash
mur sync mcp
```

Auggie supports MCP servers via `~/.augment/settings.json`.

### Hooks ❌

Auggie doesn't support hooks natively. When you run `mur sync hooks`, Auggie is skipped:

```
  ✗ Auggie: no event mapping defined for this target
```

### Patterns ✅

Patterns can be synced to Auggie's instruction context.

### Skills ✅

Skills sync to Auggie's configuration.

## When to Use Auggie

Auggie is good for:

- Free coding assistance
- Quick code generation
- Learning and experimentation

Currently, murmur prefers Gemini CLI for the free tier, but once Auggie integration is complete, you'll be able to:

```yaml
routing:
  prefer_free: [gemini, auggie]  # Try Gemini first, then Auggie
```

## Roadmap

1. ✅ Basic tool definition
2. ✅ MCP sync support
3. 🔜 Full routing integration
4. 🔜 Pattern extraction from sessions
5. 🔜 Advanced capabilities detection

## Direct Usage

```bash
auggie "write a hello world in Rust"
```

Through murmur (when enabled):

```bash
mur run -t auggie -p "write a hello world in Rust"
```

## See Also

- [Augment Code](https://www.augmentcode.com/)
- [Smart Routing](../concepts/routing.md)
