# Changelog

All notable changes to mur-core will be documented in this file.

## [0.9.0] - 2026-02-08

### Fixed
- **JSON Pattern Extraction** - Claude's JSON pattern arrays now parse correctly into individual pattern files instead of saving as one blob (#1)
- **Doctor CLI Detection** - Now finds AI CLIs installed outside PATH (e.g., `~/.npm-global/bin/`, `~/go/bin/`)

### Added
- Pattern name validation to filter low-quality extractions (now-need-update, wait-changed-line, etc.)
- Homebrew formula (`Formula/mur.rb`) for easier installation
- Unit tests for JSON extraction and pattern name validation

### Changed
- README overhauled with better installation docs, feature highlights, and command reference

---

## [0.8.0] - 2026-02-07

### Added

#### Dashboard & Visualization
- `mur serve` - Enhanced web dashboard with charts, sync status, modals
- `mur dashboard` - Generate static HTML report
- `mur dashboard -o file.html` - Save to file
- `mur dashboard --open` - Generate and open in browser

#### Status & Diagnostics
- `mur status` - Quick terminal overview of patterns, sync, stats, hooks
- `mur status --verbose` - Detailed breakdown
- `mur doctor` - Diagnose setup issues and common problems
- `mur doctor --fix` - Auto-fix issues where possible

#### Pattern Management
- `mur new <name>` - Create pattern from template with auto-inferred tags
- `mur edit <name>` - Open pattern in $EDITOR
- `mur search <query>` - Search patterns by name, description, tags
- `mur search --tag backend` - Filter by tag
- `mur transcripts` - Browse Claude Code session transcripts
- `mur transcripts --project X` - Filter by project
- `mur transcripts show <id>` - View session content

#### Import/Export
- `mur import file.yaml` - Import patterns from local files
- `mur import https://...` - Import from URLs
- `mur import *.yaml` - Glob pattern support
- `mur import --dry-run` - Preview without saving
- `mur import --force` - Overwrite existing patterns

#### Version
- `mur version` - Show version, commit, build info
- Version now displayed in dashboard

### Changed
- Dashboard shows all 8 sync targets with status
- Dashboard includes tool usage breakdown and cost savings
- README updated with all new commands

## [0.7.0] - 2026-02-07

### Added

#### Context-Aware Pattern Injection for Native CLIs
- `mur context` - Output relevant patterns for current project (used by hooks)
- Claude Code UserPromptSubmit hook now injects context-aware patterns
- Gemini CLI BeforeAgent hook now injects context-aware patterns
- Works with `claude` and `gemini` directly - no need for `mur run`
- Detects project type (Go, Swift, Python, Node.js) and matches patterns

#### IDE Integration (Static Rules)
- Continue: `~/.continue/rules/mur-patterns.md`
- Cursor: `~/.cursor/rules/mur-patterns.md`
- Windsurf: `~/.windsurf/rules/mur-patterns.md`
- `mur sync` now syncs to all 8 targets (5 CLIs + 3 IDEs)

#### Automatic Pattern Extraction
- `mur learn extract --accept-all` - Auto-save patterns above confidence threshold
- `mur learn extract --quiet` - Silent mode for hooks
- `mur learn extract --min-confidence 0.7` - Set custom threshold
- Session-end hook now auto-extracts high-confidence patterns
- Patterns are automatically synced to CLIs after extraction

#### Improved Pattern Recommendation
- `mur run` now uses semantic search when available
- Auto-indexes patterns in background when cache is stale
- Better context-aware pattern matching

#### Silent Mode
- `mur sync --quiet` - Silent mode for hooks

### Changed
- Session-end hook now runs `mur learn extract --auto --accept-all --quiet`
- Semantic search auto-initializes when embeddings are available
- `mur init` now suggests setting up semantic search
- UserPromptSubmit hook now runs `mur context` for dynamic pattern injection

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
