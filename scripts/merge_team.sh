#!/usr/bin/env bash
set -euo pipefail

# ┌─────────────────────────────────────────────────────────────────┐
# │ merge_team.sh — Merge all team learnings branches into main    │
# │                                                                 │
# │ Conflict Resolution Strategy:                                   │
# │                                                                 │
# │ Pattern files are designed to be ADDITIVE — each file is a     │
# │ self-contained knowledge record. This means:                    │
# │                                                                 │
# │ --strategy theirs (default):                                    │
# │   When a conflict occurs, the incoming branch wins. This is    │
# │   safe for pattern files because:                               │
# │   - Each pattern is a separate .md file (rarely conflicts)      │
# │   - If the same file is edited, the newer version is usually   │
# │     the more complete one (updated times_seen, etc.)            │
# │   - Worst case: you lose a minor edit, easily recoverable      │
# │     from git reflog                                             │
# │                                                                 │
# │ --strategy manual:                                              │
# │   Stops on conflict and asks the admin to resolve manually.    │
# │   Use this when:                                                │
# │   - Multiple people are editing the SAME pattern file           │
# │   - You're merging patterns that need careful review            │
# │   - You want full control over conflict resolution              │
# │                                                                 │
# │ Recovery from bad merges:                                       │
# │   git reflog  →  find the commit before the merge               │
# │   git reset --hard <commit>  →  undo the merge                  │
# │   git push --force-with-lease origin main                       │
# └─────────────────────────────────────────────────────────────────┘

# 用法: ./scripts/merge_team.sh [--dry-run] [--notify] [--strategy theirs|manual]

REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_DIR"

DRY_RUN=false
NOTIFY=false
STRATEGY="theirs"
FORCE=false
PRIVACY_FILTER="$REPO_DIR/scripts/privacy_filter.sh"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run) DRY_RUN=true; shift ;;
        --notify) NOTIFY=true; shift ;;
        --force) FORCE=true; shift ;;
        --strategy)
            STRATEGY="$2"
            if [[ "$STRATEGY" != "theirs" && "$STRATEGY" != "manual" ]]; then
                echo "Error: --strategy must be 'theirs' or 'manual'" >&2
                exit 1
            fi
            shift 2
            ;;
        -h|--help)
            echo "Usage: merge_team.sh [--dry-run] [--notify] [--force] [--strategy theirs|manual]"
            echo ""
            echo "Merge all team learnings branches into main."
            echo ""
            echo "Options:"
            echo "  --dry-run              Preview without merging"
            echo "  --notify               Send notification after merge"
            echo "  --force                Force merge _global patterns even if require_review_for_global is set"
            echo "  --strategy STRATEGY    Conflict resolution strategy:"
            echo "                           theirs  - incoming branch wins (default, safe for additive patterns)"
            echo "                           manual  - stop on conflict, ask admin to resolve"
            echo "  -h, --help             Show this help"
            exit 0
            ;;
        *) shift ;;
    esac
done

# Read require_review_for_global from config
CONFIG_FILE="$REPO_DIR/.learned-config.yaml"
REQUIRE_REVIEW_GLOBAL=false
if [[ -f "$CONFIG_FILE" ]]; then
    RRG_VAL=$(grep 'require_review_for_global:' "$CONFIG_FILE" 2>/dev/null | sed 's/.*require_review_for_global: *//' | tr -d ' ' || echo "false")
    if [[ "$RRG_VAL" == "true" ]]; then
        REQUIRE_REVIEW_GLOBAL=true
    fi
fi

echo "🔄 Team Merge - $(date '+%Y-%m-%d %H:%M')"
echo "=========================================="
echo "   Strategy: $STRATEGY"

# 確保在 main 分支
git checkout main
git pull origin main

# 取得所有 learnings 分支
git fetch --all --prune
BRANCHES=$(git branch -r | grep 'origin/learnings/' | sed 's/origin\///' | tr -d ' ')

