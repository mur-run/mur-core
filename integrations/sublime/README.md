# Murmur AI - Sublime Text Plugin

Sublime Text integration for [murmur-ai](https://github.com/anthropics/murmur-ai) — learn and share coding patterns with AI agents.

## Features

- **Sync All**: Synchronize patterns across your projects
- **Extract Learnings**: Automatically extract patterns from your codebase
- **Show Stats**: View murmur statistics in the output panel
- **List Patterns**: Browse all learned patterns

## Requirements

- Sublime Text 3 or 4
- [murmur-ai CLI](https://github.com/anthropics/murmur-ai) (`mur`) must be installed and available in PATH

## Installation

### Manual Installation

1. Open Sublime Text
2. Go to **Preferences → Browse Packages...**
3. Create a folder called `MurmurAI`
4. Copy all files from this directory into `MurmurAI/`:
   - `MurmurAI.py`
   - `Main.sublime-menu`
   - `Default.sublime-commands`
   - `MurmurAI.sublime-settings`

```bash
# Or via command line (macOS)
cp -r integrations/sublime ~/Library/Application\ Support/Sublime\ Text/Packages/MurmurAI
```

### Package Control (Coming Soon)

Package Control support will be added in a future release.

## Usage

### Command Palette

Open Command Palette (`Cmd+Shift+P` / `Ctrl+Shift+P`) and type:

| Command | Description |
|---------|-------------|
| `Murmur: Sync All` | Run `mur sync all` |
| `Murmur: Extract Learnings` | Run `mur learn extract --auto` |
| `Murmur: Show Stats` | Display statistics in output panel |
| `Murmur: List Patterns` | List all patterns in output panel |

### Menu

**Tools → Murmur AI** menu provides access to all commands.

### Output

All command output is shown in the **Murmur** output panel:
- View → Show Console to see the panel
- Or it opens automatically after running commands

## Troubleshooting

### "mur: command not found"

Make sure `mur` CLI is installed and in your PATH:

```bash
which mur
# Should show path like /usr/local/bin/mur
```

If `mur` is installed but not in PATH, you can set the path in settings:

1. **Preferences → Package Settings → MurmurAI → Settings**
2. Set `mur_path` to the full path of your `mur` binary

### Commands not appearing

1. Ensure all files are in the correct location:
   ```
   ~/Library/Application Support/Sublime Text/Packages/MurmurAI/
   ```
2. Restart Sublime Text
3. Check the console (`View → Show Console`) for errors

### Plugin not loading

Check for Python syntax errors in the console:
1. **View → Show Console**
2. Look for any import or syntax errors related to MurmurAI

## Settings

Edit settings via **Preferences → Package Settings → MurmurAI → Settings**:

```json
{
    // Path to mur CLI (leave empty to use PATH)
    "mur_path": "",
    
    // Show output panel automatically after commands
    "show_output_panel": true,
    
    // Clear output panel before each command
    "clear_output_before_run": true
}
```

## Development

To modify the plugin:

1. Edit files in `~/Library/Application Support/Sublime Text/Packages/MurmurAI/`
2. Sublime Text will automatically reload the plugin on save
3. Check the console for any errors

## License

MIT
