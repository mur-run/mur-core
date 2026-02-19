# VS Code Extension

The MUR Patterns extension brings pattern injection directly into VS Code.

## Installation

### From Marketplace

Search for "MUR Patterns" in the VS Code Extensions view, or:

```
ext install mur-run.mur-patterns
```

### From Source

```bash
git clone https://github.com/mur-run/mur-vscode
cd mur-vscode
npm install
npm run compile
code --install-extension mur-patterns-*.vsix
```

## Requirements

- mur-core CLI installed (`brew install mur` or `go install github.com/mur-run/mur-core/cmd/mur@latest`)

## Commands

| Command | Shortcut | Description |
|---------|----------|-------------|
| `MUR: Inject Patterns` | `Cmd+Shift+M` | Inject relevant patterns for current file |
| `MUR: Search Patterns` | `Cmd+Shift+F M` | Search your pattern library |
| `MUR: Show Pattern Stats` | - | View usage analytics |
| `MUR: Learn from Current File` | - | Extract patterns from code |
| `MUR: Rate Pattern` | - | Give feedback on pattern effectiveness |

## Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `mur.autoInject` | `false` | Auto-inject patterns when opening files |
| `mur.murPath` | `mur` | Path to MUR CLI binary |
| `mur.maxPatterns` | `5` | Maximum patterns to inject |

## Usage

### Inject Patterns

1. Open any code file
2. Press `Cmd+Shift+M` (or `Ctrl+Shift+M` on Windows/Linux)
3. Relevant patterns appear in the Output panel

Patterns are selected based on:
- File extension (language)
- Project type
- Content similarity

### Search Patterns

1. Press `Cmd+Shift+F M`
2. Enter a natural language query
3. Results appear in the Output panel

Example queries:
- "Swift async testing"
- "Docker multi-stage builds"
- "Go error handling"

### Learn from Code

1. Select code in the editor (or leave empty for full file)
2. Right-click → "MUR: Learn from Current File"
3. Patterns are extracted and shown

### Context Menu

Right-click in any editor to access:
- Inject Patterns
- Learn from Current File

## Status Bar

The extension shows a status bar item:
- `$(lightbulb) MUR` - Click to view stats
- `$(lightbulb) MUR (3)` - Number of injected patterns

## Troubleshooting

### "MUR command not found"

Set the full path in settings:

```json
{
  "mur.murPath": "/opt/homebrew/bin/mur"
}
```

### No patterns found

1. Check you have patterns: `mur list`
2. Run indexing: `mur index`
3. Verify search works: `mur search "test"`

## Links

- [mur-vscode GitHub](https://github.com/mur-run/mur-vscode)
- [mur-core CLI](https://github.com/mur-run/mur-core)
