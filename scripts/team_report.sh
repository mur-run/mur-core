#!/usr/bin/env bash
set -euo pipefail

# 產生團隊學習報告
# 用法: ./scripts/team_report.sh [--days N]

REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LEARNED_DIR="$REPO_DIR/learned"
PRIVACY_FILTER="$SCRIPT_DIR/privacy_filter.sh"
DAYS=1

while [[ $# -gt 0 ]]; do
    case "$1" in
        --days) DAYS="$2"; shift 2 ;;
        -h|--help)
            echo "Usage: team_report.sh [--days N]"
            echo ""
            echo "Generate a team learning report for the last N days."
            echo ""
            echo "Options:"
            echo "  --days N     Number of days to look back (default: 1)"
            echo "  -h, --help   Show this help"
            exit 0
            ;;
        *) shift ;;
    esac
done

# macOS date
if [[ "$(uname -s)" == "Darwin" ]]; then
    SINCE=$(date -v-${DAYS}d '+%Y-%m-%d')
else
    SINCE=$(date -d "$DAYS days ago" '+%Y-%m-%d')
fi

echo "━━━━━━━━━━━━━━━━━━━━"
echo "🧠 團隊學習日報"
echo "📅 $(date '+%Y-%m-%d')"
echo "━━━━━━━━━━━━━━━━━━━━"
echo ""

# 統計新增 patterns（最近 N 天修改的 .md 檔案）
TOTAL=0
DOMAINS=""

# Privacy filter: only include files that pass privacy rules in the report
REPORT_FILES=$(find "$LEARNED_DIR" -name "*.md" -newer "$LEARNED_DIR/../setup/install.sh" -mtime -${DAYS} 2>/dev/null || true)
if [[ -x "$PRIVACY_FILTER" && -n "$REPORT_FILES" ]]; then
    REPORT_FILES=$(echo "$REPORT_FILES" | "$PRIVACY_FILTER" 2>/dev/null || echo "$REPORT_FILES")
fi

echo "$REPORT_FILES" | while read f; do
    [[ -z "$f" ]] && continue
    # 提取 domain 和檔名
    REL=$(echo "$f" | sed "s|$LEARNED_DIR/||")
    DOMAIN=$(echo "$REL" | cut -d'/' -f1)
    NAME=$(basename "$f" .md)
    
    # 嘗試從 frontmatter 提取標題
    TITLE=$(head -20 "$f" | grep '^# ' | head -1 | sed 's/^# //')
    TITLE="${TITLE:-$NAME}"
    
    echo "  • *${DOMAIN}*: $TITLE"
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━"
