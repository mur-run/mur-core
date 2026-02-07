package cmd

import (
	"fmt"

	"github.com/mur-run/mur-core/internal/team"
	"github.com/spf13/cobra"
)

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Manage team knowledge sharing",
	Long: `Git-based team knowledge sharing for patterns, hooks, and skills.

The team repo structure:
  team-repo/
  ├── patterns/     # Shared patterns
  ├── hooks/        # Shared hooks
  ├── skills/       # Shared skills
  └── mcp/          # Shared MCP config`,
}

var teamInitCmd = &cobra.Command{
	Use:   "init <repo-url>",
	Short: "Initialize team repo connection",
	Long: `Clone and connect to a team knowledge repo.

Examples:
  mur team init https://github.com/myteam/murmur-knowledge.git
  mur team init git@github.com:myteam/murmur-knowledge.git`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoURL := args[0]

		fmt.Printf("Cloning team repo: %s\n", repoURL)
		fmt.Println("")

		if err := team.Clone(repoURL); err != nil {
			return fmt.Errorf("failed to initialize team repo: %w", err)
		}

		// Ensure directory structure
		if err := team.EnsureStructure(); err != nil {
			fmt.Printf("⚠ Warning: could not create directory structure: %v\n", err)
		}

		dir, _ := team.TeamDir()
		fmt.Printf("✓ Team repo initialized at %s\n", dir)
		fmt.Println("")
		fmt.Println("Next steps:")
		fmt.Println("  mur team status    — Check repo status")
		fmt.Println("  mur team pull      — Pull latest changes")
		fmt.Println("  mur learn sync     — Sync patterns with team")

		return nil
	},
}

var teamPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull latest team changes",
	Long:  `Pull the latest changes from the team repo.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Pulling team changes...")

		if err := team.Pull(); err != nil {
			return fmt.Errorf("pull failed: %w", err)
		}

		fmt.Println("✓ Team repo updated")

		return nil
	},
}

var teamPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push local changes to team",
	Long:  `Push local team repo changes to remote.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		message, _ := cmd.Flags().GetString("message")

		fmt.Println("Pushing team changes...")

		if err := team.Push(message); err != nil {
			return fmt.Errorf("push failed: %w", err)
		}

		fmt.Println("✓ Changes pushed to team repo")

		return nil
	},
}

var teamSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Bidirectional sync with team",
	Long:  `Pull latest changes and push local modifications.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Syncing with team repo...")
		fmt.Println("")

		if err := team.Sync(); err != nil {
			return fmt.Errorf("sync failed: %w", err)
		}

		fmt.Println("✓ Team repo synced")

		return nil
	},
}

var teamStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show team repo status",
	Long:  `Display the current status of the team repo.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		status, err := team.Status()
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		fmt.Println("Team Repo Status")
		fmt.Println("================")
		fmt.Println("")

		if !status.Initialized {
			fmt.Println("Status: Not initialized")
			fmt.Println("")
			fmt.Println("Run 'mur team init <repo-url>' to connect to a team repo")
			return nil
		}

		fmt.Printf("Status:   Initialized\n")
		fmt.Printf("Repo:     %s\n", status.RepoURL)
		fmt.Printf("Branch:   %s\n", status.Branch)
		fmt.Printf("Path:     %s\n", status.LocalPath)
		fmt.Println("")

		if status.Ahead > 0 || status.Behind > 0 {
			fmt.Printf("Sync:     %d ahead, %d behind\n", status.Ahead, status.Behind)
		} else {
			fmt.Println("Sync:     Up to date")
		}

		if len(status.Modified) > 0 {
			fmt.Println("")
			fmt.Println("Modified:")
			for _, m := range status.Modified {
				fmt.Printf("  %s\n", m)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(teamCmd)
	teamCmd.AddCommand(teamInitCmd)
	teamCmd.AddCommand(teamPullCmd)
	teamCmd.AddCommand(teamPushCmd)
	teamCmd.AddCommand(teamSyncCmd)
	teamCmd.AddCommand(teamStatusCmd)

	teamPushCmd.Flags().StringP("message", "m", "", "Commit message")
}
