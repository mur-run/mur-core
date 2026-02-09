# Semantic Search

mur v1.1+ includes intelligent pattern matching using embeddings. Instead of keyword search, mur finds patterns by *meaning*.

## How It Works

```
┌─────────────────────────────────────────────────────────────┐
│                  Semantic Search Flow                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Index Phase (mur index rebuild)                         │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │ Pattern     │───▶│ Ollama      │───▶│ Cache       │     │
│  │ Content     │    │ Embed       │    │ Vector      │     │
│  └─────────────┘    └─────────────┘    └─────────────┘     │
│                                                             │
│  2. Search Phase (mur search "query")                       │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │ Query       │───▶│ Embed       │───▶│ Cosine      │     │
│  │ Text        │    │ Query       │    │ Similarity  │     │
│  └─────────────┘    └─────────────┘    └──────┬──────┘     │
│                                               │             │
│  3. Results                                   ▼             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ swift-testing-macro (0.84)                          │   │
│  │ node-version-manager (0.72)                         │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Setup

### 1. Install Ollama

```bash
# macOS
brew install ollama

# Linux
curl -fsSL https://ollama.com/install.sh | sh
```

### 2. Start Ollama and Pull Model

```bash
ollama serve &
ollama pull nomic-embed-text
```

### 3. Build Index

```bash
mur index rebuild
```

## Commands

### Search

```bash
# Basic search
mur search "How to test async Swift code"

# Get more results
mur search --top 5 "Docker best practices"

# JSON output (for scripts)
mur search --json "database optimization" | jq

# Hook mode (for Claude Code integration)
mur search --inject "$PROMPT"
```

### Index Management

```bash
# Check status
mur index status

# Rebuild after adding patterns
mur index rebuild

# Index single pattern
mur index pattern my-new-pattern
```

## Configuration

```yaml
# ~/.mur/config.yaml

search:
  enabled: true                    # Enable semantic search
  provider: ollama                 # ollama | openai
  model: nomic-embed-text          # Embedding model
  ollama_url: http://localhost:11434
  top_k: 3                         # Default results count
  min_score: 0.6                   # Minimum similarity (0-1)
  auto_inject: true                # Auto-suggest in hooks

embeddings:
  cache_enabled: true
  cache_dir: ~/.mur/embeddings
```

## Hook Integration

When `search.auto_inject: true`, mur automatically suggests relevant patterns in Claude Code:

```
User: "How do I test async Swift code?"
        │
        ▼
[mur] 🎯 Relevant patterns: swift-testing-macro-over-xctest
[mur] 💡 Consider loading /swift--testing-macro-over-xctest for this task
        │
        ▼
Claude sees the hint and loads the skill
```

### Re-install Hooks

```bash
mur init --hooks
```

## Troubleshooting

### "Ollama not running"

```bash
ollama serve
# Or in a new terminal: ollama serve &
```

### "Model not found"

```bash
ollama pull nomic-embed-text
```

### "No indexed patterns"

```bash
mur index rebuild
```

### Search returns empty

- Check `min_score` — lower it if patterns aren't matching
- Run `mur index status` to verify patterns are indexed
- Try more specific queries

## Alternative Providers

### OpenAI

```yaml
search:
  provider: openai
  model: text-embedding-3-small
  # Set OPENAI_API_KEY environment variable
```

## Performance

| Patterns | Index Time | Search Time |
|----------|------------|-------------|
| 50       | ~2s        | <100ms      |
| 100      | ~4s        | <100ms      |
| 500      | ~20s       | <200ms      |

Embeddings are cached — subsequent searches are instant.
