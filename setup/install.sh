#!/usr/bin/env bash
set -euo pipefail

# Claude Code Learner - Setup (Solo / Team)
# 
# Solo mode (default):  ./setup/install.sh
# Team mode:            ./setup/install.sh --team
# Explicit solo:        ./setup/install.sh --solo
#
# Idempotent — safe to run multiple times.

MODE="solo"
while [[ $# -gt 0 ]]; do
    case "$1" in
        --team)  MODE="team"; shift ;;
        --solo)  MODE="solo"; shift ;;
        -h|--help)
            echo "Usage: install.sh [--solo | --team]"
            echo ""
            echo "Modes:"
            echo "  (default)  Solo mode — hooks, directory structure, auto-learn cron"
            echo "  --solo     Same as default (explicit)"
            echo "  --team     Solo + personal branch, personal dir, push/pull cron"
            echo ""
            echo "Safe to run multiple times. Already-installed steps are skipped."
            exit 0
            ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

echo "🧠 Claude Code Learner - Setup ($MODE mode)"
echo "===================================="

# 1. Verify claude is installed
if ! command -v claude &>/dev/null; then
    echo "❌ Claude Code CLI not found. Install first: https://code.claude.com"
    exit 1
fi

# 2. Verify we're in the right repo
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
LEARNED_DIR="$REPO_DIR/learned"

if [ ! -f "$REPO_DIR/SKILL.md" ]; then
    echo "❌ Not in the claude-code-learner directory"
    exit 1
fi

# Track what we did (for upgrade message)
HOOKS_WERE_INSTALLED=false
LEARN_CRON_EXISTS=false

# ── Step 1: Ensure learned/ directory structure ────────────────────
echo ""
echo "📂 Ensuring learned/ directory structure..."
for dir in _global devops web mobile backend data projects personal; do
    mkdir -p "$LEARNED_DIR/$dir"
    touch "$LEARNED_DIR/$dir/.gitkeep"
done
echo "✅ Directory structure ready"

# ── Step 2: Install Claude Code hooks (auto-merge) ────────────────
CLAUDE_SETTINGS="$HOME/.claude/settings.json"
HOOKS_FILE="$REPO_DIR/hooks/claude-code-hooks.json"

# Helper: update paths in hooks to use actual repo path
update_hook_paths() {
    local file="$1"
    python3 -c "
import json, re

with open('$file', 'r') as f:
    data = json.load(f)

def replace_paths(obj):
    if isinstance(obj, str):
        obj = obj.replace('~/clawd/skills/claude-code-learner', '$REPO_DIR')
        obj = re.sub(r'\\\$HOME/clawd/skills/claude-code-learner', '$REPO_DIR', obj)
        return obj
    elif isinstance(obj, dict):
        return {k: replace_paths(v) for k, v in obj.items()}
    elif isinstance(obj, list):
        return [replace_paths(v) for v in obj]
    return obj

data = replace_paths(data)

with open('$file', 'w') as f:
    json.dump(data, f, indent=2, ensure_ascii=False)
    f.write('\n')
"
}

# Check if hooks are already installed
hooks_already_installed() {
    if [ ! -f "$CLAUDE_SETTINGS" ]; then
        return 1
    fi
    python3 -c "
import json
with open('$CLAUDE_SETTINGS') as f:
    data = json.load(f)
hooks = data.get('hooks', {})
# Check if any hook command references claude-code-learner
import json as j
found = False
for event, groups in hooks.items():
    for group in (groups if isinstance(groups, list) else []):
        for hook in group.get('hooks', []):
            if 'claude-code-learner' in hook.get('command', ''):
                found = True
print('yes' if found else 'no')
" 2>/dev/null | grep -q "yes"
}

echo ""
echo "🔗 Installing Claude Code hooks..."

if hooks_already_installed; then
    echo "✅ Hooks already installed — skipping"
    HOOKS_WERE_INSTALLED=true
