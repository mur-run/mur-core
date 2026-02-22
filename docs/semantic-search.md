# Semantic Search

MUR Core v1.1+ includes intelligent pattern matching using embeddings. Instead of keyword search, MUR finds patterns by *meaning*.

## Quick Setup

### Option 1: Cloud (Recommended)

No GPU, no downloads. Best search quality.

```bash
export OPENAI_API_KEY=sk-xxx
```

```yaml
# ~/.mur/config.yaml
search:
  enabled: true
  provider: openai
  model: text-embedding-3-small
  api_key_env: OPENAI_API_KEY
  min_score: 0.3
```

```bash
mur index rebuild
```

**Cost:** ~$0.001 for 200 patterns. Virtually free.

### Option 2: Ollama (Free, Local)

```bash
# Install Ollama
brew install ollama          # macOS
# or: curl -fsSL https://ollama.com/install.sh | sh  # Linux

# Start and pull model
ollama serve &
ollama pull qwen3-embedding

# Build index
mur index rebuild
```

> **Note:** Previous versions used `nomic-embed-text` or `mxbai-embed-large`. We now recommend `qwen3-embedding` for significantly better search quality (MTEB 70.7 vs 64.7) and 100+ language support including Chinese.

### Interactive Setup

`mur init` will guide you through model selection:

```bash
mur init

# Choose setup mode:
#   ☁️  Cloud (recommended) - API keys, best quality, ~$0.02/month
#   🏠 Local - Ollama, free, needs ~2.7GB disk
#   🔧 Custom - pick providers individually
```

## Usage

```bash
# Search by meaning
mur search "How to test async Swift code"
# → swift-testing-macro-over-xctest (0.78)

# Natural language works great
mur search "how to sign a macOS app"
# → bitl-binary-signing-workaround (0.71)

# JSON output for scripts
mur search --json "error handling"

# Check index status
mur index status
```

## Configuration

```yaml
# ~/.mur/config.yaml
search:
  enabled: true
  provider: openai              # openai | ollama | google | voyage
  model: text-embedding-3-small # See model table below
  api_key_env: OPENAI_API_KEY   # env var name (not the key itself!)
  top_k: 3                      # Max results
  min_score: 0.3                # Minimum similarity (0.0-1.0)
  auto_inject: true             # Auto-inject to AI CLI prompts
```

## Embedding Providers

| Provider | Model | Cost | Quality | GPU? |
|----------|-------|------|---------|------|
| **OpenAI** | `text-embedding-3-small` | $0.02/1M tokens | ⭐⭐⭐⭐ | No |
| **OpenAI** | `text-embedding-3-large` | $0.13/1M tokens | ⭐⭐⭐⭐⭐ | No |
| **Google** | `text-embedding-004` | Free tier (1500 req/min) | ⭐⭐⭐⭐ | No |
| **Voyage** | `voyage-3-large` | $0.18/1M tokens | ⭐⭐⭐⭐⭐ | No |
| **Ollama** | `qwen3-embedding` | Free | ⭐⭐⭐⭐ | Local |
| **Ollama** | `mxbai-embed-large` (legacy) | Free | ⭐⭐⭐ | Local |
| **Ollama** | `nomic-embed-text` (legacy) | Free | ⭐⭐ | Local |

**Recommended:** `text-embedding-3-small` for best cost/quality ratio. 200 patterns costs ~$0.001.

### Provider Configuration

<details>
<summary>OpenAI</summary>

```yaml
search:
  provider: openai
  model: text-embedding-3-small
  api_key_env: OPENAI_API_KEY
  min_score: 0.3
```

```bash
export OPENAI_API_KEY=sk-...
```
</details>

<details>
<summary>Google (nearly free)</summary>

```yaml
search:
  provider: google
  model: text-embedding-004
  api_key_env: GEMINI_API_KEY
  min_score: 0.3
```

```bash
export GEMINI_API_KEY=...
```
</details>

<details>
<summary>Voyage (best for code)</summary>

```yaml
search:
  provider: voyage
  model: voyage-3-large
  api_key_env: VOYAGE_API_KEY
  min_score: 0.3
```

```bash
export VOYAGE_API_KEY=...
```
</details>

<details>
<summary>Ollama (free, local)</summary>

```yaml
search:
  provider: ollama
  model: qwen3-embedding
  ollama_url: http://localhost:11434
  min_score: 0.5
```

```bash
ollama pull qwen3-embedding
```
</details>

## Document Expansion

Improve search quality by generating search queries for each pattern using a local LLM:

```bash
# Requires an LLM model (e.g., llama3.2:3b via Ollama)
ollama pull llama3.2:3b
mur index rebuild --expand
```

This generates 5 likely search queries per pattern (e.g., "how to sign macos binary" for `bitl-binary-signing-workaround`), making natural language searches much more effective.

**Expansion is cached** — subsequent rebuilds reuse generated queries unless you delete `~/.mur/embeddings/expanded_queries.json`.

## LLM Configuration

The LLM used for extraction and expansion is configured separately:

```yaml
# ~/.mur/config.yaml
learning:
  llm:
    provider: ollama              # ollama | openai | gemini | claude
    model: llama3.2:3b            # Non-reasoning model recommended
```

| Provider | Model | Cost | Speed |
|----------|-------|------|-------|
| **Ollama** | `llama3.2:3b` | Free | ~1s/pattern |
| **OpenAI** | `gpt-4o-mini` | $0.15/1M in | Fast |
| **Gemini** | `gemini-2.0-flash` | $0.10/1M in | Fastest |
| **Claude** | `claude-haiku` | $0.25/1M in | Fast |

> **Avoid reasoning models** (deepseek-r1, qwq) for expansion — they "think out loud" and waste tokens.

## Automatic Injection

When enabled, MUR automatically suggests relevant patterns in Claude Code:

```
claude "fix this async test"
# → [mur] 🎯 Relevant patterns: swift-testing-macro-over-xctest
```

Enable in config:
```yaml
search:
  auto_inject: true
```

## Directory Sync Format

v1.1+ uses individual skill directories instead of one large file:

```
Before: ~/.claude/skills/mur-patterns.md (35KB, ~8,750 tokens)
After:  ~/.claude/skills/swift--testing-macro/SKILL.md (~150 tokens)
```

This reduces token usage by **90%+** — Claude loads only needed patterns.

```bash
mur sync                  # Uses directory format (default)
mur sync --format single  # Legacy single-file format
```

## Troubleshooting

**"OpenAI API key required"**
```bash
export OPENAI_API_KEY=sk-...
# Or set api_key_env in config to use a different env var
```

**"Ollama not running"**
```bash
ollama serve &
mur index rebuild
```

**"No embeddings found"**
```bash
mur index status
mur index rebuild
```

**Switching from nomic-embed-text or mxbai-embed-large to qwen3-embedding?**
```bash
ollama pull qwen3-embedding
# Update config.yaml: model: qwen3-embedding
mur index rebuild   # Full rebuild needed (different vector space)
```
