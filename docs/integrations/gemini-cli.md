# Gemini CLI Integration

Gemini CLI is Google's free AI coding assistant for the command line.

## Overview

| Property | Value |
|----------|-------|
| **Tool Name** | `gemini` |
| **Binary** | `gemini` |
| **Tier** | Free |
| **Capabilities** | coding, simple-qa |
| **Config Path** | `~/.gemini/settings.json` |

## Installation

```bash
npm install -g @anthropic-ai/gemini-cli
# or
brew install gemini-cli
```

Verify installation:

```bash
gemini --version
```

## Configuration in Murmur

```yaml
# ~/.mur/config.yaml
tools:
  gemini:
    enabled: true
    binary: gemini
    tier: free
    capabilities: [coding, simple-qa]
    flags: []
```

## When Murmur Routes to Gemini

Gemini is selected for:

- **Simple questions** (complexity < 0.5 in auto mode)
- **Quick explanations** - "what is X?", "explain Y"
- **Basic coding** - simple functions, snippets
- **Cost savings** - when you want free usage

## Sync Features

### MCP Servers ✅

```bash
mur sync mcp
```

Syncs to `~/.gemini/settings.json`.

### Hooks ✅

```bash
mur sync hooks
```

Gemini CLI hook events:

| Murmur Event | Gemini Event |
|--------------|--------------|
| `UserPromptSubmit` | `BeforeAgent` |
| `Stop` | `AfterAgent` |
| `BeforeTool` | `BeforeTool` |
| `AfterTool` | `AfterTool` |

### Patterns ✅

```bash
mur learn sync
```

Patterns sync to Gemini's instruction context.

### Skills ✅

Skills sync to Gemini's configuration.

## Advantages

### Free Usage

No API costs for most tasks. Perfect for:

- Learning and experimentation
- Simple Q&A
- Quick code snippets
- Documentation lookup

### Fast Responses

Gemini tends to be faster for simple queries.

### Good for Simple Tasks

Excellent for straightforward coding tasks that don't need Claude's advanced reasoning.

## Limitations

- Less capable for complex architecture tasks
- May not handle nuanced debugging as well
- Tool use capabilities are more limited

## Direct Usage

```bash
gemini "write a function to reverse a string"
```

Or through murmur:

```bash
mur run -t gemini -p "write a function to reverse a string"
```

## Cost Savings

With murmur's smart routing:

```bash
mur stats
```

```
Cost Analysis:
  Saved by routing: $8.20 (66 prompts → free tool)
```

Every simple question routed to Gemini instead of Claude saves money.

## Troubleshooting

### Gemini Not Found

```bash
which gemini
# If not found:
npm install -g @google/gemini-cli
```

### Rate Limiting

Gemini CLI has generous free limits, but if you hit them:

1. Wait for reset (usually hourly)
2. Use `mur run -t claude` for urgent tasks

### Config Sync Issues

Ensure `~/.gemini/` directory exists:

```bash
mkdir -p ~/.gemini
mur sync all
```

## See Also

- [Smart Routing](../concepts/routing.md) - How routing decides
- [Cost Savings](../commands/stats.md) - Track savings
- [Gemini CLI Docs](https://github.com/google-gemini/gemini-cli)
