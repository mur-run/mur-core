// Package hooks provides hook installation for AI CLI tools.
package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ClaudeCodeHook represents a hook entry for Claude Code.
type ClaudeCodeHook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// ClaudeCodeHookMatcher represents a hook matcher with hooks array (new format).
type ClaudeCodeHookMatcher struct {
	Matcher string           `json:"matcher"`
	Hooks   []ClaudeCodeHook `json:"hooks"`
}

// ClaudeCodeHooks represents the hooks configuration (new matcher format).
type ClaudeCodeHooks struct {
	PreToolUse       map[string][]ClaudeCodeHookMatcher `json:"PreToolUse,omitempty"`
	PostToolUse      map[string][]ClaudeCodeHookMatcher `json:"PostToolUse,omitempty"`
	UserPromptSubmit []ClaudeCodeHookMatcher            `json:"UserPromptSubmit,omitempty"`
	Stop             []ClaudeCodeHookMatcher            `json:"Stop,omitempty"`
}

// ClaudeCodeSettings represents the Claude Code settings file.
type ClaudeCodeSettings struct {
	Hooks ClaudeCodeHooks `json:"hooks"`
}

// ClaudeCodeInstalled checks if Claude Code is installed.
func ClaudeCodeInstalled() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Check if .claude directory exists
	claudeDir := filepath.Join(home, ".claude")
	_, err = os.Stat(claudeDir)
	return err == nil
}

// InstallClaudeCodeHooks installs mur hooks for Claude Code.
//
// Instead of hardcoding commands, this creates shell scripts in ~/.mur/hooks/
// and points Claude Code's settings.json at them. If scripts already exist,
// they are preserved (user customizations are kept). Settings.json hooks are
// merged — existing non-mur hooks are not removed.
func InstallClaudeCodeHooks(enableSearch bool) error {
	return InstallClaudeCodeHooksWithOptions(HookOptions{EnableSearch: enableSearch})
}

