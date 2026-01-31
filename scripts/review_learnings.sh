#!/usr/bin/env bash
# review_learnings.sh — Review and consolidate learned patterns
# Usage:
#   ./scripts/review_learnings.sh                          # Review all
#   ./scripts/review_learnings.sh --auto                   # Auto-apply suggestions
#   ./scripts/review_learnings.sh --report                 # Report only
#   ./scripts/review_learnings.sh --context devops         # Filter by domain
#   ./scripts/review_learnings.sh --project twdd-api       # Filter by project
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
LEARNED_DIR="$SKILL_DIR/learned"

# --- Parse args ---
AUTO_APPLY=false
REPORT_ONLY=false
FILTER_CONTEXT=""
FILTER_PROJECT=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --auto)     AUTO_APPLY=true; shift ;;
    --report)   REPORT_ONLY=true; shift ;;
    --context)  FILTER_CONTEXT="$2"; shift 2 ;;
    --project)  FILTER_PROJECT="$2"; shift 2 ;;
    -h|--help)
      echo "Usage: review_learnings.sh [--auto] [--report] [--context DOMAIN] [--project PROJECT]"
      echo ""
      echo "Review and consolidate learned patterns."
      echo ""
      echo "Options:"
      echo "  --auto             Auto-apply merge/upgrade suggestions"
      echo "  --report           Output report only, no changes"
      echo "  --context DOMAIN   Filter by domain (_global, devops, web, mobile, backend, data)"
      echo "  --project PROJECT  Filter by project name"
      echo "  -h, --help         Show this help"
      exit 0
      ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

# --- Build find paths ---
FIND_PATHS=()
if [[ -n "$FILTER_PROJECT" ]]; then
  FIND_PATHS=("$LEARNED_DIR/projects/$FILTER_PROJECT")
elif [[ -n "$FILTER_CONTEXT" ]]; then
  FIND_PATHS=("$LEARNED_DIR/$FILTER_CONTEXT")
else
  FIND_PATHS=("$LEARNED_DIR")
fi

# Verify paths exist
for p in "${FIND_PATHS[@]}"; do
  if [[ ! -d "$p" ]]; then
    echo "📭 Directory not found: $p" >&2
    exit 0
  fi
done

# --- Collect pattern files ---
PATTERN_FILES=""
for p in "${FIND_PATHS[@]}"; do
  FILES=$(find "$p" -name "*.md" -not -name ".gitkeep" -not -name "archived-*" -not -name "skill-*" 2>/dev/null | sort)
  PATTERN_FILES+="$FILES"$'\n'
done
PATTERN_FILES=$(echo "$PATTERN_FILES" | sed '/^$/d')
PATTERN_COUNT=$(echo "$PATTERN_FILES" | grep -c '\.md$' || true)

if [[ "$PATTERN_COUNT" -eq 0 ]]; then
  echo "📭 No patterns found."
  exit 0
fi

echo "📚 Reviewing $PATTERN_COUNT pattern(s)..." >&2

# --- Quick Stats with 3D distribution ---
echo "" >&2
echo "📊 Quick Stats:" >&2
HIGH_COUNT=0; MEDIUM_COUNT=0; LOW_COUNT=0

# Domain distribution
declare -A DOMAIN_COUNTS 2>/dev/null || true
declare -A CATEGORY_COUNTS 2>/dev/null || true

while IFS= read -r f; do
  [[ -z "$f" ]] && continue
  CONF=$(grep '^confidence:' "$f" 2>/dev/null | sed 's/confidence: *//' || echo "")
  case "$CONF" in
    HIGH)   HIGH_COUNT=$((HIGH_COUNT + 1)) ;;
    MEDIUM) MEDIUM_COUNT=$((MEDIUM_COUNT + 1)) ;;
    LOW)    LOW_COUNT=$((LOW_COUNT + 1)) ;;
  esac

  DOM=$(grep '^domain:' "$f" 2>/dev/null | sed 's/domain: *//' || echo "_global")
  CAT=$(grep '^category:' "$f" 2>/dev/null | sed 's/category: *//' || echo "pattern")

  # Count domains and categories (bash 3 compatible — use files)
  echo "$DOM" >> /tmp/ccl_review_domains_$$
  echo "$CAT" >> /tmp/ccl_review_cats_$$
done <<< "$PATTERN_FILES"

echo "   Confidence: HIGH=$HIGH_COUNT | MEDIUM=$MEDIUM_COUNT | LOW=$LOW_COUNT" >&2

# Domain distribution
echo "" >&2
echo "   📁 Domain Distribution:" >&2
if [[ -f /tmp/ccl_review_domains_$$ ]]; then
  sort /tmp/ccl_review_domains_$$ | uniq -c | sort -rn | while read count domain; do
    echo "      $domain: $count" >&2
  done
  rm -f /tmp/ccl_review_domains_$$
fi

# Category distribution
echo "" >&2
echo "   🏷️  Category Distribution:" >&2
if [[ -f /tmp/ccl_review_cats_$$ ]]; then
  sort /tmp/ccl_review_cats_$$ | uniq -c | sort -rn | while read count cat; do
    echo "      $cat: $count" >&2
  done
  rm -f /tmp/ccl_review_cats_$$
fi

