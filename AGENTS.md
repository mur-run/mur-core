# AGENTS.md

## Project Overview

**murmur-ai** - Multi-AI CLI unified management layer + cross-tool learning system.

Binary name: `mur`

## Tech Stack

- **Language:** Go 1.25+
- **CLI Framework:** Cobra
- **Config:** YAML (viper)

## Project Structure

```
cmd/mur/           # CLI entrypoint
  cmd/             # Cobra commands
    root.go        # Root command
    init.go        # mur init
    run.go         # mur run
    config.go      # mur config
    health.go      # mur health  
    learn.go       # mur learn
    sync.go        # mur sync
internal/          # Internal packages
  config/          # Config management
  health/          # Health checks
  learn/           # Learning system
  run/             # Tool runner
  sync/            # Sync to tools
pkg/               # Public packages (if any)
openspec/          # OpenSpec workflow
  VISION.md        # Project vision
  changes/         # Change specs
  archive/         # Completed specs
  decisions/       # ADRs
```

## Development Workflow

We use **OpenSpec** for changes:

1. Create change spec in `openspec/changes/XX-name.md`
2. Implement following the spec
3. `go build ./...` + `go test ./...`
4. After merge, move spec to `openspec/archive/`

## Commands

```bash
# Build
go build -o mur ./cmd/mur

# Test
go test ./...

# Install locally
go install ./cmd/mur

# Run
./mur health
./mur run -p "hello"
```

## Key Decisions

- Config lives at `~/.murmur/`
- Source of truth for patterns/hooks/MCP
- Syncs to individual CLI configs (Claude, Gemini, etc.)
