# Troubleshooting

Common issues and solutions.

## Installation Issues

### "command not found: mur"

**macOS/Linux (Go install):**
```bash
export PATH="$HOME/go/bin:$PATH"
# Add to ~/.zshrc or ~/.bashrc for persistence
```

**Windows:**
Add `%USERPROFILE%\go\bin` to PATH:
1. Search "Environment Variables"
2. Edit User Variables → Path → Add new entry

### "LC_UUID" error on macOS

Use `CGO_ENABLED=0` when installing:
```bash
CGO_ENABLED=0 go install github.com/mur-run/mur-core/cmd/mur@latest
```

### Verify Installation

```bash
mur doctor
```

## Semantic Search Issues

### "Ollama not running"

```bash
# Start Ollama
ollama serve &

# Verify it's running
curl http://localhost:11434/api/tags
```

### "No embeddings found"

```bash
# Check index status
mur index status

# Rebuild index
mur index rebuild
```

### Search returns no results

- Check `min_score` in config (lower it if needed)
- Ensure patterns have content (not just names)
- Rebuild index: `mur index rebuild`

## Sync Issues

### "Not logged in"

```bash
mur login
# Or for CI/automation:
mur login --api-key mur_xxx_...
```

### "Team not found"

```bash
# List available teams
mur cloud teams

# Select correct team
mur cloud select <team-slug>
```

### Conflicts during sync

Use interactive resolution:
```bash
mur cloud sync
# Choose: [i] Interactive, [s] Server, [l] Local
```

## Hook Issues

### Hooks not working

1. Check if hooks are installed:
```bash
cat ~/.claude/settings.json | grep hooks
```

2. Reinstall hooks:
```bash
mur init --hooks
```

### Claude Code not seeing patterns

1. Check sync status:
```bash
mur status
```

2. Force sync:
```bash
mur sync
```

3. Check skills directory:
```bash
ls ~/.claude/skills/
```

## Learning/Extraction Issues

### "No transcripts found"

Transcripts are created by Claude Code. Make sure you've:
1. Used Claude Code at least once
2. Check transcript location:
```bash
mur transcripts --list
```

### LLM extraction fails

1. Check Ollama is running (if using Ollama)
2. Check API key is set (if using OpenAI/Gemini/Claude)
3. Try a different provider:
```bash
mur learn extract --llm openai
```

## Performance Issues

### Slow search

- Reduce `top_k` in config
- Use smaller embedding model
- Reduce number of patterns

### High memory usage

- Run `mur clean` to remove temp files
- Rebuild index: `mur index rebuild`

## Data Locations

| Platform | Location |
|----------|----------|
| macOS/Linux | `~/.mur/` |
| Windows | `%USERPROFILE%\.mur\` |

**Contents:**
```
.mur/
├── config.yaml      # Configuration
├── patterns/        # Your patterns
├── hooks/           # Hook scripts
├── embeddings/      # Search index
├── stats.jsonl      # Usage stats
└── repo/            # Git sync (optional)
```

## Getting Help

1. Check docs: `mur web`
2. Run diagnostics: `mur doctor`
3. GitHub issues: https://github.com/mur-run/mur-core/issues
4. Discord: https://discord.gg/mur

## Debug Mode

For verbose output:
```bash
mur --verbose <command>
```

Check logs:
```bash
# macOS/Linux
cat ~/.mur/mur.log

# Windows
type %USERPROFILE%\.mur\mur.log
```
