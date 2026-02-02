# murmur-ai

A Clawdbot skill for driving the local [Claude Code](https://code.claude.com/) CLI in headless mode, with **continuous learning** via Claude Code Hooks.

## What it does

- **Run Claude Code** headlessly (`claude -p`) with PTY allocation and permission control
- **Learn in real-time** — Claude Code hooks prompt Claude to save non-obvious discoveries as structured patterns
- **3D categorization** — Patterns organized by Domain × Category × Project
- **Auto-sync** — Mature patterns automatically become Claude Code native skills (`~/.claude/skills/`)

## File Structure

```
skills/murmur-ai/
├── SKILL.md                        # Skill definition (triggers, docs, examples)
├── README.md                       # This file
├── hooks/
│   ├── claude-code-hooks.json      # Claude Code hooks config (auto-merged by install.sh)
│   └── on-prompt-reminder.md       # Injected reminder for real-time learning
├── scripts/
│   ├── claude_code_run.sh          # Shell wrapper with PTY + flag forwarding
│   ├── extract_patterns.sh         # Extract patterns from transcripts (manual/legacy)
│   ├── review_learnings.sh         # Review and consolidate patterns
│   ├── auto_learn.sh              # Automated pipeline: review + sync + commit
│   ├── inject_claude_md.sh         # Inject rules into project CLAUDE.md
│   ├── export_instincts.sh         # Export patterns (md/json)
│   ├── on_session_stop.sh          # Stop hook (lightweight, < 100ms)
│   ├── sync_to_claude_code.sh      # Sync mature patterns to ~/.claude/skills/
│   ├── status.sh                   # Show pattern statistics and status
│   ├── test.sh                     # Smoke tests for the skill
│   ├── merge_team.sh              # Merge team branches (supports --strategy)
│   ├── privacy_filter.sh          # Privacy filtering (pipe or --check mode)
│   ├── validate.sh                # Pattern file validation (--fix, --json)
│   └── spec_report.sh            # Spec-driven development learning report
├── templates/
│   ├── pattern-template.md         # Pattern record template
│   └── skill-template.md           # Auto-generated skill template
└── learned/                        # 3D pattern storage
    ├── _global/                    # Cross-domain patterns
    │   └── debug/                  # Example: node-heap-oom-debug.md
    ├── devops/                     # DevOps patterns
    ├── web/                        # Frontend/web patterns
    ├── mobile/                     # Mobile dev patterns
    ├── backend/                    # Backend/API patterns
    │   └── pattern/                # Example: postgres-advisory-lock.md
    ├── data/                       # Data engineering patterns
    └── projects/                   # Project-specific patterns
        └── {project-name}/
```

## Prerequisites

- Claude Code CLI installed and authenticated (`claude --version`)
- macOS or Linux
- Python 3 (for JSON parsing in some scripts)

## Setup: Install Hooks

The key feature is **real-time learning via Claude Code hooks**. The installer supports two modes:

### Solo Mode (Default)

```bash
./setup/install.sh
```

This will:
1. Ensure `learned/` directory structure exists
2. **Smart-merge** hooks into `~/.claude/settings.json` (backs up existing settings)
3. Update hook paths to use the correct absolute repo path
4. Optionally set up auto-learn cron (daily at 3am)
5. Patterns go on the current branch (usually main)

### Team Mode

```bash
./setup/install.sh --team
```

Everything solo does, plus:
1. Create your personal branch (`learnings/{name}`)
2. Create personal pattern directory (`learned/personal/{name}/`)
3. Optionally set up auto-push cron (every 30 min) and auto-pull cron (every hour)

### Upgrading Solo → Team

```bash
./setup/install.sh --team
```

The installer is idempotent — already-installed steps are skipped. Running `--team` after solo only adds the team-specific parts (branch, personal dir, push/pull cron). Your existing patterns on main become the team's shared knowledge base.

### Manual Setup (if you prefer)

View the hooks config and merge manually:

```bash
cat ~/clawd/skills/murmur-ai/hooks/claude-code-hooks.json
# Then add the "hooks" section to ~/.claude/settings.json
```

### What the hooks do

| Hook | Purpose | Performance |
|------|---------|-------------|
| **UserPromptSubmit** | Reminds Claude to save non-obvious patterns | Instant (cat file) |
| **Stop** | Detects new pattern files, logs to stderr | < 100ms |

## How It Works

### Real-time Learning Flow

1. You start a Claude Code session
2. On every prompt, the hook reminds Claude about pattern saving
3. When Claude discovers something non-obvious, it creates a `.md` file in `learned/{domain}/{category}/`
4. The Stop hook detects new files
5. Periodically, `auto_learn.sh` reviews, merges, and syncs mature patterns

### 3D Pattern Organization

Patterns have three dimensions:

- **Domain**: `_global`, `devops`, `web`, `mobile`, `backend`, `data`
- **Category**: `debug`, `pattern`, `style`, `security`, `performance`, `ops`
- **Project**: Optional project-specific scope (e.g., `twdd-api`)

### Pattern Lifecycle

```
Discover → Save → Reinforce → Review → Sync to ~/.claude/skills/
  (hook)   (file)  (times_seen++)  (merge/archive)  (native skill)
```

A pattern becomes a native Claude Code skill when:
- **Confidence**: HIGH
- **Times seen**: ≥ 3

## Usage

### Running Claude Code

```bash
# Read-only analysis
./scripts/claude_code_run.sh -p "Summarize this project." --permission-mode plan

# Fix bugs
./scripts/claude_code_run.sh -p "Fix auth bug" --permission-mode acceptEdits --allowedTools "Read,Edit,Bash(npm test)"
```

### Managing Patterns

```bash
# Review all patterns with stats
./scripts/review_learnings.sh --report

# Review by domain or project
./scripts/review_learnings.sh --context devops --report
./scripts/review_learnings.sh --project twdd-api --report

# Auto-review with AI analysis
./scripts/review_learnings.sh --auto
```

### Syncing to Claude Code

```bash
# Preview what would be synced
./scripts/sync_to_claude_code.sh --dry-run

# Sync mature patterns (HIGH + seen ≥3)
./scripts/sync_to_claude_code.sh

# Force sync all HIGH confidence
./scripts/sync_to_claude_code.sh --force

# Clean orphaned synced skills
./scripts/sync_to_claude_code.sh --clean
```

### Injecting into Projects

```bash
# Inject all HIGH patterns
./scripts/inject_claude_md.sh --target ~/Projects/myapp

# Only relevant domains + _global
./scripts/inject_claude_md.sh --target ~/Projects/myapp --contexts devops,backend

# Include project-specific patterns
./scripts/inject_claude_md.sh --target ~/Projects/myapp --project twdd-api
```

### Exporting

```bash
# All patterns as markdown
./scripts/export_instincts.sh

# HIGH confidence only, as JSON
./scripts/export_instincts.sh --min-confidence HIGH --format json -o export.json

# Filter by domain or project
./scripts/export_instincts.sh --context devops
./scripts/export_instincts.sh --project twdd-api
```

### Pattern Status

```bash
# Human-readable summary
./scripts/status.sh

# Machine-readable JSON
./scripts/status.sh --json
```

Shows: total patterns, confidence breakdown, domain breakdown, top-5 most-seen, promotion candidates, last sync/auto-learn times.

### Spec-Driven Development Report

```bash
# Full report
./scripts/spec_report.sh

# For a specific project
./scripts/spec_report.sh --project ~/Projects/myapp

# Machine-readable JSON output
./scripts/spec_report.sh --json
```

Shows: spec file overview (archived vs active, latest date), pattern extraction stats by confidence and domain, learning pipeline comparison (spec-originated vs coding-originated patterns), promotion candidates, and tool installation status.

### Smoke Tests

```bash
./scripts/test.sh
```

Validates: script --help flags, --dry-run support, directory structure, template frontmatter, hooks JSON validity.

### Manual Extraction (Legacy)

With hooks installed, extraction is automatic. For manual use:

```bash
echo "conversation text" | ./scripts/extract_patterns.sh
./scripts/extract_patterns.sh -f transcript.txt --context devops --category debug
./scripts/extract_patterns.sh -f transcript.txt --project twdd-api
```

## Automation (Cron)

```bash
# Daily: review + sync + commit (hooks handle extraction)
./scripts/auto_learn.sh

# With legacy extraction from memory files
./scripts/auto_learn.sh --days 7

# Preview
./scripts/auto_learn.sh --dry-run
```

Cron setup:
```bash
# Via Clawdbot
clawdbot cron add --schedule "0 3 * * *" --command "cd ~/clawd/skills/murmur-ai && ./scripts/auto_learn.sh"
```

## Team Collaboration (Hub + Spoke)

This skill supports multi-person collaboration using a Hub + Spoke branching model.

### Architecture

```
        ┌─────────────────┐
        │   main branch   │  ← Hub: stable, merged patterns
        │  (shared truth) │
        └────────┬────────┘
           ┌─────┼─────┐
           ▼     ▼     ▼
     learnings/ learnings/ learnings/
      alice     bob       carol       ← Spokes: personal branches
```

### Quick Start (Each Team Member)

```bash
cd ~/clawd/skills/murmur-ai
./setup/install.sh
```

This will:
1. Create your personal branch (`learnings/{name}`)
2. Set up personal pattern directory (`learned/personal/{name}/`)
3. Install Claude Code hooks
4. Show cron setup for auto-push/pull

### Admin Operations (Hub Manager)

```bash
# Merge all team branches into main
./scripts/merge_team.sh

# Preview without merging
./scripts/merge_team.sh --dry-run

# Use manual conflict resolution instead of auto-resolve
./scripts/merge_team.sh --strategy manual

# Generate team learning report
./scripts/team_report.sh
./scripts/team_report.sh --days 7
```

### Conflict Resolution

The `merge_team.sh` script supports two conflict resolution strategies via `--strategy`:

| Strategy | Behavior | When to Use |
|----------|----------|-------------|
| `theirs` (default) | Incoming branch wins | Safe for additive pattern files — each pattern is a separate file, conflicts are rare |
| `manual` | Stops on conflict, asks admin to resolve | When multiple people edit the same pattern file, or you need full control |

**Why "theirs" is the default:** Pattern files are additive and self-contained. If two people modify the same pattern, the newer version (from the incoming branch) is usually more complete (updated `times_seen`, `last_seen`, etc.). Worst case, you lose a minor edit that's easily recoverable.

**Recovering from bad merges:**
```bash
# Find the commit before the merge
git reflog

# Undo the merge
git reset --hard <commit-before-merge>

# Force push (with safety)
git push --force-with-lease origin main
```

### Branch Convention

| Branch | Purpose | Who pushes |
|--------|---------|------------|
| `main` | Stable, reviewed patterns | Admin via `merge_team.sh` |
| `learnings/{name}` | Individual's real-time patterns | Each member (auto or manual) |

### Privacy Enforcement

Privacy rules from `.learned-config.yaml` are enforced across all sharing/export scripts:

- **`share_personal: false`** — Personal patterns (`learned/personal/`) are excluded from merges, exports, syncs, and reports
- **`exclude_keywords`** — Files containing sensitive keywords (password, api_key, secret, token) are automatically filtered
- **`require_review_for_global: true`** — Global patterns (`_global/`) require review before merging (skipped unless `--force`)

The `scripts/privacy_filter.sh` script implements these rules and is integrated into `merge_team.sh`, `export_instincts.sh`, `sync_to_claude_code.sh`, and `team_report.sh`.

```bash
# Check a single file
./scripts/privacy_filter.sh --check learned/backend/pattern/some-file.md

# Filter a file list (pipe mode)
find learned/ -name "*.md" | ./scripts/privacy_filter.sh --verbose
```

### Pattern Validation

Validate all pattern files for correctness:

```bash
./scripts/validate.sh              # Validate all patterns
./scripts/validate.sh --fix        # Auto-fix common issues
./scripts/validate.sh --json       # Machine-readable output
```

Checks: frontmatter structure, required fields, valid confidence/domain/category values, date formats, required sections (`# Title`, `## Problem / Trigger`, `## Solution`), and filename conventions.

Auto-fix (`--fix`) handles: missing `times_seen` (→ 1), missing `last_seen` (→ copy from `first_seen`), lowercase confidence (→ uppercase).

### Configuration

See `.learned-config.yaml` for team sync, privacy, and merge settings.

### Uninstall

```bash
./setup/uninstall.sh
```

## Example Patterns

The repo ships with two example patterns to illustrate the format:

- **`learned/_global/debug/node-heap-oom-debug.md`** — Node.js OOM debugging when heap looks fine (external memory / ArrayBuffer)
- **`learned/backend/pattern/postgres-advisory-lock.md`** — Using PostgreSQL advisory locks to prevent duplicate job execution

These are real-world patterns that demonstrate the frontmatter format, section structure, and confidence levels.

## Spec-Driven Development

### Overview

Integrate spec-driven development tools with the learning system. Write specs before coding, then automatically extract architectural decisions, tech stack choices, and requirement patterns as reusable learnings.

### Quick Start

```bash
# 1. Install spec tools
npm install -g @fission-ai/openspec@latest
uv tool install specify-cli --from git+https://github.com/github/spec-kit.git

# 2. Initialize for your project
./scripts/spec_init.sh --project ~/Projects/myapp --superpowers
```

### Workflow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│    SPEC      │────▶│   EXECUTE    │────▶│    LEARN     │
│              │     │              │     │              │
│ Write specs  │     │ Implement    │     │ Extract      │
│ with OpenSpec│     │ against spec │     │ patterns     │
│ or Spec Kit  │     │ (hooks keep  │     │ from specs   │
│              │     │  you on      │     │ and save as  │
│              │     │  track)      │     │ learnings    │
└──────────────┘     └──────────────┘     └──────────────┘
```

### Supported Tools

| Tool | Purpose | Install Command |
|------|---------|----------------|
| **Superpowers** | Claude Code plugin for enhanced spec workflows | Inside Claude Code: `/plugin marketplace add obra/superpowers-marketplace` |
| **OpenSpec** | Spec authoring & change management | `npm install -g @fission-ai/openspec@latest` |
| **Spec Kit** | GitHub's spec-driven development toolkit | `uv tool install specify-cli --from git+https://github.com/github/spec-kit.git` |

### How It Works

- **`hooks/on-prompt-reminder.md`** includes `[SpecAwareness]` — reminds Claude to reference specs during implementation
- **`hooks/on-spec-complete.md`** includes `[SpecLearning]` — triggers pattern extraction after spec phases
- **`scripts/spec_init.sh`** — one-command setup for any project
- **`.learned-config.yaml`** `integrations` section — configure which tools, auto-extraction settings, and spec directories

## Design Influences

- **Claudeception** — Quality threshold: only extract genuinely non-obvious knowledge
- **Reflect System** — Three confidence levels, correct once → never repeat
- **Continuous Learning v2** — Pattern reinforcement, auto-upgrade to skills
- **Claude Code Hooks** — Real-time learning without post-processing delay
