# Claude Code Integration

Claude Code is Anthropic's official AI coding assistant CLI.

## Overview

| Property | Value |
|----------|-------|
| **Tool Name** | `claude` |
| **Binary** | `claude` |
| **Tier** | Paid |
| **Capabilities** | coding, analysis, complex, tool-use |
| **Config Path** | `~/.claude/settings.json` |

## Installation

```bash
npm install -g @anthropic-ai/claude-code
```

Verify installation:

```bash
claude --version
```

## Configuration in Murmur

```yaml
# ~/.mur/config.yaml
tools:
  claude:
    enabled: true
    binary: claude
    tier: paid
    capabilities: [coding, analysis, complex, tool-use]
    flags: []  # Additional flags to pass
```

## When Murmur Routes to Claude

Claude is selected for:

- **Complex tasks** (complexity ≥ 0.5 in auto mode)
- **Architecture work** - refactoring, redesign, optimization
- **Debugging** - complex bug fixing, race conditions
- **Tool use** - file operations, code execution
- **Analysis** - code review, security analysis

## Sync Features

### MCP Servers ✅

```bash
mur sync mcp
```

Syncs MCP server configuration to `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user"]
    }
  }
}
```

### Hooks ✅

```bash
mur sync hooks
```

Claude Code hook events:

| Murmur Event | Claude Event |
|--------------|--------------|
| `UserPromptSubmit` | `UserPromptSubmit` |
| `Stop` | `Stop` |
| `BeforeTool` | `PreToolUse` |
| `AfterTool` | `PostToolUse` |

### Patterns ✅

```bash
mur learn sync
```

Patterns are added to Claude's instruction context.

### Skills ✅

Skills sync to Claude's custom instructions.

## Pattern Extraction

Extract patterns from Claude Code sessions:

```bash
mur learn extract --auto
```

Murmur reads session transcripts from `~/.claude/projects/` and identifies:

- Corrections you made
- Coding standards you enforced
- Explanations you provided

## Direct Usage

You can still use Claude directly:

```bash
claude "explain this code"
```

Or through murmur:

```bash
mur run -t claude -p "explain this code"
```

## Cost Considerations

Claude Code uses Anthropic API credits. Murmur helps save costs by:

1. Routing simple questions to free tools (Gemini)
2. Tracking usage in `mur stats`
3. Showing estimated spend

## Troubleshooting

### Claude Not Found

```bash
which claude
# If empty, add npm global bin to PATH:
export PATH="$PATH:$(npm root -g)/../bin"
```

### Authentication Issues

Claude Code requires API key or OAuth:

```bash
claude login
```

### Sessions Not Found

Pattern extraction needs access to `~/.claude/projects/`. Ensure this directory exists and has session data.

## See Also

- [Smart Routing](../concepts/routing.md) - When Claude is selected
- [Pattern Extraction](../commands/learn.md#pattern-extraction) - Extract from sessions
- [Anthropic Documentation](https://docs.anthropic.com/claude-code)
