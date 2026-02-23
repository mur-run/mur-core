// Package hooks provides hook installation for AI CLI tools.
package hooks

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
)

// CurrentHookVersion is the version of mur-managed hook scripts.
// Bump this when the hook template changes to trigger auto-upgrade.
const CurrentHookVersion = 3

var hookVersionRe = regexp.MustCompile(`#\s*mur-managed-hook\s+v(\d+)`)

// parseHookVersion reads the first 5 lines of a file looking for
// "# mur-managed-hook v<N>" and returns N. Returns 0 if not found.
func parseHookVersion(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for i := 0; i < 5 && scanner.Scan(); i++ {
		if m := hookVersionRe.FindStringSubmatch(scanner.Text()); len(m) == 2 {
			v, _ := strconv.Atoi(m[1])
			return v
		}
	}
	return 0
}

// shouldUpgradeHook returns true if the hook file doesn't exist or has
// a version older than CurrentHookVersion.
func shouldUpgradeHook(path string) bool {
	return ShouldUpgradeHook(path, false)
}

// ShouldUpgradeHook returns true if the hook file should be overwritten.
// If force is true, always returns true (for --force flag).
// Otherwise returns true only if the file doesn't exist or has an older version.
func ShouldUpgradeHook(path string, force bool) bool {
	if force {
		return true
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return true
	}
	return parseHookVersion(path) < CurrentHookVersion
}

// ParseHookVersion reads the version tag from a hook file (exported for use by init).
func ParseHookVersion(path string) int {
	return parseHookVersion(path)
}

// findMurBinary finds the mur binary path.
func findMurBinary() (string, error) {
	// First try to find in PATH
	if path, err := exec.LookPath("mur"); err == nil {
		return path, nil
	}

	// Try common locations
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	commonPaths := []string{
		filepath.Join(home, "go", "bin", "mur"),
		"/usr/local/bin/mur",
		"/opt/homebrew/bin/mur",
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	// Default to just "mur" and hope it's in PATH at runtime
	return "mur", nil
}

// HookOptions configures hook installation.
type HookOptions struct {
	EnableSearch bool // Enable search hook on prompt submit
	Force        bool // Force overwrite even if hooks are up-to-date
}

// InstallAllHooks installs hooks for all supported AI tools.
func InstallAllHooks() map[string]error {
	return InstallAllHooksWithOptions(HookOptions{EnableSearch: false})
}

// InstallAllHooksWithOptions installs hooks with custom options.
func InstallAllHooksWithOptions(opts HookOptions) map[string]error {
	results := make(map[string]error)

	// Claude Code
	if ClaudeCodeInstalled() {
		if err := InstallClaudeCodeHooksWithOptions(opts); err != nil {
			results["Claude Code"] = err
		} else {
			results["Claude Code"] = nil
		}
	}

	// Gemini CLI
	if GeminiCLIInstalled() {
		if err := InstallGeminiHooks(opts.EnableSearch); err != nil {
			results["Gemini CLI"] = err
		} else {
			results["Gemini CLI"] = nil
		}
	}

	// OpenCode
	if err := InstallOpenCodeHooks(); err != nil {
		results["OpenCode"] = err
	} else {
		results["OpenCode"] = nil
	}

	// GitHub Copilot
	if err := InstallCopilotHooks(); err != nil {
		results["GitHub Copilot"] = err
	} else {
		results["GitHub Copilot"] = nil
	}

	// Continue.dev
	if ContinueDevInstalled() {
		if err := InstallContinueDevHooks(); err != nil {
			results["Continue.dev"] = err
		} else {
			results["Continue.dev"] = nil
		}
	}

	// Aider (optional - creates templates)
	if AiderInstalled() {
		if err := InstallAiderHooks(); err != nil {
			results["Aider"] = err
		} else {
			results["Aider"] = nil
		}
	}

	return results
}
