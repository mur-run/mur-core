#!/usr/bin/env bash
# check_spec_init.sh — Lightweight check if current project has OpenSpec initialized
# Called by UserPromptSubmit hook. Outputs hint to stderr if not initialized.
# Must be fast (< 50ms).

# Only check if we're in a git repo (likely a project)
git rev-parse --is-inside-work-tree &>/dev/null || exit 0

# Get project root
PROJECT_ROOT=$(git rev-parse --show-toplevel 2>/dev/null) || exit 0

# Skip if this IS the claude-code-learner repo itself
if [[ -f "$PROJECT_ROOT/SKILL.md" ]] && grep -q "claude-code-learner" "$PROJECT_ROOT/SKILL.md" 2>/dev/null; then
    exit 0
fi

# Skip if already initialized (openspec/ or .spec/ exists)
if [[ -d "$PROJECT_ROOT/openspec" ]] || [[ -d "$PROJECT_ROOT/.spec" ]]; then
    exit 0
fi

# Skip if we already reminded in this project (check marker file)
MARKER="$PROJECT_ROOT/.openspec_reminded"
if [[ -f "$MARKER" ]]; then
    # Only remind once per day
    if [[ "$(uname -s)" == "Darwin" ]]; then
        MARKER_AGE=$(( $(date +%s) - $(stat -f %m "$MARKER" 2>/dev/null || echo 0) ))
    else
        MARKER_AGE=$(( $(date +%s) - $(stat -c %Y "$MARKER" 2>/dev/null || echo 0) ))
    fi
    [[ $MARKER_AGE -lt 86400 ]] && exit 0
fi

# Check if openspec or specify is available
HAS_OPENSPEC=false
HAS_SPECKIT=false
command -v openspec &>/dev/null && HAS_OPENSPEC=true
[[ -f "$HOME/.npm-global/bin/openspec" ]] && HAS_OPENSPEC=true
command -v specify &>/dev/null && HAS_SPECKIT=true

if [[ "$HAS_OPENSPEC" == "false" ]] && [[ "$HAS_SPECKIT" == "false" ]]; then
    exit 0
fi

# Output the hint to stderr (Claude will see this)
cat >&2 <<'EOF'

[SpecInit] 🪲 This project doesn't have spec-driven development set up yet.
Consider initializing OpenSpec for structured development:

  To set up: Run `openspec init --tools claude` in the project root
  Or ask the user: "Would you like me to initialize OpenSpec for spec-driven development in this project?"

Benefits: proposal → specs → design → tasks → implement → archive
Each feature gets tracked with full traceability.

(This hint appears once per project per day)
EOF

# Create marker so we don't remind again today
touch "$MARKER" 2>/dev/null
