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

```bash
go install github.com/mur-run/mur-core/cmd/mur@latest
```

## 🎯 How It Works

```
┌─────────────────────────────────────────────────────────────┐
│                         mur init                             │
│  1. Select AI CLIs (Claude, Gemini, Codex, etc.)            │
│  2. Install Claude Code hooks (for real-time learning)      │
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
│         Patterns auto-applied                                │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     mur sync (periodic)                      │
│  • Pulls from learning repo (if configured)                 │
│  • Syncs patterns to all CLIs                               │
└─────────────────────────────────────────────────────────────┘
```

## 📋 Commands

### Essential

| Command | Description |
|---------|-------------|
| `mur init` | Interactive setup wizard |
| `mur sync` | Pull patterns + sync to CLIs |
| `mur sync --push` | Also push local changes to remote |
| `mur learn` | Add/manage learned patterns |
| `mur stats` | View learning statistics |

### Repository

| Command | Description |
|---------|-------------|
| `mur repo set <url>` | Set learning repo |
| `mur repo status` | Show repo status |
| `mur repo remove` | Remove repo config |

### Maintenance

| Command | Description |
|---------|-------------|
| `mur update` | Update mur (binary, hooks, skills) |
| `mur health` | Check AI CLI availability |

## 🔄 Learning Repo

Store patterns in a git repo for:
- Sync across machines
- Team sharing
- Backup

```bash
# Set up during init, or later:
mur repo set git@github.com:username/my-learnings.git

# Check status
mur repo status

# Sync (pull + apply)
mur sync

# Push changes
mur sync --push
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

learning:
  repo: git@github.com:username/my-learnings.git
```

## 📁 Directory Structure

```
~/.mur/
├── config.yaml     # Configuration
├── patterns/       # Learned patterns (git repo)
├── hooks/          # Hook templates
└── transcripts/    # Session logs
```

## 🤝 Supported CLIs

| CLI | Patterns | Hooks |
|-----|----------|-------|
| Claude Code | ✅ | ✅ |
| Gemini CLI | ✅ | - |
| Codex | ✅ | - |
| Auggie | ✅ | - |
| Aider | ✅ | - |

## 📖 Links

- [Documentation](./docs/)
- [Changelog](./CHANGELOG.md)
- [Issues](https://github.com/mur-run/mur-core/issues)

## 📄 License

MIT