else
    if [ ! -f "$CLAUDE_SETTINGS" ]; then
        # Case 1: No settings.json exists → create from hooks file
        mkdir -p "$HOME/.claude"
        cp "$HOOKS_FILE" "$CLAUDE_SETTINGS"
        update_hook_paths "$CLAUDE_SETTINGS"
        echo "✅ Created $CLAUDE_SETTINGS with hooks"
    else
        # Settings file exists — smart merge
        HAS_HOOKS=$(python3 -c "
import json
with open('$CLAUDE_SETTINGS') as f:
    data = json.load(f)
print('yes' if 'hooks' in data else 'no')
" 2>/dev/null || echo "error")

        if [ "$HAS_HOOKS" = "no" ]; then
            # Case 2: Settings exists but no hooks key → add hooks section
            python3 -c "
import json

with open('$CLAUDE_SETTINGS', 'r') as f:
    settings = json.load(f)

with open('$HOOKS_FILE', 'r') as f:
    hooks_data = json.load(f)

settings['hooks'] = hooks_data.get('hooks', hooks_data)

with open('$CLAUDE_SETTINGS', 'w') as f:
    json.dump(settings, f, indent=2, ensure_ascii=False)
    f.write('\n')
"
            update_hook_paths "$CLAUDE_SETTINGS"
            echo "✅ Added hooks to existing $CLAUDE_SETTINGS"
        elif [ "$HAS_HOOKS" = "yes" ]; then
            # Case 3: Settings exists AND has hooks → smart merge, backup first
            BACKUP="${CLAUDE_SETTINGS}.backup.$(date +%Y%m%d%H%M%S)"
            cp "$CLAUDE_SETTINGS" "$BACKUP"
            echo "📦 Backed up existing settings to $BACKUP"

            python3 -c "
import json

with open('$CLAUDE_SETTINGS', 'r') as f:
    settings = json.load(f)

with open('$HOOKS_FILE', 'r') as f:
    hooks_data = json.load(f)

our_hooks = hooks_data.get('hooks', hooks_data)
existing_hooks = settings.get('hooks', {})

for event, event_hooks in our_hooks.items():
    if event not in existing_hooks:
        existing_hooks[event] = event_hooks
    else:
        existing_commands = set()
        for matcher_group in existing_hooks[event]:
            for hook in matcher_group.get('hooks', []):
                existing_commands.add(hook.get('command', ''))

        for matcher_group in event_hooks:
            for hook in matcher_group.get('hooks', []):
                cmd = hook.get('command', '')
                if 'claude-code-learner' in cmd and cmd not in existing_commands:
                    existing_hooks[event].append(matcher_group)
                    break
                elif 'claude-code-learner' not in cmd:
                    existing_hooks[event].append(matcher_group)
                    break

settings['hooks'] = existing_hooks

with open('$CLAUDE_SETTINGS', 'w') as f:
    json.dump(settings, f, indent=2, ensure_ascii=False)
    f.write('\n')
"
            update_hook_paths "$CLAUDE_SETTINGS"
            echo "✅ Smart-merged hooks into $CLAUDE_SETTINGS"
        else
            echo "⚠️  Could not parse $CLAUDE_SETTINGS, please merge hooks manually from:"
            echo "  $HOOKS_FILE"
        fi
    fi
fi

echo "📂 Repo path: $REPO_DIR"

# ── Step 3: Auto-learn cron (both solo and team) ──────────────────
CRON_CMD_LEARN="cd $REPO_DIR && ./scripts/auto_learn.sh 2>/dev/null"

# Check if auto-learn cron already exists
if crontab -l 2>/dev/null | grep -q "claude-code-learner.*auto_learn"; then
    echo ""
    echo "✅ Auto-learn cron already set — skipping"
    LEARN_CRON_EXISTS=true
else
    echo ""
    echo "📅 Cron Setup: Auto-learn (daily at 3am)"
    echo "  This runs review + sync + commit on your patterns."
    read -p "  Set up auto-learn cron? [y/N]: " SETUP_LEARN
    if [[ "${SETUP_LEARN:-N}" =~ ^[Yy] ]]; then
        (crontab -l 2>/dev/null | grep -v "claude-code-learner.*auto_learn"; echo "0 3 * * * $CRON_CMD_LEARN") | crontab -
        echo "  ✅ Auto-learn cron job added (daily at 3am)"
    else
        echo "  ℹ️  Skipped. To add manually:"
        echo "    0 3 * * * $CRON_CMD_LEARN"
    fi
fi

# ── Team-only steps ───────────────────────────────────────────────
if [ "$MODE" = "team" ]; then
    echo ""
    echo "👥 Team mode — additional setup"
    echo "--------------------------------"

    # Get user name
    CURRENT_USER=$(git config user.name 2>/dev/null || whoami)
    read -p "  Your name [$CURRENT_USER]: " USER_NAME
    USER_NAME="${USER_NAME:-$CURRENT_USER}"
    USER_SLUG=$(echo "$USER_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -cd 'a-z0-9-')

    echo "  📝 Setting up for: $USER_NAME ($USER_SLUG)"

    # Create personal branch
    BRANCH="learnings/$USER_SLUG"
    echo "  🔀 Creating branch: $BRANCH"
    git checkout -b "$BRANCH" 2>/dev/null || git checkout "$BRANCH"

    # Create personal learned directory
    mkdir -p "$LEARNED_DIR/personal/$USER_SLUG"
    touch "$LEARNED_DIR/personal/$USER_SLUG/.gitkeep"
    echo "  ✅ Personal directory: learned/personal/$USER_SLUG/"

    # Auto-push cron (every 30 min)
    CRON_CMD_PUSH="cd $REPO_DIR && git add learned/ && git diff --cached --quiet || git commit -m 'learn: auto-commit by $USER_SLUG' && git push origin $BRANCH 2>/dev/null"

    if crontab -l 2>/dev/null | grep -q "claude-code-learner.*auto-commit"; then
        echo "  ✅ Auto-push cron already set — skipping"
    else
        echo ""
        read -p "  Set up auto-push cron (every 30 min)? [y/N]: " SETUP_PUSH
        if [[ "${SETUP_PUSH:-N}" =~ ^[Yy] ]]; then
            (crontab -l 2>/dev/null | grep -v "claude-code-learner.*auto-commit"; echo "*/30 * * * * $CRON_CMD_PUSH") | crontab -
            echo "  ✅ Auto-push cron job added (every 30 min)"
        else
            echo "  ℹ️  Skipped. To add manually:"
            echo "    */30 * * * * $CRON_CMD_PUSH"
        fi
    fi

    # Auto-pull cron (every hour)
    CRON_CMD_PULL="cd $REPO_DIR && git fetch origin main && git merge origin/main --no-edit 2>/dev/null; $REPO_DIR/scripts/sync_to_claude_code.sh 2>/dev/null"

    if crontab -l 2>/dev/null | grep -q "claude-code-learner.*sync_to_claude_code"; then
        echo "  ✅ Auto-pull cron already set — skipping"
    else
        read -p "  Set up auto-pull cron (every hour)? [y/N]: " SETUP_PULL
        if [[ "${SETUP_PULL:-N}" =~ ^[Yy] ]]; then
            (crontab -l 2>/dev/null | grep -v "claude-code-learner.*sync_to_claude_code"; echo "0 * * * * $CRON_CMD_PULL") | crontab -
            echo "  ✅ Auto-pull cron job added (every hour)"
        else
            echo "  ℹ️  Skipped. To add manually:"
            echo "    0 * * * * $CRON_CMD_PULL"
        fi
    fi
fi

# ── Final message ─────────────────────────────────────────────────
echo ""
echo "======================================"

if [ "$MODE" = "solo" ]; then
    echo "✅ Solo setup complete!"
    echo ""
    echo "Next steps:"
    echo "  1. Start using Claude Code normally — patterns are extracted automatically"
    echo "  2. Run '$REPO_DIR/scripts/status.sh' to see pattern statistics"
    echo "  3. Run '$REPO_DIR/scripts/sync_to_claude_code.sh' to sync mature patterns"
    echo ""
    echo "Running in solo mode. When your team grows, run:"
    echo "  ./setup/install.sh --team"
elif [ "$MODE" = "team" ]; then
    if [ "$HOOKS_WERE_INSTALLED" = true ]; then
        echo "✅ Upgraded from solo to team mode!"
        echo ""
        echo "Your existing patterns on main are now the team's shared knowledge base."
    else
        echo "✅ Team setup complete!"
    fi
    echo ""
    echo "Next steps:"
    echo "  1. Start using Claude Code normally — patterns are extracted automatically"
    echo "  2. Run 'git push origin $BRANCH' to share your learnings"
    echo "  3. Run '$REPO_DIR/scripts/sync_to_claude_code.sh' to sync team patterns locally"
    echo "  4. Run '$REPO_DIR/scripts/status.sh' to see pattern statistics"
    echo "  5. Admin: run '$REPO_DIR/scripts/merge_team.sh' to merge all branches"
fi
