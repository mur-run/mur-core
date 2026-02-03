# Murmur 🔊

[![CI](https://github.com/karajanchang/murmur-ai/actions/workflows/ci.yml/badge.svg)](https://github.com/karajanchang/murmur-ai/actions/workflows/ci.yml)
[![Release](https://github.com/karajanchang/murmur-ai/actions/workflows/release.yml/badge.svg)](https://github.com/karajanchang/murmur-ai/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/karajanchang/murmur-ai)](https://goreportcard.com/report/github.com/karajanchang/murmur-ai)

**Unified Multi-AI CLI Management + Cross-Tool Learning System**

Every AI CLI tool is an isolated island. Murmur unifies them.

## Features

- **Multi-tool runner** — One command to run any AI, no need to remember each tool's flags
- **Unified MCP management** — Configure once, sync everywhere
- **Cross-tool learning** — What Claude learns, Gemini knows too
- **Team knowledge base** — Share patterns across team, new members inherit experience automatically
- **Smart cost routing** — Simple tasks auto-route to free tools, complex ones go to Claude

## Installation

### Go Install (Recommended)

```bash
go install github.com/karajanchang/murmur-ai/cmd/mur@latest
```

### Download from Releases

```bash
# macOS (Apple Silicon)
curl -L https://github.com/karajanchang/murmur-ai/releases/latest/download/mur-darwin-arm64.tar.gz | tar xz
sudo mv mur /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/karajanchang/murmur-ai/releases/latest/download/mur-darwin-amd64.tar.gz | tar xz
sudo mv mur /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/karajanchang/murmur-ai/releases/latest/download/mur-linux-amd64.tar.gz | tar xz
sudo mv mur /usr/local/bin/

# Linux (arm64)
curl -L https://github.com/karajanchang/murmur-ai/releases/latest/download/mur-linux-arm64.tar.gz | tar xz
sudo mv mur /usr/local/bin/

# Windows (amd64) - download and extract mur-windows-amd64.zip from releases
```

### Build from Source

```bash
git clone https://github.com/karajanchang/murmur-ai.git
cd murmur-ai
go build -o mur ./cmd/mur
sudo mv mur /usr/local/bin/
```

### Shell Completion

Enable tab completion for mur commands:

```bash
# Bash
mur completion bash > /etc/bash_completion.d/mur

# Bash (macOS with Homebrew)
mur completion bash > $(brew --prefix)/etc/bash_completion.d/mur

# Zsh (add to fpath)
mur completion zsh > "${fpath[1]}/_mur"

# Fish
mur completion fish > ~/.config/fish/completions/mur.fish

# PowerShell
mur completion powershell | Out-String | Invoke-Expression
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
