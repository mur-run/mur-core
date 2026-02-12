# Aider

[Aider](https://aider.chat) is an AI pair programming tool that works in your terminal.

## Installation

```bash
mur init --hooks
```

This creates templates at `~/.aider/`:
- `conventions.md` - Pattern injection file
- `aider.conf.yml` - Configuration template

## Project Setup

For each project:

1. Copy the templates:
```bash
cp -r ~/.aider .aider
```

2. Inject patterns:
```bash
mur inject --tool aider > .aider/conventions.md
```

3. Start aider with config:
```bash
aider --config .aider/aider.conf.yml
```

## Configuration

The generated `aider.conf.yml`:

```yaml
read:
  - .aider/conventions.md

auto-commits: true
```

## Updating Patterns

Refresh patterns anytime:

```bash
mur inject --tool aider > .aider/conventions.md
```

Or automate in your shell:

```bash
alias aider-mur='mur inject --tool aider > .aider/conventions.md && aider --config .aider/aider.conf.yml'
```

## conventions.md Format

```markdown
# MUR Patterns

## Swift Testing
Use Swift Testing macros instead of XCTest...

## Error Handling
Always wrap errors with context...
```

Aider reads this as context for all conversations.

## Tips

- Re-inject patterns when switching projects
- Add `.aider/` to `.gitignore` for personal patterns
- Commit `.aider/` for team-shared patterns
