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
- **Web Dashboard** — Visualize stats, patterns, and tool health
- **Notifications** — Slack/Discord alerts when patterns are learned
- **Editor Plugins** — VS Code, Sublime, JetBrains, Neovim integrations

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
```

### Build from Source

```bash
git clone https://github.com/karajanchang/murmur-ai.git
cd murmur-ai
go build -o mur ./cmd/mur
sudo mv mur /usr/local/bin/
```

### Shell Completion

```bash
# Bash
mur completion bash > /etc/bash_completion.d/mur

# Zsh
mur completion zsh > "${fpath[1]}/_mur"

# Fish
mur completion fish > ~/.config/fish/completions/mur.fish
```

## Quick Start

```bash
# Initialize
mur init

# Check available AI tools
mur health

# Run a prompt (auto-routes to best tool)
mur run -p "explain this code"

# See routing decision without running
mur run -p "refactor this" --explain

# Sync settings to all AI tools
mur sync all

# Extract patterns from recent sessions
mur learn extract --auto

# Start web dashboard
mur serve
```

## Commands

### Core

| Command | Description |
|---------|-------------|
| `mur init` | Initialize configuration |
| `mur health` | Check AI tool availability |
| `mur run -p "prompt"` | Run prompt with smart routing |
| `mur run -t claude -p "prompt"` | Run with specific tool |
| `mur run --explain` | Show routing decision |

### Sync

| Command | Description |
|---------|-------------|
| `mur sync all` | Sync everything (MCP, hooks, patterns, skills) |
| `mur sync mcp` | Sync MCP configuration |
| `mur sync hooks` | Sync hooks configuration |
| `mur sync patterns` | Sync learned patterns |
| `mur sync skills` | Sync skills |

### Learn

| Command | Description |
|---------|-------------|
| `mur learn list` | List learned patterns |
| `mur learn add <name>` | Add a new pattern |
| `mur learn get <name>` | Show pattern details |
| `mur learn delete <name>` | Delete a pattern |
| `mur learn extract --auto` | Auto-extract patterns from sessions |
| `mur learn sync` | Sync patterns to AI tools |
| `mur learn init <repo>` | Initialize learning repo for cross-machine sync |
| `mur learn push` | Push patterns to your branch |
| `mur learn pull` | Pull shared patterns from main |
| `mur learn auto-merge` | Auto-create PRs for high-confidence patterns |

### Team (Git-based)

| Command | Description |
|---------|-------------|
| `mur team init <repo>` | Initialize team repo |
| `mur team pull` | Pull team changes |
| `mur team push` | Push changes to team |
| `mur team sync` | Bidirectional sync |
| `mur team status` | Show team repo status |

### Stats & Dashboard

| Command | Description |
|---------|-------------|
| `mur stats` | Show usage statistics |
| `mur stats --tool claude` | Stats for specific tool |
| `mur stats --period week` | Stats for time period |
| `mur stats --json` | JSON output |
| `mur serve` | Start web dashboard (localhost:8383) |
| `mur serve --port 9000` | Custom port |

### Config

| Command | Description |
|---------|-------------|
| `mur config show` | Show configuration |
| `mur config default <tool>` | Set default tool |
| `mur config routing <mode>` | Set routing mode |
| `mur config notifications` | Configure Slack/Discord |

### Notifications

| Command | Description |
|---------|-------------|
| `mur notify test` | Test notification webhooks |
| `mur notify test --slack` | Test Slack only |
| `mur notify test --discord` | Test Discord only |

### Skills

| Command | Description |
|---------|-------------|
| `mur skills list` | List available skills |
| `mur skills show <name>` | Show skill details |
| `mur skills import <path>` | Import skill file |
| `mur skills import --superpowers` | Import from Superpowers plugin |

## Supported AI Tools

| Tool | Status | Tier | Sync Support |
|------|--------|------|--------------|
| Claude Code | ✅ | Paid | MCP, Hooks, Patterns, Skills |
| Gemini CLI | ✅ | Free | MCP, Hooks, Patterns, Skills |
| Auggie | ✅ | Free | MCP, Patterns, Skills |
| Codex | ✅ | Paid | Patterns (instructions.md) |
| OpenCode | ✅ | Free | MCP, Patterns, Skills |
| Aider | ✅ | Free/Paid | Patterns (conventions) |
| Continue | ✅ | Free | MCP, Patterns, Skills |
| Cursor | ✅ | Paid | Patterns, Skills |

## Editor Integrations

| Editor | Location | Installation |
|--------|----------|--------------|
| VS Code | `integrations/vscode/` | Copy to extensions or F5 to debug |
| Sublime Text | `integrations/sublime/` | Copy to `Packages/MurmurAI/` |
| JetBrains | `integrations/jetbrains/` | `./gradlew buildPlugin` |
| Neovim | `integrations/neovim/` | Use lazy.nvim or packer |

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
  gemini:
    enabled: true
    binary: gemini
    tier: free
  # ... more tools

learning:
  repo: "git@github.com:yourorg/murmur-learnings.git"
  branch: "your-machine-name"
  auto_push: true
  pull_from_main: true

notifications:
  enabled: true
  slack:
    webhook_url: "https://hooks.slack.com/services/..."
  discord:
    webhook_url: "https://discord.com/api/webhooks/..."
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

## Cross-Machine Learning Sync

Share patterns across your machines:

```bash
# First machine
mur learn init git@github.com:you/murmur-learnings.git
mur learn extract --auto
mur learn push

# Second machine
mur learn init git@github.com:you/murmur-learnings.git
mur learn pull  # Get patterns from main branch
```

Each machine gets its own branch (auto-detected from hostname). Merge to `main` via PR to share with everyone.

## Documentation

Full documentation available at `docs/` or after running:

```bash
cd docs && mkdocs serve
```

## License

MIT
