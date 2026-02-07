// Package learning provides learning repo sync functionality.
package learning

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/learn"
)

// RepoDir returns the path to the learning repo (~/.mur/learning-repo/).
func RepoDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".mur", "learning-repo"), nil
}

// IsInitialized checks if the learning repo has been initialized.
func IsInitialized() bool {
	dir, err := RepoDir()
	if err != nil {
		return false
	}
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// DefaultBranch returns the default branch name (hostname).
func DefaultBranch() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "local"
	}
	// Sanitize hostname for git branch name
	hostname = strings.ReplaceAll(hostname, " ", "-")
	hostname = strings.ToLower(hostname)
	return hostname
}

// GetBranch returns the configured branch or default (hostname).
func GetBranch() (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return DefaultBranch(), nil
	}
	if cfg.Learning.Branch != "" {
		return cfg.Learning.Branch, nil
	}
	return DefaultBranch(), nil
}

// InitRepo clones the learning repo and sets up the branch.
func InitRepo(repoURL string) error {
	dir, err := RepoDir()
	if err != nil {
		return err
	}

	// Check if already initialized
	if IsInitialized() {
		return fmt.Errorf("learning repo already initialized at %s", dir)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return fmt.Errorf("cannot create directory: %w", err)
	}

	// Clone the repo
	cmd := exec.Command("git", "clone", repoURL, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Get branch name
	branch, err := GetBranch()
	if err != nil {
		branch = DefaultBranch()
	}

	// Create and checkout the branch
	cmd = exec.Command("git", "checkout", "-B", branch)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}

	// Save repo URL to config
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}
	cfg.Learning.Repo = repoURL
	if cfg.Learning.Branch == "" {
		cfg.Learning.Branch = branch
	}
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("cannot save config: %w", err)
	}

	// Sync local patterns to repo
	if err := syncPatternsToRepo(); err != nil {
		return fmt.Errorf("cannot sync patterns: %w", err)
	}

	return nil
}

// Push commits and pushes patterns to the configured branch.
func Push() error {
	if !IsInitialized() {
		return fmt.Errorf("learning repo not initialized (run: mur learn init <repo-url>)")
	}

	dir, err := RepoDir()
	if err != nil {
		return err
	}

	branch, err := GetBranch()
	if err != nil {
		return err
	}

	// Sync local patterns to repo directory
	if err := syncPatternsToRepo(); err != nil {
		return fmt.Errorf("cannot sync patterns: %w", err)
	}

	// Check if there are changes
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}

	if len(strings.TrimSpace(string(output))) == 0 {
		return nil // No changes to push
	}

	// Add all changes
	cmd = exec.Command("git", "add", "-A")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	// Commit
	hostname, _ := os.Hostname()
	commitMsg := fmt.Sprintf("Update patterns from %s", hostname)
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		// Ignore if nothing to commit
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil
		}
		return fmt.Errorf("git commit failed: %w", err)
	}

	// Push to origin
	cmd = exec.Command("git", "push", "-u", "origin", branch)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	return nil
}

// Pull fetches and merges patterns from the main branch.
func Pull() error {
	if !IsInitialized() {
		return fmt.Errorf("learning repo not initialized (run: mur learn init <repo-url>)")
	}

	dir, err := RepoDir()
	if err != nil {
		return err
	}

	// Fetch from origin
	cmd := exec.Command("git", "fetch", "origin", "main")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// main might not exist yet, try master
		cmd = exec.Command("git", "fetch", "origin", "master")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git fetch failed: %w", err)
		}
	}

	// Try to merge from origin/main (or origin/master)
	cmd = exec.Command("git", "merge", "origin/main", "--no-edit", "--allow-unrelated-histories")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		// Try master instead
		cmd = exec.Command("git", "merge", "origin/master", "--no-edit", "--allow-unrelated-histories")
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			// If merge fails, it might just mean no main branch exists yet
			return nil
		}
	}

	// Import patterns from repo to local
	if err := syncPatternsFromRepo(); err != nil {
		return fmt.Errorf("cannot import patterns: %w", err)
	}

	return nil
}

// Sync pushes to own branch and pulls from main.
func Sync() error {
	// First push local changes
	if err := Push(); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	// Then pull from main
	cfg, err := config.Load()
	if err == nil && cfg.Learning.PullFromMain {
		if err := Pull(); err != nil {
			return fmt.Errorf("pull failed: %w", err)
		}
	}

	return nil
}

// syncPatternsToRepo copies local patterns to the repo directory.
func syncPatternsToRepo() error {
	patternsDir, err := learn.PatternsDir()
	if err != nil {
		return err
	}

	repoDir, err := RepoDir()
	if err != nil {
		return err
	}

	repoPatternsDir := filepath.Join(repoDir, "patterns")
	if err := os.MkdirAll(repoPatternsDir, 0755); err != nil {
		return err
	}

	// List local patterns
	entries, err := os.ReadDir(patternsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No patterns yet
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		srcPath := filepath.Join(patternsDir, entry.Name())
		dstPath := filepath.Join(repoPatternsDir, entry.Name())

		if err := copyFile(srcPath, dstPath); err != nil {
			continue // Skip files we can't copy
		}
	}

	return nil
}