# Frequently seen patterns
echo "" >&2
echo "🔁 Frequently seen patterns:" >&2
while IFS= read -r f; do
  [[ -z "$f" ]] && continue
  TIMES=$(grep '^times_seen:' "$f" | sed 's/times_seen: *//' || echo "1")
  if [[ "$TIMES" -gt 1 ]]; then
    NAME=$(grep '^name:' "$f" | sed 's/name: *//')
    CONF=$(grep '^confidence:' "$f" | sed 's/confidence: *//')
    DOM=$(grep '^domain:' "$f" | sed 's/domain: *//' || echo "")
    echo "   - $NAME (seen ${TIMES}x, $CONF, domain: $DOM)" >&2
  fi
done <<< "$PATTERN_FILES"

if [[ "$REPORT_ONLY" == "true" ]]; then
  echo "" >&2
  echo "📋 Report mode — listing all patterns:" >&2
  echo "" >&2
  while IFS= read -r f; do
    [[ -z "$f" ]] && continue
    NAME=$(grep '^name:' "$f" | sed 's/name: *//')
    CONF=$(grep '^confidence:' "$f" | sed 's/confidence: *//')
    SCORE=$(grep '^score:' "$f" | sed 's/score: *//')
    CAT=$(grep '^category:' "$f" | sed 's/category: *//')
    DOM=$(grep '^domain:' "$f" | sed 's/domain: *//' || echo "")
    TIMES=$(grep '^times_seen:' "$f" | sed 's/times_seen: *//' || echo "1")
    TITLE=$(grep '^# ' "$f" | head -1 | sed 's/^# //')
    echo "  [$CONF] $NAME (score: $SCORE, seen: ${TIMES}x, domain: $DOM, cat: $CAT)"
    echo "         $TITLE"
  done <<< "$PATTERN_FILES"
  echo ""
  exit 0
fi

# --- Gather all patterns into one text block ---
ALL_PATTERNS=""
while IFS= read -r f; do
  [[ -z "$f" ]] && continue
  ALL_PATTERNS+="
=== FILE: $(basename "$f") ===
$(cat "$f")
"
done <<< "$PATTERN_FILES"

# --- Use claude -p for deep analysis ---
REVIEW_PROMPT='You are a learning pattern reviewer. Analyse these patterns and provide actionable recommendations.

For each recommendation, output a JSON array with objects like:
{
  "action": "merge|archive|upgrade|keep",
  "targets": ["pattern-slug-1", "pattern-slug-2"],
  "reason": "Why this action",
  "details": "Additional details or merged content"
}

Actions:
- merge: Two or more patterns are about the same thing, combine them
- archive: Pattern is outdated, too trivial, or superseded
- upgrade: Pattern has enough confidence/evidence to become a Claude Code skill
- keep: Pattern is fine as-is

Upgrade criteria:
- confidence: HIGH
- times_seen >= 2
- The knowledge is genuinely useful and actionable

Be conservative. Only suggest actions you are confident about.
Output ONLY the JSON array.

---
PATTERNS:
'

echo "" >&2
echo "🤖 Running deep analysis with Claude..." >&2

RESULT=$(printf '%s\n%s' "$REVIEW_PROMPT" "$ALL_PATTERNS" | claude -p --output-format json 2>/dev/null) || {
  echo "Error: claude -p failed." >&2
  exit 1
}

REC_COUNT=$(echo "$RESULT" | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    print(len(data))
except:
    print(0)
" 2>/dev/null || echo "0")

if [[ "$REC_COUNT" -eq 0 ]]; then
  echo "✅ All patterns look good. No changes recommended." >&2
  exit 0
fi

echo "" >&2
echo "💡 $REC_COUNT recommendation(s):" >&2
echo "" >&2

echo "$RESULT" | python3 -c "
import json, sys
data = json.load(sys.stdin)
for i, rec in enumerate(data):
    action = rec.get('action', '?')
    targets = ', '.join(rec.get('targets', []))
    reason = rec.get('reason', '')
    icon = {'merge': '🔀', 'archive': '🗄️', 'upgrade': '⬆️', 'keep': '✅'}.get(action, '❓')
    print(f'  {icon} [{action.upper()}] {targets}')
    print(f'     {reason}')
    print()
" 2>/dev/null

# --- Auto-apply if requested ---
if [[ "$AUTO_APPLY" == "true" ]]; then
  echo "🔧 Auto-applying recommendations..." >&2
  
  echo "$RESULT" | python3 -c "
import json, sys, os, glob

data = json.load(sys.stdin)
learned_dir = '$LEARNED_DIR'

def find_pattern_file(slug):
    for root, dirs, files in os.walk(learned_dir):
        for f in files:
            if f.endswith('.md') and slug in f:
                return os.path.join(root, f)
    return None

for rec in data:
    action = rec.get('action', '')
    targets = rec.get('targets', [])
    
    if action == 'archive':
        for slug in targets:
            path = find_pattern_file(slug)
            if path:
                dirname = os.path.dirname(path)
                archived = os.path.join(dirname, 'archived-' + os.path.basename(path))
                os.rename(path, archived)
                print(f'  🗄️  Archived: {os.path.basename(path)}', file=sys.stderr)
    
    elif action == 'upgrade':
        for slug in targets:
            path = find_pattern_file(slug)
            if path:
                print(f'  ⬆️  UPGRADE CANDIDATE: {os.path.basename(path)}', file=sys.stderr)
                print(f'     → Will be synced by sync_to_claude_code.sh', file=sys.stderr)
" 2>/dev/null
  
  echo "" >&2
  echo "✅ Auto-apply complete." >&2
else
  echo "💡 Run with --auto to apply these recommendations." >&2
fi

echo "" >&2
echo "🎓 Review complete." >&2
