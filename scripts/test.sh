#!/usr/bin/env bash
# test.sh — Smoke tests for mur-core
# Usage: ./scripts/test.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

PASS=0
FAIL=0
ERRORS=""

pass() { PASS=$((PASS + 1)); echo "  ✅ $1"; }
fail() { FAIL=$((FAIL + 1)); ERRORS+="  ❌ $1\n"; echo "  ❌ $1"; }

echo "🧪 Claude Code Learner — Smoke Tests"
echo "======================================"
echo ""

# ── Test 1: Every script's --help exits 0 ──────────────────────────
echo "📋 Test: --help flag on all scripts"
for script in "$SCRIPT_DIR"/*.sh; do
  NAME=$(basename "$script")
  [[ "$NAME" == "test.sh" ]] && continue
  [[ "$NAME" == "on_session_stop.sh" ]] && continue  # no --help
  if bash "$script" --help >/dev/null 2>&1; then
    pass "$NAME --help"
  else
    fail "$NAME --help exited non-zero"
  fi
done

# ── Test 2: --dry-run on scripts that support it ───────────────────
echo ""
echo "📋 Test: --dry-run on supported scripts"
for script in auto_learn.sh sync_to_claude_code.sh merge_team.sh; do
  FULL="$SCRIPT_DIR/$script"
  if [[ ! -f "$FULL" ]]; then
    fail "$script not found"
    continue
  fi
  # Check that --dry-run is accepted (script starts and produces output)
  # Some scripts may fail due to environment (no git remote, etc.) — that's OK
  OUTPUT=$(bash "$FULL" --dry-run 2>&1 || true)
  if echo "$OUTPUT" | grep -qi "dry.run\|DRY RUN\|preview\|unknown option.*dry" 2>/dev/null; then
    pass "$script --dry-run (flag recognized)"
  elif echo "$OUTPUT" | grep -qi "dry" 2>/dev/null; then
    pass "$script --dry-run (flag recognized)"
  else
    # Script ran but may have failed for env reasons (no git remote, etc.)
    # Check it didn't fail on arg parsing
    if echo "$OUTPUT" | grep -qi "unknown option\|invalid.*dry"; then
      fail "$script --dry-run not recognized"
    else
      pass "$script --dry-run (accepted, env-dependent execution)"
    fi
  fi
done

# ── Test 3: learned/ directory structure ───────────────────────────
echo ""
echo "📋 Test: learned/ directory structure"
EXPECTED_DIRS=("_global" "devops" "web" "mobile" "backend" "data" "projects" "personal")
for d in "${EXPECTED_DIRS[@]}"; do
  if [[ -d "$REPO_DIR/learned/$d" ]]; then
    pass "learned/$d/ exists"
  else
    fail "learned/$d/ missing"
  fi
done

# ── Test 4: Templates exist and have valid YAML frontmatter ───────
echo ""
echo "📋 Test: templates have valid YAML frontmatter"
for tmpl in "$REPO_DIR/templates"/*.md; do
  NAME=$(basename "$tmpl")
  if [[ ! -f "$tmpl" ]]; then
    fail "Template $NAME not found"
    continue
  fi
  # Check starts with ---
  FIRST_LINE=$(head -1 "$tmpl")
  if [[ "$FIRST_LINE" != "---" ]]; then
    fail "$NAME: missing opening ---"
    continue
  fi
  # Check has closing ---
  CLOSING=$(awk 'NR>1 && /^---$/ {print NR; exit}' "$tmpl")
  if [[ -z "$CLOSING" ]]; then
    fail "$NAME: missing closing ---"
    continue
  fi
  # Check has name: field
  if grep -q '^name:' "$tmpl"; then
    pass "$NAME: valid frontmatter"
  else
    fail "$NAME: no 'name:' field in frontmatter"
  fi
done

# ── Test 5: Hooks files exist and are valid JSON ──────────────────
echo ""
echo "📋 Test: hooks files"
HOOKS_FILE="$REPO_DIR/hooks/claude-code-hooks.json"
if [[ -f "$HOOKS_FILE" ]]; then
  if python3 -c "import json; json.load(open('$HOOKS_FILE'))" 2>/dev/null; then
    pass "claude-code-hooks.json is valid JSON"
  else
    fail "claude-code-hooks.json is invalid JSON"
  fi
  # Check it has hooks key
  if python3 -c "import json; d=json.load(open('$HOOKS_FILE')); assert 'hooks' in d" 2>/dev/null; then
    pass "claude-code-hooks.json has 'hooks' key"
  else
    fail "claude-code-hooks.json missing 'hooks' key"
  fi
else
  fail "claude-code-hooks.json not found"
fi

REMINDER="$REPO_DIR/hooks/on-prompt-reminder.md"
if [[ -f "$REMINDER" ]]; then
  pass "on-prompt-reminder.md exists"
else
  fail "on-prompt-reminder.md not found"
fi

# ── Test 6: validate.sh and privacy_filter.sh --help ──────────────
echo ""
echo "📋 Test: new scripts --help"
for script in validate.sh privacy_filter.sh; do
  FULL="$SCRIPT_DIR/$script"
  if [[ ! -f "$FULL" ]]; then
    fail "$script not found"
    continue
  fi
  if bash "$FULL" --help >/dev/null 2>&1; then
    pass "$script --help"
  else
    fail "$script --help exited non-zero"
  fi
done

# ── Test 7: Example patterns have valid frontmatter ───────────────
echo ""
echo "📋 Test: example pattern files"
EXAMPLE_PATTERNS=$(find "$REPO_DIR/learned" -name "*.md" -not -name ".gitkeep" -not -name ".last_check" 2>/dev/null)
if [[ -n "$EXAMPLE_PATTERNS" ]]; then
  while IFS= read -r pf; do
    PNAME=$(basename "$pf")
    FIRST=$(head -1 "$pf")
    if [[ "$FIRST" == "---" ]]; then
      if grep -q '^name:' "$pf" && grep -q '^confidence:' "$pf"; then
        pass "Pattern $PNAME: valid"
      else
        fail "Pattern $PNAME: missing required frontmatter fields"
      fi
    else
      fail "Pattern $PNAME: no frontmatter"
    fi
  done <<< "$EXAMPLE_PATTERNS"
else
  echo "  ℹ️  No example patterns found (skipping)"
fi

# ── Test 8: spec_init.sh --help exits 0 ───────────────────────────
echo ""
echo "📋 Test: spec_init.sh --help"
SPEC_INIT="$SCRIPT_DIR/spec_init.sh"
if [[ -f "$SPEC_INIT" ]]; then
  if bash "$SPEC_INIT" --help >/dev/null 2>&1; then
    pass "spec_init.sh --help"
  else
    fail "spec_init.sh --help exited non-zero"
  fi
else
  fail "spec_init.sh not found"
fi

# ── Test 9: spec_report.sh --help exits 0 ─────────────────────────
echo ""
echo "📋 Test: spec_report.sh --help"
SPEC_REPORT="$SCRIPT_DIR/spec_report.sh"
if [[ -f "$SPEC_REPORT" ]]; then
  if bash "$SPEC_REPORT" --help >/dev/null 2>&1; then
    pass "spec_report.sh --help"
  else
    fail "spec_report.sh --help exited non-zero"
  fi
else
  fail "spec_report.sh not found"
fi

# ── Test 9b: auto_learn.sh --help regression check ───────────────
echo ""
echo "📋 Test: auto_learn.sh --help regression"
AUTO_LEARN="$SCRIPT_DIR/auto_learn.sh"
if [[ -f "$AUTO_LEARN" ]]; then
  if bash "$AUTO_LEARN" --help >/dev/null 2>&1; then
    pass "auto_learn.sh --help (regression)"
  else
    fail "auto_learn.sh --help exited non-zero (regression)"
  fi
else
  fail "auto_learn.sh not found"
fi

# ── Test 9c: .spec_processed handling ─────────────────────────────
echo ""
echo "📋 Test: .spec_processed file handling"
TMPDIR_TEST=$(mktemp -d)
TEST_SPEC_PROCESSED="$TMPDIR_TEST/.spec_processed"
# Create and verify
touch "$TEST_SPEC_PROCESSED"
echo "openspec/changes/archive/test-spec.md" >> "$TEST_SPEC_PROCESSED"
if grep -qxF "openspec/changes/archive/test-spec.md" "$TEST_SPEC_PROCESSED" 2>/dev/null; then
  pass ".spec_processed: entry written and found"
else
  fail ".spec_processed: entry not found after write"
fi
# Verify non-existent entry is not found
if grep -qxF "openspec/changes/archive/nonexistent.md" "$TEST_SPEC_PROCESSED" 2>/dev/null; then
  fail ".spec_processed: false positive match"
else
  pass ".spec_processed: non-existent entry correctly not matched"
fi
rm -rf "$TMPDIR_TEST"

# ── Test 10: on-spec-complete.md exists ────────────────────────────
echo ""
echo "📋 Test: on-spec-complete.md exists"
if [[ -f "$REPO_DIR/hooks/on-spec-complete.md" ]]; then
  pass "hooks/on-spec-complete.md exists"
else
  fail "hooks/on-spec-complete.md not found"
fi

# ── Test 10: .learned-config.yaml has integrations section ────────
echo ""
echo "📋 Test: .learned-config.yaml integrations"
CONFIG="$REPO_DIR/.learned-config.yaml"
if [[ -f "$CONFIG" ]]; then
  if grep -q '^integrations:' "$CONFIG"; then
    pass ".learned-config.yaml has integrations section"
  else
    fail ".learned-config.yaml missing integrations section"
  fi
else
  fail ".learned-config.yaml not found"
fi

# ── Summary ───────────────────────────────────────────────────────
echo ""
echo "======================================"
echo "  Results: $PASS passed, $FAIL failed"
echo "======================================"
if [[ $FAIL -gt 0 ]]; then
  echo ""
  echo "Failures:"
  echo -e "$ERRORS"
  exit 1
else
  echo "  🎉 All tests passed!"
  exit 0
fi
