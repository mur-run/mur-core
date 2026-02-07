# Changelog

All notable changes to mur-core will be documented in this file.

## [0.6.0] - 2026-02-07

### Added

#### Learning Repository
- `mur repo set <url>` - Set learning repository
- `mur repo status` - Show repository status
- `mur repo remove` - Remove repository configuration
- `mur init` now asks about learning repo during setup
- Patterns can be stored in a git repo for sync across machines

#### Simplified Commands
- `mur sync` now pulls from repo + syncs to CLIs
- `mur sync --push` pushes local changes to remote
- Hidden advanced commands for cleaner help output

### Changed
- `mur init` is now fully interactive with repo setup
- Simplified `mur --help` to show only essential commands
- Renamed config directory: `~/.murmur` → `~/.mur`

## [0.5.0] - 2026-02-07

### Added

#### Export & Inject Commands
- `mur export` - Export patterns to YAML, JSON, or Markdown format
- `mur export --format json/md` - Choose output format
- `mur export --tag <tag>` - Filter by tag
- `mur export --min-effectiveness <score>` - Filter by effectiveness
- `mur inject <dir>` - Inject patterns into project CLAUDE.md/AGENTS.md
- `mur inject --file AGENTS.md` - Use different target file
- `mur inject --dry-run` - Preview without writing

#### Claude Code Hooks
- `mur init --hooks` - One-command Claude Code hooks installation
- Automatically creates `~/.mur/hooks/` with learning prompts
- Merges hooks into `~/.claude/settings.json`
- Backs up existing settings before modification

### Changed
- **Renamed from mur-cli to mur-core** - Unified naming across the ecosystem
- Updated all import paths and documentation

## [0.4.0] - 2026-02-07

### Added

#### Pattern Auto-Injection
- `mur run` now automatically injects relevant patterns based on context
- Project detection for Go, Swift, Node.js, and Python projects
- `--no-inject` flag to disable pattern injection
- `--verbose` flag to show injection details

#### Effectiveness Tracking
- Track pattern usage and success rates automatically
- `mur feedback <rating> <pattern>` - rate pattern helpfulness
- `mur pattern-stats` - view pattern effectiveness statistics
- Automatic effectiveness score updates from tracking data

#### Semantic Search
- `mur embed index` - index patterns for semantic search
- `mur embed search <query>` - test semantic matching
- `mur embed status` - view embedding status
- `mur embed rehash` - rebuild all embeddings
- Support for Ollama (local) and OpenAI embedding providers
- Automatic fallback to keyword matching when embeddings unavailable

#### Lifecycle Management
- `mur lifecycle evaluate` - check patterns for deprecation
- `mur lifecycle apply` - apply recommended lifecycle changes
- `mur lifecycle deprecate/archive/reactivate` - manual control
- `mur lifecycle list` - view patterns by status
- `mur lifecycle cleanup` - delete old archived patterns
- Auto-deprecate patterns with <30% effectiveness
- Auto-archive patterns with <10% effectiveness
- Stale pattern detection (>90 days unused)

### Changed
- Pattern injection integrated into `mur run` command
- Improved routing decision output with pattern info

## [0.3.0] - 2026-02-04

### Added
- Pattern Schema v2 with multi-dimensional tags
- Security sanitizer for prompt injection detection
- Migration tool for v1 → v2 pattern conversion
- `mur lint` for pattern validation
- `mur migrate` for schema migration

## [0.2.0] - 2026-02-03

### Added
- Team sync functionality
- Pattern learning from sessions
- Notification support (Slack, Discord)

## [0.1.0] - 2026-02-02

### Added
- Initial release
- Multi-AI CLI routing (Claude, Gemini, Auggie)
- Automatic complexity-based routing
- Usage statistics tracking
- Configuration management
