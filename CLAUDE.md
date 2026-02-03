# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**murmur-ai** is a unified management layer for multiple AI CLI tools. It acts as a single source of truth for configurations, learned patterns, hooks, and MCP servers, syncing them across different AI tools (Claude Code, Gemini CLI, Auggie, etc.).

Binary name: `mur`

## Development Commands

```bash
# Build the binary
go build -o mur ./cmd/mur

# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/config
go test ./cmd/mur/cmd

# Install locally
go install ./cmd/mur

# Run the binary
./mur health
./mur run -p "hello"
```

## Architecture

### Unified Configuration Model

murmur-ai centralizes configuration at `~/.murmur/` and syncs to individual AI CLI tools:

- **Source of truth**: `~/.murmur/config.yaml`, `~/.murmur/hooks.json`, `~/.murmur/mcp.json`
- **Sync targets**: `~/.claude/`, `~/.gemini/`, etc.

This means changes flow **from** murmur **to** AI tools, not the reverse.

### Code Structure

```
cmd/mur/           # CLI entrypoint
  cmd/             # Cobra commands (init, run, config, health, learn, sync)
internal/          # Core business logic
  config/          # Config management (read/write ~/.murmur/)
  health/          # Health checks for AI tool availability
  learn/           # Learning system (extract patterns from sessions)
  run/             # Tool runner (execute prompts via AI tools)
  sync/            # Sync logic (push to AI tool configs)
pkg/               # Public packages (if any)
openspec/          # OpenSpec-driven development workflow
```

### OpenSpec Workflow

This project uses **OpenSpec** for systematic change management:

1. **Before coding**: Create a change spec in `openspec/changes/XX-name.md`
2. **Implementation**: Follow the spec, reference it in commits
3. **After merge**: Move completed spec to `openspec/archive/`

The `openspec/VISION.md` contains the project roadmap. The `openspec/decisions/` directory would contain ADRs (Architecture Decision Records).

## Tech Stack

- **Language**: Go 1.25+
- **CLI Framework**: Cobra (github.com/spf13/cobra)
- **Config**: YAML (viper will be added as needed)

## Key Concepts

### Multi-Tool Runner

`mur run -p "prompt"` can execute against any AI tool. The routing logic is in `internal/run/`.

### Cross-Tool Learning

The learning system (`internal/learn/`) extracts patterns from AI coding sessions and can export them to multiple AI tools' skill/pattern directories.

### Sync Strategy

- **Hooks**: `~/.murmur/hooks.json` → each AI CLI's hooks config
- **MCP**: `~/.murmur/mcp.json` → each AI CLI's MCP server config
- **Learned patterns**: Generated skills/patterns → `~/.claude/learnings/`, etc.

## Configuration

Default tool and enabled AI tools are set in `~/.murmur/config.yaml`:

```yaml
default_tool: claude

tools:
  claude:
    enabled: true
    binary: claude
  gemini:
    enabled: true
    binary: gemini
```

Use `mur config show` and `mur config default <tool>` to view/modify.
