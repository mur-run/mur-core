# Continue.dev

[Continue.dev](https://continue.dev) is an open-source AI code assistant for VS Code and JetBrains.

## Installation

```bash
mur init --hooks
```

This adds:
- Custom command `/mur-patterns`
- Custom command `/mur-search`
- Context provider `@mur`

## Usage

### Custom Commands

In Continue chat, type:

```
/mur-patterns
```

This injects relevant patterns into your conversation.

```
/mur-search Swift testing
```

This searches your pattern library.

### Context Provider

Add patterns to any prompt with:

```
@mur How do I implement this?
```

## Manual Configuration

If automatic installation doesn't work, edit `~/.continue/config.json`:

```json
{
  "customCommands": [
    {
      "name": "mur-patterns",
      "description": "Inject relevant mur patterns into context",
      "prompt": "{{#if (shell \"mur inject --format md\")}}The following patterns may be relevant:\n\n{{shell \"mur inject --format md\"}}{{/if}}"
    },
    {
      "name": "mur-search",
      "description": "Search mur patterns",
      "prompt": "{{#if input}}{{shell \"mur search --format md '{{input}}'\"}}{{/if}}"
    }
  ],
  "contextProviders": [
    {
      "name": "mur",
      "params": {
        "command": "mur inject --format md"
      }
    }
  ]
}
```

## Tips

- Use `/mur-patterns` at the start of complex tasks
- Use `@mur` when you need persistent pattern context
- Use `/mur-search` for specific pattern lookup
