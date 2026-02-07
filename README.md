# mur 🔊

**Continuous learning for AI assistants.**

mur makes your AI CLIs smarter over time. Learn once, remember forever.

## 🚀 Quick Start

```bash
# Install
go install github.com/mur-run/mur-core/cmd/mur@latest

# Setup (interactive)
mur init

# That's it! Use your AI CLI as normal
claude -p "fix this bug"
gemini -p "explain this code"
```

mur works invisibly in the background. Your patterns are synced to all CLIs.

## 📦 Installation

### Go Install (recommended)

```bash
go install github.com/mur-run/mur-core/cmd/mur@latest
```

### From Source

```bash
git clone https://github.com/mur-run/mur-core.git
cd mur-core
go install ./cmd/mur
```

## 🎯 How It Works

```
┌─────────────────────────────────────────────────────────────┐
│                         mur init                             │
│  • Detects your AI CLIs (Claude, Gemini, Codex, etc.)       │
│  • Installs learning hooks                                   │
│  • Syncs patterns to all CLIs                               │
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
│         Patterns auto-applied                                │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     mur sync (periodic)                      │
│  • Syncs new patterns to all CLIs                           │
│  • ~/.claude/skills/mur-patterns.md                         │
│  • ~/.gemini/skills/mur-patterns.md                         │
│  • etc.                                                      │
└─────────────────────────────────────────────────────────────┘
```

## 📋 Commands

### Essential

| Command | Description |
|---------|-------------|
| `mur init` | Interactive setup wizard |
| `mur sync` | Sync patterns to all CLIs |
| `mur learn` | Add/manage learned patterns |
| `mur stats` | View learning statistics |

### Maintenance

| Command | Description |
|---------|-------------|
| `mur update` | Update mur (binary, hooks, skills) |
| `mur health` | Check AI CLI availability |

### Advanced (use `mur run`)

```bash
# Smart routing (auto-select best CLI)
mur run -p "fix this bug"

# Force specific CLI
mur run -t claude -p "explain this"
```

## 🔧 Configuration

After `mur init`, config is at `~/.mur/config.yaml`:

```yaml
default_tool: claude

tools:
  claude:
    enabled: true
  gemini:
    enabled: true
```

## 📁 Directory Structure

```
~/.mur/
├── config.yaml     # Configuration
├── patterns/       # Learned patterns
├── hooks/          # Hook templates
└── transcripts/    # Session logs
```

## 🤝 Supported CLIs

| CLI | Patterns Sync | Hooks |
|-----|--------------|-------|
| Claude Code | ✅ | ✅ |
| Gemini CLI | ✅ | - |
| Codex | ✅ | - |
| Auggie | ✅ | - |
| Aider | ✅ | - |

## 📖 Learn More

- [Documentation](./docs/)
- [Changelog](./CHANGELOG.md)
- [Issues](https://github.com/mur-run/mur-core/issues)

## 📄 License

MIT
