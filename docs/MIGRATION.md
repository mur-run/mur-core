# Migration Guide

## v1.0.x → v1.1.0

v1.1 introduces semantic search and a new sync format. Here's how to upgrade.

### 1. Update MUR

```bash
CGO_ENABLED=0 go install github.com/mur-run/mur-core/cmd/mur@latest
mur version  # Should show 1.1.0+
```

### 2. Migrate Patterns to v2 Schema

The new schema adds versioning and embedding metadata:

```bash
# Preview changes (no modifications)
mur migrate --dry-run

# Apply migration
mur migrate

# Verify
mur doctor --check patterns
```

**What changes:**
- Adds `schema_version: 2`
- Adds `version: "1.0.0"` 
- Adds `embedding_hash` for search cache
- Adds `resources` for L3 tracking

### 3. Set Up Semantic Search (Optional)

```bash
# Install Ollama
brew install ollama
ollama serve &
ollama pull nomic-embed-text

# Build embedding index
mur index rebuild
mur index status  # Verify all patterns indexed
```

### 4. Re-install Hooks

To get the new search hook:

```bash
mur init --hooks
```

This adds:
```json
{
  "command": "mur search --inject \"$PROMPT\" 2>/dev/null || true"
}
```

### 5. Sync Patterns

The new directory format reduces token usage by 90%+:

```bash
# Uses new format (default)
mur sync

# Remove old single-file format
mur sync --clean-old

# Or keep both formats
mur sync --format directory
mur sync --format single
```

**Before:**
```
~/.claude/skills/mur-patterns.md  # 35KB, ~8,750 tokens
```

**After:**
```
~/.claude/skills/
├── mur-index/SKILL.md           # 300 tokens (catalog)
├── swift--testing-macro/
│   ├── SKILL.md                 # 150 tokens (summary)
│   └── examples.md              # Full content (on-demand)
└── ...
```

## Configuration Changes

### New: search section

```yaml
# ~/.mur/config.yaml

search:
  enabled: true
  provider: ollama
  model: nomic-embed-text
  top_k: 3
  min_score: 0.6
  auto_inject: true
```

### New: sync section

```yaml
sync:
  format: directory      # directory (new) | single (legacy)
  prefix_domain: true    # swift--pattern-name format
  l3_threshold: 500      # chars before splitting to examples.md
```

## Rollback

If you need to go back:

```bash
# Use legacy sync format
mur sync --format single

# Disable semantic search
mur config set search.enabled false

# Patterns are backward compatible
# v1.0.x can read v2 patterns (ignores new fields)
```

## Troubleshooting

### "MUR migrate: no patterns found"

Patterns should be in `~/.mur/patterns/`:
```bash
ls ~/.mur/patterns/
```

### "MUR index: ollama not running"

```bash
ollama serve
# Then retry
mur index rebuild
```

### Hooks not working

Re-install:
```bash
mur init --hooks
cat ~/.claude/settings.json | jq '.hooks'
```

## Breaking Changes

None — v1.1 is fully backward compatible with v1.0.x patterns and configuration.
