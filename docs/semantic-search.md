# Semantic Search

MUR Core v1.1+ includes intelligent pattern matching using embeddings. Instead of keyword search, MUR finds patterns by *meaning*.

## Quick Setup

### Option 1: Ollama (Free, Local)

```bash
# Install Ollama
brew install ollama          # macOS
# or: curl -fsSL https://ollama.com/install.sh | sh  # Linux

# Start and pull model
ollama serve &
ollama pull nomic-embed-text

# Build index
mur index rebuild
```

### Option 2: OpenAI (No GPU Required)

```bash
export OPENAI_API_KEY=sk-xxx
```

```yaml
# ~/.mur/config.yaml
search:
  provider: openai
  model: text-embedding-3-small
  min_score: 0.7                 # OpenAI scores are higher
```

```bash
mur index rebuild
```

## Usage

```bash
# Search by meaning
mur search "How to test async Swift code"
# → swift-testing-macro-over-xctest (0.58)

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
  provider: ollama              # ollama | openai
  model: nomic-embed-text       # See model table below
  ollama_url: http://localhost:11434
  top_k: 3                      # Max results
  min_score: 0.5                # Minimum similarity (0.0-1.0)
  auto_inject: false
    # false = inject patterns by project/tags only
    # true  = also inject semantically similar patterns
```

## Embedding Models

| Provider | Model | min_score | Cost |
|----------|-------|-----------|------|
| ollama | `nomic-embed-text` | 0.5 | Free |
| openai | `text-embedding-3-small` | 0.7 | $0.02/1M tokens |
| openai | `text-embedding-3-large` | 0.7 | $0.13/1M tokens |

**Ollama vs OpenAI:**
- Ollama: Free, offline, needs local GPU, scores ~0.5-0.6
- OpenAI: Paid, needs internet, no GPU, scores ~0.7-0.9

## Automatic Injection

When enabled, MUR Core automatically suggests relevant patterns in Claude Code:

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

**Slow search?**
- Reduce `top_k` in config
- Use smaller embedding model
