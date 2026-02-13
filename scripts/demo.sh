#!/bin/bash
# MUR Core Demo Script
# 執行: chmod +x scripts/demo.sh && ./scripts/demo.sh

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

run_cmd() {
    echo -e "${GREEN}${BOLD}\$ $1${NC}"
    sleep 0.8
    eval "$1"
}

scene() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}${BOLD}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    sleep 1
}

wait_key() {
    echo ""
    echo -e "${BOLD}[Press Enter to continue]${NC}"
    read
}

clear
echo ""
echo -e "${BOLD}🎬 MUR Core Demo${NC}"
echo ""
echo "Press Enter to start recording..."
read
clear

# ===== Scene 1 =====
scene "1. Check Version"
run_cmd "mur version"
wait_key

# ===== Scene 2 =====
scene "2. Initialize Hooks"
run_cmd "mur init --hooks"
wait_key

# ===== Scene 3 =====
scene "3. Health Check"
run_cmd "mur doctor"
wait_key

# ===== Scene 4 =====
scene "4. Semantic Search"
run_cmd 'mur search "Go error handling"'
wait_key

# ===== Scene 5 =====
scene "5. Sync to All Tools"
run_cmd "mur sync --cli"
wait_key

# ===== Scene 6 =====
scene "6. Status"
run_cmd "mur status"
wait_key

# ===== Scene 7 =====
scene "7. Dashboard"
echo -e "${GREEN}${BOLD}\$ mur serve${NC}"
echo ""
echo "Starting dashboard at http://localhost:3377"
mur serve &
PID=$!
sleep 2
open http://localhost:3377

echo ""
echo "Press Enter when done with dashboard..."
read
kill $PID 2>/dev/null || true

# ===== Done =====
clear
scene "✅ Demo Complete!"
echo "Now open: https://app.mur.run"
echo ""
