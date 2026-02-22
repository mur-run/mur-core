# Configuration Guide

MUR configuration is stored in:
- **macOS/Linux:** `~/.mur/config.yaml`
- **Windows:** `%USERPROFILE%\.mur\config.yaml`

Run `mur init` for interactive setup, or edit the file directly.

## Full Configuration Reference

```yaml
schema_version: 2

# Default AI tool
default_tool: claude

# Tool-specific settings
tools:
  claude:
    enabled: true
  gemini:
    enabled: true

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🔍 Semantic Search
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Cost: ~$0.001 for 200 patterns with OpenAI.
search:
  enabled: true
  provider: openai                # openai | ollama | google | voyage
  model: text-embedding-3-small  # See provider table below
  api_key_env: OPENAI_API_KEY    # env var name (not the key!)
  min_score: 0.3                 # OpenAI: 0.3 | Ollama: 0.5
  top_k: 3
  auto_inject: true

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🧠 Learning & Extraction
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Cost: ~$0.02 for 200 patterns with paid models.
learning:
  auto_extract: true
  sync_to_tools: true
  llm:
    provider: ollama              # ollama | openai | gemini | claude
    model: llama3.2:3b            # See provider table below
    # api_key_env: OPENAI_API_KEY # For cloud providers

# Sync settings
sync:
  format: directory               # directory (recommended) | single
  clean_old: false

# Cloud sync (requires mur.run account)
server:
  url: https://api.mur.run

# Pattern consolidation
consolidation:
  enabled: true
  schedule: weekly
  auto_merge: keep-best

# Community sharing
community:
  share_enabled: true
  auto_share_on_push: true

# Cache settings
cache:
  community:
    ttl_minutes: 60
    cleanup: on_sync
```

## Embedding Providers

| Provider | Model | Cost | Quality | Config |
|----------|-------|------|---------|--------|
| **OpenAI** | `text-embedding-3-small` | $0.02/1M tokens | ⭐⭐⭐⭐ | `api_key_env: OPENAI_API_KEY` |
| **Google** | `text-embedding-004` | Free tier | ⭐⭐⭐⭐ | `api_key_env: GEMINI_API_KEY` |
| **Voyage** | `voyage-3-large` | $0.18/1M tokens | ⭐⭐⭐⭐⭐ | `api_key_env: VOYAGE_API_KEY` |
| **Ollama** | `qwen3-embedding` | Free | ⭐⭐⭐ | `ollama_url: http://localhost:11434` |

## LLM Providers (Extraction & Expansion)

| Provider | Model | Cost | Config |
|----------|-------|------|--------|
| **Ollama** | `llama3.2:3b` | Free | No API key needed |
| **OpenAI** | `gpt-4o-mini` | $0.15/1M in | `api_key_env: OPENAI_API_KEY` |
| **Gemini** | `gemini-2.0-flash` | $0.10/1M in | `api_key_env: GEMINI_API_KEY` |
| **Claude** | `claude-haiku` | $0.25/1M in | `api_key_env: ANTHROPIC_API_KEY` |

## API Keys

API keys are set via environment variables (never stored in config):

```bash
# Add to ~/.zshrc or ~/.bashrc
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="..."
export ANTHROPIC_API_KEY="sk-ant-..."
export VOYAGE_API_KEY="..."
```

The `api_key_env` field in config specifies which env var to read.

## View & Change Settings

```bash
mur config show                      # View current config
mur config default gemini            # Set default tool
mur config set search.provider openai
mur config set search.model text-embedding-3-small
```

## Configuration Locations

| Path | Purpose |
|------|---------|
| `~/.mur/config.yaml` | Main configuration |
| `~/.mur/patterns/` | Learned patterns |
| `~/.mur/embeddings/` | Embedding cache + expanded queries |
| `~/.mur/hooks/` | Hook scripts (on-stop.sh, on-prompt.sh) |
| `~/.mur/transcripts/` | Session transcripts |
| `~/.mur/tracking/` | Usage tracking |
