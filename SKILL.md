---
name: murmur-ai
description: "Drive the locally installed Claude Code CLI in headless mode (`claude -p`) for codebase analysis, refactoring, bug fixing, test generation, code review, and structured output. Includes continuous learning with Hook-based real-time pattern extraction, 3D categorization (Domain/Category/Project), and auto-sync to Claude Code native skills."
---

# Claude Code Learner

Run the local **Claude Code** CLI from OpenClaw in headless (non-interactive) mode, with **continuous learning** that captures non-obvious patterns in real time.

## Prerequisites

```bash
claude --version   # verify Claude Code is installed
```

## Bundled Script

`scripts/claude_code_run.sh` — a shell wrapper that:
- Allocates a PTY via `script(1)` (handles macOS vs Linux syntax automatically)
- Forwards all common Claude Code flags
- Streams output to stdout

## Quick Start

```bash
# Simple prompt (read-only plan mode)
./skills/murmur-ai/scripts/claude_code_run.sh \
  -p "Summarize this project." \
  --permission-mode plan

# With a working directory
./skills/murmur-ai/scripts/claude_code_run.sh \
  -p "Find and fix the null pointer bug in src/." \
  --workdir ~/Projects/myapp \
  --permission-mode acceptEdits \
  --allowedTools "Read,Edit,Bash"
```

## Headless Mode (`claude -p`)

Claude Code's `-p` flag runs a single prompt non-interactively and exits.

```bash
claude -p "Your prompt here"
```

**Important:** Claude Code sometimes expects a TTY. The bundled script wraps calls with `script(1)` to provide a pseudo-terminal.

### PTY Handling

| OS    | Command |
|-------|---------|
| macOS | `script -q /dev/null claude -p "..."` |
| Linux | `script -q -c 'claude -p "..."' /dev/null` |

## Key Flags

### `--permission-mode`

| Mode | Effect |
|------|--------|
| `plan` | Read-only analysis. Best for exploration and review. |
| `acceptEdits` | Can read and edit files. Needed for refactoring, bug fixing. |

### `--allowedTools`

Common: `Read`, `Edit`, `Write`, `Bash`, `WebSearch`
Scope bash: `Bash(npm test)`, `Bash(git diff *)`

### `--output-format`

| Format | Use case |
|--------|----------|
| `text` | Human-readable (default) |
| `json` | Machine-parseable |

## Common Recipes

```bash
# Analyse a repo (read-only)
claude -p "Summarize architecture" --permission-mode plan

# Fix bugs with test verification
claude -p "Fix failing auth tests" --permission-mode acceptEdits --allowedTools "Read,Edit,Bash(npm test)"

# Code review (JSON output)
claude -p "Review staged diff for bugs" --permission-mode plan --output-format json
```

---

## 🎓 Continuous Learning System

### Architecture Overview

```
┌─────────────────────────────────────────────────┐
│                 Claude Code Session              │
│                                                  │
│  UserPromptSubmit Hook                           │
│  → injects on-prompt-reminder.md to stderr       │
│  → Claude sees learning instructions every turn  │
│                                                  │
│  Claude discovers non-obvious pattern            │
│  → writes .md file to learned/{domain}/{cat}/    │
│                                                  │
│  Stop Hook                                       │
│  → on_session_stop.sh detects new files          │
└─────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────┐
│               auto_learn.sh (cron)               │
│  1. Review & merge patterns                      │
│  2. sync_to_claude_code.sh → ~/.claude/skills/   │
│  3. Git commit                                   │
└─────────────────────────────────────────────────┘
```

### 3D Directory Structure

Patterns are organized in three dimensions: **Domain × Category × Project**

```
learned/
├── _global/           # Cross-domain patterns
│   ├── debug/
│   ├── pattern/
│   └── ...
├── devops/            # DevOps-specific
├── web/               # Frontend/web
├── mobile/            # Mobile dev
├── backend/           # Backend/API
├── data/              # Data engineering
└── projects/          # Project-specific
    └── {project-name}/
        ├── debug/
        ├── security/
        └── ...
```

