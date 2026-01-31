#!/usr/bin/env bash
set -euo pipefail

# privacy_filter.sh — Filter pattern files based on .learned-config.yaml privacy rules
#
# Reads privacy settings from .learned-config.yaml (no yq required).
# Filters OUT files that violate privacy rules:
#   - Files in learned/personal/ when share_personal: false
#   - Files containing exclude_keywords (password, api_key, secret, token)
#
# Usage:
#   find learned/ -name "*.md" | ./scripts/privacy_filter.sh
#   find learned/ -name "*.md" | ./scripts/privacy_filter.sh --verbose
#   ./scripts/privacy_filter.sh --check learned/backend/pattern/some-file.md
#
# Exit codes (--check mode):
#   0 = file is safe to share
#   1 = file is filtered out (privacy violation)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
CONFIG_FILE="$REPO_DIR/.learned-config.yaml"

VERBOSE=false
CHECK_FILE=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --verbose|-v) VERBOSE=true; shift ;;
        --check)      CHECK_FILE="$2"; shift 2 ;;
        -h|--help)
            echo "Usage: privacy_filter.sh [--verbose] [--check FILE]"
            echo ""
            echo "Filter pattern files based on .learned-config.yaml privacy rules."
            echo ""
            echo "Pipe mode (default):"
            echo "  find learned/ -name '*.md' | ./scripts/privacy_filter.sh"
            echo "  Reads file paths from stdin, outputs safe paths to stdout."
            echo ""
            echo "Check mode:"
            echo "  ./scripts/privacy_filter.sh --check FILE"
            echo "  Exit 0 if safe, exit 1 if filtered."
            echo ""
            echo "Options:"
            echo "  --verbose, -v   Show what was filtered and why (to stderr)"
            echo "  --check FILE    Check a single file"
            echo "  -h, --help      Show this help"
            exit 0
            ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

# ── Read config ───────────────────────────────────────────────────

# Default values
SHARE_PERSONAL=false
EXCLUDE_KEYWORDS=("password" "api_key" "secret" "token")

if [[ -f "$CONFIG_FILE" ]]; then
    # Read share_personal
    SP_VAL=$(grep 'share_personal:' "$CONFIG_FILE" 2>/dev/null | sed 's/.*share_personal: *//' | tr -d ' ' || echo "false")
    if [[ "$SP_VAL" == "true" ]]; then
        SHARE_PERSONAL=true
    fi

    # Read exclude_keywords (lines after exclude_keywords: that start with "    - ")
    CUSTOM_KEYWORDS=()
    IN_KEYWORDS=false
    while IFS= read -r line; do
        if echo "$line" | grep -q 'exclude_keywords:'; then
            IN_KEYWORDS=true
            continue
        fi
        if [[ "$IN_KEYWORDS" == true ]]; then
            if echo "$line" | grep -q '^\s*-'; then
                KW=$(echo "$line" | sed 's/^\s*- *//' | tr -d ' ')
                if [[ -n "$KW" ]]; then
                    CUSTOM_KEYWORDS+=("$KW")
                fi
            else
                IN_KEYWORDS=false
            fi
        fi
    done < "$CONFIG_FILE"

    if [[ ${#CUSTOM_KEYWORDS[@]} -gt 0 ]]; then
        EXCLUDE_KEYWORDS=("${CUSTOM_KEYWORDS[@]}")
    fi
fi

# ── Filter logic ──────────────────────────────────────────────────

# Check a single file. Returns 0 if safe, 1 if filtered.
# Sets FILTER_REASON on failure.
FILTER_REASON=""

check_file() {
    local filepath="$1"
    FILTER_REASON=""

    # Check if file exists
    if [[ ! -f "$filepath" ]]; then
        FILTER_REASON="file not found"
        return 1
    fi

    # Rule 1: Personal files when share_personal is false
    if [[ "$SHARE_PERSONAL" == false ]]; then
        if echo "$filepath" | grep -q 'learned/personal/'; then
            FILTER_REASON="personal file (share_personal: false)"
            return 1
        fi
    fi

    # Rule 2: Exclude keywords in content
    for kw in "${EXCLUDE_KEYWORDS[@]}"; do
        if grep -qi "$kw" "$filepath" 2>/dev/null; then
            FILTER_REASON="contains exclude keyword: $kw"
            return 1
        fi
    done

    return 0
}

# ── Single file check mode ────────────────────────────────────────

if [[ -n "$CHECK_FILE" ]]; then
    if check_file "$CHECK_FILE"; then
        if [[ "$VERBOSE" == true ]]; then
            echo "PASS: $CHECK_FILE" >&2
        fi
        exit 0
    else
        if [[ "$VERBOSE" == true ]]; then
            echo "FILTERED: $CHECK_FILE — $FILTER_REASON" >&2
        fi
        exit 1
    fi
fi

# ── Pipe mode: read file paths from stdin ─────────────────────────

TOTAL=0
PASSED=0
FILTERED=0

while IFS= read -r filepath; do
    [[ -z "$filepath" ]] && continue
    TOTAL=$((TOTAL + 1))

    if check_file "$filepath"; then
        echo "$filepath"
        PASSED=$((PASSED + 1))
    else
        FILTERED=$((FILTERED + 1))
        if [[ "$VERBOSE" == true ]]; then
            echo "FILTERED: $filepath — $FILTER_REASON" >&2
        fi
    fi
done

if [[ "$VERBOSE" == true ]]; then
    echo "" >&2
    echo "Privacy filter: $TOTAL total, $PASSED passed, $FILTERED filtered" >&2
fi
