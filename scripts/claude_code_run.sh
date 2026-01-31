#!/usr/bin/env bash
# claude_code_run.sh — Run Claude Code CLI in headless mode with PTY allocation.
# Handles macOS vs Linux `script` syntax automatically.
set -euo pipefail

###############################################################################
# Defaults
###############################################################################
PROMPT=""
PERMISSION_MODE=""
ALLOWED_TOOLS=""
OUTPUT_FORMAT=""
APPEND_SYSTEM_PROMPT=""
WORKDIR=""

###############################################################################
# Usage
###############################################################################
usage() {
  cat <<EOF
Usage: $(basename "$0") -p "prompt" [options]

Required:
  -p "prompt"                       Prompt to send to Claude Code

Options:
  --permission-mode MODE            plan | acceptEdits (default: none/ask)
  --allowedTools "Tool1,Tool2"      Comma-separated tool allowlist
  --output-format FORMAT            text | json (default: text)
  --append-system-prompt "text"     Extra system instructions
  --workdir /path/to/repo           Working directory for Claude Code
  -h, --help                        Show this help

Examples:
  $(basename "$0") -p "Summarize this project." --permission-mode plan
  $(basename "$0") -p "Fix the bug." --workdir ~/project --allowedTools "Read,Edit,Bash"
EOF
  exit 0
}

###############################################################################
# Parse arguments
###############################################################################
while [[ $# -gt 0 ]]; do
  case "$1" in
    -p)
      PROMPT="$2"; shift 2 ;;
    --permission-mode)
      PERMISSION_MODE="$2"; shift 2 ;;
    --allowedTools)
      ALLOWED_TOOLS="$2"; shift 2 ;;
    --output-format)
      OUTPUT_FORMAT="$2"; shift 2 ;;
    --append-system-prompt)
      APPEND_SYSTEM_PROMPT="$2"; shift 2 ;;
    --workdir)
      WORKDIR="$2"; shift 2 ;;
    -h|--help)
      usage ;;
    *)
      echo "Error: Unknown argument: $1" >&2
      usage ;;
  esac
done

###############################################################################
# Validate
###############################################################################
if [[ -z "$PROMPT" ]]; then
  echo "Error: -p \"prompt\" is required." >&2
  usage
fi

# Check claude is installed
if ! command -v claude &>/dev/null; then
  echo "Error: 'claude' CLI not found. Install Claude Code first." >&2
  echo "  See: https://code.claude.com/docs/en/overview" >&2
  exit 1
fi

###############################################################################
# Build the claude command
###############################################################################
CMD_ARGS=(-p "$PROMPT")

if [[ -n "$PERMISSION_MODE" ]]; then
  CMD_ARGS+=(--permission-mode "$PERMISSION_MODE")
fi

if [[ -n "$ALLOWED_TOOLS" ]]; then
  # Split comma-separated tools into separate --allowedTools flags
  IFS=',' read -ra TOOLS <<< "$ALLOWED_TOOLS"
  for tool in "${TOOLS[@]}"; do
    CMD_ARGS+=(--allowedTools "$tool")
  done
fi

if [[ -n "$OUTPUT_FORMAT" ]]; then
  CMD_ARGS+=(--output-format "$OUTPUT_FORMAT")
fi

if [[ -n "$APPEND_SYSTEM_PROMPT" ]]; then
  CMD_ARGS+=(--append-system-prompt "$APPEND_SYSTEM_PROMPT")
fi

###############################################################################
# Change to working directory if specified
###############################################################################
if [[ -n "$WORKDIR" ]]; then
  if [[ ! -d "$WORKDIR" ]]; then
    echo "Error: Working directory does not exist: $WORKDIR" >&2
    exit 1
  fi
  cd "$WORKDIR"
fi

###############################################################################
# Run with PTY via script(1)
# macOS: script -q /dev/null command args...
# Linux: script -q -c 'command args...' /dev/null
###############################################################################

# Build the full command string for script -c (Linux) or direct args (macOS)
# We need to properly quote arguments for shell execution
build_quoted_cmd() {
  local cmd="claude"
  for arg in "${CMD_ARGS[@]}"; do
    # Escape single quotes in arg and wrap in single quotes
    cmd+=" '${arg//\'/\'\\\'\'}'"
  done
  echo "$cmd"
}

OS="$(uname -s)"
QUOTED_CMD="$(build_quoted_cmd)"

case "$OS" in
  Darwin)
    # macOS: script -q /dev/null <command>
    exec script -q /dev/null bash -c "$QUOTED_CMD"
    ;;
  Linux)
    # Linux: script -q -c '<command>' /dev/null
    exec script -q -c "$QUOTED_CMD" /dev/null
    ;;
  *)
    echo "Warning: Unknown OS '$OS', running without PTY wrapper." >&2
    exec claude "${CMD_ARGS[@]}"
    ;;
esac
