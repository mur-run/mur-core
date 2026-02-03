# 23 - Neovim Plugin

## Summary
Create a Neovim plugin for murmur-ai integration, allowing users to run murmur commands directly from their editor.

## Motivation
Many developers use Neovim as their primary editor. Providing a native plugin allows seamless integration with the murmur-ai workflow without leaving the editor.

## Design

### Directory Structure
```
integrations/neovim/
├── lua/
│   └── murmur/
│       ├── init.lua       # Main module
│       └── commands.lua   # Command implementations
├── plugin/
│   └── murmur.vim         # Auto-load commands
└── README.md
```

### Commands
| Command | Description | Implementation |
|---------|-------------|----------------|
| `:MurmurSync` | Sync all sources | `mur sync all` |
| `:MurmurExtract` | Extract patterns from current buffer | `mur learn extract --auto` |
| `:MurmurStats` | Show statistics | Floating window with stats |
| `:MurmurPatterns` | Browse patterns | Telescope picker (fallback to split) |

### Implementation Details
- Use `vim.fn.jobstart` for async command execution
- Notifications via `vim.notify` for status updates
- Floating windows for stats display
- Optional Telescope integration for pattern browsing

## Dependencies
- Neovim 0.7+ (for lua API)
- `mur` CLI in PATH
- Optional: telescope.nvim for enhanced pattern browsing

## Testing
- Manual testing in Neovim
- Verify all commands execute correctly
- Test with/without Telescope installed

## Rollout
1. Create plugin structure
2. Implement core commands
3. Add optional Telescope integration
4. Document installation methods