// InstallClaudeCodeHooksWithOptions installs mur hooks for Claude Code with full options.
func InstallClaudeCodeHooksWithOptions(opts HookOptions) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	murBin, err := findMurBinary()
	if err != nil {
		murBin = "mur"
	}

	hooksDir := filepath.Join(home, ".mur", "hooks")
	settingsPath := filepath.Join(home, ".claude", "settings.json")

	// Ensure hooks directory exists
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("cannot create hooks directory: %w", err)
	}

	// Install /mur:in and /mur:out as Claude Code slash commands
	commandsDir := filepath.Join(home, ".claude", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return fmt.Errorf("cannot create commands directory: %w", err)
	}

	murInCmd := filepath.Join(commandsDir, "mur:in.md")
	murInContent := `---
description: "Start recording this conversation for workflow extraction. Events (prompts, tool calls, stops) will be captured until you run /mur:out."
allowed-tools: Bash(mur:*)
---

Start a mur session recording by running:

` + "```bash\nmur session start --source claude-code\n```" + `

After starting, confirm to the user:
- The session ID
- That recording is active
- Remind them to use ` + "`/mur:out`" + ` when done
`
	if err := os.WriteFile(murInCmd, []byte(murInContent), 0644); err != nil {
		return fmt.Errorf("cannot write mur:in.md command: %w", err)
	}
	fmt.Printf("  + Installed /mur:in command at %s\n", murInCmd)

	murOutCmd := filepath.Join(commandsDir, "mur:out.md")
	murOutContent := `---
description: "Stop recording and analyze the captured conversation to extract a reusable workflow."
allowed-tools: Bash(mur:*)
---

Stop the active mur session recording by running:

` + "```bash\nmur session stop\n```" + `

After stopping, show the user:
- Session duration and number of events captured
- Ask if they want to analyze: ` + "`mur session stop --analyze`" + `
- Or export: ` + "`mur session export <session-id> --format skill`" + `

If --analyze fails due to missing API key, tell the user to set ANTHROPIC_API_KEY or OPENAI_API_KEY.
`
	if err := os.WriteFile(murOutCmd, []byte(murOutContent), 0644); err != nil {
		return fmt.Errorf("cannot write mur:out.md command: %w", err)
	}
	fmt.Printf("  + Installed /mur:out command at %s\n", murOutCmd)

	// Clean up old session hook scripts (replaced by slash commands)
	for _, old := range []string{"mur-session-in.sh", "mur-session-out.sh"} {
		oldPath := filepath.Join(hooksDir, old)
		if _, err := os.Stat(oldPath); err == nil {
			os.Remove(oldPath)
			fmt.Printf("  - Removed old %s (replaced by slash command)\n", old)
		}
	}

	// Create default hook scripts (only if outdated or forced)
	stopScript := filepath.Join(hooksDir, "on-stop.sh")
	if ShouldUpgradeHook(stopScript, opts.Force) {
		content := fmt.Sprintf(`#!/bin/bash
# mur-managed-hook v%d
# Read hook input from stdin (Claude Code passes JSON)
INPUT=$(cat /dev/stdin 2>/dev/null || echo '{}')

# Record stop event to active session (if recording)
if [ -f ~/.mur/session/active.json ]; then
  STOP_REASON=$(echo "$INPUT" | jq -r '.stop_reason // "turn_end"' 2>/dev/null)
  %s session record --type assistant --content "[stop: $STOP_REASON]" 2>/dev/null || true
fi

# Lightweight sync (blocking, fast)
%s sync --quiet 2>/dev/null || true

# LLM extract in background (non-blocking)
(%s learn extract --llm --auto --accept-all --quiet 2>/dev/null &) || true

# Load user customizations if they exist
[ -f ~/.mur/hooks/on-stop.local.sh ] && source ~/.mur/hooks/on-stop.local.sh
`, CurrentHookVersion, murBin, murBin, murBin)
		if err := os.WriteFile(stopScript, []byte(content), 0755); err != nil {
			return fmt.Errorf("cannot write on-stop.sh: %w", err)
		}
		fmt.Printf("  + Created/upgraded %s (v%d)\n", stopScript, CurrentHookVersion)
	} else {
		fmt.Printf("  ~ Kept existing %s (v%d)\n", stopScript, parseHookVersion(stopScript))
	}

	promptScript := filepath.Join(hooksDir, "on-prompt.sh")
	if ShouldUpgradeHook(promptScript, opts.Force) {
		content := fmt.Sprintf(`#!/bin/bash
# mur-managed-hook v%d
# Read hook input from stdin (Claude Code passes JSON)
INPUT=$(cat /dev/stdin 2>/dev/null || echo '{}')

# Inject context-aware patterns based on current project
%s context --compact 2>/dev/null || true

# Record user prompt to active session (if recording)
if [ -f ~/.mur/session/active.json ]; then
  PROMPT=$(echo "$INPUT" | jq -r '.prompt // empty' 2>/dev/null)
  if [ -n "$PROMPT" ]; then
    %s session record --type user --content "$PROMPT" 2>/dev/null || true
  fi
fi
`, CurrentHookVersion, murBin, murBin)
		if err := os.WriteFile(promptScript, []byte(content), 0755); err != nil {
			return fmt.Errorf("cannot write on-prompt.sh: %w", err)
		}
		fmt.Printf("  + Created/upgraded %s (v%d)\n", promptScript, CurrentHookVersion)
	} else {
		fmt.Printf("  ~ Kept existing %s (v%d)\n", promptScript, parseHookVersion(promptScript))
	}

	// Create PostToolUse hook script for session recording
	onToolScript := filepath.Join(hooksDir, "on-tool.sh")
	if ShouldUpgradeHook(onToolScript, opts.Force) {
		content := fmt.Sprintf(`#!/bin/bash
# mur-managed-hook v%d
# Record tool usage to active session (if recording)
if [ -f ~/.mur/session/active.json ]; then
  INPUT=$(cat /dev/stdin 2>/dev/null || echo '{}')
  TOOL=$(echo "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null)
  TOOL_INPUT=$(echo "$INPUT" | jq -c '.tool_input // {}' 2>/dev/null)
  if [ -n "$TOOL" ]; then
    %s session record --type tool_call --tool "$TOOL" --content "$TOOL_INPUT" 2>/dev/null || true
  fi
fi
`, CurrentHookVersion, murBin)
		if err := os.WriteFile(onToolScript, []byte(content), 0755); err != nil {
			return fmt.Errorf("cannot write on-tool.sh: %w", err)
		}
		fmt.Printf("  + Created/upgraded %s (v%d)\n", onToolScript, CurrentHookVersion)
	} else {
		fmt.Printf("  ~ Kept existing %s (v%d)\n", onToolScript, parseHookVersion(onToolScript))
	}

	reminderFile := filepath.Join(hooksDir, "on-prompt-reminder.md")
	if _, err := os.Stat(reminderFile); os.IsNotExist(err) {
		content := fmt.Sprintf("[ContinuousLearning] If during this task you discover something non-obvious (a debugging technique, a workaround, a pattern), save it:\n\n  %s learn add --name \"pattern-name\" --content \"description\"\n\nOr create a file in ~/.mur/patterns/\n\nOnly save if: it required discovery, it helps future tasks, and it's verified.\n", murBin)
		if err := os.WriteFile(reminderFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("cannot write on-prompt-reminder.md: %w", err)
		}
	}

	// Read existing settings (preserve non-hook fields like permissions, etc.)
	var rawSettings map[string]json.RawMessage
	if data, err := os.ReadFile(settingsPath); err == nil {
		_ = json.Unmarshal(data, &rawSettings)
	}
	if rawSettings == nil {
		rawSettings = make(map[string]json.RawMessage)
	}

	// Read existing hooks to merge
	var existingHooks map[string]json.RawMessage
	if raw, ok := rawSettings["hooks"]; ok {
		_ = json.Unmarshal(raw, &existingHooks)
	}
	if existingHooks == nil {
		existingHooks = make(map[string]json.RawMessage)
	}

	// Build mur hook entries — pointing to shell scripts
	stopMatcher := ClaudeCodeHookMatcher{
		Matcher: "",
		Hooks: []ClaudeCodeHook{
			{Type: "command", Command: fmt.Sprintf("bash %s", stopScript)},
		},
	}

	promptHooks := []ClaudeCodeHook{
		{Type: "command", Command: fmt.Sprintf("cat %s >&2", reminderFile)},
	}
	if opts.EnableSearch {
		promptHooks = append(promptHooks, ClaudeCodeHook{
			Type:    "command",
			Command: fmt.Sprintf("%s search --inject \"$PROMPT\" 2>/dev/null || true", murBin),
		})
	}
	promptMatcher := ClaudeCodeHookMatcher{
		Matcher: "",
		Hooks:   promptHooks,
	}

	// PostToolUse matcher for session recording
	postToolMatcher := ClaudeCodeHookMatcher{
		Matcher: "",
		Hooks: []ClaudeCodeHook{
			{Type: "command", Command: fmt.Sprintf("bash %s", onToolScript)},
		},
	}

	// Merge: replace mur-managed matchers, keep user-added non-mur matchers
	existingHooks["Stop"] = mustMarshal(mergeMurMatcherSet(existingHooks["Stop"], stopMatcher))
	existingHooks["UserPromptSubmit"] = mustMarshal(mergeMurMatcherSet(existingHooks["UserPromptSubmit"], promptMatcher))
	existingHooks["PostToolUse"] = mustMarshal(mergeMurMatcherSet(existingHooks["PostToolUse"], postToolMatcher))

	// Write back
	rawSettings["hooks"] = mustMarshal(existingHooks)

	// Ensure .claude directory exists
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0755); err != nil {
		return fmt.Errorf("cannot create .claude directory: %w", err)
	}

	data, err := json.MarshalIndent(rawSettings, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("cannot write settings: %w", err)
	}

	fmt.Printf("✓ Installed Claude Code hooks at %s\n", settingsPath)
	fmt.Println("  + Stop hook → on-stop.sh (learn + sync)")
	fmt.Println("  + Prompt hook → on-prompt-reminder.md")
	fmt.Println("  + PostToolUse hook → on-tool.sh (record tool calls)")
	fmt.Println("  + Slash commands → /mur:in, /mur:out (session recording)")
	if opts.EnableSearch {
		fmt.Println("  + Search hook (suggests patterns on prompt)")
	}

	return nil
}

