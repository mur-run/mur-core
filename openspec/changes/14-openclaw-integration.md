# 14: OpenClaw Integration

## Summary

Add integration support for OpenClaw, enabling murmur-ai to be used as a skill within OpenClaw's agent framework.

## Motivation

OpenClaw is an agentic framework that supports custom skills. By providing a murmur-ai skill, OpenClaw users can:
- Sync AI tool configurations across all their CLI tools
- Extract learned patterns from coding sessions
- View usage statistics
- Leverage murmur-ai's unified management layer

## Changes

### 1. New `mur output` Command

Add a new command for programmatic JSON output:

```bash
mur output --json          # Full status dump
mur output sync --json     # Sync results as JSON
mur output stats --json    # Stats as JSON
mur output learn --json    # Patterns as JSON
```

This provides structured output for automation tools.

### 2. OpenClaw Skill Wrapper

Create `integrations/openclaw/` directory with:
- `SKILL.md` — Skill documentation for OpenClaw
- `package.json` — OpenClaw skill metadata
- `README.md` — Integration guide

### 3. Skill Commands

The OpenClaw skill exposes:
- `/murmur:sync` — Trigger `mur sync all`
- `/murmur:learn` — Trigger `mur learn extract --auto --dry-run`
- `/murmur:stats` — Display `mur stats --json`

## Implementation

### File: `cmd/mur/cmd/output.go`

New command that wraps existing commands with JSON output mode.

### File: `integrations/openclaw/SKILL.md`

```markdown
# Murmur AI Skill

Unified management for AI CLI tools.

## Commands
- /murmur:sync — Sync MCP, hooks, and patterns to all AI tools
- /murmur:learn — Extract patterns from recent coding sessions
- /murmur:stats — Show usage statistics
```

## Testing

- `mur output --json` returns valid JSON
- `mur output stats --json` matches `mur stats --json`
- OpenClaw can invoke skill commands successfully

## Acceptance Criteria

- [ ] `mur output --json` works
- [ ] `integrations/openclaw/` exists with SKILL.md, package.json, README.md
- [ ] `go build ./...` succeeds
- [ ] Documentation updated
