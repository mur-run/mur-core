# MUR sync

Synchronize configuration across all AI CLI tools.

## Usage

```bash
mur sync <target> [flags]
```

## Targets

| Target | Description |
|--------|-------------|
| `all` | Sync everything (MCP, hooks, skills) |
| `mcp` | Sync MCP server configuration |
| `hooks` | Sync hook configuration |
| `skills` | Sync skill definitions |

## How It Works

MUR Core reads your unified configuration from `~/.mur/config.yaml` and writes it to each AI tool's config file:

```
~/.mur/config.yaml
        ↓
   mur sync all
        ↓
┌───────────────────────────────────┐
│ ~/.claude/settings.json          │
│ ~/.gemini/settings.json          │
│ ~/.augment/settings.json         │
└───────────────────────────────────┘
```

## Examples

### Sync Everything

```bash
mur sync all
```

Output:

```
Syncing configuration to AI tools...

MCP Servers:
  ✓ Claude Code: synced 3 MCP servers
  ✓ Gemini CLI: synced 3 MCP servers
  ✓ Auggie: synced 3 MCP servers

Hooks:
  ✓ Claude Code: synced 4 hooks
  ✓ Gemini CLI: synced 4 hooks
  ✗ Auggie: no event mapping defined for this target

Skills:
  ✓ Claude Code: synced 2 skills
  ✓ Gemini CLI: synced 2 skills
```

### Sync Only MCP

```bash
mur sync mcp
```

### Sync Only Hooks

```bash
mur sync hooks
```

## MCP Configuration

Define MCP servers once in `~/.mur/config.yaml`:

```yaml
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
    sqlite:
      command: npx
      args: ["-y", "@modelcontextprotocol/server-sqlite", "~/data/app.db"]
```

Then sync to all tools:

```bash
mur sync mcp
```

## Hook Configuration

Hooks let you run commands at various points in the AI session:

```yaml
hooks:
  UserPromptSubmit:
    - matcher: ""
      hooks:
        - type: command
          command: echo "Prompt submitted"
  Stop:
    - matcher: ""
      hooks:
        - type: command
          command: mur learn extract --auto
```

### Event Mapping

Different AI tools use different event names. MUR Core translates:

| MUR Core Event | Claude Code | Gemini CLI |
|--------------|-------------|------------|
| UserPromptSubmit | UserPromptSubmit | BeforeAgent |
| Stop | Stop | AfterAgent |
| BeforeTool | PreToolUse | BeforeTool |
| AfterTool | PostToolUse | AfterTool |

## Supported Tools

| Tool | MCP | Hooks | Skills |
|------|-----|-------|--------|
| Claude Code | ✅ | ✅ | ✅ |
| Gemini CLI | ✅ | ✅ | ✅ |
| Auggie | ✅ | ❌ | ✅ |

## See Also

- [Cross-CLI Sync](../concepts/cross-cli-sync.md) - Detailed sync explanation
- [Configuration](../getting-started/configuration.md) - Full config reference
