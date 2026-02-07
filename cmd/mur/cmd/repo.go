package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage learning repository",
	Long: `Manage the git repository for storing learned patterns.

Examples:
  mur repo set git@github.com:user/my-learnings.git
  mur repo status
  mur repo remove`,
}

var repoSetCmd = &cobra.Command{
	Use:   "set <repo-url>",
	Short: "Set or change the learning repository",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRepoSet,
}

var repoStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show learning repository status",
	RunE:  runRepoStatus,
}

var repoRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove learning repository configuration",
	RunE:  runRepoRemove,
}

func init() {
	rootCmd.AddCommand(repoCmd)
	repoCmd.AddCommand(repoSetCmd)
	repoCmd.AddCommand(repoStatusCmd)
	repoCmd.AddCommand(repoRemoveCmd)
}

func runRepoSet(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var repoURL string
	if len(args) > 0 {
		repoURL = args[0]
	} else {
		prompt := &survey.Input{
			Message: "Learning repo URL:",
			Help:    "e.g., git@github.com:username/my-learnings.git",
		}
		if err := survey.AskOne(prompt, &repoURL); err != nil {
			return err
		}
	}

	if repoURL == "" {
		return fmt.Errorf("repo URL is required")
	}

	patternsDir := filepath.Join(home, ".mur", "repo")

	// Check if patterns dir exists and has content
	if entries, err := os.ReadDir(patternsDir); err == nil && len(entries) > 0 {
		// Check if it's already a git repo
		gitDir := filepath.Join(patternsDir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			// Already a git repo, update remote
			fmt.Println("Updating remote origin...")
			cmd := exec.Command("git", "-C", patternsDir, "remote", "set-url", "origin", repoURL)
			if err := cmd.Run(); err != nil {
				// Try adding remote instead
				cmd = exec.Command("git", "-C", patternsDir, "remote", "add", "origin", repoURL)
				cmd.Run()
			}
		} else {
			// Has content but not a git repo - backup and clone
			backupDir := patternsDir + ".backup"
			fmt.Printf("Backing up existing patterns to %s\n", backupDir)
			os.Rename(patternsDir, backupDir)

			// Clone new repo
			fmt.Println("Cloning repository...")
			cmd := exec.Command("git", "clone", repoURL, patternsDir)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				// Restore backup on failure
				os.Rename(backupDir, patternsDir)
				return fmt.Errorf("failed to clone: %w", err)
			}
		}
	} else {
		// Empty or doesn't exist - just clone
		os.MkdirAll(filepath.Dir(patternsDir), 0755)
		fmt.Println("Cloning repository...")
		cmd := exec.Command("git", "clone", repoURL, patternsDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone: %w", err)
		}
	}

	// Save repo URL to config
	if err := saveRepoConfig(home, repoURL); err != nil {
		fmt.Printf("⚠ Warning: couldn't save config: %v\n", err)
	}

	fmt.Println()
	fmt.Println("✅ Learning repo configured!")
	fmt.Printf("   %s\n", repoURL)
	fmt.Println()
	fmt.Println("Run 'mur sync' to sync patterns to your CLIs.")

	return nil
}

func runRepoStatus(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	patternsDir := filepath.Join(home, ".mur", "repo")
	gitDir := filepath.Join(patternsDir, ".git")

	// Check if it's a git repo
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		fmt.Println("No learning repo configured.")
		fmt.Println()
		fmt.Println("Set one with: mur repo set <url>")
		return nil
	}

	// Get remote URL
	remoteCmd := exec.Command("git", "-C", patternsDir, "remote", "get-url", "origin")
	remote, err := remoteCmd.Output()
	if err != nil {
		fmt.Println("Learning repo: (local only, no remote)")
	} else {
		fmt.Printf("Learning repo: %s\n", strings.TrimSpace(string(remote)))
	}

	// Get current branch
	branchCmd := exec.Command("git", "-C", patternsDir, "rev-parse", "--abbrev-ref", "HEAD")
	branch, _ := branchCmd.Output()
	fmt.Printf("Branch: %s\n", strings.TrimSpace(string(branch)))

	// Get status
	statusCmd := exec.Command("git", "-C", patternsDir, "status", "--short")
	status, _ := statusCmd.Output()
	if len(status) > 0 {
		fmt.Println("Changes:")
		fmt.Print(string(status))
	} else {
		fmt.Println("Status: Clean")
	}

	// Count patterns
	count := 0
	filepath.Walk(patternsDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, ".yaml") {
			count++
		}
		return nil
	})
	fmt.Printf("Patterns: %d\n", count)

	return nil
}

func runRepoRemove(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	patternsDir := filepath.Join(home, ".mur", "repo")
	gitDir := filepath.Join(patternsDir, ".git")

	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		fmt.Println("No learning repo configured.")
		return nil
	}

	var confirm bool
	prompt := &survey.Confirm{
		Message: "Remove learning repo configuration? (patterns will be kept)",
		Default: false,
	}
	if err := survey.AskOne(prompt, &confirm); err != nil {
		return err
	}

	if !confirm {
		fmt.Println("Cancelled.")
		return nil
	}

	// Remove .git directory
	if err := os.RemoveAll(gitDir); err != nil {
		return fmt.Errorf("failed to remove git config: %w", err)
	}

	// Remove from config
	saveRepoConfig(home, "")

	fmt.Println("✅ Learning repo removed. Patterns are still available locally.")

	return nil
}

func saveRepoConfig(home, repoURL string) error {
	configPath := filepath.Join(home, ".mur", "config.yaml")

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	// Update learning section
	learning, ok := config["learning"].(map[string]interface{})
	if !ok {
		learning = make(map[string]interface{})
	}

	if repoURL != "" {
		learning["repo"] = repoURL
	} else {
		delete(learning, "repo")
	}
	config["learning"] = learning

	// Write back
	newData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, newData, 0644)
}

// SetupLearningRepo is called from init to optionally set up a learning repo
func SetupLearningRepo(home string) error {
	var useRepo bool
	prompt := &survey.Confirm{
		Message: "Use a git repo for patterns? (enables sync across machines)",
		Default: false,
	}
	if err := survey.AskOne(prompt, &useRepo); err != nil {
		return err
	}

	if !useRepo {
		return nil
	}

	var repoURL string
	urlPrompt := &survey.Input{
		Message: "Repo URL:",
		Help:    "e.g., git@github.com:username/my-learnings.git",
	}
	if err := survey.AskOne(urlPrompt, &repoURL); err != nil {
		return err
	}

	if repoURL == "" {
		fmt.Println("  Skipped (no URL provided)")
		return nil
	}

	// Clone the repo
	patternsDir := filepath.Join(home, ".mur", "repo")
	os.MkdirAll(filepath.Dir(patternsDir), 0755)

	fmt.Println("  Cloning repository...")
	cmd := exec.Command("git", "clone", repoURL, patternsDir)
	if err := cmd.Run(); err != nil {
		fmt.Printf("  ⚠ Clone failed: %v\n", err)
		return nil
	}

	// Save to config
	saveRepoConfig(home, repoURL)

	fmt.Println("  ✓ Learning repo configured")
	return nil
}
