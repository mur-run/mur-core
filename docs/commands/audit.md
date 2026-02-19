# mur audit

View the audit trail of pattern injections and operations.

## Usage

```bash
# Show recent audit entries
mur audit

# Filter by pattern name
mur audit --pattern "api-retry-pattern"

# Limit results
mur audit --limit 20
```

## Output

```
📋 Pattern Audit Log (last 10 entries)
───────────────────────────────────────

2026-02-19 19:30:12  inject  api-retry-pattern       source: hook
2026-02-19 19:28:45  inject  go-error-handling        source: hook
2026-02-19 18:15:00  share   docker-compose-setup     source: cli
2026-02-19 17:42:33  inject  swift-async-await        source: hook
2026-02-19 17:42:33  inject  swift-concurrency        source: hook
2026-02-19 16:00:00  verify  *                        source: cli
```

## What Gets Logged

| Action | When |
|--------|------|
| `inject` | Pattern injected into an AI prompt via hooks |
| `load` | Pattern loaded with hash verification |
| `share` | Pattern shared to community |
| `modify` | Pattern content modified |
| `verify` | `mur verify` command run |

## Storage

- Location: `~/.mur/audit/audit.jsonl`
- Format: append-only JSONL (one JSON object per line)
- Auto-rotation: log rotates to `audit-YYYY-MM.jsonl` when exceeding 10MB
- Each entry includes: timestamp, pattern ID/name, action, source, prompt hash (SHA256, not the actual prompt)

## Privacy

The audit log stores a **SHA256 hash** of prompts, not the actual prompt text. This lets you correlate injection events without storing sensitive content.
