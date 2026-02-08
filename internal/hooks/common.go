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

// InstallAllHooks installs hooks for all supported AI tools.
func InstallAllHooks() map[string]error {
	results := make(map[string]error)

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

	return results
}
