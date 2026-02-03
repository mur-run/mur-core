# Murmur AI - VS Code Extension

VS Code integration for [murmur-ai](https://github.com/anthropics/murmur-ai) — learn and share coding patterns with AI agents.

## Features

- **Sync All**: Synchronize patterns across your projects
- **Extract Learnings**: Automatically extract patterns from your codebase
- **Show Stats**: View murmur statistics in the output panel
- **List Patterns**: Browse all learned patterns
- **Status Bar**: See pattern count at a glance

## Requirements

- [murmur-ai CLI](https://github.com/anthropics/murmur-ai) (`mur`) must be installed and available in PATH

## Installation

### From Source (Development)

1. Clone the repository:
   ```bash
   git clone https://github.com/anthropics/murmur-ai
   cd murmur-ai/integrations/vscode
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Compile:
   ```bash
   npm run compile
   ```

4. Open in VS Code and press F5 to launch Extension Development Host

### Install Locally via VSIX

```bash
npm run vscode:prepublish
vsce package
code --install-extension murmur-ai-0.1.0.vsix
```

## Commands

Open Command Palette (`Cmd+Shift+P` / `Ctrl+Shift+P`) and type:

| Command | Description |
|---------|-------------|
| `Murmur: Sync All` | Run `mur sync all` |
| `Murmur: Extract Learnings` | Run `mur learn extract --auto` |
| `Murmur: Show Stats` | Display statistics in output panel |
| `Murmur: List Patterns` | List all patterns in output panel |

## Status Bar

The extension shows a status bar item on the left:
- Displays pattern count (e.g., `📝 42 patterns`)
- Click to view patterns list
- Updates automatically after sync

## Output

All command output is shown in the **Murmur** output channel:
- View → Output → Select "Murmur" from dropdown
- Or click on any command result notification

## Troubleshooting

### "mur: command not found"

Make sure `mur` CLI is installed and in your PATH:
```bash
which mur
# Should show path like /usr/local/bin/mur
```

### Extension not activating

1. Open Command Palette
2. Run "Developer: Show Running Extensions"
3. Check if "Murmur AI" is listed

## Development

```bash
# Watch mode
npm run watch

# Lint
npm run lint

# Compile
npm run compile
```

## License

MIT
