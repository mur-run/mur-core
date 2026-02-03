# Murmur AI Skill

Unified management layer for AI CLI tools. Sync configurations, extract learned patterns, and view usage statistics across Claude Code, Gemini CLI, and Auggie.

## Requirements

- `mur` CLI installed and in PATH
- murmur-ai configured (`~/.murmur/config.yaml`)

## Commands

### /murmur:sync

Synchronize all configurations to AI CLI tools.

**What it syncs:**
- MCP server configurations
- Hook configurations  
- Learned patterns
- Skills

**Usage:**
```
/murmur:sync
```

**Example output:**
```json
{
  "success": true,
  "results": {
    "mcp": [
      {"target": "claude", "success": true, "message": "2 servers synced"}
    ],
    "patterns": [
      {"target": "claude", "success": true, "message": "synced 5 patterns"}
    ]
  }
}
```

### /murmur:learn

Extract patterns from recent coding sessions.

**What it does:**
- Scans Claude Code session transcripts from the last 7 days
- Identifies reusable patterns (code snippets, workflows, preferences)
- Reports found patterns without auto-saving (dry-run mode)

**Usage:**
```
/murmur:learn
```

**Example output:**
```json
{
  "sessions_scanned": 12,
  "patterns_found": [
    {
      "name": "swift-error-handling",
      "category": "code",
      "confidence": 0.85,
      "preview": "Use Result<T, Error> for..."
    }
  ]
}
```

### /murmur:stats

Display usage statistics for AI CLI tools.

**What it shows:**
- Tool usage counts (Claude, Gemini, Auggie)
- Estimated costs
- Routing decisions
- Usage trends

**Usage:**
```
/murmur:stats
```

**Example output:**
```json
{
  "total_calls": 142,
  "by_tool": {
    "claude": {"calls": 89, "cost_usd": 2.45},
    "gemini": {"calls": 45, "cost_usd": 0.12},
    "auggie": {"calls": 8, "cost_usd": 0.00}
  },
  "period": "all"
}
```

## Implementation Notes

This skill wraps the `mur` CLI. Commands execute:
- `/murmur:sync` → `mur output sync --json`
- `/murmur:learn` → `mur learn extract --auto --dry-run`
- `/murmur:stats` → `mur output stats --json`

## Learn More

- [murmur-ai on GitHub](https://github.com/mur-run/mur-cli)
- [Full Documentation](https://github.com/mur-run/mur-cli#readme)
