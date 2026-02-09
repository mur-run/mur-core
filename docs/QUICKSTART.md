# Quick Start Guide

Get started with mur in 5 minutes.

## Installation

```bash
# Go install (recommended)
CGO_ENABLED=0 go install github.com/mur-run/mur-core/cmd/mur@latest

# Add to PATH if needed
export PATH="$HOME/go/bin:$PATH"
```

## Initialize

```bash
mur init --hooks
```

This creates `~/.mur/config.yaml` and sets up CLI hooks for automatic learning.

## You're Done!

Just use your AI CLI normally:

```bash
# mur automatically injects relevant patterns
claude "fix this authentication bug"
gemini "explain this error"
```

## Check Status

```bash
# See overview
mur status

# Diagnose issues
mur doctor
```

## Create Your First Pattern

```bash
# Interactive creation
mur new swift-error-handling

# Opens in your editor with template
```

Or create manually:

```yaml
# ~/.mur/patterns/swift-error-handling.yaml
name: swift-error-handling
description: Handle errors with Result types
content: |
  When handling errors in Swift:
  1. Use Result<Success, Error> for functions that can fail
  2. Prefer throwing functions for synchronous operations
  3. Use async/await with try for async operations

tags:
  confirmed: [swift, error-handling]
  
applies:
  languages: [swift]
  keywords: [error, Result, throw, catch]

schema_version: 2
```

## Extract Patterns from Sessions

```bash
# See available sessions
mur transcripts

# Extract patterns using LLM
mur learn extract --llm

# Auto-accept high-confidence patterns
mur learn extract --llm --accept-all
```

## Sync to All AI Tools

```bash
mur sync

# Output:
# ✓ Claude Code: Synced 12 patterns
# ✓ Gemini CLI: Synced 12 patterns
# ✓ Codex: Synced 12 patterns
# ... (8 tools total)
```

## Multi-Machine Sync (Optional)

Sync patterns across machines via git:

```bash
# Set up learning repo
mur learn init git@github.com:you/patterns.git

# Push patterns
mur learn push

# Pull on other machines
mur learn pull
```

## LLM Extraction Setup

For better pattern extraction, configure an LLM:

```yaml
# ~/.mur/config.yaml
learning:
  llm:
    provider: ollama          # ollama | openai | gemini | claude
    model: deepseek-r1:8b     # recommended for Ollama
```

Or use API providers:

```bash
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="..."
export ANTHROPIC_API_KEY="sk-ant-..."
```

## Common Commands

| Command | Description |
|---------|-------------|
| `mur status` | Overview of patterns, sync status |
| `mur doctor` | Diagnose and fix issues |
| `mur sync` | Sync patterns to all AI tools |
| `mur new <name>` | Create new pattern |
| `mur search <query>` | Search patterns |
| `mur transcripts` | Browse Claude Code sessions |
| `mur learn extract --llm` | Extract patterns with AI |
| `mur serve` | Web dashboard |

## What's Next?

- **View stats**: `mur stats`
- **Web dashboard**: `mur serve`
- **All commands**: `mur --help`
- **GitHub**: https://github.com/mur-run/mur-core
