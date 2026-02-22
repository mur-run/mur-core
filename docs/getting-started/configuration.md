# Configuration

MUR Core's configuration lives at `~/.mur/config.yaml`. Run `mur init` for interactive setup.

## Full Configuration Reference

```yaml
# ~/.mur/config.yaml

schema_version: 2

# Default tool when routing is disabled
default_tool: claude

# Smart routing settings
routing:
  mode: auto  # auto | manual | cost-first | quality-first
  complexity_threshold: 0.5  # 0.0-1.0

# AI tool definitions
tools:
  claude:
    enabled: true
    binary: claude
    tier: paid
    capabilities: [coding, analysis, complex]
  gemini:
    enabled: true
    binary: gemini
    tier: free
    capabilities: [coding, simple-qa]

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🔍 Semantic Search
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Embedding model for pattern search.
# Cost: ~$0.001 for 200 patterns with OpenAI.
#
# Providers: openai | ollama | google | voyage
search:
  enabled: true
  provider: openai                    # openai (recommended) | ollama (free)
  model: text-embedding-3-small      # OpenAI: text-embedding-3-small
                                      # Ollama: qwen3-embedding
                                      # Google: text-embedding-004
  api_key_env: OPENAI_API_KEY        # env var name (cloud providers only)
  ollama_url: http://localhost:11434  # Ollama only
  top_k: 3
  min_score: 0.3                     # OpenAI: 0.3 | Ollama: 0.5
  auto_inject: true

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🧠 Learning & Extraction
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# LLM for pattern extraction and search query expansion.
# Cost: ~$0.02 for 200 patterns with paid models.
#
# Providers: ollama | openai | gemini | claude
learning:
  auto_extract: true
  sync_to_tools: true
  llm:
    provider: ollama                  # ollama (free) | openai | gemini | claude
    model: llama3.2:3b                # Ollama: llama3.2:3b
                                      # OpenAI: gpt-4o-mini
                                      # Gemini: gemini-2.0-flash
                                      # Claude: claude-haiku
    api_key_env: ""                   # env var name (cloud providers only)

# MCP server configuration (synced to all tools)
mcp:
  servers:
    filesystem:
      command: npx
      args: ["-y", "@modelcontextprotocol/server-filesystem", "/home/user"]

# Pattern Consolidation
consolidation:
  enabled: true
  schedule: weekly
  auto_merge: keep-best

# Community Sharing
community:
  share_enabled: true
  auto_share_on_push: true
```

## Setup Modes

`mur init` offers three modes:

| Mode | Embedding | LLM | Cost | Quality |
|------|-----------|-----|------|---------|
| ☁️ Cloud | OpenAI | OpenAI | ~$0.02/mo | ⭐⭐⭐⭐⭐ |
| 🏠 Local | Ollama | Ollama | Free | ⭐⭐⭐ |
| 🔧 Custom | Mix & match | Mix & match | Varies | Varies |

## Routing Modes

### `auto` (default)

Analyzes prompt complexity and routes accordingly:

- **Complexity < threshold** → Free tool (Gemini)
- **Complexity ≥ threshold** → Paid tool (Claude)

### `manual`

Always uses `default_tool`. No automatic routing.

### `cost-first` / `quality-first`

Aggressive preference for cost savings or quality.

## Environment Variables

Cloud providers read API keys from environment variables:

```bash
# In your shell profile (.zshrc, .bashrc, etc.)
export OPENAI_API_KEY=sk-...
export GEMINI_API_KEY=...
export ANTHROPIC_API_KEY=sk-ant-...
export VOYAGE_API_KEY=...
```

The `api_key_env` config field specifies which env var to use (not the key itself).

## View & Change Settings

```bash
mur config show              # View current config
mur config default gemini    # Set default tool
mur config routing cost-first # Set routing mode
mur config set search.enabled true
```

## Configuration Locations

| File | Purpose |
|------|---------|
| `~/.mur/config.yaml` | Main configuration |
| `~/.mur/patterns/` | Learned patterns |
| `~/.mur/embeddings/` | Embedding cache |
| `~/.mur/hooks/` | Hook scripts |
| `~/.mur/transcripts/` | Session transcripts |
