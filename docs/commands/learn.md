# mur learn

Manage learned patterns in your knowledge base.

## Usage

```bash
mur learn <subcommand> [flags]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `list` | List all patterns |
| `add <name>` | Add a new pattern |
| `get <name>` | Show pattern details |
| `delete <name>` | Delete a pattern |
| `sync` | Sync patterns to AI tools |
| `extract` | Extract patterns from sessions |
| `init <repo>` | Initialize learning repo |
| `push` | Push patterns to repo |
| `pull` | Pull shared patterns |
| `repo-sync` | Full repo sync (push + pull) |

## Pattern Management

### List Patterns

```bash
mur learn list
```

Output:

```
Learned Patterns
================

  error-handling-go     [golang/pattern]  85%
    Always wrap errors with context using fmt.Errorf
  api-response-format   [api/convention]  90%
    Use consistent JSON response format
  test-naming           [testing/pattern] 75%
    Test functions should be named Test_<function>_<scenario>

Total: 3 patterns
```

Filter by domain or category:

```bash
mur learn list --domain golang
mur learn list --category pattern
```

### Add Pattern

Interactive mode:

```bash
mur learn add my-pattern
```

From stdin:

```bash
echo "Always use context.Context as first parameter" | mur learn add go-context --stdin
```

### View Pattern

```bash
mur learn get error-handling-go
```

### Delete Pattern

```bash
mur learn delete old-pattern
mur learn delete old-pattern --force  # Skip confirmation
```

## Pattern Extraction

Extract patterns automatically from your AI coding sessions.

### Auto-Extract

Scan recent sessions and extract patterns:

```bash
mur learn extract --auto
```

Preview without saving:

```bash
mur learn extract --auto --dry-run
```

### From Specific Session

```bash
mur learn extract --session abc123
```

### Interactive

Choose from recent sessions:

```bash
mur learn extract
```

## Sync to AI Tools

Patterns are injected into AI tool instructions so all tools benefit:

```bash
mur learn sync
```

Output:

```
Syncing patterns to AI tools...

  ✓ Claude Code: synced 5 patterns
  ✓ Gemini CLI: synced 5 patterns
  ✓ Auggie: synced 5 patterns
```

Clean up orphaned patterns:

```bash
mur learn sync --cleanup
```

## Learning Repository

Share patterns across machines using a git repository.

### Initialize

```bash
mur learn init git@github.com:user/my-patterns.git
```

### Push Local Patterns

```bash
mur learn push
```

### Pull Shared Patterns

```bash
mur learn pull
```

### Full Sync

```bash
mur learn repo-sync
```

## Pattern Structure

Patterns have these fields:

| Field | Description |
|-------|-------------|
| `name` | Unique identifier |
| `description` | Short summary |
| `domain` | Area (golang, python, api, etc.) |
| `category` | Type (pattern, convention, antipattern) |
| `confidence` | How reliable (0.0-1.0) |
| `content` | The actual pattern content |

## Domains

- `golang`, `python`, `javascript`, `typescript`, `rust`
- `api`, `database`, `testing`
- `architecture`, `security`
- `general`

## Categories

- `pattern` - Best practices
- `convention` - Style conventions
- `antipattern` - What to avoid
- `snippet` - Code templates
- `explanation` - Concepts

## See Also

- [Patterns Concept](../concepts/patterns.md) - How patterns work
- [Team Command](team.md) - Team pattern sharing
