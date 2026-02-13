#!/bin/bash
# 設定錄製用的 Terminal 環境

# 清除畫面
clear

# 簡化 prompt
export PS1="$ "

# 設定視窗大小 (適合 1280x720 錄製)
printf '\e[8;30;100t'

# 設定字體提示
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🎬 Recording Setup Complete"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  ✓ Window size: 100x30"
echo "  ✓ Simple prompt: $"
echo ""
echo "  Next steps:"
echo "  1. Cmd + 放大字體 (按 3-4 次)"
echo "  2. Cmd+Shift+5 選區域錄製"
echo "  3. 執行 ./scripts/demo.sh"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
