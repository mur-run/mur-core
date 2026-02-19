# OpenAI Codex CLI

MUR Core syncs patterns to Codex CLI via the instructions file.

## How It Works

Codex reads instructions from `~/.codex/instructions.md`. MUR appends patterns in a dedicated section:

```
~/.codex/
└── instructions.md    # Your instructions + MUR patterns
```

MUR Core wraps its content in markers:
```markdown
<!-- mur:start -->
## Learned Patterns (mur)
...patterns...
<!-- mur:end -->
```

This preserves your existing instructions.

## Setup

```bash
# Install Codex (if not installed)
npm i -g @openai/codex

# Initialize mur (if not done)
mur init

# Sync patterns to Codex
mur sync
```

## Verify Integration

```bash
cat ~/.codex/instructions.md
# Should show your patterns in the mur section

mur status
# Should show "Codex" as a sync target
```

## Usage

Once synced, Codex will:
- Read patterns from the MUR section
- Apply learned conventions
- Follow best practices from your patterns

### Manual Re-sync

```bash
mur sync
```

### Auto-sync

```bash
mur sync auto enable
```

## Your Own Instructions

Add your custom instructions above the MUR section:

```markdown
# My Codex Instructions

- Use TypeScript
- Follow our style guide

<!-- mur:start -->
## Learned Patterns (mur)
...
<!-- mur:end -->
```

MUR Core will preserve everything outside its markers.

## Troubleshooting

### Patterns overwriting my instructions

This shouldn't happen — MUR Core only modifies content between its markers. If it does:

1. Check for duplicate marker tags
2. Run `mur sync` to fix

## Related

- [Claude Code Integration](./claude-code.md)
- [OpenCode Integration](./opencode.md)
- [All Integrations](../index.md)
