#!/usr/bin/env bash
set -euo pipefail

# validate.sh — Validate pattern files in learned/
#
# Checks YAML frontmatter, required fields, valid values, required sections,
# and filename conventions.
#
# Usage:
#   ./scripts/validate.sh                    # Validate all patterns
#   ./scripts/validate.sh --fix              # Auto-fix common issues
#   ./scripts/validate.sh --json             # Machine-readable output
#   ./scripts/validate.sh learned/some.md    # Validate specific file(s)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
LEARNED_DIR="$REPO_DIR/learned"

FIX=false
JSON_OUTPUT=false
FILES=()

while [[ $# -gt 0 ]]; do
    case "$1" in
        --fix)   FIX=true; shift ;;
        --json)  JSON_OUTPUT=true; shift ;;
        -h|--help)
            echo "Usage: validate.sh [--fix] [--json] [FILE ...]"
            echo ""
            echo "Validate pattern .md files in learned/ directory."
            echo ""
            echo "Checks:"
            echo "  1. YAML frontmatter (opens/closes with ---)"
            echo "  2. Required fields: name, confidence, category, domain, first_seen, last_seen, times_seen"
            echo "  3. Valid confidence: HIGH, MEDIUM, LOW"
            echo "  4. Valid domain: _global, devops, web, mobile, backend, data"
            echo "  5. Valid category: debug, pattern, style, security, performance, ops"
            echo "  6. Date format: YYYY-MM-DD"
            echo "  7. times_seen: numeric >= 1"
            echo "  8. Required sections: # Title, ## Problem / Trigger, ## Solution"
            echo "  9. Filename convention: {date}-{name}.md or {name}.md"
            echo ""
            echo "Options:"
            echo "  --fix        Auto-fix common issues (missing times_seen, last_seen, lowercase confidence)"
            echo "  --json       Machine-readable JSON output"
            echo "  -h, --help   Show this help"
            exit 0
            ;;
        *)
            if [[ -f "$1" ]]; then
                FILES+=("$1")
            else
                echo "Unknown option or file not found: $1" >&2
                exit 1
            fi
            shift
            ;;
    esac
done