// mergeMurMatcherSet replaces all mur-managed matchers in an existing hook array
// with the given set, preserving any non-mur matchers.
func mergeMurMatcherSet(existing json.RawMessage, murMatchers ...ClaudeCodeHookMatcher) []ClaudeCodeHookMatcher {
	var matchers []ClaudeCodeHookMatcher
	if existing != nil {
		_ = json.Unmarshal(existing, &matchers)
	}

	// Keep non-mur matchers (those with a non-empty matcher or no mur references)
	var kept []ClaudeCodeHookMatcher
	for _, m := range matchers {
		if m.Matcher != "" && !isMurMatcher(m) {
			kept = append(kept, m)
		}
	}

	// Add all mur matchers
	return append(kept, murMatchers...)
}

// isMurMatcher checks if a matcher was created by mur.
func isMurMatcher(m ClaudeCodeHookMatcher) bool {
	for _, h := range m.Hooks {
		if strings.Contains(h.Command, ".mur/") ||
			strings.Contains(h.Command, "mur ") ||
			strings.HasPrefix(h.Command, "mur\t") {
			return true
		}
	}
	return false
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

// UninstallClaudeCodeHooks removes mur hooks from Claude Code.
func UninstallClaudeCodeHooks() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")

	// Read existing settings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to uninstall
		}
		return fmt.Errorf("cannot read settings: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("cannot parse settings: %w", err)
	}

	// Remove hooks
	delete(settings, "hooks")

	// Write back
	newData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, newData, 0644); err != nil {
		return fmt.Errorf("cannot write settings: %w", err)
	}

	fmt.Println("✓ Removed Claude Code hooks")
	return nil
}
