#!/usr/bin/env bash
# auto_learn.sh — Automated learning pipeline for Clawdbot cron
# Usage:
#   ./scripts/auto_learn.sh              # Review + merge + sync + commit
#   ./scripts/auto_learn.sh --days 7     # Include extraction from memory (legacy)
#   ./scripts/auto_learn.sh --dry-run    # Preview only
#
# With hooks installed, extraction happens in real-time.
# This script focuses on: review + merge + upgrade + sync + commit.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
LEARNED_DIR="$SKILL_DIR/learned"
MEMORY_DIR="$HOME/clawd/memory"

# --- Defaults ---
DAYS=0
DRY_RUN=false
DO_COMMIT=true
DO_EXTRACT=false

# --- Parse args ---
while [[ $# -gt 0 ]]; do
  case "$1" in
    --days)
      DAYS="$2"
      DO_EXTRACT=true
      shift 2
      ;;
    --dry-run)
      DRY_RUN=true
      DO_COMMIT=false
      shift
      ;;
    --no-commit)
      DO_COMMIT=false
      shift
      ;;
    --commit)
      DO_COMMIT=true
      shift
      ;;
    -h|--help)
      echo "Usage: auto_learn.sh [--days N] [--dry-run] [--no-commit]"
      echo ""
      echo "Automated learning pipeline — reviews, merges, syncs patterns."
      echo "With hooks installed, extraction is real-time. This script handles:"
      echo "  1. Review & merge patterns (if > 5 exist)"
      echo "  2. Sync mature patterns to ~/.claude/skills/"
      echo "  3. Git commit changes"
      echo ""
      echo "Options:"
      echo "  --days N       Also extract from last N days of memory (legacy mode)"
      echo "  --dry-run      Preview only, no changes"
      echo "  --no-commit    Skip git commit"
      echo "  -h, --help     Show this help"
      exit 0
      ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

# --- Summary counters ---
PATTERNS_BEFORE=0
PATTERNS_AFTER=0
SKILLS_SYNCED=0
ERRORS=0

echo "🤖 Auto-learn pipeline starting..." >&2
echo "   Learned dir: $LEARNED_DIR" >&2
echo "   Dry run: $DRY_RUN" >&2
echo "" >&2

# --- Count existing patterns (across all subdirs) ---
PATTERNS_BEFORE=$(find "$LEARNED_DIR" -name "*.md" -not -name ".gitkeep" -not -name "archived-*" -not -name "skill-*" -not -name ".last_check" 2>/dev/null | wc -l | tr -d ' ')
echo "📊 Current patterns: $PATTERNS_BEFORE" >&2

# --- Step 1: Legacy extraction (only if --days specified) ---
if [[ "$DO_EXTRACT" == "true" && "$DAYS" -gt 0 ]]; then
  echo "" >&2
  echo "📂 Step 1: Legacy extraction from memory (last $DAYS days)..." >&2

  if [[ -d "$MEMORY_DIR" ]]; then
    for i in $(seq 0 $((DAYS - 1))); do
      TARGET_DATE=$(date -v-${i}d +%Y-%m-%d 2>/dev/null || date -d "-${i} days" +%Y-%m-%d 2>/dev/null)
      DAILY_FILE="$MEMORY_DIR/${TARGET_DATE}.md"
      if [[ -f "$DAILY_FILE" ]]; then
        FILE_SIZE=$(wc -c < "$DAILY_FILE" | tr -d ' ')
        if [[ $FILE_SIZE -ge 100 ]]; then
          echo "   Processing: $(basename "$DAILY_FILE")" >&2
          if [[ "$DRY_RUN" == "true" ]]; then
            echo "   [DRY RUN] Would extract from $(basename "$DAILY_FILE")" >&2
          else
            "$SCRIPT_DIR/extract_patterns.sh" -f "$DAILY_FILE" -s "$(basename "$DAILY_FILE" .md)" 2>&1 || {
              echo "   ⚠️  Error processing $(basename "$DAILY_FILE")" >&2
              ERRORS=$((ERRORS + 1))
            }
          fi
        fi
      fi
    done
  else
    echo "   ⚠️  Memory directory not found: $MEMORY_DIR" >&2
  fi
else
  echo "📂 Step 1: Skipping extraction (hooks handle real-time extraction)" >&2
fi

# --- Step 1.5: Extract patterns from spec artifacts ---
SPEC_SCANNED=0
SPEC_NEW=0

# Check if auto_extract_from_specs is enabled in config
CONFIG_FILE="$SKILL_DIR/.learned-config.yaml"
SPEC_EXTRACT_ENABLED=false
if [[ -f "$CONFIG_FILE" ]]; then
  if grep -q 'auto_extract_from_specs: *true' "$CONFIG_FILE" 2>/dev/null; then
    SPEC_EXTRACT_ENABLED=true
  fi
fi

