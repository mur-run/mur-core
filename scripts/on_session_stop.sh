#!/usr/bin/env bash
# on_session_stop.sh — Lightweight Stop hook for Claude Code
# Detects new pattern files and stages them for git.
# MUST complete in < 100ms. No claude -p calls.

LEARNED_DIR="$HOME/clawd/skills/mur-core/learned"
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

# --- Check for new spec artifacts (lightweight, < 100ms) ---
REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || true)
if [ -n "$REPO_ROOT" ]; then
    SPEC_HINT=false
    # Check for new files in openspec/changes/ or .spec/
    if [ -d "$REPO_ROOT/openspec/changes" ]; then
        NEW_SPECS=$(find "$REPO_ROOT/openspec/changes" -name "*.md" -newer "$LAST_CHECK" -type f 2>/dev/null | head -1)
        [ -n "$NEW_SPECS" ] && SPEC_HINT=true
    fi
    if [ "$SPEC_HINT" = "false" ] && [ -d "$REPO_ROOT/.spec" ]; then
        NEW_SPECS=$(find "$REPO_ROOT/.spec" -name "*.md" -newer "$LAST_CHECK" -type f 2>/dev/null | head -1)
        [ -n "$NEW_SPECS" ] && SPEC_HINT=true
    fi
    if [ "$SPEC_HINT" = "true" ]; then
        echo "[ContinuousLearning] New spec artifacts detected — run auto_learn.sh to extract patterns" >&2
    fi
fi
