# OpenClaw Integration for Murmur AI

This directory contains the OpenClaw skill wrapper for murmur-ai, enabling OpenClaw agents to manage AI CLI tool configurations.

## Installation

### 1. Install murmur-ai

```bash
# Using go install
go install github.com/mur-run/mur-core/cmd/mur@latest

# Or build from source
git clone https://github.com/mur-run/mur-core.git
cd murmur-ai
go build -o mur ./cmd/mur
mv mur /usr/local/bin/
```

### 2. Configure murmur-ai

```bash
mur init
```

This creates `~/.mur/config.yaml` with default settings.

### 3. Register with OpenClaw

Copy or symlink this directory to your OpenClaw skills folder, or reference it in your OpenClaw configuration.

## Usage

Once installed, OpenClaw can use these commands:

| Command | Description |
|---------|-------------|
| `/murmur:sync` | Sync all configurations to AI CLI tools |
| `/murmur:learn` | Extract patterns from coding sessions |
| `/murmur:stats` | Show usage statistics |

## Files

- `SKILL.md` — Skill documentation (read by OpenClaw)
- `package.json` — Skill metadata and command definitions
- `README.md` — This file

## Manual Testing

You can test the underlying commands directly:

```bash
# Test sync
mur output sync --json

# Test learn
mur learn extract --auto --dry-run

# Test stats
mur output stats --json
```

## Troubleshooting

### "mur: command not found"

Ensure `mur` is in your PATH:

```bash
which mur
# Should output: /usr/local/bin/mur or similar
```

### No statistics showing

Statistics are recorded when you use `mur run`. Try:

```bash
mur run claude "hello world"
mur stats
```

### Sync fails

Check your murmur config:

```bash
cat ~/.mur/config.yaml
```

Ensure the AI CLI tools are installed and their config paths are correct.

## Contributing

Issues and PRs welcome at [murmur-ai](https://github.com/mur-run/mur-core).
