#!/usr/bin/env bash
# spec_init.sh — Initialize spec-driven development for a project
# Usage:
#   ./scripts/spec_init.sh --project ~/Projects/BitL
#   ./scripts/spec_init.sh --project ~/Projects/BitL --tool openspec
#   ./scripts/spec_init.sh --project ~/Projects/BitL --tool speckit
#   ./scripts/spec_init.sh --project ~/Projects/BitL --tool both
#   ./scripts/spec_init.sh --project ~/Projects/BitL --superpowers
set -euo pipefail

PROJECT=""
TOOL="auto"
SUPERPOWERS=false

usage() {
  cat <<EOF
spec_init.sh — Initialize spec-driven development for a project

Usage:
  $(basename "$0") --project PATH [--tool openspec|speckit|both] [--superpowers]

Options:
  --project PATH       Target project directory (required)
  --tool TOOL          Which spec tool to set up: openspec, speckit, both, or auto (default: auto)
  --superpowers        Print instructions for installing Superpowers plugin in Claude Code
  -h, --help           Show this help message

Examples:
  $(basename "$0") --project ~/Projects/BitL
  $(basename "$0") --project ~/Projects/BitL --tool openspec
  $(basename "$0") --project ~/Projects/BitL --tool both --superpowers
EOF
  exit 0
}

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --project)
      PROJECT="$2"
      shift 2
      ;;
    --tool)
      TOOL="$2"
      shift 2
      ;;
    --superpowers)
      SUPERPOWERS=true
      shift
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo "Unknown option: $1" >&2
      echo "Use --help for usage." >&2
      exit 1
      ;;
  esac
done

if [[ -z "$PROJECT" ]]; then
  echo "Error: --project PATH is required." >&2
  echo "Use --help for usage." >&2
  exit 1
fi

# Resolve and validate project path
PROJECT="$(cd "$PROJECT" 2>/dev/null && pwd || echo "$PROJECT")"
if [[ ! -d "$PROJECT" ]]; then
  echo "Error: Project directory does not exist: $PROJECT" >&2
  exit 1
fi

# Detect installed tools
HAS_OPENSPEC=false
HAS_SPECKIT=false

if command -v openspec &>/dev/null; then
  HAS_OPENSPEC=true
fi
if command -v specify &>/dev/null; then
  HAS_SPECKIT=true
fi

# Auto-detect which tools to use
if [[ "$TOOL" == "auto" ]]; then
  if $HAS_OPENSPEC && $HAS_SPECKIT; then
    TOOL="both"
  elif $HAS_OPENSPEC; then
    TOOL="openspec"
  elif $HAS_SPECKIT; then
    TOOL="speckit"
  else
    echo "Error: No spec tools found. Install openspec or specify first." >&2
    echo "  npm install -g @fission-ai/openspec@latest" >&2
    echo "  uv tool install specify-cli --from git+https://github.com/github/spec-kit.git" >&2
    exit 1
  fi
fi

echo "🔧 Spec-Driven Development Init"
echo "================================"
echo "Project: $PROJECT"
echo "Tool:    $TOOL"
echo ""

SETUP_OPENSPEC=false
SETUP_SPECKIT=false

# Initialize OpenSpec
if [[ "$TOOL" == "openspec" || "$TOOL" == "both" ]]; then
  SETUP_OPENSPEC=true
  echo "📦 Setting up OpenSpec..."
  if ! $HAS_OPENSPEC; then
    echo "  ⚠️  openspec not found. Install with: npm install -g @fission-ai/openspec@latest"
  else
    cd "$PROJECT"
    if openspec init 2>/dev/null; then
      echo "  ✅ openspec init completed"
    else
      echo "  ℹ️  openspec init failed or requires interactive mode. Creating structure manually..."
      mkdir -p openspec/changes/archive openspec/templates
      if [[ ! -f openspec/openspec.yaml ]]; then
        cat > openspec/openspec.yaml <<YAML
# OpenSpec configuration
version: "1.0"
project: "$(basename "$PROJECT")"
YAML
      fi
      echo "  ✅ openspec/ directory structure created"
    fi
  fi
  echo ""
fi

# Initialize Spec Kit
if [[ "$TOOL" == "speckit" || "$TOOL" == "both" ]]; then
  SETUP_SPECKIT=true
  echo "📦 Setting up Spec Kit..."
  if ! $HAS_SPECKIT; then
    echo "  ⚠️  specify not found. Install with: uv tool install specify-cli --from git+https://github.com/github/spec-kit.git"
  else
    cd "$PROJECT"
    if specify init . --ai claude 2>/dev/null; then
      echo "  ✅ specify init completed"
    elif specify init --here --ai claude 2>/dev/null; then
      echo "  ✅ specify init completed (--here mode)"
    else
      echo "  ⚠️  specify init failed. You may need to run it interactively:"
      echo "     cd $PROJECT && specify init . --ai claude"
    fi
  fi
  echo ""
fi

# Superpowers instructions
if $SUPERPOWERS; then
  echo "🦸 Superpowers (Claude Code Plugin)"
  echo "  To install Superpowers in Claude Code, run these slash commands inside a session:"
  echo ""
  echo "    /plugin marketplace add obra/superpowers-marketplace"
  echo "    /plugin install superpowers@superpowers-marketplace"
  echo ""
fi

# Update .learned-config.yaml integrations section
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG_FILE="$SKILL_DIR/.learned-config.yaml"

if [[ -f "$CONFIG_FILE" ]]; then
  if ! grep -q '^integrations:' "$CONFIG_FILE"; then
    echo "📝 Adding integrations section to .learned-config.yaml..."
    cat >> "$CONFIG_FILE" <<YAML

integrations:
  # Spec-driven development tools
  spec_tool: ${TOOL}
  superpowers: true

  # Auto-extraction settings
  auto_extract_from_specs: true
  extract_decisions: true
  extract_rejected_alternatives: true

  # Spec artifact directories to scan
  spec_dirs:
    - openspec/changes/archive/
    - .spec/
YAML
    echo "  ✅ integrations section added"
  else
    echo "  ℹ️  integrations section already exists in .learned-config.yaml"
  fi
else
  echo "  ⚠️  .learned-config.yaml not found at $CONFIG_FILE"
fi

echo ""
echo "✅ Summary"
echo "=========="
$SETUP_OPENSPEC && echo "  • OpenSpec: initialized in $PROJECT/openspec/"
$SETUP_SPECKIT && echo "  • Spec Kit: initialized in $PROJECT/.spec/"
$SUPERPOWERS && echo "  • Superpowers: instructions printed (install via Claude Code session)"
echo "  • Config: .learned-config.yaml updated"
echo ""
echo "Next steps:"
echo "  1. Start a Claude Code session in your project"
echo "  2. Create specs before implementing features"
echo "  3. Patterns will be auto-extracted when specs complete"
