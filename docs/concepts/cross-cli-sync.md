# Cross-CLI Sync

Configure once, apply everywhere. Murmur unifies configuration across all your AI CLI tools.

## The Problem

Every AI CLI tool has its own configuration:

```
~/.claude/settings.json      ← Claude Code config
~/.gemini/settings.json      ← Gemini CLI config
~/.augment/settings.json     ← Auggie config
```

You want the same MCP servers, hooks, and patterns in all of them. Without murmur, you're manually copying configuration between files.

## The Solution

Define everything once in `~/.mur/config.yaml`:

```yaml
mcp:
  servers:
    filesystem:
      command: npx
      args: ["-y", "@modelcontextprotocol/server-filesystem", "/home/user"]
    github:
      command: npx
      args: ["-y", "@modelcontextprotocol/server-github"]

hooks:
  UserPromptSubmit:
    - matcher: ""
      hooks:
        - type: command
          command: echo "Starting..."
```

Then sync:

```bash
mur sync all
```

All tools now have identical configuration.

## What Syncs

### MCP Servers

Model Context Protocol servers provide AI tools with access to external data:

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
    database:
      command: npx
      args: ["-y", "@modelcontextprotocol/server-sqlite", "~/data/app.db"]
```

After `mur sync mcp`, all tools can access these servers.

### Hooks

Hooks run commands at specific points in the AI session:

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
  BeforeTool:
    - matcher: "file"
      hooks:
        - type: command
          command: git status
```

#### Event Translation

Different tools use different event names. Murmur translates:

| Murmur Event | Claude Code | Gemini CLI | Auggie |
|--------------|-------------|------------|--------|
| `UserPromptSubmit` | `UserPromptSubmit` | `BeforeAgent` | — |
| `Stop` | `Stop` | `AfterAgent` | — |
| `BeforeTool` | `PreToolUse` | `BeforeTool` | — |
| `AfterTool` | `PostToolUse` | `AfterTool` | — |

### Patterns

Learned patterns are injected into tool instructions:

```bash
mur learn sync
```

This adds your patterns to each tool's system prompt or instruction file.

### Skills

Skill definitions (custom capabilities) sync too:

```yaml
skills:
  code-review:
    description: "Review code for issues"
    instructions: |
      When reviewing code:
      1. Check for security issues
      2. Look for performance problems
      3. Suggest improvements
```

## Sync Commands

| Command | What it Syncs |
|---------|---------------|
| `mur sync all` | Everything |
| `mur sync mcp` | MCP servers only |
| `mur sync hooks` | Hooks only |
| `mur sync skills` | Skills only |

## How It Works

```
~/.mur/config.yaml
         │
         ▼
   ┌─────────────┐
   │  mur sync   │
   │    all      │
   └─────────────┘
         │
         ├──────────────────────────────────┐
         │                                  │
         ▼                                  ▼
┌─────────────────┐                ┌─────────────────┐
│ Read murmur     │                │ For each target │
│ config          │                │ CLI tool...     │
└─────────────────┘                └─────────────────┘
                                           │
         ┌─────────────────────────────────┼───────────────────────────────────┐
         │                                 │                                   │
         ▼                                 ▼                                   ▼
┌─────────────────┐               ┌─────────────────┐               ┌─────────────────┐
│ ~/.claude/      │               │ ~/.gemini/      │               │ ~/.augment/     │
│ settings.json   │               │ settings.json   │               │ settings.json   │
└─────────────────┘               └─────────────────┘               └─────────────────┘
```

## Supported Tools

| Tool | MCP | Hooks | Skills | Patterns |
|------|-----|-------|--------|----------|
| Claude Code | ✅ | ✅ | ✅ | ✅ |
| Gemini CLI | ✅ | ✅ | ✅ | ✅ |
| Auggie | ✅ | ❌ | ✅ | ✅ |

## Automatic Sync

Enable auto-sync after pattern extraction:

```yaml
learning:
  sync_to_tools: true
```

Now when you run `mur learn extract --auto`, patterns automatically sync.

## Troubleshooting

### Config Not Syncing

Check if the target config file exists:

```bash
ls -la ~/.claude/settings.json
ls -la ~/.gemini/settings.json
```

Murmur creates the file if it doesn't exist, but the directory must be writable.

### Hooks Not Working

Some tools (like Auggie) don't support hooks. Check the compatibility table above.

### MCP Servers Not Available

Ensure the MCP server packages are installed:

```bash
npx -y @modelcontextprotocol/server-filesystem --help
```

## See Also

- [sync Command](../commands/sync.md) - Sync command reference
- [Configuration](../getting-started/configuration.md) - Full config reference
- [Patterns](patterns.md) - How patterns work
