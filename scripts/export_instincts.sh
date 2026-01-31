#!/usr/bin/env bash
# export_instincts.sh — Export learned patterns for sharing
# Usage:
#   ./scripts/export_instincts.sh                           # All to stdout (md)
#   ./scripts/export_instincts.sh --min-confidence HIGH     # Only high confidence
#   ./scripts/export_instincts.sh --context devops           # Filter by domain
#   ./scripts/export_instincts.sh --project twdd-api         # Filter by project
#   ./scripts/export_instincts.sh --format json --output out.json
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
LEARNED_DIR="$SKILL_DIR/learned"

# --- Defaults ---
OUTPUT=""
MIN_CONFIDENCE="LOW"
FORMAT="md"
FILTER_CONTEXT=""
FILTER_PROJECT=""

# --- Confidence ordering ---
confidence_level() {
  case "$1" in
    HIGH)   echo 3 ;;
    MEDIUM) echo 2 ;;
    LOW)    echo 1 ;;
    *)      echo 0 ;;
  esac
}

# --- Parse args ---
while [[ $# -gt 0 ]]; do
  case "$1" in
    --output|-o)         OUTPUT="$2"; shift 2 ;;
    --min-confidence)    MIN_CONFIDENCE="$2"; shift 2 ;;
    --format)            FORMAT="$2"; shift 2 ;;
    --context)           FILTER_CONTEXT="$2"; shift 2 ;;
    --project)           FILTER_PROJECT="$2"; shift 2 ;;
    -h|--help)
      echo "Usage: export_instincts.sh [--output FILE] [--min-confidence LEVEL] [--format md|json] [--context DOMAIN] [--project PROJECT]"
      echo ""
      echo "Export learned patterns for sharing."
      echo ""
      echo "Options:"
      echo "  --output FILE            Output file (default: stdout)"
      echo "  --min-confidence LEVEL   Minimum confidence: HIGH, MEDIUM, LOW (default: LOW)"
      echo "  --format FORMAT          Output format: md or json (default: md)"
      echo "  --context DOMAIN         Filter by domain (_global, devops, web, mobile, backend, data)"
      echo "  --project PROJECT        Filter by project name"
      echo "  -h, --help               Show this help"
      exit 0
      ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

if [[ "$FORMAT" != "md" && "$FORMAT" != "json" ]]; then
  echo "Error: --format must be 'md' or 'json'" >&2; exit 1
fi

MIN_LEVEL=$(confidence_level "$MIN_CONFIDENCE")

# --- Build search path ---
SEARCH_DIR="$LEARNED_DIR"
if [[ -n "$FILTER_PROJECT" ]]; then
  SEARCH_DIR="$LEARNED_DIR/projects/$FILTER_PROJECT"
elif [[ -n "$FILTER_CONTEXT" ]]; then
  SEARCH_DIR="$LEARNED_DIR/$FILTER_CONTEXT"
fi

if [[ ! -d "$SEARCH_DIR" ]]; then
  echo "📭 Directory not found: $SEARCH_DIR" >&2; exit 0
fi

# --- Collect patterns ---
# Privacy filter: exclude files that violate privacy rules before exporting
PRIVACY_FILTER="$SCRIPT_DIR/privacy_filter.sh"
RAW_FILES=$(find "$SEARCH_DIR" -name "*.md" -not -name ".gitkeep" -not -name "archived-*" -not -name "skill-*" 2>/dev/null | sort)

if [[ -x "$PRIVACY_FILTER" ]]; then
  PATTERN_FILES=$(echo "$RAW_FILES" | "$PRIVACY_FILTER" 2>/dev/null || echo "$RAW_FILES")
else
  PATTERN_FILES="$RAW_FILES"
fi

if [[ -z "$PATTERN_FILES" ]]; then
  echo "📭 No patterns found in $SEARCH_DIR/" >&2; exit 0
fi

# --- Filter by confidence ---
FILTERED_FILES=()
while IFS= read -r f; do
  [[ -z "$f" ]] && continue
  CONF=$(grep '^confidence:' "$f" 2>/dev/null | sed 's/confidence: *//' || echo "LOW")
  LEVEL=$(confidence_level "$CONF")
  if [[ $LEVEL -ge $MIN_LEVEL ]]; then
    FILTERED_FILES+=("$f")
  fi
done <<< "$PATTERN_FILES"