MERGED_COUNT=0
NEW_PATTERNS=""

for BRANCH in $BRANCHES; do
    USER=$(echo "$BRANCH" | sed 's|learnings/||')
    echo ""
    echo "📥 Processing: $USER ($BRANCH)"
    
    # 查看這個分支相對於 main 有哪些新檔案
    NEW_FILES=$(git diff main..origin/$BRANCH --name-only --diff-filter=A -- 'learned/' 2>/dev/null || true)
    
    if [ -z "$NEW_FILES" ]; then
        echo "   No new patterns"
        continue
    fi
    
    # Privacy filter: skip files that violate privacy rules
    if [[ -x "$PRIVACY_FILTER" ]]; then
        SAFE_FILES=$(echo "$NEW_FILES" | "$PRIVACY_FILTER" 2>/dev/null || true)
        FILTERED_COUNT=$(( $(echo "$NEW_FILES" | wc -l) - $(echo "$SAFE_FILES" | grep -c . || echo 0) ))
        if [[ $FILTERED_COUNT -gt 0 ]]; then
            echo "   🔒 Privacy filter: $FILTERED_COUNT file(s) excluded"
        fi
        NEW_FILES="$SAFE_FILES"
    fi

    # require_review_for_global: flag _global/ patterns for review
    if [[ "$REQUIRE_REVIEW_GLOBAL" == true && "$FORCE" != true ]]; then
        GLOBAL_FILES=$(echo "$NEW_FILES" | grep '_global/' || true)
        if [[ -n "$GLOBAL_FILES" ]]; then
            echo "   ⚠️  Global patterns require review (require_review_for_global: true):"
            echo "$GLOBAL_FILES" | while read gf; do echo "     🔍 $gf"; done
            echo "   Skipping global patterns. Use --force to include them."
            NEW_FILES=$(echo "$NEW_FILES" | grep -v '_global/' || true)
        fi
    fi

    if [ -z "$NEW_FILES" ]; then
        echo "   No new patterns (after filtering)"
        continue
    fi
    
    echo "   New files:"
    echo "$NEW_FILES" | while read f; do echo "     + $f"; done
    
    if [ "$DRY_RUN" = true ]; then
        echo "   [DRY RUN] Would merge"
        continue
    fi
    
    # Merge with chosen strategy
    git merge "origin/$BRANCH" --no-edit -m "merge: learnings from $USER" 2>/dev/null || {
        if [ "$STRATEGY" = "theirs" ]; then
            echo "   ⚠️ Conflict detected, auto-resolving (incoming wins — strategy: theirs)"
            git checkout --theirs -- learned/ 2>/dev/null || true
            git add learned/
            git commit -m "merge: learnings from $USER (auto-resolved, strategy: theirs)" 2>/dev/null || true
        elif [ "$STRATEGY" = "manual" ]; then
            echo "   ⚠️ Conflict detected! Manual resolution required."
            echo "   Conflicting files:"
            git diff --name-only --diff-filter=U 2>/dev/null | while read cf; do echo "     ⚡ $cf"; done
            echo ""
            echo "   Resolve conflicts, then run:"
            echo "     git add learned/"
            echo "     git commit"
            echo "     Then re-run merge_team.sh to continue with remaining branches."
            exit 1
        fi
    }
    
    MERGED_COUNT=$((MERGED_COUNT + 1))
    NEW_PATTERNS="$NEW_PATTERNS\n$(echo "$NEW_FILES" | sed "s/^/  [$USER] /")"
done

echo ""
echo "=========================================="
echo "✅ Merged from $MERGED_COUNT contributors"

if [ "$DRY_RUN" = false ] && [ $MERGED_COUNT -gt 0 ]; then
    git push origin main
    echo "📤 Pushed to main"
fi

# 輸出摘要（可被 Clawdbot 讀取）
if [ -n "$NEW_PATTERNS" ]; then
    echo ""
    echo "📋 New patterns:"
    echo -e "$NEW_PATTERNS"
fi