**Domains:** `_global`, `devops`, `web`, `mobile`, `backend`, `data`
**Categories:** `debug`, `pattern`, `style`, `security`, `performance`, `ops`

### 🔧 Hook Installation

Run `setup/install.sh` to automatically merge hooks into `~/.claude/settings.json`:

```bash
# Solo mode (default) — hooks, directory structure, auto-learn cron
./setup/install.sh

# Team mode — solo + personal branch, push/pull cron
./setup/install.sh --team

# Upgrade from solo to team (idempotent, skips already-installed steps)
./setup/install.sh --team
```

This smart-merges hooks (backs up existing settings), updates paths, and optionally sets up cron jobs.

**Solo mode** (default): Patterns go on the current branch. No git branches or push/pull cron.
**Team mode** (`--team`): Creates a personal `learnings/{name}` branch and `learned/personal/{name}/` directory, plus push/pull cron jobs.

The hooks do two things:
- **UserPromptSubmit**: Reminds Claude to save non-obvious discoveries
- **Stop**: Lightweight check for new pattern files (< 100ms)

### Pattern Lifecycle

1. **Discover** — Claude finds something non-obvious during a session
2. **Save** — Claude writes a pattern file (via hook prompt)
3. **Reinforce** — Same pattern seen again → `times_seen++`, confidence may upgrade
4. **Review** — Periodic review merges duplicates, archives stale patterns
5. **Sync** — HIGH confidence + seen ≥3× → synced to `~/.claude/skills/` as native skill

### Scripts

#### Extract patterns (manual/legacy)
```bash
echo "conversation" | ./scripts/extract_patterns.sh
./scripts/extract_patterns.sh -f transcript.txt --context devops --category debug
./scripts/extract_patterns.sh -f transcript.txt --project twdd-api
```

#### Review learned patterns
```bash
./scripts/review_learnings.sh --report                # Stats + listing
./scripts/review_learnings.sh --context backend       # Filter by domain
./scripts/review_learnings.sh --project twdd-api      # Filter by project
./scripts/review_learnings.sh --auto                  # Auto-apply suggestions
```

#### Sync to Claude Code native skills
```bash
./scripts/sync_to_claude_code.sh              # Sync HIGH + seen ≥3
./scripts/sync_to_claude_code.sh --dry-run    # Preview
./scripts/sync_to_claude_code.sh --force      # All HIGH (ignore times_seen)
./scripts/sync_to_claude_code.sh --clean      # Remove orphaned synced skills
```

#### Inject into project CLAUDE.md
```bash
./scripts/inject_claude_md.sh --target ~/Projects/myapp
./scripts/inject_claude_md.sh --target ~/Projects/myapp --contexts devops,backend
./scripts/inject_claude_md.sh --target ~/Projects/myapp --project twdd-api
```

#### Auto-learn pipeline (cron)
```bash
./scripts/auto_learn.sh              # Review + sync + commit
./scripts/auto_learn.sh --days 7     # Also extract from memory (legacy)
./scripts/auto_learn.sh --dry-run    # Preview
```

#### Spec-driven development report
```bash
./scripts/spec_report.sh                    # Full report
./scripts/spec_report.sh --project ~/Projects/myapp  # For a specific project
./scripts/spec_report.sh --json             # Machine-readable output
```

Shows: spec file overview (archived vs active), pattern extraction stats by confidence and domain, learning pipeline (spec vs coding patterns), and tool installation status.

#### Export patterns
```bash
./scripts/export_instincts.sh --min-confidence HIGH
./scripts/export_instincts.sh --context devops --format json -o export.json
./scripts/export_instincts.sh --project twdd-api
```

#### Privacy filtering
```bash
# Filter a file list through privacy rules from .learned-config.yaml
find learned/ -name "*.md" | ./scripts/privacy_filter.sh
find learned/ -name "*.md" | ./scripts/privacy_filter.sh --verbose

# Check a single file (exit 0=safe, exit 1=filtered)
./scripts/privacy_filter.sh --check learned/backend/pattern/some-file.md
```

