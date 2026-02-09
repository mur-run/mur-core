# mur 🔊

**Your AI assistant's memory.**

mur captures patterns from your coding sessions and injects them back into your AI tools. Learn once, remember forever.

## Why mur?

You're using Claude Code, Gemini CLI, maybe Cursor or Windsurf. But:

- Each session starts from scratch — the AI forgets your preferences
- Patterns you discover stay in your head (or get lost)
- Every AI tool is an isolated island with no shared memory

**mur fixes this.**

## Key Features

<div class="grid cards" markdown>

-   :material-brain: **Continuous Learning**

    ---

    Extract patterns from your coding sessions automatically. LLM-powered extraction finds insights humans miss.

    [:octicons-arrow-right-24: Learn more](concepts/patterns.md)

-   :material-sync: **Universal Sync**

    ---

    Patterns sync to 8+ AI tools: Claude, Gemini, Codex, Auggie, Aider, Continue, Cursor, Windsurf.

    [:octicons-arrow-right-24: Learn more](concepts/cross-cli-sync.md)

-   :material-ghost: **Zero Friction**

    ---

    Install hooks once, then forget about it. Use your AI CLI normally — mur works invisibly.

    [:octicons-arrow-right-24: Quick start](getting-started/quick-start.md)

-   :material-server: **Local First**

    ---

    All data stays on your machine. Optional git sync for multi-machine setups.

    [:octicons-arrow-right-24: Configuration](getting-started/configuration.md)

</div>

## Quick Example

```bash
# One-time setup
mur init --hooks

# Use your AI CLI normally — mur injects relevant patterns
claude "fix this SwiftUI bug"
# → mur automatically injects your Swift/SwiftUI patterns

# Extract patterns from sessions (runs automatically via hooks)
mur learn extract --llm

# Sync to all AI tools
mur sync
```

## Supported Tools

| Tool | Hooks | Static Sync |
|------|-------|-------------|
| [Claude Code](integrations/claude-code.md) | ✅ | ✅ |
| [Gemini CLI](integrations/gemini-cli.md) | ✅ | ✅ |
| [Codex](integrations/auggie.md) | — | ✅ |
| [Auggie](integrations/auggie.md) | — | ✅ |
| [Aider](integrations/auggie.md) | — | ✅ |
| Continue | — | ✅ |
| Cursor | — | ✅ |
| Windsurf | — | ✅ |

## Get Started

[Installation :material-download:](getting-started/installation.md){ .md-button .md-button--primary }
[Quick Start :material-rocket:](getting-started/quick-start.md){ .md-button }
