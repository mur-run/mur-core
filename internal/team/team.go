// Package team provides Git-based team knowledge sharing for murmur-ai.
package team

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mur-run/mur-core/internal/config"
)

// TeamStatus represents the current state of the team repo.
type TeamStatus struct {
	Initialized bool
	RepoURL     string
	Branch      string
	LocalPath   string
	Ahead       int
	Behind      int
	Modified    []string
}

// TeamDir returns the path to ~/.mur/team/
func TeamDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".mur", "team"), nil
}

// IsInitialized checks if the team repo is configured and cloned.
func IsInitialized() bool {
	dir, err := TeamDir()
	if err != nil {
		return false
	}

	// Check if .git directory exists
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// Clone clones the team repo to ~/.mur/team/
func Clone(repoURL string) error {
	dir, err := TeamDir()
	if err != nil {
		return err
	}

	// Check if already initialized
	if IsInitialized() {
		return fmt.Errorf("team repo already initialized at %s", dir)
	}

	// Create parent directory
	parent := filepath.Dir(dir)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("cannot create directory: %w", err)
	}

	// Get branch from config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("cannot load config: %w", err)
	}

	branch := cfg.Team.Branch
	if branch == "" {
		branch = "main"
	}

	// Clone the repo
	args := []string{"clone", "--branch", branch, repoURL, dir}
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %s\n%w", string(output), err)
	}

	// Update config with repo URL
	cfg.Team.Repo = repoURL
	cfg.Team.Branch = branch
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("cannot save config: %w", err)
	}

	return nil
}

// Pull pulls the latest changes from the remote.
func Pull() error {
	if !IsInitialized() {
		return fmt.Errorf("team repo not initialized, run 'mur team init <repo-url>' first")
	}

	output, err := runGit("pull", "--rebase")
	if err != nil {
		return fmt.Errorf("git pull failed: %s\n%w", output, err)
	}

	return nil
}

// Push pushes local changes to the remote.
func Push(message string) error {
	if !IsInitialized() {
		return fmt.Errorf("team repo not initialized, run 'mur team init <repo-url>' first")
	}

	// Stage all changes
	if _, err := runGit("add", "-A"); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	// Check if there are changes to commit
	status, err := runGit("status", "--porcelain")
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}

	if strings.TrimSpace(status) == "" {
		return fmt.Errorf("no changes to push")
	}

	// Commit
	if message == "" {
		message = "Update team knowledge base"
	}
	if _, err := runGit("commit", "-m", message); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	// Push
	output, err := runGit("push")
	if err != nil {
		return fmt.Errorf("git push failed: %s\n%w", output, err)
	}

	return nil
}

// Sync performs bidirectional sync (pull then push).
func Sync() error {
	if !IsInitialized() {
		return fmt.Errorf("team repo not initialized, run 'mur team init <repo-url>' first")
	}

	// First pull
	if err := Pull(); err != nil {
		// Ignore "Already up to date" type messages
		if !strings.Contains(err.Error(), "Already up to date") {
			return fmt.Errorf("pull failed: %w", err)
		}
	}

	// Then try to push if there are changes
	if err := Push("Sync team knowledge base"); err != nil {
		// Ignore "no changes to push"
		if !strings.Contains(err.Error(), "no changes to push") {
			return fmt.Errorf("push failed: %w", err)
		}
	}

	return nil
}

// Status returns the current team repo status.
func Status() (*TeamStatus, error) {
	dir, err := TeamDir()
	if err != nil {
		return nil, err
	}

	status := &TeamStatus{
		Initialized: IsInitialized(),
		LocalPath:   dir,
	}

	if !status.Initialized {
		return status, nil
	}

	// Get remote URL
	if url, err := runGit("remote", "get-url", "origin"); err == nil {
		status.RepoURL = strings.TrimSpace(url)
	}

	// Get current branch
	if branch, err := runGit("branch", "--show-current"); err == nil {
		status.Branch = strings.TrimSpace(branch)
	}

	// Fetch to get latest remote state
	_, _ = runGit("fetch")

	// Get ahead/behind count
	if output, err := runGit("rev-list", "--left-right", "--count", "HEAD...@{upstream}"); err == nil {
		parts := strings.Fields(strings.TrimSpace(output))
		if len(parts) == 2 {
			_, _ = fmt.Sscanf(parts[0], "%d", &status.Ahead)
			_, _ = fmt.Sscanf(parts[1], "%d", &status.Behind)
		}
	}

	// Get modified files
	if output, err := runGit("status", "--porcelain"); err == nil {
		lines := strings.Split(strings.TrimSpace(output), "\n")
		for _, line := range lines {
			if line = strings.TrimSpace(line); line != "" {
				status.Modified = append(status.Modified, line)
			}
		}
	}

	return status, nil
}

// PatternsDir returns the path to team patterns directory.
func PatternsDir() (string, error) {
	dir, err := TeamDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "patterns"), nil
}

// HooksDir returns the path to team hooks directory.
func HooksDir() (string, error) {
	dir, err := TeamDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "hooks"), nil
}

// SkillsDir returns the path to team skills directory.
func SkillsDir() (string, error) {
	dir, err := TeamDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "skills"), nil
}

// MCPDir returns the path to team MCP directory.
func MCPDir() (string, error) {
	dir, err := TeamDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mcp"), nil
}

// EnsureStructure creates the team repo directory structure.
func EnsureStructure() error {
	dirs := []func() (string, error){
		PatternsDir,
		HooksDir,
		SkillsDir,
		MCPDir,
	}

	for _, dirFn := range dirs {
		dir, err := dirFn()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create directory %s: %w", dir, err)
		}
	}

	return nil
}

// runGit executes a git command in the team directory.
func runGit(args ...string) (string, error) {
	dir, err := TeamDir()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}
