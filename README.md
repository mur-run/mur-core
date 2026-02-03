# Murmur 🔊

**Unified Multi-AI CLI Management + Cross-Tool Learning System**

Every AI CLI tool is an isolated island. Murmur unifies them.

## Features

- **Multi-tool runner** — One command to run any AI, no need to remember each tool's flags
- **Unified MCP management** — Configure once, sync everywhere
- **Cross-tool learning** — What Claude learns, Gemini knows too
- **Team knowledge base** — Share patterns across team, new members inherit experience automatically
- **Smart cost routing** — Simple tasks auto-route to free tools, complex ones go to Claude

## Installation

```bash
# Go install
go install github.com/karajanchang/murmur-ai/cmd/mur@latest

# Or download binary from releases
curl -L https://github.com/karajanchang/murmur-ai/releases/latest/download/mur-darwin-arm64.tar.gz | tar xz
sudo mv mur /usr/local/bin/
```

## Quick Start

```bash
# Initialize
mur init

# Check available AI tools
mur health

# Run a prompt (auto-routes to best tool)
mur run -p "explain this code"

# Run with specific tool
mur run -t gemini -p "write a haiku"

# See routing decision without running
mur run -p "refactor this" --explain

# List learned patterns
mur learn list

# Extract patterns from recent sessions
mur learn extract --auto

# Sync to all AI tools
mur sync all
```

## Commands

| Command | Description |
|---------|-------------|
| `mur init` | Initialize configuration |
| `mur run -p "prompt"` | Run prompt with smart routing |
| `mur run -t claude -p "prompt"` | Run with specific tool |
| `mur run -p "prompt" --explain` | Show routing decision |
| `mur config show` | Show configuration |
| `mur config default <tool>` | Set default tool |
| `mur config routing <mode>` | Set routing mode (auto/manual/cost-first/quality-first) |
| `mur health` | Check AI tool availability |
| `mur learn list` | List learned patterns |
| `mur learn add <name>` | Add a new pattern |
| `mur learn extract --auto` | Auto-extract patterns from sessions |
| `mur learn sync` | Sync patterns to AI tools |
| `mur sync all` | Sync everything (MCP, hooks, patterns) |
| `mur sync mcp` | Sync MCP configuration |
| `mur sync hooks` | Sync hooks configuration |

## Supported AI Tools

| Tool | Status | Tier |
|------|--------|------|
| Claude Code | ✅ Supported | Paid |
| Gemini CLI | ✅ Supported | Free |
| Auggie | 🔜 Coming | Free |
| Codex | 🔜 Coming | Paid |
| OpenCode | 🔜 Coming | Free |

## Configuration

Config file: `~/.murmur/config.yaml`

```yaml
default_tool: claude

routing:
  mode: auto  # auto | manual | cost-first | quality-first
  complexity_threshold: 0.5

tools:
  claude:
    enabled: true
    binary: claude
    tier: paid
    capabilities: [coding, analysis, complex]
  gemini:
    enabled: true
    binary: gemini
    tier: free
    capabilities: [coding, simple-qa]

learning:
  auto_extract: true
  sync_to_tools: true
```

## How Smart Routing Works

Murmur analyzes your prompt and routes to the best tool:

| Prompt Type | Routed To | Reason |
|-------------|-----------|--------|
| "what is git?" | Gemini (free) | Simple Q&A |
| "explain this error" | Gemini (free) | Basic explanation |
| "refactor this architecture" | Claude (paid) | Complex analysis |
| "debug this race condition" | Claude (paid) | Advanced debugging |

Override with `-t <tool>` or set `routing.mode: manual` to disable.

## License

MIT
