# Cursor IDE

MUR syncs patterns to Cursor IDE via the rules system.

## How It Works

Cursor supports custom rules in `~/.cursor/rules/`. MUR syncs patterns there:

```
~/.cursor/rules/
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

# Sync patterns to Cursor
mur sync
```

That's it! Cursor will automatically load patterns from the rules directory.

## Verify Integration

Check that patterns are synced:

```bash
ls ~/.cursor/rules/
# Should show mur-index/ and pattern directories

mur status
# Should show "Cursor" as a sync target
```

## Usage

Once synced, Cursor's AI features will have access to your patterns. The AI can:
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
2. Verify files exist: `ls ~/.cursor/rules/`
3. Restart Cursor to reload rules

### Old patterns showing

```bash
# Clean and resync
mur sync --clean-old
```

## Related

- [Continue.dev Integration](./continue-dev.md)
- [VS Code Extension](./vscode.md)
- [All Integrations](../index.md)
