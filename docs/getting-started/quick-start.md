# Quick Start

This guide will get you up and running with MUR Core in under 5 minutes.

## Initialize

First, create your configuration:

```bash
mur init
```

This creates `~/.mur/config.yaml` with sensible defaults.

## Check Available Tools

See which AI CLI tools are installed and working:

```bash
mur health
```

Example output:

```
AI Tool Health Check
====================

  ✓ claude      Claude Code     ~/.npm-global/bin/claude
  ✓ gemini      Gemini CLI      /usr/local/bin/gemini
  ✗ auggie      Auggie          not found

Available: 2/3 tools
```

## Run Your First Prompt

```bash
mur run -p "what is a goroutine in Go?"
```

MUR Core analyzes your prompt and routes it to the most appropriate tool:

```
→ gemini (auto: complexity 0.15 < 0.50 threshold, using free tool)

A goroutine is a lightweight thread managed by the Go runtime...
```

## Try Different Prompts

```bash
# Simple question → routes to free tool
mur run -p "explain REST APIs"

# Complex task → routes to paid tool
mur run -p "refactor this authentication module for OAuth2 support"

# See the routing decision without running
mur run -p "debug this memory leak" --explain
```

## Force a Specific Tool

Override automatic routing when needed:

```bash
# Always use Claude for this prompt
mur run -t claude -p "write a haiku about coding"

# Always use Gemini
mur run -t gemini -p "complex analysis" 
```

## What's Next?

- **[Configuration](configuration.md)** - Customize routing behavior
- **[Smart Routing](../concepts/routing.md)** - Understand how routing works
- **[Pattern Learning](../concepts/patterns.md)** - Extract and share patterns
- **[Cross-CLI Sync](../concepts/cross-cli-sync.md)** - Sync MCP and hooks
