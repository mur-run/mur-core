# mur 🔊

[![Go Version](https://img.shields.io/github/go-mod/go-version/mur-run/mur-core)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/mur-run/mur-core)](https://github.com/mur-run/mur-core/releases)
[![License](https://img.shields.io/github/license/mur-run/mur-core)](./LICENSE)

**Your AI assistant's memory.**

mur captures patterns from your coding sessions and injects them back into your AI tools. Learn once, remember forever. Works invisibly — just use your CLI as normal.

## ✨ Features

- **🧠 Continuous Learning** — Extract patterns from Claude Code, Gemini CLI sessions
- **🔄 Universal Sync** — Patterns sync to 8+ AI tools (Claude, Gemini, Codex, Cursor, etc.)
- **🔌 Zero Friction** — Install hooks once, then forget about it
- **📊 Dashboard** — Web UI for pattern management and analytics
- **🔒 Local First** — All data stays on your machine (optional git sync)

## 🚀 Quick Start

```bash
# Install
CGO_ENABLED=0 go install github.com/mur-run/mur-core/cmd/mur@latest

# Add to PATH (if needed)
export PATH="$HOME/go/bin:$PATH"

# Setup
mur init --hooks

# Done! Use your AI CLI normally — mur works invisibly
claude "fix this bug"
```

## 📦 Installation

### From Source (Recommended)

```bash
CGO_ENABLED=0 go install github.com/mur-run/mur-core/cmd/mur@latest

# Verify
mur version
```

### From Git

```bash
git clone https://github.com/mur-run/mur-core.git
cd mur-core
make install
```

### PATH Setup

```bash
# Add to your shell config (~/.zshrc, ~/.bashrc, etc.)
export PATH="$HOME/go/bin:$PATH"
```

<details>
<summary>📋 Troubleshooting</summary>

**"command not found: mur"**
```bash
export PATH="$HOME/go/bin:$PATH"
```

**"LC_UUID" error on macOS**  
Use `CGO_ENABLED=0` when installing (already included above).

**Check installation**
```bash
mur doctor
```

</details>

## 🎯 How It Works

```
┌──────────────────────────────────────────────┐
│  You use AI CLIs normally                     │
│                                               │
│  $ claude "explain this code"                 │
│  $ gemini "fix the bug"                       │
└──────────────────────────────────────────────┘
                    │
                    ▼
┌──────────────────────────────────────────────┐
│  mur hooks inject relevant patterns           │
│                                               │
│  [context: your-project-patterns.md]          │
│  [context: learned-from-last-week.md]         │
└──────────────────────────────────────────────┘
                    │
                    ▼
┌──────────────────────────────────────────────┐
│  AI responds with project context             │
│                                               │
│  "Based on your navigation pattern, use..."   │
└──────────────────────────────────────────────┘
                    │
                    ▼
┌──────────────────────────────────────────────┐
│  mur learns from the session (optional)       │
│                                               │
│  $ mur learn extract --auto                   │
└──────────────────────────────────────────────┘
```

## 📋 Commands

### Core

| Command | Description |
|---------|-------------|
| `mur init` | Interactive setup wizard |
| `mur init --hooks` | Quick setup with CLI hooks |
| `mur status` | Overview of patterns, sync status |
| `mur doctor` | Diagnose and fix issues |
| `mur sync` | Sync patterns to all AI tools |

### Patterns

| Command | Description |
|---------|-------------|
| `mur new <name>` | Create new pattern |
| `mur edit <name>` | Edit pattern in $EDITOR |
| `mur search <query>` | Search patterns |
| `mur copy <name>` | Copy pattern to clipboard |
| `mur examples` | Install example patterns |

### Learning

| Command | Description |
|---------|-------------|
| `mur transcripts` | Browse Claude Code sessions |
| `mur learn extract` | Extract patterns from sessions |
| `mur learn extract --auto` | Auto-extract high-confidence patterns |
| `mur import <file>` | Import from file or URL |

### Dashboard

| Command | Description |
|---------|-------------|
| `mur serve` | Start web dashboard |
| `mur dashboard` | Generate static HTML report |
| `mur stats` | View usage statistics |

<details>
<summary>📖 All Commands</summary>

```
mur
├── init           # Setup wizard
├── status         # Quick overview
├── doctor         # Diagnose issues
├── sync           # Sync to AI tools
├── new            # Create pattern
├── edit           # Edit pattern
├── search         # Search patterns
├── copy           # Copy to clipboard
├── examples       # Install examples
├── import         # Import patterns
├── export         # Export patterns
├── config         # View/edit config
├── transcripts    # Browse sessions
├── serve          # Web dashboard
├── dashboard      # Static report
├── stats          # Usage stats
├── clean          # Cleanup old files
├── version        # Show version
├── web            # Open docs/GitHub
└── learn
    ├── list       # List patterns
    ├── get        # Show pattern
    ├── add        # Add pattern
    ├── delete     # Delete pattern
    ├── sync       # Sync to CLIs
    └── extract    # Extract from sessions
```

</details>

## 🔄 Supported Tools

**AI CLIs** (with hooks for real-time injection):
- Claude Code
- Gemini CLI

**AI CLIs** (static sync):
- Codex
- Auggie
- Aider

**IDEs** (static sync):
- Continue
- Cursor
- Windsurf

## 🔧 Configuration

```yaml
# ~/.mur/config.yaml

default_tool: claude

tools:
  claude:
    enabled: true
  gemini:
    enabled: true

learning:
  repo: git@github.com:you/patterns.git  # Optional: sync across machines
  auto_push: true
  llm:
    provider: ollama           # ollama | openai | gemini | claude
    model: deepseek-r1:8b      # LLM model for extraction
    ollama_url: http://localhost:11434
    openai_url: https://api.openai.com/v1  # or Groq, Together, etc.
```

Set your default LLM, then just run:
```bash
mur learn extract --llm          # Uses config default
mur learn extract --llm openai   # Override with OpenAI
mur learn extract --llm gemini   # Override with Gemini
```

### Remote Ollama (LAN Setup)

Run Ollama on a powerful server and access it from other machines:

**On the server (e.g., Mac mini):**
```bash
# Make Ollama listen on all interfaces
launchctl setenv OLLAMA_HOST "0.0.0.0"
brew services restart ollama
```

**On client machines:**
```yaml
# ~/.mur/config.yaml
learning:
  llm:
    provider: ollama
    model: deepseek-r1:8b
    ollama_url: http://192.168.1.100:11434  # Server IP
```

This way, laptops can use LLM extraction without running models locally.

### Recommended Models

| Provider | Model | Notes |
|----------|-------|-------|
| Ollama | `deepseek-r1:8b` | Best for extraction, 5GB |
| Ollama | `qwen2.5:14b` | Good for code, 9GB |
| OpenAI | `gpt-4o-mini` | Cheap & fast |
| Gemini | `gemini-2.0-flash` | Free tier available |
| Claude | `claude-sonnet-4-20250514` | Best quality |

## 📊 Dashboard

```bash
# Interactive web dashboard
mur serve
# → http://localhost:8080

# Static HTML report
mur dashboard -o report.html
```

Features:
- Pattern browser with search/filter
- Usage statistics and charts
- Sync status across tools
- One-click pattern editing

## 📁 Directory Structure

```
~/.mur/
├── config.yaml      # Configuration
├── patterns/        # Your patterns (YAML)
├── hooks/           # CLI hook scripts
├── stats.jsonl      # Usage statistics
└── repo/            # Git sync repo (optional)
```

## 🤝 Contributing

Issues and PRs welcome!

```bash
git clone https://github.com/mur-run/mur-core.git
cd mur-core
make check  # lint + test
```

## 📄 License

MIT — use it however you want.
