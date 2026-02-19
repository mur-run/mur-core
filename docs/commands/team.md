# MUR team

Manage team knowledge base and shared patterns.

## Usage

```bash
mur team <subcommand> [flags]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `init <repo>` | Initialize team repository |
| `push` | Push patterns to team repo |
| `pull` | Pull patterns from team repo |
| `sync` | Full sync (push + pull) |
| `status` | Show sync status |

## Team Knowledge Sharing

Share learned patterns across your team so everyone benefits from collective experience.

```
Developer A                 Team Repo                  Developer B
┌─────────────┐            ┌──────────┐              ┌─────────────┐
│ learns      │   push     │          │    pull     │ inherits    │
│ pattern     │ ────────►  │ patterns │  ────────►  │ pattern     │
│             │            │          │              │             │
└─────────────┘            └──────────┘              └─────────────┘
```

## Setup

### Initialize Team Repo

```bash
mur team init git@github.com:myorg/team-patterns.git
```

This creates a connection to a shared repository for your team's patterns.

### Configure in `~/.mur/config.yaml`

```yaml
team:
  repo_url: git@github.com:myorg/team-patterns.git
  auto_sync: true      # Sync automatically
  sync_interval: 24h   # How often to sync
  branch: main         # Branch to use
```

## Workflow

### Push Your Patterns

Share patterns you've learned:

```bash
mur team push
```

This pushes patterns marked as `team: true` to the shared repo.

### Pull Team Patterns

Get patterns others have shared:

```bash
mur team pull
```

### Full Sync

Push and pull in one command:

```bash
mur team sync
```

### Check Status

See what's pending:

```bash
mur team status
```

```
Team Sync Status
================

Local patterns:     12
Team patterns:      45
Pending push:       2
Pending pull:       3

Last sync: 2024-01-15 10:30:00
```

## Pattern Visibility

When adding patterns, you can mark them for team sharing:

```bash
mur learn add my-pattern
# ... during interactive prompts:
# Share with team? [y/N] y
```

Or edit patterns:

```yaml
# ~/.mur/patterns/my-pattern.yaml
name: my-pattern
team: true  # ← Share with team
domain: golang
category: pattern
content: |
  Always handle errors explicitly...
```

## Onboarding New Team Members

New developers automatically inherit team knowledge:

```bash
# New developer joins
mur init
mur team init git@github.com:myorg/team-patterns.git
mur team pull

# Now they have all team patterns!
mur learn list
# Shows 45 patterns from the team
```

## Conflict Resolution

If the same pattern exists locally and in team:

1. **Team version wins** by default
2. Local modifications are preserved in `*.local` files
3. Use `mur team pull --keep-local` to preserve local versions

## Best Practices

1. **Be selective** - Only share high-confidence, proven patterns
2. **Add context** - Include descriptions and examples
3. **Review periodically** - Prune outdated patterns
4. **Namespace** - Use consistent domain/category naming

## See Also

- [learn Command](learn.md) - Pattern management
- [Patterns Concept](../concepts/patterns.md) - How patterns work
