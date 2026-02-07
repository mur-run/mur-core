# mur 🔊

**Continuous learning for AI assistants.**

mur makes your AI CLIs smarter over time. Learn once, remember forever.

## 🚀 Quick Start

```bash
# Install
CGO_ENABLED=0 go install github.com/mur-run/mur-core/cmd/mur@latest
export PATH="$HOME/go/bin:$PATH"

# Setup (interactive)
mur init

# Check status
mur status

# That's it! Use your AI CLI as normal
claude -p "fix this bug"
gemini -p "explain this code"
```

mur works invisibly in the background. Your patterns are synced to all CLIs.

## 📦 Installation

```bash
# Install (recommended)
CGO_ENABLED=0 go install github.com/mur-run/mur-core/cmd/mur@latest

# Make sure ~/go/bin is in your PATH
export PATH="$HOME/go/bin:$PATH"

# Verify installation
mur version
```

<details>
<summary>Troubleshooting</summary>

**"command not found: mur"**
```bash
# Add go/bin to PATH
export PATH="$HOME/go/bin:$PATH"
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
```

**"LC_UUID" error on macOS**
```bash
# Use CGO_ENABLED=0
CGO_ENABLED=0 go install github.com/mur-run/mur-core/cmd/mur@latest
```

**Build from source**
```bash
git clone https://github.com/mur-run/mur-core.git
cd mur-core
go build -o ~/go/bin/mur ./cmd/mur
```

</details>

## 📋 Commands

### Setup & Status

| Command | Description |
|---------|-------------|
| `mur init` | Interactive setup wizard |
| `mur init --hooks` | Setup with CLI hooks for auto-learning |
| `mur status` | Quick overview of patterns, sync, stats |
| `mur doctor` | Diagnose setup issues |
| `mur doctor --fix` | Auto-fix common issues |

### Pattern Management

| Command | Description |
|---------|-------------|
| `mur new <name>` | Create new pattern from template |
| `mur edit <name>` | Edit pattern in $EDITOR |
| `mur learn list` | List all patterns |
| `mur learn get <name>` | Show pattern details |
| `mur learn delete <name>` | Delete a pattern |
| `mur lint` | Validate all patterns |
| `mur lint <name>` | Validate specific pattern |

### Learning & Extraction

| Command | Description |
|---------|-------------|
| `mur transcripts` | Browse Claude Code sessions |
| `mur transcripts --project X` | Filter by project |
| `mur transcripts show <id>` | View session content |
| `mur learn extract` | Extract patterns from sessions |
| `mur learn extract --auto` | Auto-extract from recent sessions |

### Import & Export

| Command | Description |
|---------|-------------|
| `mur import file.yaml` | Import patterns from file |
| `mur import https://...` | Import from URL |
| `mur export` | Export all patterns (YAML) |
| `mur export --format json` | Export as JSON |
| `mur export -o file.yaml` | Export to file |

### Sync & Deploy

| Command | Description |
|---------|-------------|
| `mur sync` | Sync patterns to all CLIs/IDEs |
| `mur sync --push` | Also push to learning repo |
| `mur repo set <url>` | Set learning repo |
| `mur repo status` | Show repo status |

### Dashboard & Analytics

| Command | Description |
|---------|-------------|
| `mur serve` | Start interactive web dashboard |
| `mur serve -p 3000` | Custom port |
| `mur dashboard` | Generate static HTML report |
| `mur dashboard -o report.html` | Save to file |
| `mur stats` | View usage statistics |

### Maintenance

| Command | Description |
|---------|-------------|
| `mur update` | Update mur components |
| `mur health` | Check AI CLI availability |

## 🎯 How It Works

```
┌─────────────────────────────────────────────────────────────┐
│                         mur init                             │
│  1. Select AI CLIs (Claude, Gemini, Codex, etc.)            │
│  2. Install hooks (for real-time learning)                  │
│  3. Set up learning repo (optional, for sync)               │
│  4. Sync patterns to all CLIs                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Use Any CLI Normally                      │
│                                                              │
│   claude -p "fix bug"     gemini -p "explain"               │
│         │                        │                           │
│         └────────┬───────────────┘                          │
│                  ▼                                           │
│         Patterns auto-injected                               │
└─────────────────────────────────────────────────────────────┘
```

## 🔄 Sync Targets

mur syncs patterns to 8 targets:

**CLIs (dynamic injection via hooks):**
- Claude Code (`~/.claude/skills/mur/`)
- Gemini CLI (`~/.gemini/skills/mur/`)

**CLIs (static sync):**
- Codex (`~/.codex/instructions.md`)
- Auggie (`~/.augment/skills/mur/`)
- Aider (`~/.aider/mur-patterns.md`)

**IDEs (static sync):**
- Continue (`~/.continue/rules/mur/`)
- Cursor (`~/.cursor/rules/mur/`)
- Windsurf (`~/.windsurf/rules/mur/`)

## 📊 Dashboard

View your patterns and analytics:

```bash
# Interactive dashboard (opens browser)
mur serve

# Static HTML report
mur dashboard -o report.html
open report.html
```

## 🔧 Configuration

Config location: `~/.mur/config.yaml`

```yaml
default_tool: claude

tools:
  claude:
    enabled: true
  gemini:
    enabled: true

learning:
  repo: git@github.com:username/my-learnings.git
```

## 📁 Directory Structure

```
~/.mur/
├── config.yaml     # Configuration
├── patterns/       # Your learned patterns
├── stats.jsonl     # Usage statistics
└── repo/           # Learning repo (if configured)
```

## 📖 Links

- [Changelog](./CHANGELOG.md)
- [Issues](https://github.com/mur-run/mur-core/issues)

## 📄 License

MIT