Enforces `share_personal`, `exclude_keywords`, and `require_review_for_global` from `.learned-config.yaml`. Integrated into `merge_team.sh`, `export_instincts.sh`, `sync_to_claude_code.sh`, and `team_report.sh`.

#### Pattern validation
```bash
./scripts/validate.sh              # Validate all patterns
./scripts/validate.sh --fix        # Auto-fix common issues
./scripts/validate.sh --json       # Machine-readable output
./scripts/validate.sh FILE ...     # Validate specific files
```

Checks: YAML frontmatter, required fields, valid confidence/domain/category values, date formats, required sections, filename conventions.

#### Pattern status
```bash
./scripts/status.sh              # Human-readable stats
./scripts/status.sh --json       # Machine-readable JSON
```

#### Smoke tests
```bash
./scripts/test.sh                # Validate scripts, structure, templates, hooks
```

### Confidence Levels

| Level | Meaning | Criteria |
|-------|---------|----------|
| HIGH | Verified fix | Corrected a real problem, confirmed working |
| MEDIUM | Useful approach | Confirmed helpful, needs more validation |
| LOW | Observation | Noticed pattern, not yet proven |

## Team Collaboration

### 安裝（每人一次）
```bash
./setup/install.sh
```

### 日常使用
正常使用 Claude Code，patterns 自動提取和同步。

### 管理者操作
```bash
# 合併所有人的 learnings
./scripts/merge_team.sh

# 查看報告
./scripts/team_report.sh

# 預覽（不實際合併）
./scripts/merge_team.sh --dry-run

# 手動解決衝突（而非自動 theirs）
./scripts/merge_team.sh --strategy manual
```

### 分支規範
- `main` — 團隊共享的穩定 patterns
- `learnings/{name}` — 個人的即時推送
- 從不直接 push 到 main，透過 merge_team.sh 合併

## Spec-Driven Development Integration

### Supported Tools

| Tool | Purpose | Install |
|------|---------|---------|
| **Superpowers** | Claude Code plugin for spec-driven workflows | `/plugin marketplace add obra/superpowers-marketplace` (inside Claude Code) |
| **OpenSpec** | Specification authoring and change management | `npm install -g @fission-ai/openspec@latest` |
| **Spec Kit** | GitHub's spec-driven development toolkit | `uv tool install specify-cli --from git+https://github.com/github/spec-kit.git` |

### Setup

Initialize spec-driven development for any project:

```bash
./scripts/spec_init.sh --project ~/Projects/myapp
./scripts/spec_init.sh --project ~/Projects/myapp --tool both --superpowers
```

### Workflow: Spec → Execute → Learn

1. **Spec** — Write specifications using OpenSpec or Spec Kit before coding
2. **Execute** — Implement against the spec; Claude Code hooks remind you to reference it
3. **Learn** — When a spec phase completes, `on-spec-complete.md` prompts extraction of:
   - Architectural decisions (why this approach?)
   - Rejected alternatives (why NOT that approach?)
   - Reusable requirement patterns

### Auto-Extraction from Specs

The `[SpecAwareness]` hook (in `on-prompt-reminder.md`) ensures Claude:
- References the current spec during implementation
- Notes deviations from the spec as patterns
- Checks for learnings after completing spec tasks

The `[SpecLearning]` hook (in `on-spec-complete.md`) triggers after spec completion to extract architectural decisions, tech stack choices, and requirement patterns.

### Configuration

In `.learned-config.yaml`, the `integrations` section controls:
- `spec_tool`: Which tool to use (`auto`, `openspec`, `speckit`, `both`)
- `auto_extract_from_specs`: Enable/disable automatic pattern extraction
- `extract_decisions`: Save architectural decisions as patterns
- `spec_dirs`: Directories to scan for completed spec artifacts

## ⚠️ Rules

1. Use `--permission-mode plan` for any read-only / analysis task.
2. Keep `--allowedTools` as narrow as possible.
3. Never run Claude Code inside the OpenClaw workspace (`~/clawd/`).
4. For long-running tasks, prefer background mode with progress monitoring.
