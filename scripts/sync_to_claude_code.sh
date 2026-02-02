#!/usr/bin/env bash
# sync_to_claude_code.sh — Sync mature patterns to ~/.claude/skills/
# Usage:
#   ./scripts/sync_to_claude_code.sh              # Sync HIGH + times_seen >= 3
#   ./scripts/sync_to_claude_code.sh --dry-run    # Preview only
#   ./scripts/sync_to_claude_code.sh --force      # Sync all HIGH (ignore times_seen)
#   ./scripts/sync_to_claude_code.sh --clean      # Remove orphaned synced skills
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
LEARNED_DIR="$SKILL_DIR/learned"
CLAUDE_SKILLS_DIR="$HOME/.claude/skills"

# --- Defaults ---
DRY_RUN=false
FORCE=false
CLEAN=false
MIN_TIMES=3

# --- Parse args ---
while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)  DRY_RUN=true; shift ;;
    --force)    FORCE=true; MIN_TIMES=1; shift ;;
    --clean)    CLEAN=true; shift ;;
    -h|--help)
      echo "Usage: sync_to_claude_code.sh [--dry-run] [--force] [--clean]"
      echo ""
      echo "Sync mature learned patterns to ~/.claude/skills/ as native skills."
      echo ""
      echo "Options:"
      echo "  --dry-run   Preview what would be synced"
      echo "  --force     Sync all HIGH confidence (ignore times_seen threshold)"
      echo "  --clean     Remove synced skills whose source patterns no longer exist"
      echo "  -h, --help  Show this help"
      exit 0
      ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

mkdir -p "$CLAUDE_SKILLS_DIR"

# Privacy filter: check each pattern before syncing
PRIVACY_FILTER="$SCRIPT_DIR/privacy_filter.sh"

SYNCED=0
SKIPPED=0
CLEANED=0
PRIVACY_SKIPPED=0

echo "🔄 Syncing learned patterns to Claude Code skills..." >&2
echo "   Source: $LEARNED_DIR" >&2
echo "   Target: $CLAUDE_SKILLS_DIR" >&2
echo "   Min times_seen: $MIN_TIMES" >&2
echo "" >&2

# --- Helper: read YAML frontmatter field ---
get_field() {
  local file="$1" field="$2"
  grep "^${field}:" "$file" 2>/dev/null | sed "s/^${field}: *//" | head -1
}

# --- Helper: read section content ---
get_section() {
  local file="$1" section="$2"
  sed -n "/^## ${section}/,/^## /{/^## /!p}" "$file" | sed '/^$/d' | sed '/^---$/d'
}

# --- Sync patterns ---
while IFS= read -r f; do
  [[ -z "$f" ]] && continue

  P_CONF=$(get_field "$f" "confidence")
  P_TIMES=$(get_field "$f" "times_seen")
  P_TIMES=${P_TIMES:-1}
  P_NAME=$(get_field "$f" "name")

  [[ -z "$P_NAME" ]] && continue

  # Privacy filter: skip files that violate privacy rules
  if [[ -x "$PRIVACY_FILTER" ]]; then
    if ! "$PRIVACY_FILTER" --check "$f" >/dev/null 2>&1; then
      PRIVACY_SKIPPED=$((PRIVACY_SKIPPED + 1))
      continue
    fi
  fi

  # Filter: HIGH confidence + times_seen threshold
  if [[ "$P_CONF" != "HIGH" ]]; then
    continue
  fi
  if [[ "$P_TIMES" -lt "$MIN_TIMES" ]]; then
    SKIPPED=$((SKIPPED + 1))
    continue
  fi

  # Read content
  P_TITLE=$(grep '^# ' "$f" | head -1 | sed 's/^# //')
  P_PROBLEM=$(get_section "$f" "Problem / Trigger")
  P_SOLUTION=$(get_section "$f" "Solution")
  P_DOMAIN=$(get_field "$f" "domain")
  P_CATEGORY=$(get_field "$f" "category")

  SKILL_NAME="learned-${P_NAME}"
  SKILL_OUT_DIR="$CLAUDE_SKILLS_DIR/${SKILL_NAME}"
  SKILL_OUT="$SKILL_OUT_DIR/SKILL.md"

  if [[ "$DRY_RUN" == "true" ]]; then
    echo "  [DRY RUN] Would sync: $P_NAME → $SKILL_OUT" >&2
    SYNCED=$((SYNCED + 1))
    continue
  fi

  mkdir -p "$SKILL_OUT_DIR"
  cat > "$SKILL_OUT" <<EOF
---
name: ${SKILL_NAME}
description: "${P_PROBLEM}"
---

# ${P_TITLE:-$SKILL_NAME}

## When to Apply
${P_PROBLEM}

## What to Do
${P_SOLUTION}

## Metadata
- Domain: ${P_DOMAIN:-_global}
- Category: ${P_CATEGORY:-pattern}
- Confidence: HIGH
- Times seen: ${P_TIMES}
- Source: murmur-ai
EOF

  echo "  ✅ Synced: $P_NAME → $SKILL_OUT" >&2
  SYNCED=$((SYNCED + 1))

done < <(find "$LEARNED_DIR" -name "*.md" -not -name ".gitkeep" -not -name "archived-*" -not -name "skill-*" 2>/dev/null | sort)

# --- Clean orphaned skills ---
if [[ "$CLEAN" == "true" ]]; then
  echo "" >&2
  echo "🧹 Cleaning orphaned synced skills..." >&2

  for skill_dir in "$CLAUDE_SKILLS_DIR"/learned-*/; do
    [[ -d "$skill_dir" ]] || continue
    SKILL_SLUG=$(basename "$skill_dir" | sed 's/^learned-//')

    # Check if source pattern still exists
    FOUND=$(find "$LEARNED_DIR" -name "*.md" -exec grep -l "^name: ${SKILL_SLUG}$" {} \; 2>/dev/null | head -1)
    if [[ -z "$FOUND" ]]; then
      if [[ "$DRY_RUN" == "true" ]]; then
        echo "  [DRY RUN] Would remove: $skill_dir" >&2
      else
        rm -rf "$skill_dir"
        echo "  🗑️  Removed orphaned: $skill_dir" >&2
      fi
      CLEANED=$((CLEANED + 1))
    fi
  done
fi

echo "" >&2
echo "=== SYNC SUMMARY ===" >&2
echo "  Synced:  $SYNCED" >&2
echo "  Skipped: $SKIPPED (below threshold)" >&2
if [[ $PRIVACY_SKIPPED -gt 0 ]]; then
  echo "  Privacy: $PRIVACY_SKIPPED (filtered by privacy rules)" >&2
fi
if [[ "$CLEAN" == "true" ]]; then
  echo "  Cleaned: $CLEANED" >&2
fi
echo "====================" >&2
