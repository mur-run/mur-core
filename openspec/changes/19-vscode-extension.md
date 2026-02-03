# Change Spec: VS Code Extension

**ID:** 19  
**Status:** Complete  
**Created:** 2026-02-03

## Summary

Create a VS Code extension for murmur-ai that provides IDE integration for common CLI commands.

## Motivation

Developers using VS Code should be able to:
- Sync patterns without leaving the IDE
- Extract learnings from current workspace
- View stats and patterns in the editor
- See pattern count at a glance in status bar

## Design

### Directory Structure

```
integrations/vscode/
├── package.json          # Extension manifest
├── src/
│   └── extension.ts      # Main entry point
├── tsconfig.json         # TypeScript config
└── README.md             # Usage documentation
```

### Commands

| Command ID | Title | Action |
|------------|-------|--------|
| `murmur.sync` | Murmur: Sync All | `mur sync all` |
| `murmur.learn.extract` | Murmur: Extract Learnings | `mur learn extract --auto` |
| `murmur.stats` | Murmur: Show Stats | `mur stats --json` → Output panel |
| `murmur.patterns` | Murmur: List Patterns | `mur pattern list` → Output panel |

### Status Bar

- Position: Left side
- Shows: Pattern count (e.g., "📝 42 patterns")
- Click: Opens patterns list
- Updates: On activation and after sync

### Implementation Details

1. **CLI Execution**: Use `child_process.exec` to call `mur` CLI
2. **Output**: Use `vscode.window.createOutputChannel('Murmur')` for command output
3. **Error Handling**: Show errors via `vscode.window.showErrorMessage`
4. **Status Bar**: Use `vscode.window.createStatusBarItem`

## Package.json Highlights

```json
{
  "name": "murmur-ai",
  "displayName": "Murmur AI",
  "publisher": "karajanchang",
  "version": "0.1.0",
  "engines": { "vscode": "^1.80.0" },
  "activationEvents": ["onStartupFinished"],
  "main": "./out/extension.js"
}
```

## Testing

1. Open extension folder in VS Code
2. Press F5 to launch Extension Development Host
3. Test each command via Command Palette (Cmd+Shift+P)
4. Verify status bar shows pattern count

## Checklist

- [x] Create change spec
- [x] Create extension directory structure
- [x] Implement extension.ts
- [x] Configure package.json
- [x] Add tsconfig.json
- [x] Write README.md
- [x] Test locally (compiles successfully)
- [x] Commit and push