# If no files specified, find all .md in learned/
if [[ ${#FILES[@]} -eq 0 ]]; then
    while IFS= read -r f; do
        [[ -n "$f" ]] && FILES+=("$f")
    done < <(find "$LEARNED_DIR" -name "*.md" -not -name ".gitkeep" -not -name ".last_check" 2>/dev/null | sort)
fi

VALID_DOMAINS=("_global" "devops" "web" "mobile" "backend" "data")
VALID_CATEGORIES=("debug" "pattern" "style" "security" "performance" "ops")
VALID_CONFIDENCE=("HIGH" "MEDIUM" "LOW")

TOTAL=0
PASSED=0
FAILED=0

# JSON accumulator
JSON_RESULTS=()

# Get frontmatter content (between first and second ---)
get_frontmatter() {
    local file="$1"
    awk 'NR==1 && /^---$/ {in_fm=1; next} in_fm && /^---$/ {exit} in_fm {print}' "$file"
}

# Get a field from frontmatter
get_field() {
    local fm="$1" field="$2"
    echo "$fm" | grep "^${field}:" | sed "s/^${field}: *//" | head -1
}

# Check if value is in array
in_array() {
    local val="$1"; shift
    for item in "$@"; do
        [[ "$item" == "$val" ]] && return 0
    done
    return 1
}

validate_file() {
    local file="$1"
    local -a issues=()
    local -a fixed=()
    local basename_file
    basename_file=$(basename "$file")

    TOTAL=$((TOTAL + 1))

    # 1. Has YAML frontmatter
    local first_line
    first_line=$(head -1 "$file")
    if [[ "$first_line" != "---" ]]; then
        issues+=("Missing YAML frontmatter (no opening ---)")
        # Can't check fields without frontmatter
        report_file "$file" "${issues[@]}"
        return
    fi

    local closing_line
    closing_line=$(awk 'NR>1 && /^---$/ {print NR; exit}' "$file")
    if [[ -z "$closing_line" ]]; then
        issues+=("Missing closing --- in frontmatter")
        report_file "$file" "${issues[@]}"
        return
    fi

    local fm
    fm=$(get_frontmatter "$file")

    # 2. Required fields
    local name conf category domain first_seen last_seen times_seen
    name=$(get_field "$fm" "name")
    conf=$(get_field "$fm" "confidence")
    category=$(get_field "$fm" "category")
    domain=$(get_field "$fm" "domain")
    first_seen=$(get_field "$fm" "first_seen")
    last_seen=$(get_field "$fm" "last_seen")
    times_seen=$(get_field "$fm" "times_seen")

    [[ -z "$name" ]] && issues+=("Missing required field: name")
    [[ -z "$conf" ]] && issues+=("Missing required field: confidence")
    [[ -z "$category" ]] && issues+=("Missing required field: category")
    [[ -z "$domain" ]] && issues+=("Missing required field: domain")
    [[ -z "$first_seen" ]] && issues+=("Missing required field: first_seen")

    # 3. Valid confidence (with auto-fix for case)
    if [[ -n "$conf" ]]; then
        local conf_upper
        conf_upper=$(echo "$conf" | tr '[:lower:]' '[:upper:]')
        if ! in_array "$conf_upper" "${VALID_CONFIDENCE[@]}"; then
            issues+=("Invalid confidence: '$conf' (must be HIGH, MEDIUM, or LOW)")
        elif [[ "$conf" != "$conf_upper" ]]; then
            if [[ "$FIX" == true ]]; then
                sed -i '' "s/^confidence: *${conf}/confidence: ${conf_upper}/" "$file" 2>/dev/null || \
                sed -i "s/^confidence: *${conf}/confidence: ${conf_upper}/" "$file"
                fixed+=("Fixed confidence: $conf → $conf_upper")
            else
                issues+=("Confidence should be uppercase: '$conf' → '$conf_upper'")
            fi
        fi
    fi

    # 4. Valid domain
    if [[ -n "$domain" ]]; then
        if ! in_array "$domain" "${VALID_DOMAINS[@]}"; then
            issues+=("Invalid domain: '$domain' (must be one of: ${VALID_DOMAINS[*]})")
        fi
    fi

    # 5. Valid category
    if [[ -n "$category" ]]; then
        if ! in_array "$category" "${VALID_CATEGORIES[@]}"; then
            issues+=("Invalid category: '$category' (must be one of: ${VALID_CATEGORIES[*]})")
        fi
    fi

    # 6. Date format for first_seen and last_seen
    if [[ -n "$first_seen" ]]; then
        if ! echo "$first_seen" | grep -qE '^[0-9]{4}-[0-9]{2}-[0-9]{2}$'; then
            issues+=("Invalid date format for first_seen: '$first_seen' (expected YYYY-MM-DD)")
        fi
    fi
    if [[ -n "$last_seen" ]]; then
        if ! echo "$last_seen" | grep -qE '^[0-9]{4}-[0-9]{2}-[0-9]{2}$'; then
            issues+=("Invalid date format for last_seen: '$last_seen' (expected YYYY-MM-DD)")
        fi
    fi

    # Auto-fix: missing last_seen → copy from first_seen
    if [[ -z "$last_seen" && -n "$first_seen" ]]; then
        if [[ "$FIX" == true ]]; then
            # Insert last_seen after first_seen line
            sed -i '' "s/^first_seen: *${first_seen}/first_seen: ${first_seen}\nlast_seen: ${first_seen}/" "$file" 2>/dev/null || \
            sed -i "s/^first_seen: *${first_seen}/first_seen: ${first_seen}\nlast_seen: ${first_seen}/" "$file"
            fixed+=("Added last_seen: $first_seen (copied from first_seen)")
        else
            issues+=("Missing required field: last_seen")
        fi
    elif [[ -z "$last_seen" && -z "$first_seen" ]]; then
        issues+=("Missing required field: last_seen")
    fi

    # 7. times_seen is numeric >= 1 (with auto-fix)
    if [[ -z "$times_seen" ]]; then
        if [[ "$FIX" == true ]]; then
            # Insert times_seen before closing ---
            sed -i '' "/^---$/,/^---$/{
                /^---$/!b
                :a
                n
                /^---$/{i\\
times_seen: 1
                b}
                ba
            }" "$file" 2>/dev/null || true
            # Simpler approach: add before the second ---
            if ! grep -q '^times_seen:' "$file"; then
                awk 'NR==1{print;next} /^---$/ && !done{print "times_seen: 1"; done=1} {print}' "$file" > "$file.tmp" && mv "$file.tmp" "$file"
            fi
            fixed+=("Added times_seen: 1")
        else
            issues+=("Missing required field: times_seen")
        fi
    elif ! echo "$times_seen" | grep -qE '^[0-9]+$'; then
        issues+=("times_seen is not numeric: '$times_seen'")
    elif [[ "$times_seen" -lt 1 ]]; then
        issues+=("times_seen must be >= 1 (got: $times_seen)")
    fi

    # 8. Required sections
    local content
    content=$(cat "$file")
    if ! echo "$content" | grep -q '^# '; then
        issues+=("Missing required section: # Title")
    fi
    if ! echo "$content" | grep -q '^## Problem'; then
        issues+=("Missing required section: ## Problem / Trigger")
    fi
    if ! echo "$content" | grep -q '^## Solution'; then
        issues+=("Missing required section: ## Solution")
    fi

    # 9. Filename convention
    if ! echo "$basename_file" | grep -qE '^([0-9]{4}-[0-9]{2}-[0-9]{2}-)?[a-z0-9][a-z0-9_-]*\.md$'; then
        issues+=("Filename doesn't match convention ({date}-{name}.md or {name}.md): $basename_file")
    fi

    # Report fixes
    if [[ ${#fixed[@]} -gt 0 ]]; then
        for fx in "${fixed[@]}"; do
            if [[ "$JSON_OUTPUT" != true ]]; then
                echo "  🔧 $fx"
            fi
        done
    fi

    if [[ ${#issues[@]} -gt 0 ]]; then
        report_file "$file" "${issues[@]}"
    else
        report_file "$file"
    fi
}

report_file() {
    local file="$1"; shift
    local issues=("$@")

    local relpath
    relpath=$(echo "$file" | sed "s|$REPO_DIR/||")

    if [[ ${#issues[@]} -eq 0 || ( ${#issues[@]} -eq 1 && -z "${issues[0]}" ) ]]; then
        PASSED=$((PASSED + 1))
        if [[ "$JSON_OUTPUT" == true ]]; then
            JSON_RESULTS+=("{\"file\":\"$relpath\",\"status\":\"ok\",\"issues\":[]}")
        else
            echo "✅ $relpath"
        fi
    else
        FAILED=$((FAILED + 1))
        if [[ "$JSON_OUTPUT" == true ]]; then
            local issues_json="["
            local first=true
            for issue in "${issues[@]}"; do
                [[ -z "$issue" ]] && continue
                if [[ "$first" == true ]]; then
                    first=false
                else
                    issues_json+=","
                fi
                # Escape quotes in issue text
                local escaped
                escaped=$(echo "$issue" | sed 's/"/\\"/g')
                issues_json+="\"$escaped\""
            done
            issues_json+="]"
            JSON_RESULTS+=("{\"file\":\"$relpath\",\"status\":\"fail\",\"issues\":$issues_json}")
        else
            echo "❌ $relpath"
            for issue in "${issues[@]}"; do
                [[ -n "$issue" ]] && echo "   - $issue"
            done
        fi
    fi
}

# ── Main ──────────────────────────────────────────────────────────

if [[ ${#FILES[@]} -eq 0 ]]; then
    if [[ "$JSON_OUTPUT" == true ]]; then
        echo '{"total":0,"passed":0,"failed":0,"results":[]}'
    else
        echo "📭 No pattern files found in learned/"
    fi
    exit 0
fi

if [[ "$JSON_OUTPUT" != true ]]; then
    echo "🔍 Validating ${#FILES[@]} pattern file(s)..."
    [[ "$FIX" == true ]] && echo "   (--fix mode: auto-fixing common issues)"
    echo ""
fi

for f in "${FILES[@]}"; do
    validate_file "$f"
done

# ── Summary ───────────────────────────────────────────────────────

if [[ "$JSON_OUTPUT" == true ]]; then
    echo -n "{\"total\":$TOTAL,\"passed\":$PASSED,\"failed\":$FAILED,\"results\":["
    for i in "${!JSON_RESULTS[@]}"; do
        if [[ $i -gt 0 ]]; then
            echo -n ","
        fi
        echo -n "${JSON_RESULTS[$i]}"
    done
    echo "]}"
else
    echo ""
    echo "======================================"
    echo "  Results: $TOTAL checked, $PASSED passed, $FAILED failed"
    echo "======================================"
    if [[ $FAILED -gt 0 ]]; then
        exit 1
    else
        echo "  🎉 All patterns valid!"
    fi
fi
