#!/usr/bin/env bash
# extract_patterns.sh — Extract learnable patterns from conversations/transcripts
# Usage:
#   echo "conversation text" | ./scripts/extract_patterns.sh
#   ./scripts/extract_patterns.sh -f /path/to/transcript.txt
#   ./scripts/extract_patterns.sh -f transcript.txt -s "session" --context devops --category debug
#   ./scripts/extract_patterns.sh -f transcript.txt --project twdd-api --category security
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
LEARNED_DIR="$SKILL_DIR/learned"
TEMPLATE="$SKILL_DIR/templates/pattern-template.md"

# --- Parse args ---
INPUT_FILE=""
SESSION_NAME="unknown"
CONTEXT=""
PROJECT=""
CATEGORY=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    -f|--file)      INPUT_FILE="$2"; shift 2 ;;
    -s|--session)   SESSION_NAME="$2"; shift 2 ;;
    --context)      CONTEXT="$2"; shift 2 ;;
    --project)      PROJECT="$2"; shift 2 ;;
    --category)     CATEGORY="$2"; shift 2 ;;
    -h|--help)
      echo "Usage: extract_patterns.sh [-f FILE] [-s SESSION] [--context DOMAIN] [--project PROJECT] [--category CAT]"
      echo ""
      echo "Extract learnable patterns from a conversation or transcript."
      echo ""
      echo "Options:"
      echo "  -f, --file FILE       Read input from file (default: stdin)"
      echo "  -s, --session NAME    Session name for source tracking"
      echo "  --context DOMAIN      Domain: _global, devops, web, mobile, backend, data"
      echo "  --project PROJECT     Project name (saves under learned/projects/{project}/)"
      echo "  --category CATEGORY   Category: debug, pattern, style, security, performance, ops"
      echo "  -h, --help            Show this help"
      exit 0
      ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

# --- Validate context/category ---
VALID_DOMAINS="_global devops web mobile backend data"
VALID_CATEGORIES="debug pattern style security performance ops"

if [[ -n "$CONTEXT" ]] && ! echo "$VALID_DOMAINS" | grep -qw "$CONTEXT"; then
  echo "Error: Invalid context '$CONTEXT'. Valid: $VALID_DOMAINS" >&2; exit 1
fi
if [[ -n "$CATEGORY" ]] && ! echo "$VALID_CATEGORIES" | grep -qw "$CATEGORY"; then
  echo "Error: Invalid category '$CATEGORY'. Valid: $VALID_CATEGORIES" >&2; exit 1
fi

# --- Read input ---
if [[ -n "$INPUT_FILE" ]]; then
  if [[ ! -f "$INPUT_FILE" ]]; then
    echo "Error: File not found: $INPUT_FILE" >&2; exit 1
  fi
  INPUT=$(cat "$INPUT_FILE")
else
  if [[ -t 0 ]]; then
    echo "Error: No input. Pipe text or use -f FILE." >&2; exit 1
  fi
  INPUT=$(cat)
fi

if [[ -z "$INPUT" ]]; then
  echo "Error: Empty input." >&2; exit 1
fi

# --- Today's date (macOS compatible) ---
TODAY=$(date +%Y-%m-%d)

# --- Build context hint for AI ---
CONTEXT_HINT=""
if [[ -n "$CONTEXT" ]]; then
  CONTEXT_HINT="Domain hint: $CONTEXT. "
fi
if [[ -n "$PROJECT" ]]; then
  CONTEXT_HINT+="Project: $PROJECT. "
fi
if [[ -n "$CATEGORY" ]]; then
  CONTEXT_HINT+="Category hint: $CATEGORY. "
fi

# --- Analysis prompt ---
ANALYSIS_PROMPT="You are a learning pattern extractor. ${CONTEXT_HINT}Analyse the following conversation/transcript and extract NON-OBVIOUS patterns worth remembering.

