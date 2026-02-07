# mur 🔊

[![CI](https://github.com/mur-run/mur-core/actions/workflows/ci.yml/badge.svg)](https://github.com/mur-run/mur-core/actions/workflows/ci.yml)
[![Release](https://github.com/mur-run/mur-core/actions/workflows/release.yml/badge.svg)](https://github.com/mur-run/mur-core/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/mur-run/mur-core)](https://goreportcard.com/report/github.com/mur-run/mur-core)

**Unified AI CLI with Continuous Learning**

Every AI CLI is an isolated island. **mur** unifies them with a learning engine that makes your patterns smarter over time.

## ✨ What's New in v0.4.0

- **Auto Pattern Injection** — Relevant patterns injected automatically based on context
- **Effectiveness Tracking** — Patterns learn from your feedback and usage
- **Semantic Search** — Find patterns by meaning, not just keywords
- **Lifecycle Management** — Auto-deprecate underperforming patterns
- **Cross-CLI Learning** — Learn from Claude, Gemini, Codex sessions together
- **Pattern Suggestions** — Extract patterns from your session histories

## 🚀 Quick Start

```bash
# Install
go install github.com/mur-run/mur-core/cmd/mur@latest

# Initialize
mur init

# Run with auto pattern injection
mur run -p "fix the login bug"
→ claude (auto: complexity 0.65) [3 patterns]

# Give feedback
mur feedback helpful swift-error-handling
```

## 📦 Installation

### Homebrew (macOS)

```bash
brew install mur-run/tap/mur
```

### Go Install

```bash
go install github.com/mur-run/mur-core/cmd/mur@latest
```

### Download Binary

```bash
# macOS (Apple Silicon)
curl -L https://github.com/mur-run/mur-core/releases/latest/download/mur_Darwin_arm64.tar.gz | tar xz
sudo mv mur /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/mur-run/mur-core/releases/latest/download/mur_Darwin_x86_64.tar.gz | tar xz
sudo mv mur /usr/local/bin/

# Linux
curl -L https://github.com/mur-run/mur-core/releases/latest/download/mur_Linux_x86_64.tar.gz | tar xz
sudo mv mur /usr/local/bin/
```

## 🎯 Core Features

### Smart Routing

```bash
mur run -p "what is git?"           # → Free tool (simple question)
mur run -p "refactor this module"   # → Paid tool (complex task)
mur run -p "fix bug" -t claude      # Force specific tool
mur run -p "test" --explain         # Show routing decision
```

### Pattern Auto-Injection

Patterns are automatically injected based on your project context:

```bash
cd ~/Projects/MySwiftApp
mur run -p "fix the auth error"

# mur detects: Swift project, error-handling context
# Auto-injects: swift-error-handling, auth-patterns
→ claude [2 patterns]
```

Disable with `--no-inject`, see details with `--verbose`.

### Effectiveness Tracking

```bash
# After using mur run, give feedback
mur feedback helpful swift-error-handling
mur feedback unhelpful debugging-tips -c "too generic"

# View pattern stats
mur pattern-stats
swift-error-handling        █████████░ 90%
  Uses: 12   Success: 92% | 👍 8 😐 2 👎 0
```

### Semantic Search

```bash
# Index patterns for semantic search
mur embed index

# Search by meaning
mur embed search "handle errors in network requests"
1. network-error-handling  ██████████ 95%
2. retry-logic            ████████░░ 78%
```

Supports local (Ollama) or cloud (OpenAI) embeddings.

### Lifecycle Management

```bash
# Check for low-performing patterns
mur lifecycle evaluate
⚠️ debugging-tips: active → deprecated
   Reason: low effectiveness: 25% (threshold: 30%)

# Apply changes
mur lifecycle apply
```

### Pattern Suggestion

```bash
# Extract patterns from session files
mur suggest scan ~/sessions/
Found 8 suggestions

1. swift-error-handling  ████████░░ 80%
   Handle Result types with proper error propagation

# Interactive review
mur suggest scan --interactive
```

### Cross-CLI Learning

```bash
# Learn from all your AI CLI histories
mur cross-learn status
Claude Code:    ✓ 42 session files
Gemini CLI:     ✓ 18 session files

mur cross-learn scan
📚 Claude Code: 8 suggestions
📚 Gemini CLI: 3 suggestions
```

### Sync to All CLIs

```bash
mur sync patterns
✓ Claude Code: synced 12 patterns
✓ Gemini CLI: synced 12 patterns
✓ Codex: synced 12 patterns
✓ Aider: synced 12 patterns
```

## 📋 Command Reference

### Run & Route

| Command | Description |
|---------|-------------|
| `mur run -p "prompt"` | Run with smart routing + pattern injection |
| `mur run -t claude -p "..."` | Force specific tool |
| `mur run --explain` | Show routing decision |
| `mur run --no-inject` | Disable pattern injection |
| `mur run -v` | Verbose (show injected patterns) |

### Patterns

| Command | Description |
|---------|-------------|
| `mur learn list` | List all patterns |
| `mur learn add <name>` | Add pattern interactively |
| `mur learn get <name>` | Show pattern details |
| `mur lint` | Validate all patterns |
| `mur migrate` | Migrate to Schema v2 |

### Feedback & Stats

| Command | Description |
|---------|-------------|
| `mur feedback helpful <pattern>` | Rate pattern as helpful |
| `mur feedback unhelpful <pattern>` | Rate as unhelpful |
| `mur pattern-stats` | View effectiveness stats |
| `mur pattern-stats --update` | Refresh effectiveness scores |

### Semantic Search

| Command | Description |
|---------|-------------|
| `mur embed index` | Index patterns for search |
| `mur embed search "query"` | Semantic search |
| `mur embed status` | Show index status |
| `mur embed rehash` | Rebuild index |

### Lifecycle

| Command | Description |
|---------|-------------|
| `mur lifecycle evaluate` | Check for deprecation |
| `mur lifecycle apply` | Apply recommendations |
| `mur lifecycle deprecate <name>` | Manually deprecate |
| `mur lifecycle reactivate <name>` | Reactivate pattern |
| `mur lifecycle cleanup` | Delete old archived |

### Suggestions

| Command | Description |
|---------|-------------|
| `mur suggest scan <dir>` | Scan for pattern suggestions |
| `mur suggest scan -i` | Interactive review |
| `mur suggest accept <name>` | Accept suggestion |
| `mur suggest list` | List pending suggestions |

### Cross-CLI Learning

| Command | Description |
|---------|-------------|
| `mur cross-learn status` | Show available CLI sources |
| `mur cross-learn scan` | Learn from all CLIs |
| `mur cross-learn scan -s claude` | Learn from specific CLI |

### Sync

| Command | Description |
|---------|-------------|
| `mur sync all` | Sync everything |
| `mur sync patterns` | Sync patterns to all CLIs |
| `mur sync mcp` | Sync MCP config |
| `mur sync hooks` | Sync hooks |

### Other

| Command | Description |
|---------|-------------|
| `mur init` | Initialize config |
| `mur health` | Check tool availability |
| `mur stats` | Usage statistics |
| `mur serve` | Start web dashboard |

## 🔧 Configuration

`~/.murmur/config.yaml`:

```yaml
default_tool: claude

routing:
  mode: auto  # auto | manual | cost-first
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

# Embedding provider for semantic search
embedding:
  provider: ollama  # ollama | openai
  model: nomic-embed-text

# Lifecycle thresholds
lifecycle:
  deprecate_threshold: 0.3
  archive_threshold: 0.1
  stale_after_days: 90
```

## 🤖 Supported AI CLIs

| CLI | Sync Support | Learning |
|-----|--------------|----------|
| Claude Code | MCP, Hooks, Patterns | ✅ Sessions |
| Gemini CLI | MCP, Hooks, Patterns | ✅ Sessions |
| Codex | instructions.md | ✅ Sessions |
| Aider | conventions/ | ✅ Sessions |
| Continue | MCP, Patterns | ✅ Sessions |
| Cursor | Patterns | - |
| OpenCode | MCP, Patterns | - |
| Auggie | MCP, Patterns | - |

## 🏗️ Architecture

```
~/.murmur/
├── config.yaml           # Main configuration
├── patterns/             # Pattern storage (Schema v2)
│   ├── swift-error.yaml
│   └── debugging.yaml
├── tracking/             # Usage & effectiveness data
│   └── usage.jsonl
├── embeddings/           # Semantic search cache
│   └── embeddings.json
└── suggestions/          # Pending suggestions
```

## 📖 Pattern Schema v2

```yaml
id: uuid
name: swift-error-handling
description: Handle Swift errors with Result types
content: |
  When handling errors in Swift...

tags:
  confirmed: [swift, error-handling]
  inferred:
    - tag: ios
      confidence: 0.92

applies:
  languages: [swift]
  frameworks: [swiftui]
  keywords: [error, Result, throw]

security:
  trust_level: owner
  reviewed: true

learning:
  effectiveness: 0.85
  usage_count: 42

lifecycle:
  status: active
```

## 🔗 Links

- **Website**: [mur.run](https://mur.run)
- **Documentation**: [docs/](./docs/)
- **Issues**: [GitHub Issues](https://github.com/mur-run/mur-core/issues)
- **Releases**: [GitHub Releases](https://github.com/mur-run/mur-core/releases)

## 📄 License

MIT
