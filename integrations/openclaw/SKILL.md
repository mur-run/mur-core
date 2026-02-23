---
name: mur-skill
description: "OpenClaw skill for mur-core — continuous learning for AI assistants. Sync patterns, search knowledge, record sessions, and manage your learning across all AI CLI tools."
---

# mur-core OpenClaw Skill

Continuous learning for AI assistants. Learn once, remember forever.

Wraps the [mur CLI](https://github.com/mur-run/mur-core) v1.12+.

## Requirements

```bash
go install github.com/mur-run/mur-core/cmd/mur@latest
mur init   # Interactive setup
```

## Commands

### /mur:status

Show mur status overview — config, patterns count, hooks, cloud connection.

```bash
mur status
```

### /mur:sync

Smart sync based on your plan. Cloud sync for Pro+, git sync for Free, then pushes to local AI CLIs.

```bash
mur sync
```

Options:
- `mur sync --cloud` — force cloud sync
- `mur sync --git` — force git sync
- `mur sync --cli` — only sync to local CLIs (no remote)

### /mur:search <query>

Semantic search across local and community patterns.

```bash
mur search "Swift async testing"
mur search --community "API retry patterns"
mur search --json "error handling"
```

### /mur:learn

List and manage learned patterns.

```bash
mur learn list                    # List all patterns
mur learn extract                 # Extract from recent sessions (dry-run)
mur learn extract --auto --save   # Auto-extract and save
mur learn add <name>              # Add a new pattern
mur learn get <name>              # Show a pattern
mur learn delete <name>           # Delete a pattern
```

### /mur:stats

Show pattern usage analytics — counts, costs, trends.

```bash
mur stats
```

### /mur:session (Recording)

Record conversation segments for workflow extraction. Use `/mur:in` and `/mur:out` markers in AI tools.

```bash
mur session start --source claude-code   # Start recording
mur session stop --analyze               # Stop + analyze
mur session list                         # List recordings
mur session analyze <id>                 # LLM analysis
mur session ui <id>                      # Web workflow editor
mur session export <id> --format skill   # Export as skill
```

**Typical flow:**
1. `/mur:in` → `mur session start`
2. Conversation happens (events captured)
3. `/mur:out` → `mur session stop --analyze --open`
4. Edit in web UI → Save → Export

### /mur:doctor

Diagnose setup issues and fix common problems.

```bash
mur doctor
```

### /mur:feedback

Rate pattern effectiveness to improve recommendations.

```bash
mur feedback helpful <pattern-name>
mur feedback unhelpful <pattern-name>
```

## Other Useful Commands

| Command | Purpose |
|---------|---------|
| `mur consolidate` | Score health, detect duplicates, resolve conflicts |
| `mur community` | Browse and copy community patterns |
| `mur dashboard` | Generate static HTML dashboard |
| `mur serve` | Start local dashboard server |
| `mur import` | Import patterns from external sources |
| `mur index rebuild` | Rebuild embedding index |
| `mur config` | View/edit configuration |

## Pattern Storage

Patterns live in `~/.mur/patterns/` with domain organization:

```
~/.mur/patterns/
├── _global/        # Cross-domain patterns
├── backend/        # Backend/API
├── devops/         # DevOps
├── web/            # Frontend
└── projects/       # Project-specific
```

## Learn More

- [GitHub](https://github.com/mur-run/mur-core)
- [Documentation](https://github.com/mur-run/mur-core#readme)
