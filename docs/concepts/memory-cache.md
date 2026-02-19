# In-Process Memory Cache

How MUR Core makes pattern matching instant.

## The Problem

Every `mur hook` invocation reads pattern files from disk, parses YAML, and (for semantic search) loads embedding vectors from numpy files. With 500+ patterns, this adds 20-50ms of I/O overhead — repeated on every prompt.

## The Solution

Load everything once into RAM at startup. Serve all lookups from memory.

```
Before:  Prompt → Read 500 YAML files → Parse → Match → Inject
After:   Prompt → map[id]*Pattern lookup → Match → Inject
         (~15ms)                          (<0.1ms)
```

## Architecture

```
MemoryCache
├── PatternCache
│   ├── patterns map[string]*Pattern    // ID → Pattern
│   ├── byName map[string]string        // name → ID (case-insensitive)
│   └── index map[string][]string       // tag → Pattern IDs (inverted index)
│
└── EmbeddingMatrix
    ├── data []float32      // contiguous: [p0_d0, p0_d1, ..., p1_d0, ...]
    ├── normed []float32    // pre-normalized unit vectors
    ├── ids []string        // pattern ID at each row
    └── dim int             // vector dimensionality (768)
```

### Why Pre-Normalized Vectors?

Normal cosine similarity:
```
cos(A, B) = (A·B) / (|A| × |B|)    ← 3 operations
```

With unit vectors (|A| = |B| = 1):
```
cos(A, B) = A·B                     ← 1 dot product
```

We normalize once at load time, then every similarity lookup is a single dot product over contiguous memory — CPU cache friendly and auto-vectorizable by the Go compiler.

## Memory Usage

| Patterns | YAML Cache | Embeddings (768d × 2) | Total |
|----------|-----------|----------------------|-------|
| 100 | 300 KB | 600 KB | ~1 MB |
| 500 | 1.5 MB | 3 MB | ~4.5 MB |
| 1,000 | 3 MB | 6 MB | ~9 MB |
| 5,000 | 15 MB | 30 MB | ~45 MB |

Even 5,000 patterns use less RAM than a single Chrome tab.

## Lazy Loading

By default, embeddings load on-demand (first semantic search call). This keeps startup fast for users who only use keyword matching:

```
Startup with lazy loading:   ~5-20ms (patterns only)
First semantic search:       +50-100ms (loads embeddings)
Subsequent searches:         <1ms
```

## Configuration

The cache is automatic — no configuration needed. To disable for debugging:

```bash
mur run --no-cache "your prompt"
```

## What's NOT Cached

- **Ollama embedding API calls** — each new prompt still needs a fresh embedding from Ollama
- **LLM extraction** — pattern learning still calls the configured LLM
- **Community API calls** — network requests are never cached in-process

The cache only optimizes local pattern and embedding **file reads**.
