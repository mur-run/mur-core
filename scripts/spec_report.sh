#!/usr/bin/env bash
# spec_report.sh — Show spec-driven development learning report
# Usage:
#   ./scripts/spec_report.sh                    # Full report
#   ./scripts/spec_report.sh --project PATH     # For a specific project
#   ./scripts/spec_report.sh --json             # Machine-readable output
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
LEARNED_DIR="$SKILL_DIR/learned"
CONFIG_FILE="$SKILL_DIR/.learned-config.yaml"

# --- Defaults ---
PROJECT_PATH=""
JSON_OUTPUT=false

# --- Parse args ---
while [[ $# -gt 0 ]]; do
  case "$1" in
    --project)
      PROJECT_PATH="$2"
      shift 2
      ;;
    --json)
      JSON_OUTPUT=true
      shift
      ;;
    -h|--help)
      echo "Usage: spec_report.sh [--project PATH] [--json]"
      echo ""
      echo "Show spec-driven development learning report."
      echo ""
      echo "Options:"
      echo "  --project PATH   Scan a specific project's spec directories"
      echo "  --json           Machine-readable JSON output"
      echo "  -h, --help       Show this help"
      exit 0
      ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

# --- Determine scan root ---
if [[ -n "$PROJECT_PATH" ]]; then
  SCAN_ROOT="$(cd "$PROJECT_PATH" 2>/dev/null && pwd)" || {
    echo "Error: Cannot access project path: $PROJECT_PATH" >&2
    exit 1
  }
else
  SCAN_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo "$PWD")
fi

# --- 1. Spec Overview ---
OPENSPEC_ARCHIVE_DIR="$SCAN_ROOT/openspec/changes/archive"
OPENSPEC_ACTIVE_DIR="$SCAN_ROOT/openspec/changes"
SPEC_KIT_DIR="$SCAN_ROOT/.spec"

ARCHIVED_SPECS=0
ACTIVE_SPECS=0
SPEC_KIT_FILES=0
LATEST_SPEC_DATE="N/A"

