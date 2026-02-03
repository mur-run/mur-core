# 18 - Shell Completion

## Summary
Enable shell completion support for bash, zsh, fish, and PowerShell using Cobra's built-in completion system.

## Motivation
Tab completion improves CLI usability significantly. Users can discover commands and flags without reading documentation.

## Design

### Commands
- `mur completion bash` — Output bash completion script
- `mur completion zsh` — Output zsh completion script  
- `mur completion fish` — Output fish completion script
- `mur completion powershell` — Output PowerShell completion script

### Implementation
Cobra (1.6+) automatically provides a built-in `completion` command with subcommands for each shell. No custom code needed — it just works out of the box.

## Files Changed
- `README.md` — Add installation instructions for shell completion

## Testing
```bash
# Build
go build ./...

# Test bash completion output
./mur completion bash | head -20

# Test zsh completion output  
./mur completion zsh | head -20
```

## Installation Instructions (for README)
```bash
# Bash
mur completion bash > /etc/bash_completion.d/mur

# Zsh (add to fpath)
mur completion zsh > "${fpath[1]}/_mur"

# Fish
mur completion fish > ~/.config/fish/completions/mur.fish

# PowerShell
mur completion powershell | Out-String | Invoke-Expression
```

## Status: DONE
Cobra's built-in completion command works perfectly. Only README update needed.