if [[ ${#FILTERED_FILES[@]} -eq 0 ]]; then
  echo "📭 No patterns match minimum confidence: $MIN_CONFIDENCE" >&2; exit 0
fi

CONTEXT_LABEL=""
[[ -n "$FILTER_CONTEXT" ]] && CONTEXT_LABEL=" (domain: $FILTER_CONTEXT)"
[[ -n "$FILTER_PROJECT" ]] && CONTEXT_LABEL=" (project: $FILTER_PROJECT)"

echo "📦 Exporting ${#FILTERED_FILES[@]} pattern(s)${CONTEXT_LABEL} (min: $MIN_CONFIDENCE, format: $FORMAT)..." >&2

generate_md() {
  echo "# Claude Code Learner — Exported Patterns"
  echo ""
  echo "Exported: $(date +%Y-%m-%d)"
  echo "Minimum confidence: $MIN_CONFIDENCE"
  [[ -n "$FILTER_CONTEXT" ]] && echo "Domain filter: $FILTER_CONTEXT"
  [[ -n "$FILTER_PROJECT" ]] && echo "Project filter: $FILTER_PROJECT"
  echo "Total patterns: ${#FILTERED_FILES[@]}"
  echo ""
  echo "---"
  echo ""

  for f in "${FILTERED_FILES[@]}"; do
    P_NAME=$(grep '^name:' "$f" | sed 's/name: *//')
    P_CONF=$(grep '^confidence:' "$f" | sed 's/confidence: *//')
    P_SCORE=$(grep '^score:' "$f" | sed 's/score: *//' || echo "N/A")
    P_CAT=$(grep '^category:' "$f" | sed 's/category: *//' || echo "N/A")
    P_DOM=$(grep '^domain:' "$f" | sed 's/domain: *//' || echo "N/A")
    P_TIMES=$(grep '^times_seen:' "$f" | sed 's/times_seen: *//' || echo "1")
    P_TITLE=$(grep '^# ' "$f" | head -1 | sed 's/^# //')
    P_PROBLEM=$(sed -n '/^## Problem/,/^##/{/^##/!p}' "$f" | sed '/^$/d')
    P_SOLUTION=$(sed -n '/^## Solution/,/^##/{/^##/!p}' "$f" | sed '/^$/d')

    echo "## $P_TITLE"
    echo ""
    echo "- **Name:** \`$P_NAME\`"
    echo "- **Confidence:** $P_CONF (score: $P_SCORE)"
    echo "- **Domain:** $P_DOM"
    echo "- **Category:** $P_CAT"
    echo "- **Times seen:** $P_TIMES"
    echo ""
    if [[ -n "$P_PROBLEM" ]]; then
      echo "### Problem"
      echo "$P_PROBLEM"
      echo ""
    fi
    if [[ -n "$P_SOLUTION" ]]; then
      echo "### Solution"
      echo "$P_SOLUTION"
      echo ""
    fi
    echo "---"
    echo ""
  done
}

generate_json() {
  python3 -c "
import json, re, sys, os

files = '''$(printf '%s\n' "${FILTERED_FILES[@]}")'''.strip().split('\n')
patterns = []

for f in files:
    f = f.strip()
    if not f or not os.path.isfile(f):
        continue
    with open(f) as fh:
        content = fh.read()

    def get_field(name):
        m = re.search(r'^' + name + r':\s*(.+)$', content, re.MULTILINE)
        return m.group(1).strip() if m else ''

    def get_section(name):
        m = re.search(r'^## ' + name + r'[^\n]*\n(.*?)(?=^## |\Z)', content, re.MULTILINE | re.DOTALL)
        return m.group(1).strip() if m else ''

    title_m = re.search(r'^# (.+)$', content, re.MULTILINE)
    title = title_m.group(1) if title_m else get_field('name')

    patterns.append({
        'name': get_field('name'),
        'title': title,
        'confidence': get_field('confidence'),
        'score': float(get_field('score') or 0),
        'domain': get_field('domain'),
        'category': get_field('category'),
        'project': get_field('project'),
        'times_seen': int(get_field('times_seen') or 1),
        'first_seen': get_field('first_seen'),
        'last_seen': get_field('last_seen'),
        'problem': get_section('Problem'),
        'solution': get_section('Solution'),
    })

print(json.dumps({
    'exported': '$(date +%Y-%m-%d)',
    'min_confidence': '$MIN_CONFIDENCE',
    'domain_filter': '$FILTER_CONTEXT',
    'project_filter': '$FILTER_PROJECT',
    'count': len(patterns),
    'patterns': patterns
}, indent=2, ensure_ascii=False))
"
}

if [[ "$FORMAT" == "md" ]]; then
  if [[ -n "$OUTPUT" ]]; then
    generate_md > "$OUTPUT"
    echo "✅ Exported to $OUTPUT" >&2
  else
    generate_md
  fi
elif [[ "$FORMAT" == "json" ]]; then
  if [[ -n "$OUTPUT" ]]; then
    generate_json > "$OUTPUT"
    echo "✅ Exported to $OUTPUT" >&2
  else
    generate_json
  fi
fi
