# Changelog

All notable changes to mur-core will be documented in this file.

## [1.8.1] - 2026-02-15

### Fixed
- `mur update` now auto-refreshes Homebrew tap before upgrade
  - Ensures users always get the latest version
  - No more "already up to date" when new version exists

## [1.8.0] - 2026-02-15

### Added
- **🌍 Community Auto-Share**: Share patterns with the community
  - `mur init` now asks about community sharing (default: enabled)
  - Auto-share on `mur sync --push` when enabled
  - Patterns help 10,000+ developers worldwide
  - Get ⭐ recognition and a public profile

- **🔐 Secret Scanning**: Automatic security protection
  - Scans patterns for API keys, passwords, tokens before sharing
  - Supports 30+ secret types (AWS, GitHub, OpenAI, Stripe, etc.)
  - Patterns with secrets are automatically blocked
  - Server-side validation for defense in depth

- **🔄 Config Migration**: Smooth upgrades between versions
  - `schema_version` tracking for config files
  - `mur update config` migrates to latest schema
  - `mur init` now merges new fields (preserves existing settings)
  - `mur init --force` for clean reinstall

### Changed
- `mur init` no longer overwrites existing config (use `--force` to override)
- Config defaults now include community sharing settings

### Fixed
- MergeConfig preserves schema version for proper migration detection

## [1.7.2] - 2026-02-15

### Added
- **🍺 Smart Update**: Auto-detect installation method
  - `mur update` now detects Homebrew vs Go installation
  - Homebrew users: runs `brew upgrade mur`
  - Go users: runs `go install ...@latest`
  - No more manual detection needed

- **🔐 Team Subscription Check**: Cloud sync validates subscription
  - Warns when team subscription is expired
  - Graceful degradation for expired teams

### Documentation
- Rewrote README for better onboarding experience
- Added demo GIF showing core workflow

## [1.7.1] - 2026-02-14

### Added
- **🦞 OpenClaw Integration**: Learn from OpenClaw conversations
  - OpenClaw transcript support (`~/.openclaw/agents/*/sessions/*.jsonl`)
  - Unified session list shows both Claude Code and OpenClaw sessions
  - Auto-detection of OpenClaw sessions

- **🏷️ Pattern Tags**: Better organization
  - LLM extracts 2-5 relevant tags per pattern
  - Tags help categorize and discover patterns

### Fixed
- `--accept-all` flag now works without interactive prompts
- Claude Code and Gemini hooks use correct `learn` command

### Changed
- Performance: replaced bubble sort with `sort.Slice`

### Documentation
- Added Community V3, Collections, and Import documentation

## [1.7.0] - 2026-02-13

### Added
- **🌍 Community V3**: Full community features
  - `mur community share <name>` — Share patterns to community
  - `mur community featured` — View featured patterns
  - `mur community user <login>` — View user profiles
  - Category and tags support for patterns
  
- **📚 Collections**: Curate pattern collections
  - `mur collection list` — List public collections
  - `mur collection show <id>` — View collection contents
  - `mur collection create <name>` — Create new collection
  - `--visibility` flag for public/private collections

- **📥 Import from Gist**: Import patterns from GitHub
  - `mur import gist <url>` — Import from GitHub Gist
  - Supports `pattern.yaml` or README + code files
  - Auto-extracts name, description, content
  - `--share` flag to share directly to community

### Changed
- **🔍 Enhanced Search**: Tech stack filtering
  - `mur config set tech_stack "swift,go,docker"`
  - Community search auto-filters by your stack

## [1.6.2] - 2026-02-13

### Added
- **🔍 Community Search**: Search community patterns from CLI
  - `mur search --community "query"` — Include community in search
  - `mur search --community-only "query"` — Search only community
  - `mur search --local "query"` — Search only local patterns
  - Tech stack filtering via `tech_stack` config

- **⚙️ Tech Stack Config**: Filter community results by your stack
  - `mur config set tech_stack "swift,go,docker"`
  - Automatically filters community search results