if [[ "$SPEC_EXTRACT_ENABLED" == "true" ]]; then
  echo "" >&2
  echo "📜 Step 1.5: Scanning spec artifacts for patterns..." >&2

  # Determine the current git repo root (scan specs from there, not the skill repo)
  REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo "$HOME")

  SPEC_PROCESSED="$LEARNED_DIR/.spec_processed"
  touch "$SPEC_PROCESSED"

  # Directories to scan for spec artifacts
  SPEC_SCAN_DIRS=(
    "$REPO_ROOT/openspec/changes/archive/"
    "$REPO_ROOT/.spec/"
  )

  for SCAN_DIR in "${SPEC_SCAN_DIRS[@]}"; do
    if [[ ! -d "$SCAN_DIR" ]]; then
      continue
    fi
    while IFS= read -r -d '' SPEC_FILE; do
      SPEC_SCANNED=$((SPEC_SCANNED + 1))
      # Check if already processed
      REL_PATH="${SPEC_FILE#$REPO_ROOT/}"
      if grep -qxF "$REL_PATH" "$SPEC_PROCESSED" 2>/dev/null; then
        continue
      fi
      # New spec file — extract patterns
      SPEC_NEW=$((SPEC_NEW + 1))
      if [[ "$DRY_RUN" == "true" ]]; then
        echo "   [DRY RUN] Would extract from: $REL_PATH" >&2
      else
        echo "   Extracting from: $REL_PATH" >&2
        "$SCRIPT_DIR/extract_patterns.sh" -f "$SPEC_FILE" -s "spec:$REL_PATH" --context _global --category pattern 2>&1 || {
          echo "   ⚠️  Error extracting from $REL_PATH" >&2
          ERRORS=$((ERRORS + 1))
        }
      fi
      # Mark as processed
      echo "$REL_PATH" >> "$SPEC_PROCESSED"
    done < <(find "$SCAN_DIR" -name "*.md" -type f -print0 2>/dev/null)
  done

  echo "   Spec files scanned: $SPEC_SCANNED, new patterns extracted: $SPEC_NEW" >&2
else
  echo "" >&2
  echo "📜 Step 1.5: Spec extraction disabled (auto_extract_from_specs != true)" >&2
fi

# --- Step 2: Review & merge if enough patterns ---
PATTERNS_AFTER=$(find "$LEARNED_DIR" -name "*.md" -not -name ".gitkeep" -not -name "archived-*" -not -name "skill-*" -not -name ".last_check" 2>/dev/null | wc -l | tr -d ' ')

echo "" >&2
echo "📊 Step 2: Review check (patterns: $PATTERNS_AFTER)..." >&2

if [[ $PATTERNS_AFTER -gt 5 ]]; then
  echo "   Pattern count > 5, running review..." >&2
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "   [DRY RUN] Would run: review_learnings.sh --auto" >&2
  else
    "$SCRIPT_DIR/review_learnings.sh" --auto 2>&1 || {
      echo "   ⚠️  Review encountered an error" >&2
      ERRORS=$((ERRORS + 1))
    }
  fi
else
  echo "   Not enough patterns yet ($PATTERNS_AFTER ≤ 5), skipping review." >&2
fi

# --- Step 3: Sync mature patterns to Claude Code skills ---
echo "" >&2
echo "🔄 Step 3: Syncing to Claude Code skills..." >&2

if [[ "$DRY_RUN" == "true" ]]; then
  "$SCRIPT_DIR/sync_to_claude_code.sh" --dry-run 2>&1 || {
    echo "   ⚠️  Sync encountered an error" >&2
    ERRORS=$((ERRORS + 1))
  }
else
  SYNC_OUTPUT=$("$SCRIPT_DIR/sync_to_claude_code.sh" --clean 2>&1) || {
    echo "   ⚠️  Sync encountered an error" >&2
    ERRORS=$((ERRORS + 1))
  }
  echo "$SYNC_OUTPUT" >&2
  SKILLS_SYNCED=$(echo "$SYNC_OUTPUT" | grep -c '✅ Synced' || true)
fi

# --- Step 4: Git commit ---
echo "" >&2
echo "📦 Step 4: Git commit..." >&2

if [[ "$DO_COMMIT" == "true" && "$DRY_RUN" != "true" ]]; then
  CHANGES=$(git -C "$HOME/clawd" status --porcelain -- "skills/claude-code-learner/learned/" 2>/dev/null || true)
  if [[ -n "$CHANGES" ]]; then
    git -C "$HOME/clawd" add "skills/claude-code-learner/learned/" 2>/dev/null
    git -C "$HOME/clawd" commit -m "auto-learn: review + sync patterns

Patterns: $PATTERNS_AFTER
Skills synced: $SKILLS_SYNCED" 2>/dev/null && {
      echo "   ✅ Changes committed." >&2
    } || {
      echo "   ⚠️  Git commit failed." >&2
    }
  else
    echo "   No changes to commit." >&2
  fi
else
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "   [DRY RUN] Would commit changes." >&2
  else
    echo "   Commit disabled." >&2
  fi
fi

# --- Summary ---
PATTERNS_FINAL=$(find "$LEARNED_DIR" -name "*.md" -not -name ".gitkeep" -not -name "archived-*" -not -name "skill-*" -not -name ".last_check" 2>/dev/null | wc -l | tr -d ' ')

echo "" >&2
echo "==============================" >&2
echo "  AUTO-LEARN SUMMARY" >&2
echo "==============================" >&2
echo "  Patterns before:       $PATTERNS_BEFORE" >&2
echo "  Patterns after:        $PATTERNS_FINAL" >&2
echo "  Skills synced:         $SKILLS_SYNCED" >&2
echo "  Errors:                $ERRORS" >&2
echo "  Dry run:               $DRY_RUN" >&2
echo "==============================" >&2
