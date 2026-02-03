# Change Spec: Sublime Text Plugin

## ID
21-sublime-plugin

## Status
Implemented

## Summary
Create a Sublime Text plugin for murmur-ai that allows users to run mur CLI commands directly from the editor.

## Motivation
Many developers use Sublime Text as their primary editor. Providing a native plugin makes it convenient to use murmur-ai features without switching to the terminal.

## Design

### Directory Structure
```
integrations/sublime/
├── MurmurAI.py              # Main plugin file
├── Main.sublime-menu        # Menu entries
├── Default.sublime-commands # Command palette
├── MurmurAI.sublime-settings # Default settings
└── README.md                # Documentation
```

### Commands
| Command | Description | CLI |
|---------|-------------|-----|
| `murmur_sync` | Sync all patterns | `mur sync all` |
| `murmur_learn_extract` | Extract learnings | `mur learn extract --auto` |
| `murmur_stats` | Show statistics | `mur stats` |
| `murmur_patterns` | List patterns | `mur patterns list` |

### Menu Integration
- Tools > Murmur AI > Sync All
- Tools > Murmur AI > Extract Learnings
- Tools > Murmur AI > Show Stats
- Tools > Murmur AI > List Patterns

### Implementation
- Uses `subprocess` to run mur CLI
- Output displayed in Sublime's output panel
- Async execution to avoid blocking the UI
- Status bar messages for feedback

## Files Changed
- `integrations/sublime/MurmurAI.py` - Main plugin
- `integrations/sublime/Main.sublime-menu` - Menu entries
- `integrations/sublime/Default.sublime-commands` - Command palette
- `integrations/sublime/MurmurAI.sublime-settings` - Settings
- `integrations/sublime/README.md` - Documentation

## Testing
- Manual testing in Sublime Text
- Verify all commands execute correctly
- Verify output panel displays results

## Notes
- Requires `mur` CLI in PATH
- Compatible with Sublime Text 3 and 4
