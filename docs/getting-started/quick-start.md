# Quick Start

Get up and running with MUR Core in under 5 minutes.

## Install

```bash
# macOS (Homebrew)
brew tap mur-run/tap && brew install mur

# Or via Go
CGO_ENABLED=0 go install github.com/mur-run/mur-core/cmd/mur@latest
```

## Initialize

```bash
mur init
```

The setup wizard will guide you through:

1. **AI CLI detection** — finds Claude Code, Gemini CLI, etc.
2. **Model setup** — choose cloud (recommended) or local
3. **Hook installation** — enables real-time pattern injection

```
📦 Model Setup
? Choose setup mode:
  > ☁️  Cloud (recommended) - API keys, best quality, ~$0.02/month
    🏠 Local - Ollama, free, needs ~2.7GB disk
    🔧 Custom - pick providers individually
```

**Cloud mode** uses OpenAI for both search and extraction. Set your API key:

```bash
export OPENAI_API_KEY=sk-...
```

**Local mode** uses Ollama (free, no API key needed):

```bash
ollama pull mxbai-embed-large   # 669MB, for search
ollama pull llama3.2:3b          # 2GB, for extraction
```

## Quick Setup (Non-Interactive)

For quick setup with defaults:

```bash
mur init --hooks
```

This installs hooks for all detected AI CLIs with local (Ollama) defaults.

## Verify

```bash
mur status
mur doctor    # Check for issues
```

## Use It

Just use your AI CLI normally — MUR works in the background:

```bash
claude "fix this bug"
# → MUR automatically injects relevant patterns
```

## Build Search Index

```bash
mur index rebuild

# With document expansion (better natural language search)
mur index rebuild --expand
```

## What's Next?

- **[Configuration](configuration.md)** — Customize providers and models
- **[Semantic Search](../semantic-search.md)** — Advanced search features
- **[Commands](../commands.md)** — Full command reference