- **💾 Community Cache**: Local cache for injected community patterns
  - Auto-caches top patterns when using `--inject`
  - 7-day TTL, 50MB max, LRU cleanup
  - `mur config set cache.community.enabled true/false`
  - Cleanup on `mur sync` (configurable)

### Changed
- **🚀 Hybrid Search (Server)**: BM25 + keyword with RRF fusion
  - Full-text search using PostgreSQL tsvector
  - Weighted ranking: name (A) > description (B) > content (C)
  - ~84% precision target (up from ~62% ILIKE-only)

## [1.6.1] - 2026-02-13

### Changed
- **🚀 Lightweight mur-index**: Patterns now load on-demand via `mur search`
  - Skills folders only contain `mur-index/SKILL.md` (instructs AI to use search)
  - Patterns stay in `~/.mur/patterns/` and load via semantic search
  - Scales to millions of community patterns without polluting skills folders
  - 90%+ token savings compared to loading all patterns upfront

### Fixed
- **🧹 Clean up old pattern directories**: `mur clean --force` removes legacy format
  - Removes individual pattern directories from skills folders
  - Sync now auto-cleans old patterns during sync

## [1.6.0] - 2026-02-12

### Added
- **📊 Pattern Analytics**: Track pattern usage and effectiveness
  - `mur stats` — View analytics overview (usage counts, effectiveness scores)
  - `mur stats <pattern>` — Detailed stats for a specific pattern
  - `mur feedback` — Rate patterns (helpful/not helpful)
  - SQLite-based local storage (pure Go, no CGO required)

- **🔍 Claude Code Search Hooks**: Automatic pattern suggestions
  - `mur init --hooks --search` — Enable search hooks
  - Suggests relevant patterns when you submit prompts
  - Runs `mur search --inject` on UserPromptSubmit