QUALITY THRESHOLD — Only extract patterns that:
- Required actual \"discovery\" or debugging to figure out
- Are NOT easily found by reading official documentation
- Represent workarounds, gotchas, or hard-won knowledge
- Would save significant time if known in advance

For each pattern found, output a JSON array. Each element:
{
  \"name\": \"kebab-case-slug\",
  \"title\": \"Human Readable Title\",
  \"confidence\": \"HIGH|MEDIUM|LOW\",
  \"score\": 0.0-1.0,
  \"category\": \"debug|pattern|style|security|performance|ops\",
  \"domain\": \"_global|devops|web|mobile|backend|data\",
  \"project\": \"project-name or empty string\",
  \"problem\": \"What situation triggers this\",
  \"solution\": \"The concrete solution or workaround\",
  \"verification\": \"How to confirm it works\",
  \"why_non_obvious\": \"Why this is not trivially discoverable\"
}

Confidence levels:
- HIGH: Verified fix that corrected a real problem
- MEDIUM: Confirmed useful approach
- LOW: Observed pattern, needs more validation

If NO patterns meet the quality threshold, output an empty array: []

IMPORTANT: Output ONLY the JSON array, no other text.

---
TRANSCRIPT:
"

# --- Call claude -p for analysis ---
echo "🔍 Analysing input for learnable patterns..." >&2

RESULT=$(printf '%s\n%s' "$ANALYSIS_PROMPT" "$INPUT" | claude -p --output-format json 2>/dev/null) || {
  echo "Error: claude -p failed. Is Claude Code installed and authenticated?" >&2
  exit 1
}

# --- Parse JSON result ---
if ! echo "$RESULT" | grep -q '^\['; then
  RESULT=$(echo "$RESULT" | sed -n '/^\[/,/^\]/p')
  if [[ -z "$RESULT" ]]; then
    echo "⚠️  No valid JSON output from analysis." >&2
    exit 1
  fi
fi

PATTERN_COUNT=$(echo "$RESULT" | grep -c '"name"' || true)

if [[ "$PATTERN_COUNT" -eq 0 ]]; then
  echo "✅ No patterns met the quality threshold. Nothing to save." >&2
  exit 0
fi

echo "📝 Found $PATTERN_COUNT pattern(s). Saving..." >&2

# --- Helper: determine output path ---
determine_output_dir() {
  local domain="$1" category="$2" project="$3"

  # Override with CLI args if provided
  [[ -n "$CONTEXT" ]] && domain="$CONTEXT"
  [[ -n "$CATEGORY" ]] && category="$CATEGORY"
  [[ -n "$PROJECT" ]] && project="$PROJECT"

  # Defaults
  [[ -z "$domain" ]] && domain="_global"
  [[ -z "$category" ]] && category="pattern"

  if [[ -n "$project" ]]; then
    echo "$LEARNED_DIR/projects/${project}/${category}"
  else
    echo "$LEARNED_DIR/${domain}/${category}"
  fi
}

