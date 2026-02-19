# GitHub Copilot

MUR Core syncs patterns to GitHub Copilot via the copilot-instructions.md file.

## How It Works

GitHub Copilot reads project-specific instructions from `.github/copilot-instructions.md`. MUR Core syncs your patterns there:

```
~/.github/
└── copilot-instructions.md    # Patterns for Copilot
```

## Setup

```bash
# Initialize mur (if not done)
mur init

# Sync patterns to Copilot
mur sync
```

GitHub Copilot will automatically use these instructions when generating code.

## Verify Integration

```bash
cat ~/.github/copilot-instructions.md
# Should show your patterns

mur status
# Should show "GitHub Copilot" as a sync target
```

## Usage

Once synced, GitHub Copilot will:
- Consider patterns when generating completions
- Follow conventions from your patterns
- Apply learned best practices

### Manual Re-sync

After adding or updating patterns:

```bash
mur sync
```

### Auto-sync

Enable automatic background sync:

```bash
mur sync auto enable
```

## Per-Project Instructions

You can also add project-specific instructions in your repository:

```bash
# In your project root
mkdir -p .github
echo "# Project-specific instructions" > .github/copilot-instructions.md
```

MUR's global instructions complement project-specific ones.

## Troubleshooting

### Patterns not being used

1. Ensure Copilot is enabled in VS Code/IDE
2. Check file exists: `cat ~/.github/copilot-instructions.md`
3. Restart your IDE

### Too many patterns

MUR Core truncates patterns to fit Copilot's context. High-effectiveness patterns are prioritized.

## Related

- [VS Code Extension](./vscode.md)
- [Cursor Integration](./cursor.md)
- [All Integrations](../index.md)