### Changed
- **🛠️ Hooks Infrastructure**: Unified hook installation system
  - Added Claude Code hook support
  - `HookOptions` for configurable hook behavior
  - Silent fail for missing tools (doesn't break workflow)

## [1.5.1] - 2026-02-12

### Added
- **⏰ Auto-Sync**: Automatic background pattern sync
  - `mur sync auto enable` — interactive setup
  - `mur sync auto disable` — turn off
  - `mur sync auto status` — check status
  - Cross-platform: macOS (LaunchAgent), Linux (systemd), Windows (Task Scheduler)
  - Configurable intervals: 15m, 30m, 1h, 6h, 24h

### Changed
- **📖 README**: Comprehensive update with all new sync features

## [1.5.0] - 2026-02-12

### Added
- **🔀 Interactive Conflict Resolution**: When sync conflicts occur, choose how to resolve
  - `[i]` Interactive - choose for each pattern
  - `[s]` Accept all from server
  - `[l]` Accept all from local
  - `[d]` View diff between versions
  - Resolution summary after completion

## [1.4.9] - 2026-02-12

### Added
- **🧠 Smart `mur sync`**: Automatically detects plan and syncs accordingly
  - Trial/Pro/Team/Enterprise → Cloud sync with mur.run
  - Free users → Git sync (if configured)
  - Added `--cloud`, `--git`, `--cli` flags to override

## [1.4.8] - 2026-02-12

### Added
- **☁️ Cloud Status in `mur status`**: Shows login status, active team, and last sync time
- **🔼 `mur cloud push`**: New command to upload local patterns to server
- **🔽 `mur cloud pull`**: New command to download patterns from server
- **⚠️ Session Expired Detection**: Clear message when cloud session has expired

## [1.4.7] - 2026-02-11

### Changed
- **📝 Config Template**: `search` section now commented by default (requires Ollama)

## [1.4.6] - 2026-02-11

### Changed
- **📝 Simplified Config Template**: Removed advanced sections from default init
  - Removed: `server`, `team`, `mcp` (add manually when needed)
  - Kept as comments: `notifications`, `hooks`

## [1.4.5] - 2026-02-11

### Added
- **📝 Complete Config Template**: `mur init` now generates full config with all sections
  - `server`: Cloud sync (mur.run) settings
  - `team`: Git-based team sync
  - `mcp`: Model Context Protocol integration
  - `notifications`: Slack/Discord webhooks
  - `hooks`: Custom hook configuration

## [1.4.4] - 2026-02-11

### Added
- **🔍 Search Config**: `mur init` now includes search configuration section
  - Ollama embedding model settings
  - Auto-inject option for search results

## [1.4.3] - 2026-02-11

### Changed
- **📚 README**: Improved installation docs with Homebrew as primary method
- Added path difference notes for Homebrew vs Go install

## [1.4.2] - 2026-02-11

### Added
- **🍺 Homebrew Install**: `brew tap mur-run/tap && brew install mur`
- Auto-update Homebrew formula on new releases

## [1.4.1] - 2026-02-11

### Added
- **🔑 API Key Login**: `mur login --api-key mur_xxx_...` for CI/automation
  - Create API keys at https://app.mur.run/core/settings
  - API keys never expire (unless set during creation)
  - Supports both JWT and API key authentication

## [1.2.0] - 2026-02-10

### Added
- **📱 Device Tracking**: Automatic device ID generation and tracking
  - Persistent device ID stored in `~/.mur/device_id`
  - All API requests include `X-Device-ID/Name/OS` headers
  - Device limit enforcement (Free: 1, Trial/Pro: 3, Team: 5)
- **🌍 Community Patterns**: Browse and copy patterns from the community
  - `mur community` — View popular patterns
  - `mur community search <query>` — Search community
  - `mur community copy <name>` — Copy pattern to your team
  - `mur community recent` — View recent patterns
- **📊 Referral System**: View referral stats and share link
  - `mur referral` — View referral status and share link
- **📱 Device Management**: List and manage connected devices
  - `mur devices` — List all registered devices
  - `mur devices logout <name>` — Force logout a device
- **🎮 Auggie Hooks**: Full hook support for Augment CLI (SessionStart, Stop events)
  - `mur init --hooks` now configures Auggie alongside Claude Code and Gemini CLI

### Changed
- API client now sends device headers with all authenticated requests
- 429 errors return structured `DeviceLimitError` with active device list
- Auggie moved from "static sync" to "hooks supported" in README

## [1.1.0] - 2026-02-09

### Added
- **🔍 Semantic Search**: `mur search` finds patterns by meaning using embeddings
- **🤖 OpenAI Embeddings**: Use `text-embedding-3-small/large` when no local GPU
- **📁 Directory Sync Format**: 90%+ token savings with individual skill directories
- **📊 Embedding Index**: `mur index status/rebuild` for managing vector embeddings
- **🔗 Search Hooks**: Auto-suggest relevant patterns in Claude Code prompts
- **🔄 Pattern Migration**: `mur migrate` upgrades patterns to v2 schema
- **📈 Analytics Tracking**: `mur analytics` tracks pattern usage and effectiveness

### New Commands
- `mur search <query>` — Semantic pattern search
- `mur search --inject` — Hook mode for auto-suggestions
- `mur index status` — Check embedding index health
- `mur index rebuild` — Rebuild all embeddings
- `mur migrate --dry-run` — Preview pattern migration
- `mur analytics` — View pattern usage summary
- `mur analytics top` — Show most used patterns
- `mur analytics cold` — Show patterns not used recently
- `mur analytics feedback` — Record pattern helpfulness

### Changed
- `mur sync` now defaults to directory format (individual skill folders)
- `mur init --hooks` adds semantic search hook when enabled
- Pattern schema v2 adds version, resources, and embedding_hash fields
- Default `min_score` lowered to 0.5 for better nomic-embed-text compatibility

### Configuration
```yaml
search:
  enabled: true
  provider: ollama
  model: nomic-embed-text
  auto_inject: true  # auto-suggest in hooks
  
sync:
  format: directory  # new default
  prefix_domain: true  # swift--pattern-name format
```

### Migration
```bash
# Upgrade patterns to v2 schema
mur migrate patterns --to 2

# Rebuild for semantic search
mur index rebuild

# Re-install hooks with search support
mur init --hooks
```

---

## [1.0.1] - 2026-02-08

### Added
- **Cloud sync** - `mur cloud login/logout/sync` commands for mur-server integration
- Team pattern sharing via cloud sync

### Fixed
- **Infinite recursion in token refresh** - Fixed issue where auth refresh could loop infinitely
- **Empty changes slice** - Cloud sync now correctly returns empty slice instead of nil

---

## [1.0.0] - 2026-02-08 🎉

**mur is ready for production!**

### Highlights
- **10 AI Tools Supported** — Claude, Gemini, Codex, Auggie, Aider, OpenCode, Continue, Cursor, Windsurf, GitHub Copilot
- **4 Hook Integrations** — Claude Code, Gemini CLI, OpenCode, GitHub Copilot
- **LLM-Powered Extraction** — Smart pattern extraction with Ollama, OpenAI, Gemini, Claude
- **Premium Model Routing** — Use better models for important sessions

### What's New in 1.0
- OpenCode and GitHub Copilot sync support
- OpenCode and GitHub Copilot hooks support  
- Terminal screenshots in README
- Complete documentation

### Since 0.9.x
- LLM-based pattern extraction (`mur learn extract --llm`)
- Premium model routing with min_messages and projects rules
- Smart LLM fallback (Config → Ollama → keyword)
- Doctor shows all hook and LLM status
- 56+ learned patterns

---

## [0.9.8] - 2026-02-08

### Fixed
- **API key documentation** - Clarified that `api_key_env` takes the variable NAME, not the key value itself

### Changed
- Improved error messages for missing API keys
- Doctor now shows clearer status for premium LLM configuration

---

## [0.9.7] - 2026-02-08

### Added
- **Doctor premium LLM display** - `mur doctor` now shows Premium LLM and routing configuration status
- Shows ✓ when API key is available, or "(missing ENV_VAR)" when not configured

---

## [0.9.6] - 2026-02-08

### Added
- **Premium model routing** - Use different models for important sessions
  - Configure `llm.premium` for high-quality model
  - Set `llm.routing.min_messages` or `llm.routing.projects` rules
  - Sessions matching rules use premium model automatically
- **Complete config template** - All LLM options documented and commented out
- **API key environment variables** - Configure `api_key_env` for each provider

### Changed
- Config defaults are now commented out (user enables by uncommenting)
- Hook uses `--llm` flag for better extraction quality

---

## [0.9.4] - 2026-02-08

### Added
- **Remote Ollama docs** - LAN setup for running Ollama on a server
- **Recommended models table** - Best models for each provider

---

## [0.9.3] - 2026-02-08

### Added
- **Smart LLM fallback** - Auto-detects Ollama if no LLM configured
- **Doctor LLM check** - Shows LLM extraction configuration status
- Hook now uses `--llm` for better extraction quality

### Changed
- Extraction falls back gracefully: Config LLM → Ollama → Keyword (with warning)

---

## [0.9.2] - 2026-02-08

### Added
- **OpenAI and Gemini providers** for LLM extraction
  - `--llm openai` for OpenAI-compatible APIs (OpenAI, Groq, Together, etc.)
  - `--llm gemini` for Google Gemini API
  - Configurable endpoint URL for custom OpenAI-compatible services

---

## [0.9.1] - 2026-02-08

### Added
- **LLM-based extraction** - `mur learn extract --llm` uses AI for smart pattern extraction
  - Supports Ollama (local, free) and Claude API
  - Much better quality than keyword-based extraction
  - Example: `mur learn extract --llm ollama --llm-model deepseek-r1:8b`

---

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