if [[ -d "$OPENSPEC_ARCHIVE_DIR" ]]; then
  ARCHIVED_SPECS=$(find "$OPENSPEC_ARCHIVE_DIR" -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
fi

if [[ -d "$OPENSPEC_ACTIVE_DIR" ]]; then
  # Active = files in changes/ but NOT in changes/archive/
  ACTIVE_SPECS=$(find "$OPENSPEC_ACTIVE_DIR" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
fi

if [[ -d "$SPEC_KIT_DIR" ]]; then
  SPEC_KIT_FILES=$(find "$SPEC_KIT_DIR" -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
fi

TOTAL_SPECS=$((ARCHIVED_SPECS + ACTIVE_SPECS + SPEC_KIT_FILES))

# Find latest spec date
ALL_SPEC_FILES=""
[[ -d "$OPENSPEC_ARCHIVE_DIR" ]] && ALL_SPEC_FILES+=$(find "$OPENSPEC_ARCHIVE_DIR" -name "*.md" -type f 2>/dev/null)
[[ -d "$OPENSPEC_ACTIVE_DIR" ]] && ALL_SPEC_FILES+=$(find "$OPENSPEC_ACTIVE_DIR" -maxdepth 1 -name "*.md" -type f 2>/dev/null)
[[ -d "$SPEC_KIT_DIR" ]] && ALL_SPEC_FILES+=$(find "$SPEC_KIT_DIR" -name "*.md" -type f 2>/dev/null)

if [[ -n "$ALL_SPEC_FILES" ]]; then
  # Get most recently modified spec file date
  LATEST_SPEC_DATE=$(echo "$ALL_SPEC_FILES" | xargs ls -t 2>/dev/null | head -1 | xargs stat -f "%Sm" -t "%Y-%m-%d" 2>/dev/null || echo "N/A")
fi

# --- 2. Pattern Extraction Stats ---
SPEC_PROCESSED="$LEARNED_DIR/.spec_processed"
PROCESSED_COUNT=0
if [[ -f "$SPEC_PROCESSED" ]]; then
  PROCESSED_COUNT=$(wc -l < "$SPEC_PROCESSED" | tr -d ' ')
fi

# Count patterns by confidence
ALL_PATTERNS=$(find "$LEARNED_DIR" -name "*.md" -not -name ".gitkeep" -not -name "archived-*" -not -name "skill-*" -not -name ".last_check" -not -name ".spec_processed" -type f 2>/dev/null)
TOTAL_PATTERNS=0
HIGH_COUNT=0
MEDIUM_COUNT=0
LOW_COUNT=0

# Domain counts
GLOBAL_COUNT=0
DEVOPS_COUNT=0
WEB_COUNT=0
MOBILE_COUNT=0
BACKEND_COUNT=0
DATA_COUNT=0

if [[ -n "$ALL_PATTERNS" ]]; then
  TOTAL_PATTERNS=$(echo "$ALL_PATTERNS" | wc -l | tr -d ' ')
  HIGH_COUNT=$(echo "$ALL_PATTERNS" | xargs grep -l '^confidence: HIGH' 2>/dev/null | wc -l | tr -d ' ')
  MEDIUM_COUNT=$(echo "$ALL_PATTERNS" | xargs grep -l '^confidence: MEDIUM' 2>/dev/null | wc -l | tr -d ' ')
  LOW_COUNT=$(echo "$ALL_PATTERNS" | xargs grep -l '^confidence: LOW' 2>/dev/null | wc -l | tr -d ' ')
  GLOBAL_COUNT=$(echo "$ALL_PATTERNS" | xargs grep -l '^domain: _global' 2>/dev/null | wc -l | tr -d ' ')
  DEVOPS_COUNT=$(echo "$ALL_PATTERNS" | xargs grep -l '^domain: devops' 2>/dev/null | wc -l | tr -d ' ')
  WEB_COUNT=$(echo "$ALL_PATTERNS" | xargs grep -l '^domain: web' 2>/dev/null | wc -l | tr -d ' ')
  MOBILE_COUNT=$(echo "$ALL_PATTERNS" | xargs grep -l '^domain: mobile' 2>/dev/null | wc -l | tr -d ' ')
  BACKEND_COUNT=$(echo "$ALL_PATTERNS" | xargs grep -l '^domain: backend' 2>/dev/null | wc -l | tr -d ' ')
  DATA_COUNT=$(echo "$ALL_PATTERNS" | xargs grep -l '^domain: data' 2>/dev/null | wc -l | tr -d ' ')
fi

# --- 3. Learning Pipeline ---
# Count spec-originated patterns (source line contains "spec:")
SPEC_ORIGINATED=0
CODING_ORIGINATED=0
if [[ -n "$ALL_PATTERNS" ]]; then
  SPEC_ORIGINATED=$(echo "$ALL_PATTERNS" | xargs grep -l 'Session: spec:' 2>/dev/null | wc -l | tr -d ' ')
  CODING_ORIGINATED=$((TOTAL_PATTERNS - SPEC_ORIGINATED))
fi

# Promotion candidates: MEDIUM confidence with times_seen >= 2 from specs
PROMOTION_CANDIDATES=0
if [[ -n "$ALL_PATTERNS" ]]; then
  while IFS= read -r pf; do
    [[ -z "$pf" ]] && continue
    if grep -q '^confidence: MEDIUM' "$pf" 2>/dev/null && grep -q 'Session: spec:' "$pf" 2>/dev/null; then
      TIMES=$(grep '^times_seen:' "$pf" 2>/dev/null | sed 's/times_seen: *//' || echo "1")
      if [[ "$TIMES" -ge 2 ]]; then
        PROMOTION_CANDIDATES=$((PROMOTION_CANDIDATES + 1))
      fi
    fi
  done <<< "$ALL_PATTERNS"
fi

# --- 4. Tool Usage ---
SPEC_TOOL="not configured"
SUPERPOWERS="disabled"
OPENSPEC_INSTALLED=false
SPECKIT_INSTALLED=false

if [[ -f "$CONFIG_FILE" ]]; then
  SPEC_TOOL_VAL=$(grep '^ *spec_tool:' "$CONFIG_FILE" 2>/dev/null | sed 's/.*spec_tool: *//' | sed 's/ *#.*//' || echo "")
  [[ -n "$SPEC_TOOL_VAL" ]] && SPEC_TOOL="$SPEC_TOOL_VAL"
  if grep -q '^ *superpowers: *true' "$CONFIG_FILE" 2>/dev/null; then
    SUPERPOWERS="enabled"
  fi
fi

command -v openspec >/dev/null 2>&1 && OPENSPEC_INSTALLED=true
command -v specify >/dev/null 2>&1 && SPECKIT_INSTALLED=true

# Check if claude superpowers plugin is available
SUPERPOWERS_PLUGIN=false
if [[ -d "$HOME/.claude/plugins" ]] && find "$HOME/.claude/plugins" -name "*superpowers*" -type d 2>/dev/null | grep -q .; then
  SUPERPOWERS_PLUGIN=true
fi

# --- Output ---
if [[ "$JSON_OUTPUT" == "true" ]]; then
  cat <<EOF
{
  "spec_overview": {
    "total_specs": $TOTAL_SPECS,
    "archived_specs": $ARCHIVED_SPECS,
    "active_specs": $ACTIVE_SPECS,
    "spec_kit_files": $SPEC_KIT_FILES,
    "latest_spec_date": "$LATEST_SPEC_DATE"
  },
  "pattern_stats": {
    "total_patterns": $TOTAL_PATTERNS,
    "processed_spec_files": $PROCESSED_COUNT,
    "by_confidence": {
      "HIGH": $HIGH_COUNT,
      "MEDIUM": $MEDIUM_COUNT,
      "LOW": $LOW_COUNT
    },
    "by_domain": {
      "_global": $GLOBAL_COUNT,
      "devops": $DEVOPS_COUNT,
      "web": $WEB_COUNT,
      "mobile": $MOBILE_COUNT,
      "backend": $BACKEND_COUNT,
      "data": $DATA_COUNT
    }
  },
  "learning_pipeline": {
    "from_specs": $SPEC_ORIGINATED,
    "from_coding": $CODING_ORIGINATED,
    "promotion_candidates": $PROMOTION_CANDIDATES
  },
  "tool_usage": {
    "spec_tool": "$SPEC_TOOL",
    "superpowers": "$SUPERPOWERS",
    "superpowers_plugin": $SUPERPOWERS_PLUGIN,
    "openspec_installed": $OPENSPEC_INSTALLED,
    "speckit_installed": $SPECKIT_INSTALLED
  },
  "scan_root": "$SCAN_ROOT"
}
EOF
else
  echo "========================================"
  echo "  SPEC-DRIVEN DEVELOPMENT REPORT"
  echo "========================================"
  echo ""
  echo "📁 Scan root: $SCAN_ROOT"
  echo ""

  echo "── Spec Overview ──────────────────────"
  echo "  Total spec files:      $TOTAL_SPECS"
  echo "    Archived (OpenSpec): $ARCHIVED_SPECS"
  echo "    Active (OpenSpec):   $ACTIVE_SPECS"
  echo "    Spec Kit:            $SPEC_KIT_FILES"
  echo "  Latest spec date:      $LATEST_SPEC_DATE"
  echo ""

  echo "── Pattern Extraction Stats ───────────"
  echo "  Spec files processed:  $PROCESSED_COUNT"
  echo "  Total patterns:        $TOTAL_PATTERNS"
  echo "  By confidence:"
  echo "    HIGH:   $HIGH_COUNT"
  echo "    MEDIUM: $MEDIUM_COUNT"
  echo "    LOW:    $LOW_COUNT"
  echo "  By domain:"
  echo "    _global: $GLOBAL_COUNT  devops: $DEVOPS_COUNT  web: $WEB_COUNT"
  echo "    mobile: $MOBILE_COUNT  backend: $BACKEND_COUNT  data: $DATA_COUNT"
  echo ""

  echo "── Learning Pipeline ──────────────────"
  echo "  Patterns from specs:    $SPEC_ORIGINATED"
  echo "  Patterns from coding:   $CODING_ORIGINATED"
  echo "  Promotion candidates:   $PROMOTION_CANDIDATES"
  if [[ $PROMOTION_CANDIDATES -gt 0 ]]; then
    echo "    (MEDIUM confidence spec patterns seen ≥2 times)"
  fi
  echo ""

  echo "── Tool Usage ─────────────────────────"
  echo "  spec_tool config:       $SPEC_TOOL"
  echo "  Superpowers config:     $SUPERPOWERS"
  echo "  Superpowers plugin:     $SUPERPOWERS_PLUGIN"
  echo "  OpenSpec installed:     $OPENSPEC_INSTALLED"
  echo "  Spec Kit installed:     $SPECKIT_INSTALLED"
  echo ""
  echo "========================================"
fi
