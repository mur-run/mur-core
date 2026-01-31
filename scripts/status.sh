#!/usr/bin/env bash
# status.sh — Show status and statistics of learned patterns
# Usage:
#   ./scripts/status.sh            # Human-readable output
#   ./scripts/status.sh --json     # Machine-readable JSON output
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
LEARNED_DIR="$REPO_DIR/learned"
CLAUDE_SKILLS_DIR="$HOME/.claude/skills"

JSON_MODE=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --json) JSON_MODE=true; shift ;;
    -h|--help)
      echo "Usage: status.sh [--json]"
      echo ""
      echo "Show status and statistics of learned patterns."
      echo ""
      echo "Options:"
      echo "  --json     Output in JSON format"
      echo "  -h, --help Show this help"
      exit 0
      ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

# ── Collect all pattern files ─────────────────────────────────────
PATTERN_FILES=$(find "$LEARNED_DIR" -name "*.md" \
  -not -name ".gitkeep" -not -name "archived-*" \
  -not -name "skill-*" -not -name ".last_check" 2>/dev/null | sort)

TOTAL=0
HIGH=0; MEDIUM=0; LOW=0

# Temp files for aggregation (bash 3 compatible)
TMP_DOMAINS=$(mktemp)
TMP_TOPSEEN=$(mktemp)
TMP_PROMO=$(mktemp)

while IFS= read -r f; do
  [[ -z "$f" ]] && continue
  TOTAL=$((TOTAL + 1))

  CONF=$(grep '^confidence:' "$f" 2>/dev/null | sed 's/confidence: *//' | head -1)
  DOM=$(grep '^domain:' "$f" 2>/dev/null | sed 's/domain: *//' | head -1)
  TIMES=$(grep '^times_seen:' "$f" 2>/dev/null | sed 's/times_seen: *//' | head -1)
  NAME=$(grep '^name:' "$f" 2>/dev/null | sed 's/name: *//' | head -1)
  DOM="${DOM:-_global}"
  TIMES="${TIMES:-1}"

  case "$CONF" in
    HIGH)   HIGH=$((HIGH + 1)) ;;
    MEDIUM) MEDIUM=$((MEDIUM + 1)) ;;
    LOW)    LOW=$((LOW + 1)) ;;
  esac

  echo "$DOM" >> "$TMP_DOMAINS"
  echo "$TIMES $NAME ($CONF)" >> "$TMP_TOPSEEN"

  # Promotion candidates: MEDIUM with times_seen >= 2, or HIGH with times_seen >= 2 but < 3 (not yet synced threshold)
  if [[ "$CONF" == "MEDIUM" && "$TIMES" -ge 2 ]]; then
    echo "  $NAME — $CONF, seen ${TIMES}x (→ may upgrade to HIGH)" >> "$TMP_PROMO"
  elif [[ "$CONF" == "HIGH" && "$TIMES" -ge 2 && "$TIMES" -lt 3 ]]; then
    echo "  $NAME — $CONF, seen ${TIMES}x (→ close to sync threshold of 3)" >> "$TMP_PROMO"
  fi
done <<< "$PATTERN_FILES"

# ── Domain breakdown ──────────────────────────────────────────────
DOMAIN_BREAKDOWN=""
if [[ -s "$TMP_DOMAINS" ]]; then
  DOMAIN_BREAKDOWN=$(sort "$TMP_DOMAINS" | uniq -c | sort -rn)
fi

# ── Top 5 most-seen ──────────────────────────────────────────────
TOP5=""
if [[ -s "$TMP_TOPSEEN" ]]; then
  TOP5=$(sort -rn "$TMP_TOPSEEN" | head -5)
fi

# ── Promotion candidates ─────────────────────────────────────────
PROMO_COUNT=0
PROMO_LIST=""
if [[ -s "$TMP_PROMO" ]]; then
  PROMO_COUNT=$(wc -l < "$TMP_PROMO" | tr -d ' ')
  PROMO_LIST=$(cat "$TMP_PROMO")
fi

