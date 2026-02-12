// Package hooks provides hook installation for AI CLI tools.
package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

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
		if err := InstallClaudeCodeHooks(opts.EnableSearch); err != nil {
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