// syncPatternsFromRepo imports patterns from repo to local.
func syncPatternsFromRepo() error {
	repoDir, err := RepoDir()
	if err != nil {
		return err
	}

	repoPatternsDir := filepath.Join(repoDir, "patterns")
	if _, err := os.Stat(repoPatternsDir); os.IsNotExist(err) {
		return nil // No patterns in repo
	}

	patternsDir, err := learn.PatternsDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(patternsDir, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(repoPatternsDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		srcPath := filepath.Join(repoPatternsDir, entry.Name())
		dstPath := filepath.Join(patternsDir, entry.Name())

		// Don't overwrite existing local patterns (local wins)
		if _, err := os.Stat(dstPath); err == nil {
			continue
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			continue
		}
	}

	return nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// AutoMergeResult contains the result of an auto-merge operation.
type AutoMergeResult struct {
	PatternsChecked int
	PRsCreated      int
	PRsFailed       int
	Patterns        []PatternPRResult
}

// PatternPRResult contains the result for a single pattern PR.
type PatternPRResult struct {
	Pattern learn.Pattern
	PRURL   string
	Error   error
}

// GetHighConfidencePatterns returns patterns with confidence >= threshold.
func GetHighConfidencePatterns(threshold float64) ([]learn.Pattern, error) {
	patterns, err := learn.List()
	if err != nil {
		return nil, fmt.Errorf("cannot list patterns: %w", err)
	}

	var highConfidence []learn.Pattern
	for _, p := range patterns {
		if p.Confidence >= threshold {
			highConfidence = append(highConfidence, p)
		}
	}

	return highConfidence, nil
}

// CreatePatternPR creates a GitHub PR for a pattern using gh CLI.
func CreatePatternPR(pattern learn.Pattern, dryRun bool) (string, error) {
	if !IsInitialized() {
		return "", fmt.Errorf("learning repo not initialized")
	}

	dir, err := RepoDir()
	if err != nil {
		return "", err
	}

	branch, err := GetBranch()
	if err != nil {
		return "", err
	}

	// Build PR title and body
	title := fmt.Sprintf("Add pattern: %s", pattern.Name)
	body := fmt.Sprintf(`## Pattern: %s

**Description:** %s

**Domain:** %s  
**Category:** %s  
**Confidence:** %.0f%%

### Content Preview

%s
`,
		pattern.Name,
		pattern.Description,
		pattern.Domain,
		pattern.Category,
		pattern.Confidence*100,
		truncateContent(pattern.Content, 500),
	)

	if dryRun {
		return fmt.Sprintf("[dry-run] Would create PR: %s", title), nil
	}

	// Check if gh CLI is available
	if _, err := exec.LookPath("gh"); err != nil {
		return "", fmt.Errorf("gh CLI not found (install: https://cli.github.com/)")
	}

	// Create PR using gh CLI
	cmd := exec.Command("gh", "pr", "create",
		"--title", title,
		"--body", body,
		"--base", "main",
		"--head", branch,
		"--label", "auto-merge",
		"--label", "pattern",
	)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if PR already exists
		if strings.Contains(string(output), "already exists") {
			return "", fmt.Errorf("PR already exists for this branch")
		}
		return "", fmt.Errorf("gh pr create failed: %s", string(output))
	}

	// Extract PR URL from output
	prURL := strings.TrimSpace(string(output))
	return prURL, nil
}

// AutoMerge checks patterns and creates PRs for high-confidence ones.
func AutoMerge(dryRun bool) (*AutoMergeResult, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("learning repo not initialized (run: mur learn init <repo-url>)")
	}

	// Load config for threshold
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	threshold := cfg.Learning.MergeThreshold
	if threshold == 0 {
		threshold = 0.8 // Default threshold
	}

	// Get high-confidence patterns
	patterns, err := GetHighConfidencePatterns(threshold)
	if err != nil {
		return nil, err
	}

	result := &AutoMergeResult{
		PatternsChecked: len(patterns),
	}

	if len(patterns) == 0 {
		return result, nil
	}

	// Push changes first to ensure branch is up to date
	if !dryRun {
		if err := Push(); err != nil {
			return nil, fmt.Errorf("push failed: %w", err)
		}
	}

	// Create PRs for each pattern
	for _, p := range patterns {
		prResult := PatternPRResult{Pattern: p}

		prURL, err := CreatePatternPR(p, dryRun)
		if err != nil {
			prResult.Error = err
			result.PRsFailed++
		} else {
			prResult.PRURL = prURL
			result.PRsCreated++
		}

		result.Patterns = append(result.Patterns, prResult)
	}

	return result, nil
}

// truncateContent shortens content for PR body.
func truncateContent(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
