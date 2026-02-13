# OpenCode CLI

MUR syncs patterns to OpenCode CLI via the instructions file.

## How It Works

OpenCode reads instructions from `~/.opencode/instructions.md`. MUR syncs patterns there:

```
~/.opencode/
└── instructions.md    # Patterns merged into instructions
```

## Setup

```bash
# Install OpenCode (if not installed)
brew install opencode-ai/tap/opencode

# Initialize mur (if not done)
mur init

# Sync patterns to OpenCode
mur sync
```

## Verify Integration

```bash
cat ~/.opencode/instructions.md
# Should show your patterns

mur status
# Should show "OpenCode" as a sync target
```

## Usage

Once synced, OpenCode will:
- Include patterns in its system context
- Apply learned conventions to generated code
- Follow best practices from your patterns

### Manual Re-sync

```bash
mur sync
```

### Auto-sync

```bash
mur sync auto enable
```

## Troubleshooting

### Patterns not appearing

1. Check sync succeeded: `mur sync`
2. Verify file exists: `cat ~/.opencode/instructions.md`
3. Run `opencode` again

## Related

- [Claude Code Integration](./claude-code.md)
- [Gemini CLI Integration](./gemini-cli.md)
- [All Integrations](../index.md)
