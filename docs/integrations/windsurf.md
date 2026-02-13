# Windsurf IDE

MUR syncs patterns to Windsurf IDE via the rules system.

## How It Works

Windsurf supports custom rules in `~/.windsurf/rules/`. MUR syncs patterns there:

```
~/.windsurf/rules/
├── mur-index/
│   └── SKILL.md          # Index of all patterns
├── swift--testing/
│   ├── SKILL.md          # Pattern summary
│   └── examples.md       # Detailed examples
└── go--error-handling/
    └── SKILL.md
```

## Setup

```bash
# Initialize mur (if not done)
mur init

# Sync patterns to Windsurf
mur sync
```

That's it! Windsurf will automatically load patterns from the rules directory.

## Verify Integration

Check that patterns are synced:

```bash
ls ~/.windsurf/rules/
# Should show mur-index/ and pattern directories

mur status
# Should show "Windsurf" as a sync target
```

## Usage

Once synced, Windsurf's Cascade AI will have access to your patterns. It can:
- Reference patterns when generating code
- Follow conventions from your patterns
- Suggest solutions based on learned patterns

### Manual Re-sync

After adding or updating patterns:

```bash
mur sync
```

### Auto-sync

Enable automatic background sync:

```bash
mur sync auto enable
# Choose interval: 15m, 30m, 1h, etc.
```

## Troubleshooting

### Patterns not appearing

1. Check sync succeeded: `mur sync`
2. Verify files exist: `ls ~/.windsurf/rules/`
3. Restart Windsurf to reload rules

### Old patterns showing

```bash
# Clean and resync
mur sync --clean-old
```

## Related

- [Cursor Integration](./cursor.md)
- [Continue.dev Integration](./continue-dev.md)
- [All Integrations](../index.md)
