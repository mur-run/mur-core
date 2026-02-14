# Configuration Guide

MUR configuration is stored in:
- **macOS/Linux:** `~/.mur/config.yaml`
- **Windows:** `%USERPROFILE%\.mur\config.yaml`

## Full Configuration Reference

```yaml
# Default AI tool for commands
default_tool: claude

# Tool-specific settings
tools:
  claude:
    enabled: true
  gemini:
    enabled: true
  codex:
    enabled: true
  cursor:
    enabled: true

# Sync settings
sync:
  format: directory    # directory (recommended) or single
  clean_old: false     # Remove old single-file format

# Cloud sync (requires mur.run account)
server:
  url: https://api.mur.run
  team: ""             # Team slug for team sync

# Semantic search settings
search:
  enabled: true
  provider: ollama              # ollama | openai
  model: nomic-embed-text       # Embedding model
  ollama_url: http://localhost:11434
  top_k: 3                      # Max results per search
  min_score: 0.5                # Minimum similarity (0.0-1.0)
  auto_inject: false            # Auto-inject similar patterns

# Pattern learning settings
learning:
  repo: ""                      # Git repo for sync (optional)
  auto_push: false              # Auto-push after learning
  
  llm:
    provider: ollama            # ollama | openai | gemini | claude
    model: deepseek-r1:8b       # Model for extraction
    ollama_url: http://localhost:11434
    
    # Premium model for important sessions (optional)
    premium:
      provider: gemini
      model: gemini-2.0-flash
      api_key_env: GEMINI_API_KEY
    
    # When to use premium model
    routing:
      min_messages: 20          # Long sessions
      projects: [important-project]

# Cache settings
cache:
  community:
    ttl_minutes: 60
    cleanup: on_sync            # on_sync | manual
```

## API Keys

API keys are set via environment variables (not in config for security):

```bash
# Add to ~/.zshrc or ~/.bashrc
export ANTHROPIC_API_KEY="sk-ant-..."     # Claude
export OPENAI_API_KEY="sk-..."            # OpenAI
export GEMINI_API_KEY="..."               # Gemini
```

Reference in config by variable name:

```yaml
learning:
  llm:
    provider: claude
    model: claude-sonnet-4-20250514
    api_key_env: ANTHROPIC_API_KEY    # Variable NAME, not the key
```

## Embedding Models

| Provider | Model | min_score | Notes |
|----------|-------|-----------|-------|
| ollama | `nomic-embed-text` | 0.5 | Free, local |
| openai | `text-embedding-3-small` | 0.7 | $0.02/1M tokens |
| openai | `text-embedding-3-large` | 0.7 | $0.13/1M tokens |

## LLM Models for Extraction

| Provider | Model | Notes |
|----------|-------|-------|
| ollama | `deepseek-r1:8b` | Best local, 5GB |
| ollama | `qwen2.5:14b` | Good for code, 9GB |
| openai | `gpt-4o-mini` | Cheap & fast |
| gemini | `gemini-2.0-flash` | Free tier available |
| claude | `claude-sonnet-4-20250514` | Best quality |

## Remote Ollama (LAN Setup)

Run Ollama on a server and access from other machines.

### Server Setup

**macOS:**
```bash
launchctl setenv OLLAMA_HOST "0.0.0.0"
brew services restart ollama
```

**Linux:**
```bash
sudo systemctl edit ollama
# Add: Environment="OLLAMA_HOST=0.0.0.0"
sudo systemctl restart ollama
```

### Client Config

```yaml
search:
  provider: ollama
  ollama_url: http://192.168.1.100:11434

learning:
  llm:
    provider: ollama
    ollama_url: http://192.168.1.100:11434
```
