# Command Reference

Complete list of MUR commands.

## Setup & Status

| Command | Description |
|---------|-------------|
| `mur init` | Interactive setup wizard |
| `mur init --hooks` | Quick setup with CLI hooks |
| `mur status` | Overview of patterns, sync, cloud status |
| `mur doctor` | Diagnose and fix issues |
| `mur version` | Show version |
| `mur update` | Update MUR (auto-detects Homebrew vs Go) |

## Pattern Management

| Command | Description |
|---------|-------------|
| `mur new <name>` | Create new pattern |
| `mur edit <name>` | Edit pattern in $EDITOR |
| `mur copy <name>` | Copy pattern content to clipboard |
| `mur examples` | Install example patterns |
| `mur migrate` | Migrate patterns to v2 schema |
| `mur export` | Export patterns to file |
| `mur import <file>` | Import patterns from file or URL |
| `mur import gist <url>` | Import from GitHub Gist |

## Sync

| Command | Description |
|---------|-------------|
| `mur sync` | Smart sync (cloud or git based on plan) |
| `mur sync --cloud` | Force cloud sync |
| `mur sync --git` | Force git sync |
| `mur sync --cli` | Only sync to local AI tools |
| `mur sync auto enable` | Enable background auto-sync |
| `mur sync auto disable` | Disable auto-sync |
| `mur sync auto status` | Check auto-sync status |

## Cloud (mur.run)

| Command | Description |
|---------|-------------|
| `mur login` | Login via OAuth (opens browser) |
| `mur login --api-key <key>` | Login with API key |
| `mur logout` | Logout |
| `mur whoami` | Show current user |
| `mur cloud teams` | List your teams |
| `mur cloud select <team>` | Set active team |
| `mur cloud sync` | Bidirectional cloud sync |
| `mur cloud push` | Push to server |
| `mur cloud pull` | Pull from server |
| `mur cloud pull --force` | Pull and overwrite local |

## Semantic Search

| Command | Description |
|---------|-------------|
| `mur search <query>` | Search patterns by meaning |
| `mur search --json <query>` | JSON output |
| `mur index status` | Check embedding index status |
| `mur index rebuild` | Rebuild all embeddings |

## Learning

| Command | Description |
|---------|-------------|
| `mur transcripts` | Browse Claude Code sessions |
| `mur transcripts --list` | List recent sessions |
| `mur learn extract` | Extract patterns from sessions |
| `mur learn extract --llm` | Use LLM for extraction |
| `mur learn extract --auto` | Auto-extract high-confidence |

## Community

| Command | Description |
|---------|-------------|
| `mur community` | Browse popular patterns |
| `mur community search <query>` | Search community |
| `mur community copy <name>` | Copy pattern locally |
| `mur community share <name>` | Share your pattern |
| `mur community featured` | View featured patterns |
| `mur community user <login>` | View user profile |

## Collections

| Command | Description |
|---------|-------------|
| `mur collection list` | List public collections |
| `mur collection show <id>` | View collection details |
| `mur collection create <name>` | Create new collection |

## Dashboard & Stats

| Command | Description |
|---------|-------------|
| `mur serve` | Start web dashboard (localhost:8080) |
| `mur dashboard` | Generate static HTML report |
| `mur dashboard -o report.html` | Save report to file |
| `mur stats` | View usage statistics |

## Configuration

| Command | Description |
|---------|-------------|
| `mur config` | View current config |
| `mur config edit` | Edit config in $EDITOR |
| `mur config path` | Show config file path |

## Maintenance

| Command | Description |
|---------|-------------|
| `mur clean` | Cleanup old/temp files |
| `mur clean --dry-run` | Show what would be cleaned |

## Help

| Command | Description |
|---------|-------------|
| `mur --help` | Show help |
| `mur <command> --help` | Help for specific command |
| `mur web` | Open docs in browser |
| `mur web github` | Open GitHub repo |

## Command Tree

```
mur
├── init [--hooks]
├── status
├── doctor
├── version
├── update
├── sync [--cloud|--git|--cli]
│   └── auto [enable|disable|status]
├── cloud
│   ├── teams
│   ├── select <team>
│   ├── sync
│   ├── push
│   └── pull [--force]
├── new <name>
├── edit <name>
├── copy <name>
├── search <query> [--json]
├── index [status|rebuild]
├── examples
├── migrate
├── export
├── import <file>
│   └── gist <url>
├── transcripts [--list]
├── learn
│   └── extract [--llm] [--auto]
├── community [search|copy|share|featured|user]
├── collection [list|show|create]
├── serve
├── dashboard [-o file]
├── stats
├── config [edit|path]
├── clean [--dry-run]
├── login [--api-key]
├── logout
├── whoami
└── web [github]
```
