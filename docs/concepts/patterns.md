# Patterns

Patterns are reusable knowledge snippets that murmur learns from your coding sessions and shares across AI tools.

## What Are Patterns?

When you work with AI coding assistants, you teach them things:

- "Always use `context.Context` as the first parameter in Go"
- "Our API responses follow this JSON structure"
- "We use this error handling pattern"

These learnings are **patterns**. Murmur captures them so:

1. You don't repeat yourself across sessions
2. All AI tools (Claude, Gemini, etc.) benefit
3. Your team inherits your experience

## How Patterns Work

```
Your Session                    Pattern Store                AI Tools
┌──────────────┐               ┌─────────────┐             ┌─────────────┐
│ "Always wrap │   extract     │ patterns/   │    sync     │ Claude      │
│  errors with │ ────────────► │ error-      │ ──────────► │ Gemini      │
│  context"    │               │ handling.md │             │ Auggie      │
└──────────────┘               └─────────────┘             └─────────────┘
```

## Pattern Structure

Each pattern has:

```yaml
name: error-handling-go
description: Wrap errors with context for better debugging
domain: golang
category: pattern
confidence: 0.85
created_at: 2024-01-15T10:30:00Z
updated_at: 2024-01-15T10:30:00Z
content: |
  When handling errors in Go, always wrap them with context:
  
  ```go
  if err != nil {
      return fmt.Errorf("failed to process user %s: %w", userID, err)
  }
  ```
  
  This makes debugging easier by showing the call chain.
```

## Creating Patterns

### Manual Creation

```bash
mur learn add error-handling-go
```

Interactive prompts guide you through:

- Name, description
- Domain (golang, python, api, etc.)
- Category (pattern, convention, antipattern)
- Confidence level (0.0-1.0)
- Content

### Automatic Extraction

Let murmur find patterns in your sessions:

```bash
mur learn extract --auto
```

Murmur analyzes your Claude Code session transcripts and identifies:

- Repeated corrections you made
- Coding standards you enforced
- Explanations you provided

## Pattern Domains

Organize patterns by technology area:

| Domain | Examples |
|--------|----------|
| `golang` | Error handling, concurrency patterns |
| `python` | Type hints, async patterns |
| `javascript` | Module patterns, React conventions |
| `typescript` | Type definitions, generics usage |
| `api` | REST conventions, response formats |
| `database` | Query patterns, schema conventions |
| `testing` | Test structure, mocking patterns |
| `security` | Auth patterns, input validation |
| `architecture` | Design patterns, module structure |
| `general` | Cross-cutting concerns |

## Pattern Categories

| Category | Purpose |
|----------|---------|
| `pattern` | Best practices to follow |
| `convention` | Style and naming conventions |
| `antipattern` | Things to avoid |
| `snippet` | Code templates |
| `explanation` | Concept explanations |

## Confidence Levels

Confidence indicates how reliable the pattern is:

| Level | Meaning |
|-------|---------|
| 0.9-1.0 | Well-established, always apply |
| 0.7-0.9 | Generally good, occasional exceptions |
| 0.5-0.7 | Useful but context-dependent |
| 0.3-0.5 | Experimental, use with caution |
| 0.0-0.3 | Needs validation |

## Syncing Patterns

Patterns are injected into AI tool instructions:

```bash
mur learn sync
```

This updates each tool's configuration with your patterns, so when you ask:

> "Write a function to fetch user data"

The AI knows your conventions without you having to explain them again.

## Pattern Storage

Patterns are stored in `~/.murmur/patterns/`:

```
~/.murmur/patterns/
├── error-handling-go.yaml
├── api-response-format.yaml
├── test-naming.yaml
└── .synced/
    ├── claude.json
    └── gemini.json
```

## Best Practices

1. **Be specific** - "Use fmt.Errorf with %w" is better than "handle errors properly"
2. **Include examples** - Show code, not just descriptions
3. **Set realistic confidence** - Don't mark everything as 1.0
4. **Update over time** - Patterns evolve; keep them current
5. **Prune regularly** - Remove patterns that no longer apply

## See Also

- [learn Command](../commands/learn.md) - Pattern management commands
- [team Command](../commands/team.md) - Sharing patterns with your team
- [Cross-CLI Sync](cross-cli-sync.md) - How patterns sync across tools