# ── Last sync time ────────────────────────────────────────────────
LAST_SYNC="unknown"
if [[ -d "$CLAUDE_SKILLS_DIR" ]]; then
  NEWEST=$(find "$CLAUDE_SKILLS_DIR" -name "SKILL.md" -newer "$LEARNED_DIR/.last_check" 2>/dev/null | head -1)
  if [[ -n "$NEWEST" ]]; then
    if [[ "$(uname -s)" == "Darwin" ]]; then
      LAST_SYNC=$(stat -f '%Sm' -t '%Y-%m-%d %H:%M' "$NEWEST" 2>/dev/null || echo "unknown")
    else
      LAST_SYNC=$(stat -c '%y' "$NEWEST" 2>/dev/null | cut -d'.' -f1 || echo "unknown")
    fi
  else
    # Fall back to any learned-* skill
    ANY_SKILL=$(find "$CLAUDE_SKILLS_DIR" -path "*/learned-*/SKILL.md" 2>/dev/null | head -1)
    if [[ -n "$ANY_SKILL" ]]; then
      if [[ "$(uname -s)" == "Darwin" ]]; then
        LAST_SYNC=$(stat -f '%Sm' -t '%Y-%m-%d %H:%M' "$ANY_SKILL" 2>/dev/null || echo "unknown")
      else
        LAST_SYNC=$(stat -c '%y' "$ANY_SKILL" 2>/dev/null | cut -d'.' -f1 || echo "unknown")
      fi
    else
      LAST_SYNC="never"
    fi
  fi
fi

# ── Last auto_learn run ──────────────────────────────────────────
LAST_AUTOLEARN="unknown"
AL_COMMIT=$(git -C "$HOME/clawd" log --oneline --grep="auto-learn" -1 --format="%ai" 2>/dev/null || true)
if [[ -n "$AL_COMMIT" ]]; then
  LAST_AUTOLEARN=$(echo "$AL_COMMIT" | cut -d' ' -f1,2 | cut -d'+' -f1)
else
  LAST_AUTOLEARN="never"
fi

# Cleanup
rm -f "$TMP_DOMAINS" "$TMP_TOPSEEN" "$TMP_PROMO"

# ── Output ────────────────────────────────────────────────────────
if [[ "$JSON_MODE" == "true" ]]; then
  python3 -c "
import json
domain_lines = '''$DOMAIN_BREAKDOWN'''.strip().split('\n')
domains = {}
for line in domain_lines:
    line = line.strip()
    if line:
        parts = line.split()
        if len(parts) >= 2:
            domains[parts[1]] = int(parts[0])

top5_lines = '''$TOP5'''.strip().split('\n')
top5 = []
for line in top5_lines:
    line = line.strip()
    if line:
        parts = line.split(' ', 1)
        if len(parts) >= 2:
            top5.append({'times_seen': int(parts[0]), 'pattern': parts[1]})

print(json.dumps({
    'total_patterns': $TOTAL,
    'confidence': {'HIGH': $HIGH, 'MEDIUM': $MEDIUM, 'LOW': $LOW},
    'domains': domains,
    'top_5_most_seen': top5,
    'promotion_candidates': $PROMO_COUNT,
    'last_sync': '$LAST_SYNC',
    'last_auto_learn': '$LAST_AUTOLEARN'
}, indent=2))
"
  exit 0
fi

echo "🧠 Claude Code Learner — Status"
echo "================================"
echo ""
echo "📊 Total patterns: $TOTAL"
echo ""
echo "🎯 By Confidence:"
echo "   HIGH:   $HIGH"
echo "   MEDIUM: $MEDIUM"
echo "   LOW:    $LOW"
echo ""
echo "📁 By Domain:"
if [[ -n "$DOMAIN_BREAKDOWN" ]]; then
  echo "$DOMAIN_BREAKDOWN" | while read count domain; do
    echo "   $domain: $count"
  done
else
  echo "   (none)"
fi
echo ""
echo "🔝 Top 5 Most-Seen Patterns:"
if [[ -n "$TOP5" ]]; then
  echo "$TOP5" | while read times rest; do
    echo "   ${times}x — $rest"
  done
else
  echo "   (none)"
fi
echo ""
echo "⏫ Promotion Candidates ($PROMO_COUNT):"
if [[ -n "$PROMO_LIST" ]]; then
  echo "$PROMO_LIST"
else
  echo "   (none)"
fi
echo ""
echo "🔄 Last sync: $LAST_SYNC"
echo "🤖 Last auto_learn: $LAST_AUTOLEARN"
echo "================================"