# --- Extract and save each pattern ---
INDEX=0
while [[ $INDEX -lt $PATTERN_COUNT ]]; do
  BLOCK=$(echo "$RESULT" | python3 -c "
import json, sys
data = json.load(sys.stdin)
if $INDEX < len(data):
    import json as j
    print(j.dumps(data[$INDEX]))
" 2>/dev/null) || {
    echo "Warning: Could not parse pattern $INDEX, skipping." >&2
    INDEX=$((INDEX + 1))
    continue
  }

  # Extract fields
  P_NAME=$(echo "$BLOCK" | python3 -c "import json,sys; print(json.load(sys.stdin).get('name','unknown'))")
  P_TITLE=$(echo "$BLOCK" | python3 -c "import json,sys; print(json.load(sys.stdin).get('title','Unknown Pattern'))")
  P_CONFIDENCE=$(echo "$BLOCK" | python3 -c "import json,sys; print(json.load(sys.stdin).get('confidence','LOW'))")
  P_SCORE=$(echo "$BLOCK" | python3 -c "import json,sys; print(json.load(sys.stdin).get('score',0.5))")
  P_CATEGORY=$(echo "$BLOCK" | python3 -c "import json,sys; print(json.load(sys.stdin).get('category','pattern'))")
  P_DOMAIN=$(echo "$BLOCK" | python3 -c "import json,sys; print(json.load(sys.stdin).get('domain','_global'))")
  P_PROJECT=$(echo "$BLOCK" | python3 -c "import json,sys; print(json.load(sys.stdin).get('project',''))")
  P_PROBLEM=$(echo "$BLOCK" | python3 -c "import json,sys; print(json.load(sys.stdin).get('problem',''))")
  P_SOLUTION=$(echo "$BLOCK" | python3 -c "import json,sys; print(json.load(sys.stdin).get('solution',''))")
  P_VERIFICATION=$(echo "$BLOCK" | python3 -c "import json,sys; print(json.load(sys.stdin).get('verification',''))")

  # Determine output directory
  OUT_DIR=$(determine_output_dir "$P_DOMAIN" "$P_CATEGORY" "$P_PROJECT")
  mkdir -p "$OUT_DIR"

  # Check if pattern already exists (dedup by name across all learned/)
  EXISTING=$(find "$LEARNED_DIR" -name "*.md" -exec grep -l "^name: ${P_NAME}$" {} \; 2>/dev/null | head -1 || true)

  if [[ -n "$EXISTING" ]]; then
    echo "  ♻️  Pattern '$P_NAME' already exists, updating times_seen and last_seen..." >&2
    CURRENT_TIMES=$(grep '^times_seen:' "$EXISTING" | sed 's/times_seen: *//' || echo "1")
    NEW_TIMES=$((CURRENT_TIMES + 1))
    sed -i '' "s/^last_seen: .*/last_seen: ${TODAY}/" "$EXISTING"
    sed -i '' "s/^times_seen: .*/times_seen: ${NEW_TIMES}/" "$EXISTING"
    if [[ $NEW_TIMES -ge 3 && "$P_CONFIDENCE" != "HIGH" ]]; then
      sed -i '' "s/^confidence: .*/confidence: HIGH/" "$EXISTING"
      echo "  ⬆️  Pattern '$P_NAME' upgraded to HIGH confidence (seen ${NEW_TIMES} times)" >&2
    fi
  else
    # Determine final domain/project for frontmatter
    FINAL_DOMAIN="${CONTEXT:-$P_DOMAIN}"
    FINAL_DOMAIN="${FINAL_DOMAIN:-_global}"
    FINAL_CATEGORY="${CATEGORY:-$P_CATEGORY}"
    FINAL_CATEGORY="${FINAL_CATEGORY:-pattern}"
    FINAL_PROJECT="${PROJECT:-$P_PROJECT}"

    OUTFILE="$OUT_DIR/${TODAY}-${P_NAME}.md"

    PROJECT_LINE=""
    if [[ -n "$FINAL_PROJECT" ]]; then
      PROJECT_LINE="project: ${FINAL_PROJECT}"
    fi

    cat > "$OUTFILE" <<EOF
---
name: ${P_NAME}
confidence: ${P_CONFIDENCE}
score: ${P_SCORE}
category: ${FINAL_CATEGORY}
domain: ${FINAL_DOMAIN}
${PROJECT_LINE}
first_seen: ${TODAY}
last_seen: ${TODAY}
times_seen: 1
---

# ${P_TITLE}

## Problem / Trigger
${P_PROBLEM}

## Solution
${P_SOLUTION}

## Verification
${P_VERIFICATION}

## Source
Session: ${SESSION_NAME}
EOF
    echo "  ✅ Saved: ${OUTFILE}" >&2
  fi

  INDEX=$((INDEX + 1))
done

echo "" >&2
echo "🎓 Done! $PATTERN_COUNT pattern(s) processed." >&2
