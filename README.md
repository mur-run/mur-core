> ## ⚠️ MUR has moved to v2 (Rust)
>
> This Go version (v1) is in maintenance mode. The active development continues at:
>
> **👉 [github.com/mur-run/mur](https://github.com/mur-run/mur)** — Rust rewrite with semantic search, pattern evolution, and workflow intelligence.
>
> To upgrade: `brew upgrade mur` then `mur migrate`

# MUR Core 🔮

[![Go Version](https://img.shields.io/github/go-mod/go-version/mur-run/mur-core)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/mur-run/mur-core)](https://github.com/mur-run/mur-core/releases)
[![License](https://img.shields.io/github/license/mur-run/mur-core)](./LICENSE)

**Continuous learning for AI assistants.**

MUR captures patterns from your coding sessions and injects them into your AI tools. Your AI assistant learns your conventions, remembers your fixes, and gets smarter over time — automatically.

![Demo](assets/demo.gif)

## 🤔 Why MUR?

**The Problem:** Every time you use an AI CLI, you start from scratch. It forgets your project conventions, coding patterns, and past discoveries.

**The Solution:** MUR remembers. Install once, then your AI tools automatically know your patterns and preferences.

```
Without MUR:
  You: "Use Swift Testing instead of XCTest"
  ... 3 days later ...
  You: "Use Swift Testing instead of XCTest" (again)

With MUR:
  AI already knows your testing preferences.
  Zero repetition. Continuous learning.
```

## 🚀 Quick Start

```bash
# Install (macOS)
brew tap mur-run/tap && brew install mur

# Setup hooks
mur init --hooks

# Done! Use your AI CLI normally
claude "fix this bug"
```

MUR works invisibly in the background.

<details>
<summary>Other platforms (Linux, Windows, Go install)</summary>

**Linux / macOS (Go):**
```bash
CGO_ENABLED=0 go install github.com/mur-run/mur-core/cmd/mur@latest
export PATH="$HOME/go/bin:$PATH"
mur init --hooks
```

**Windows (PowerShell):**
```powershell
$env:CGO_ENABLED=0; go install github.com/mur-run/mur-core/cmd/mur@latest
mur init --hooks
```

</details>

## 🎉 What's Next?

After installation:

```bash
# 1. Create your first pattern
mur new "prefer-typescript"

# 2. Use your AI CLI — MUR injects patterns automatically
claude "refactor this function"

# 3. Check status
mur status

# 4. Browse your patterns
mur serve   # Opens web dashboard at localhost:8080
```

## ✨ Features

| Feature | Description |
|---------|-------------|
| 🧠 **Continuous Learning** | Extract patterns from Claude Code, Gemini CLI sessions |
| 🔄 **Universal Sync** | Patterns sync to 10+ AI tools automatically |
| 🔌 **Zero Friction** | Install hooks once, then forget about it |
| 🔍 **Semantic Search** | Find patterns by meaning, not keywords |
| 📊 **Dashboard** | Web UI for pattern management |
| 🔒 **Local First** | All data on your machine, optional cloud sync |
| 🌍 **Community** | Share and discover patterns from developers worldwide |

## 📄 Pattern Format

Patterns are YAML files stored in:
- **macOS/Linux:** `~/.mur/patterns/`
- **Windows:** `%USERPROFILE%\.mur\patterns\`

```yaml
# ~/.mur/patterns/swift-testing.yaml
name: swift-testing-macro
description: Prefer Swift Testing over XCTest
content: |
  When writing tests in Swift:
  - Use @Test macro instead of func test...()
  - Use #expect() instead of XCTAssert
  - Use @Suite for test organization
tags:
  languages: [swift]
  topics: [testing]
applies:
  projects: [my-ios-app]  # Optional: limit to specific projects
```

**More examples:** `mur examples` installs sample patterns to get started.

## 🔄 How It Works

```
┌─────────────────────────────────────────┐
│ 1. You use AI CLI normally              │
│    $ claude "explain this code"         │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│ 2. MUR hooks inject relevant patterns   │
│    [context: swift-testing.yaml]        │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│ 3. AI responds with your preferences    │
│    "Using @Test as you prefer..."       │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│ 4. MUR learns from sessions (optional)  │
│    $ mur learn extract --auto           │
└─────────────────────────────────────────┘
```

**Token efficiency:** MUR uses directory-based sync, loading only relevant patterns. This reduces token usage by 90%+ compared to a single large context file.

## 📋 Core Commands

| Command | Description |
|---------|-------------|
| `mur init --hooks` | Setup with CLI hooks |
| `mur status` | Overview of patterns and sync status |
| `mur doctor` | Diagnose and fix issues |
| `mur new <name>` | Create new pattern |
| `mur edit <name>` | Edit pattern in $EDITOR |
| `mur search <query>` | Semantic search patterns |
| `mur sync` | Sync patterns to all AI tools |
| `mur serve` | Start web dashboard |
| `mur learn extract` | Extract patterns from sessions |

<details>
<summary>📖 All Commands</summary>

```
mur
├── init           # Setup wizard
├── status         # Quick overview
├── doctor         # Diagnose issues
├── sync           # Sync patterns to AI tools
├── new            # Create pattern
├── edit           # Edit pattern
├── search         # Semantic search
├── copy           # Copy pattern to clipboard
├── serve          # Web dashboard
├── learn extract  # Extract from sessions
├── community      # Browse community patterns
├── cloud          # Cloud sync (Pro/Team)
├── login/logout   # Authentication
└── update         # Update MUR
```

See [docs/commands.md](docs/commands.md) for complete reference.

</details>

## 🔄 Supported Tools

**With hooks (real-time injection):**
- Claude Code
- Gemini CLI  
- Auggie (Augment CLI)

**Static sync:**
- Codex, Aider
- Cursor, Windsurf, Continue

## 🔒 Privacy & Security

- **100% Local** — All patterns stored on your machine
  - macOS/Linux: `~/.mur/`
  - Windows: `%USERPROFILE%\.mur\`
- **No telemetry** — We don't collect usage data
- **Optional cloud** — Only if you explicitly enable it
- **Git-based sync** — Use your own repo for backup

## ☁️ Cloud Sync (Optional)

Sync across devices with [mur.run](https://mur.run):

```bash
mur login           # OAuth login
mur sync --cloud    # Sync with cloud
```

| Plan | Price | Features |
|------|-------|----------|
| Free | $0 | Local patterns, git sync |
| Pro | $9/mo | Cloud sync, 3 devices |
| Team | $49/mo | 5 members, shared patterns |

## 🔍 Semantic Search

Find patterns by meaning, not keywords:

```bash
# Option 1: Cloud (recommended, ~$0.001 for 200 patterns)
export OPENAI_API_KEY=sk-...
mur index rebuild

# Option 2: Local (free, needs Ollama)
ollama pull qwen3-embedding
mur index rebuild

# Search naturally
mur search "how to sign a macOS app"
# → bitl-binary-signing-workaround (0.71)
```

See [docs/semantic-search.md](docs/semantic-search.md) for all providers (OpenAI, Google, Voyage, Ollama) and advanced features like document expansion.

## ⚙️ Configuration

```yaml
# ~/.mur/config.yaml
tools:
  claude:
    enabled: true
  gemini:
    enabled: true

# Semantic search (cloud or local)
search:
  provider: openai                # openai | ollama | google | voyage
  model: text-embedding-3-small
  api_key_env: OPENAI_API_KEY

# Pattern extraction LLM
learning:
  llm:
    provider: ollama              # ollama | openai | gemini | claude
    model: llama3.2:3b
```

See [docs/configuration.md](docs/configuration.md) for all options.

## ⚠️ Limitations

- **Hooks require supported CLIs** — Claude Code, Gemini CLI, Auggie
- **Large libraries** — 500+ patterns may increase context usage
- **Semantic search** — Requires Ollama or OpenAI API key
- **Windows** — Some features may have limited testing

## 🗑️ Uninstall

```bash
# Remove binary
brew uninstall mur                    # If installed via Homebrew
rm $(which mur)                       # If installed via Go

# Remove data and hooks
rm -rf ~/.mur                         # macOS/Linux
rm -rf $env:USERPROFILE\.mur          # Windows (PowerShell)

# Remove Claude Code hooks (if installed)
# Edit ~/.claude/settings.json and remove the "hooks" section
```

## 💻 System Requirements

- **Platforms:** macOS, Linux, Windows
- **Go 1.21+** (only for source installation)
- **Optional:** Ollama (for semantic search & LLM extraction)

## 🤝 Contributing

Issues and PRs welcome!

```bash
git clone https://github.com/mur-run/mur-core.git
cd mur-core
make check  # lint + test
```

## 📚 Documentation

- [Configuration Guide](docs/configuration.md)
- [Semantic Search](docs/semantic-search.md)
- [Cloud Sync](docs/cloud-sync.md)
- [All Commands](docs/commands.md)
- [Troubleshooting](docs/troubleshooting.md)

## 📄 License

MIT — use it however you want.
