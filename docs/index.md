# Murmur 🔊

**Unified Multi-AI CLI Management + Cross-Tool Learning System**

Every AI CLI tool is an isolated island. Murmur unifies them.

## Why Murmur?

You're using Claude Code for complex refactoring, Gemini CLI for quick questions, maybe Auggie for free coding help. But:

- Each tool has different commands and flags
- MCP configurations don't sync between tools
- Patterns you learn with one tool stay siloed
- You're paying for Claude when Gemini would suffice

**Murmur fixes all of this.**

## Key Features

<div class="grid cards" markdown>

-   :material-router: **Smart Routing**

    ---

    Automatically routes prompts to the right tool. Simple questions → free tools. Complex tasks → paid tools.

    [:octicons-arrow-right-24: Learn more](concepts/routing.md)

-   :material-sync: **Cross-Tool Sync**

    ---

    Configure MCP servers once, sync everywhere. Hooks, patterns, all unified.

    [:octicons-arrow-right-24: Learn more](concepts/cross-cli-sync.md)

-   :material-brain: **Pattern Learning**

    ---

    Extract patterns from sessions. What Claude learns, Gemini knows too.

    [:octicons-arrow-right-24: Learn more](concepts/patterns.md)

-   :material-account-group: **Team Knowledge**

    ---

    Share patterns across your team. New members inherit experience automatically.

    [:octicons-arrow-right-24: Learn more](commands/team.md)

</div>

## Quick Example

```bash
# Initialize murmur
mur init

# Run a prompt - auto-routes to best tool
mur run -p "what is dependency injection?"
# → Routes to Gemini (free) - simple Q&A

mur run -p "refactor this authentication module for better testability"
# → Routes to Claude (paid) - complex architecture task

# See why it chose that tool
mur run -p "fix this race condition" --explain
# Shows complexity score, category, tool selection reason
```

## Supported AI Tools

| Tool | Status | Cost |
|------|--------|------|
| [Claude Code](integrations/claude-code.md) | ✅ Supported | Paid |
| [Gemini CLI](integrations/gemini-cli.md) | ✅ Supported | Free |
| [Auggie](integrations/auggie.md) | 🔜 Coming | Free |
| [OpenClaw](integrations/openclaw.md) | 🔜 Coming | Paid |

## Get Started

Ready to unify your AI CLI tools?

[Installation :material-download:](getting-started/installation.md){ .md-button .md-button--primary }
[Quick Start :material-rocket:](getting-started/quick-start.md){ .md-button }
