#!/usr/bin/env bash
# on_session_stop.sh — Lightweight Stop hook for Claude Code
# Detects new pattern files and stages them for git.
# MUST complete in < 100ms. No claude -p calls.

LEARNED_DIR="$HOME/clawd/skills/claude-code-learner/learned"
LAST_CHECK="$LEARNED_DIR/.last_check"

# Ensure .last_check exists
[ -f "$LAST_CHECK" ] || touch "$LAST_CHECK"

# Find .md files newer than .last_check (excluding .gitkeep)
RECENT_FILES=$(find "$LEARNED_DIR" -name "*.md" -not -name ".gitkeep" -newer "$LAST_CHECK" 2>/dev/null | head -10)

if [ -n "$RECENT_FILES" ]; then
    # Update timestamp
    touch "$LAST_CHECK"
    # Log to stderr (visible in Claude Code)
    echo "[ContinuousLearning] New patterns detected" >&2
fi
