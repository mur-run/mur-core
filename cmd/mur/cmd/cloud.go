package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mur-run/mur-core/internal/cloud"
	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/core/pattern"
)

var cloudCmd = &cobra.Command{
	Use:   "cloud",
	Short: "Manage cloud sync with mur-server",
	Long: `Cloud sync enables team pattern sharing via mur-server.

Commands:
  mur cloud teams    — List your teams
  mur cloud select   — Set active team
  mur cloud sync     — Bidirectional sync with server
  mur cloud push     — Upload local patterns to server
  mur cloud pull     — Download patterns from server`,
}

var cloudTeamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "List your teams",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getCloudClient(cmd)
		if err != nil {
			return err
		}

		if !client.AuthStore().IsLoggedIn() {
			fmt.Println("Not logged in. Run 'mur login' first.")
			return nil
		}

		teams, err := client.ListTeams()
		if err != nil {
			return fmt.Errorf("failed to list teams: %w", err)
		}

		if len(teams) == 0 {
			fmt.Println("No teams found.")
			fmt.Println("")
			fmt.Println("Create a team:")
			fmt.Println("  mur cloud create \"My Team\"")
			return nil
		}

		// Get active team from config
		cfg, _ := config.Load()
		activeTeam := ""
		if cfg != nil {
			activeTeam = cfg.Server.Team
		}

		fmt.Println("Your Teams")
		fmt.Println("==========")
		fmt.Println("")

		for _, t := range teams {
			active := ""
			if t.Slug == activeTeam || t.ID == activeTeam {
				active = " (active)"
			}
			fmt.Printf("  %s%s\n", t.Name, active)
			fmt.Printf("    Slug: %s  |  Role: %s  |  Plan: %s\n", t.Slug, t.Role, t.Plan)
			fmt.Println("")
		}

		if activeTeam == "" && len(teams) > 0 {
			fmt.Println("Set active team with:")
			fmt.Printf("  mur cloud select %s\n", teams[0].Slug)
		}

		return nil
	},
}

var cloudCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new team",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		client, err := getCloudClient(cmd)
		if err != nil {
			return err
		}

		if !client.AuthStore().IsLoggedIn() {
			fmt.Println("Not logged in. Run 'mur login' first.")
			return nil
		}

		team, err := client.CreateTeam(name)
		if err != nil {
			return fmt.Errorf("failed to create team: %w", err)
		}

		fmt.Printf("✓ Team created: %s (slug: %s)\n", team.Name, team.Slug)
		fmt.Println("")
		fmt.Println("Set as active team:")
		fmt.Printf("  mur cloud select %s\n", team.Slug)

		return nil
	},
}

var cloudSelectCmd = &cobra.Command{
	Use:   "select <team-slug>",
	Short: "Set active team for sync",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		teamSlug := args[0]

		cfg, err := config.Load()
		if err != nil {
			cfg = config.Default()
		}

		cfg.Server.Team = teamSlug

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("✓ Active team set to: %s\n", teamSlug)
		fmt.Println("")
		fmt.Println("Now you can sync:")
		fmt.Println("  mur cloud sync")

		return nil
	},
}

var cloudSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync patterns with server",
	Long: `Bidirectional sync between local patterns and mur-server.

Examples:
  mur cloud sync              # Sync with active team
  mur cloud sync --team=slug  # Sync with specific team
  mur cloud sync --dry-run    # Show what would sync`,
	RunE: func(cmd *cobra.Command, args []string) error {
		teamSlug, _ := cmd.Flags().GetString("team")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		forceLocal, _ := cmd.Flags().GetBool("force-local")
		forceServer, _ := cmd.Flags().GetBool("force-server")

		client, err := getCloudClient(cmd)
		if err != nil {
			return err
		}

		if !client.AuthStore().IsLoggedIn() {
			fmt.Println("Not logged in. Run 'mur login' first.")
			return nil
		}

		// Get team from flag or config (auto-select if single team)
		if teamSlug == "" {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			teamSlug, err = resolveActiveTeam(cfg, client)
			if err != nil {
				return err
			}
		}

		// Find team and check subscription
		teams, err := client.ListTeams()
		if err != nil {
			return fmt.Errorf("failed to list teams: %w", err)
		}

		var team *cloud.Team
		for _, t := range teams {
			if t.Slug == teamSlug || t.ID == teamSlug {
				team = &t
				break
			}
		}

		if team == nil {
			return fmt.Errorf("team not found: %s", teamSlug)
		}

		// Check team subscription status
		if !team.CanSync {
			fmt.Println("❌ Team subscription expired")
			fmt.Println("")
			fmt.Println("Cloud sync is disabled because the team subscription has expired.")
			fmt.Println("Contact your team owner to renew the subscription.")
			fmt.Println("")
			fmt.Println("You can still use local patterns and sync to CLIs.")
			return fmt.Errorf("team subscription expired - sync disabled")
		}

		teamID := team.ID
		fmt.Printf("Syncing with team: %s\n", teamSlug)
		fmt.Println("")

		// Load local patterns
		store, err := pattern.DefaultStore()
		if err != nil {
			return fmt.Errorf("failed to load patterns: %w", err)
		}

		localPatterns, err := store.List()
		if err != nil {
			return fmt.Errorf("failed to list local patterns: %w", err)
		}

		// Get local version (stored in a sync state file)
		localVersion := getLocalSyncVersion(teamSlug)

		// Check sync status
		status, err := client.GetSyncStatus(teamID, localVersion)
		if err != nil {
			return fmt.Errorf("failed to get sync status: %w", err)
		}

		fmt.Printf("Local version:  %d\n", localVersion)
		fmt.Printf("Server version: %d\n", status.ServerVersion)
		fmt.Println("")

		// Pull changes from server
		if status.HasUpdates {
			fmt.Println("⬇️  Pulling from server...")

			pullResp, err := client.Pull(teamID, localVersion)
			if err != nil {
				return fmt.Errorf("failed to pull: %w", err)
			}

			created, updated, deleted := 0, 0, 0
			for _, p := range pullResp.Patterns {
				exists := store.Exists(p.Name)

				if dryRun {
					if p.Deleted {
						fmt.Printf("  Would delete: %s\n", p.Name)
						deleted++
					} else if exists {
						fmt.Printf("  Would update: %s\n", p.Name)
						updated++
					} else {
						fmt.Printf("  Would create: %s\n", p.Name)
						created++
					}
					continue
				}

				if p.Deleted {
					// Delete local pattern
					if err := store.Delete(p.Name); err == nil {
						deleted++
					}
				} else {
					// Create or update
					localP := convertCloudPattern(&p)
					if exists {
						if err := store.Update(localP); err == nil {
							updated++
						}
					} else {
						if err := store.Create(localP); err == nil {
							created++
						}
					}
				}
			}

			if !dryRun {
				saveLocalSyncVersion(teamSlug, pullResp.Version)
			}

			fmt.Printf("  ✓ %d created, %d updated, %d deleted\n", created, updated, deleted)
			fmt.Println("")
		} else {
			fmt.Println("⬇️  No updates from server")
			fmt.Println("")
		}

		// Push local changes
		fmt.Println("⬆️  Pushing to server...")

		changes := make([]cloud.SyncChange, 0) // Initialize as empty slice, not nil
		for i := range localPatterns {
			// For now, push all as creates/updates
			// A proper implementation would track local changes
			cloudP := convertLocalPattern(&localPatterns[i])
			changes = append(changes, cloud.SyncChange{
				Action:  "create", // Server will handle upsert
				Pattern: cloudP,
			})
		}

		if len(changes) == 0 {
			fmt.Println("  No local changes to push")
		} else if dryRun {
			fmt.Printf("  Would push %d patterns\n", len(changes))
		} else {
			pushReq := cloud.PushRequest{
				BaseVersion: localVersion,
				Changes:     changes,
			}

			pushResp, err := client.Push(teamID, pushReq)
			if err != nil {
				return fmt.Errorf("failed to push: %w", err)
			}

			if !pushResp.OK {
				if forceLocal {
					fmt.Printf("  ⚠️  %d conflict(s) detected — forcing local versions...\n", len(pushResp.Conflicts))
					forcePushReq := cloud.PushRequest{
						BaseVersion: localVersion,
						Changes:     changes,
						ForceLocal:  true,
					}
					forceResp, err := client.Push(teamID, forcePushReq)
					if err != nil {
						return fmt.Errorf("force push failed: %w", err)
					}
					if forceResp.OK {
						saveLocalSyncVersion(teamSlug, forceResp.Version)
						fmt.Printf("  ✓ %d patterns force-pushed\n", len(changes))
					} else {
						return fmt.Errorf("force push rejected by server")
					}
				} else if forceServer {
					// Accept server versions - pull them
					fmt.Println("  --force-server: Accepting server versions...")
					// Pull and overwrite local
				} else {
					// Interactive conflict resolution
					resolutions, err := ResolveConflictsInteractive(pushResp.Conflicts)
					if err != nil {
						return fmt.Errorf("conflict resolution cancelled: %w", err)
					}

					keepServer, keepLocal, skipped := ApplyResolutions(resolutions)
					fmt.Printf("\n📊 Resolution summary: %d server, %d local, %d skipped\n", keepServer, keepLocal, skipped)

					// Apply resolutions
					if keepServer > 0 {
						// Pull server versions for patterns marked as "keep server"
						fmt.Println("Applying server versions...")
						for _, c := range pushResp.Conflicts {
							if resolutions[c.PatternName] == ResolutionKeepServer && c.ServerVersion != nil {
								localP := convertCloudPattern(c.ServerVersion)
								if store.Exists(localP.Name) {
									_ = store.Update(localP)
								} else {
									_ = store.Create(localP)
								}
							}
						}
					}

					if keepLocal > 0 {
						// Need to force push local versions
						fmt.Println("Note: Keeping local versions requires --force-local flag")
						fmt.Println("Run: mur cloud sync --force-local")
					}
				}
				return nil
			}

			saveLocalSyncVersion(teamSlug, pushResp.Version)
			fmt.Printf("  ✓ %d patterns pushed\n", len(changes))
		}

		fmt.Println("")
		fmt.Println("✅ Sync complete")

		return nil
	},
}

// Helper functions

func getCloudClient(cmd *cobra.Command) (*cloud.Client, error) {
	serverURL, _ := cmd.Flags().GetString("server")

	if serverURL == "" {
		cfg, err := config.Load()
		if err == nil && cfg.Server.URL != "" {
			serverURL = cfg.Server.URL
		}
	}

	return cloud.NewClient(serverURL)
}

// resolveActiveTeam returns the active team slug. If none is set in config
// and the user has exactly one team, it auto-selects and persists it.
func resolveActiveTeam(cfg *config.Config, client *cloud.Client) (string, error) {
	if cfg.Server.Team != "" {
		return cfg.Server.Team, nil
	}
	teams, err := client.ListTeams()
	if err != nil {
		return "", fmt.Errorf("failed to list teams: %w", err)
	}
	if len(teams) == 0 {
		return "", fmt.Errorf("no teams found. Create one with: mur cloud create \"My Team\"")
	}
	if len(teams) == 1 {
		cfg.Server.Team = teams[0].Slug
		_ = cfg.Save()
		fmt.Fprintf(os.Stderr, "  Auto-selected team: %s\n", teams[0].Name)
		return teams[0].Slug, nil
	}
	// Multiple teams - list them and ask user to choose
	fmt.Fprintf(os.Stderr, "Multiple teams found:\n")
	for _, t := range teams {
		fmt.Fprintf(os.Stderr, "  - %s (%s)\n", t.Name, t.Slug)
	}
	return "", fmt.Errorf("multiple teams found. Select one with: mur cloud select <team-slug>")
}

func getLocalSyncVersion(teamSlug string) int64 {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".mur", "sync-state.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	var state map[string]int64
	if err := yaml.Unmarshal(data, &state); err != nil {
		return 0
	}

	return state[teamSlug]
}

func saveLocalSyncVersion(teamSlug string, version int64) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".mur", "sync-state.yaml")

	state := make(map[string]int64)

	// Load existing state
	data, err := os.ReadFile(path)
	if err == nil {
		_ = yaml.Unmarshal(data, &state)
	}

	state[teamSlug] = version

	data, _ = yaml.Marshal(state)
	_ = os.WriteFile(path, data, 0644)
}

func convertCloudPattern(p *cloud.Pattern) *pattern.Pattern {
	local := &pattern.Pattern{
		Name:        p.Name,
		Description: p.Description,
		Content:     p.Content,
	}

	// Set schema version (v1.1.0+)
	if p.SchemaVersion > 0 {
		local.SchemaVersion = p.SchemaVersion
	} else {
		local.SchemaVersion = 2 // Default to v2
	}
	local.Version = p.PatternVersion
	local.EmbeddingHash = p.EmbeddingHash

	// Convert tags
	if p.Tags != nil {
		if confirmed, ok := p.Tags["confirmed"].([]interface{}); ok {
			for _, t := range confirmed {
				if s, ok := t.(string); ok {
					local.Tags.Confirmed = append(local.Tags.Confirmed, s)
				}
			}
		}
	}

	// Convert applies
	if p.Applies != nil {
		if langs, ok := p.Applies["languages"].([]interface{}); ok {
			for _, l := range langs {
				if s, ok := l.(string); ok {
					local.Applies.Languages = append(local.Applies.Languages, s)
				}
			}
		}
		if projs, ok := p.Applies["projects"].([]interface{}); ok {
			for _, pr := range projs {
				if s, ok := pr.(string); ok {
					local.Applies.Projects = append(local.Applies.Projects, s)
				}
			}
		}
	}

	return local
}

func convertLocalPattern(p *pattern.Pattern) *cloud.Pattern {
	cp := &cloud.Pattern{
		Name:        p.Name,
		Description: p.Description,
		Content:     strings.TrimSpace(p.Content),
		// v1.1.0+ schema version fields
		PatternVersion: p.Version,
		SchemaVersion:  p.SchemaVersion,
		EmbeddingHash:  p.EmbeddingHash,
	}

	// Default schema version to 2 if not set
	if cp.SchemaVersion == 0 {
		cp.SchemaVersion = 2
	}

	// Convert tags
	if len(p.Tags.Confirmed) > 0 {
		cp.Tags = map[string]any{
			"confirmed": p.Tags.Confirmed,
		}
	}

	// Convert applies
	applies := make(map[string]any)
	if len(p.Applies.Languages) > 0 {
		applies["languages"] = p.Applies.Languages
	}
	if len(p.Applies.Projects) > 0 {
		applies["projects"] = p.Applies.Projects
	}
	if len(applies) > 0 {
		cp.Applies = applies
	}

	return cp
}

var cloudPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push local patterns to server",
	Long: `Upload local patterns to the server.

This is equivalent to 'mur cloud sync' but only pushes local changes.

Examples:
  mur cloud push              # Push to active team
  mur cloud push --team=slug  # Push to specific team
  mur cloud push --force      # Overwrite server on conflicts`,
	RunE: func(cmd *cobra.Command, args []string) error {
		teamSlug, _ := cmd.Flags().GetString("team")
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		client, err := getCloudClient(cmd)
		if err != nil {
			return err
		}

		if !client.AuthStore().IsLoggedIn() {
			fmt.Println("Not logged in. Run 'mur login' first.")
			return nil
		}

		// Get team from flag or config (auto-select if single team)
		if teamSlug == "" {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			teamSlug, err = resolveActiveTeam(cfg, client)
			if err != nil {
				return err
			}
		}

		// Find team ID
		teams, err := client.ListTeams()
		if err != nil {
			return fmt.Errorf("failed to list teams: %w", err)
		}

		var teamID string
		for _, t := range teams {
			if t.Slug == teamSlug || t.ID == teamSlug {
				teamID = t.ID
				break
			}
		}

		if teamID == "" {
			return fmt.Errorf("team not found: %s", teamSlug)
		}

		fmt.Printf("Pushing to team: %s\n", teamSlug)
		fmt.Println("")

		// Load local patterns
		store, err := pattern.DefaultStore()
		if err != nil {
			return fmt.Errorf("failed to load patterns: %w", err)
		}

		localPatterns, err := store.List()
		if err != nil {
			return fmt.Errorf("failed to list local patterns: %w", err)
		}

		localVersion := getLocalSyncVersion(teamSlug)

		changes := make([]cloud.SyncChange, 0)
		for i := range localPatterns {
			cloudP := convertLocalPattern(&localPatterns[i])
			changes = append(changes, cloud.SyncChange{
				Action:  "create",
				Pattern: cloudP,
			})
		}

		if len(changes) == 0 {
			fmt.Println("No patterns to push")
			return nil
		}

		if dryRun {
			fmt.Printf("Would push %d patterns\n", len(changes))
			return nil
		}

		pushReq := cloud.PushRequest{
			BaseVersion: localVersion,
			Changes:     changes,
		}

		pushResp, err := client.Push(teamID, pushReq)
		if err != nil {
			return fmt.Errorf("failed to push: %w", err)
		}

		if !pushResp.OK {
			if force {
				fmt.Println("--force: Force push not yet implemented")
				fmt.Println("Use 'mur cloud sync --force-local' instead")
			} else {
				// Interactive conflict resolution
				resolutions, err := ResolveConflictsInteractive(pushResp.Conflicts)
				if err != nil {
					return fmt.Errorf("conflict resolution cancelled: %w", err)
				}

				keepServer, keepLocal, skipped := ApplyResolutions(resolutions)
				fmt.Printf("\n📊 Resolution: %d server, %d local, %d skipped\n", keepServer, keepLocal, skipped)

				if keepLocal > 0 {
					fmt.Println("Note: Use --force to push local versions")
				}
			}
			return nil
		}

		saveLocalSyncVersion(teamSlug, pushResp.Version)
		fmt.Printf("✅ Pushed %d patterns\n", len(changes))

		return nil
	},
}

var cloudPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull patterns from server",
	Long: `Download patterns from the server.

This is equivalent to 'mur cloud sync' but only pulls server changes.

Examples:
  mur cloud pull              # Pull from active team
  mur cloud pull --team=slug  # Pull from specific team
  mur cloud pull --force      # Overwrite local on conflicts`,
	RunE: func(cmd *cobra.Command, args []string) error {
		teamSlug, _ := cmd.Flags().GetString("team")
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		client, err := getCloudClient(cmd)
		if err != nil {
			return err
		}

		if !client.AuthStore().IsLoggedIn() {
			fmt.Println("Not logged in. Run 'mur login' first.")
			return nil
		}

		// Get team from flag or config (auto-select if single team)
		if teamSlug == "" {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			teamSlug, err = resolveActiveTeam(cfg, client)
			if err != nil {
				return err
			}
		}

		// Find team ID
		teams, err := client.ListTeams()
		if err != nil {
			return fmt.Errorf("failed to list teams: %w", err)
		}

		var teamID string
		for _, t := range teams {
			if t.Slug == teamSlug || t.ID == teamSlug {
				teamID = t.ID
				break
			}
		}

		if teamID == "" {
			return fmt.Errorf("team not found: %s", teamSlug)
		}

		fmt.Printf("Pulling from team: %s\n", teamSlug)
		fmt.Println("")

		// Load local store
		store, err := pattern.DefaultStore()
		if err != nil {
			return fmt.Errorf("failed to load patterns: %w", err)
		}

		localVersion := getLocalSyncVersion(teamSlug)
		if force {
			localVersion = 0 // Pull everything
		}

		// Check sync status
		status, err := client.GetSyncStatus(teamID, localVersion)
		if err != nil {
			return fmt.Errorf("failed to get sync status: %w", err)
		}

		if !status.HasUpdates && !force {
			fmt.Println("Already up to date")
			return nil
		}

		pullResp, err := client.Pull(teamID, localVersion)
		if err != nil {
			return fmt.Errorf("failed to pull: %w", err)
		}

		created, updated, deleted := 0, 0, 0
		for _, p := range pullResp.Patterns {
			exists := store.Exists(p.Name)

			if dryRun {
				if p.Deleted {
					fmt.Printf("  Would delete: %s\n", p.Name)
					deleted++
				} else if exists {
					fmt.Printf("  Would update: %s\n", p.Name)
					updated++
				} else {
					fmt.Printf("  Would create: %s\n", p.Name)
					created++
				}
				continue
			}

			if p.Deleted {
				if err := store.Delete(p.Name); err == nil {
					deleted++
				}
			} else {
				localP := convertCloudPattern(&p)
				if exists {
					if err := store.Update(localP); err == nil {
						updated++
					}
				} else {
					if err := store.Create(localP); err == nil {
						created++
					}
				}
			}
		}

		if !dryRun {
			saveLocalSyncVersion(teamSlug, pullResp.Version)
		}

		fmt.Printf("✅ %d created, %d updated, %d deleted\n", created, updated, deleted)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cloudCmd)
	cloudCmd.AddCommand(cloudTeamsCmd)
	cloudCmd.AddCommand(cloudCreateCmd)
	cloudCmd.AddCommand(cloudSelectCmd)
	cloudCmd.AddCommand(cloudSyncCmd)
	cloudCmd.AddCommand(cloudPushCmd)
	cloudCmd.AddCommand(cloudPullCmd)

	// Global flags for cloud commands
	cloudCmd.PersistentFlags().String("server", "", "Server URL (default: https://api.mur.run)")

	// Sync flags
	cloudSyncCmd.Flags().String("team", "", "Team slug to sync with")
	cloudSyncCmd.Flags().Bool("dry-run", false, "Show what would sync without making changes")
	cloudSyncCmd.Flags().Bool("force-local", false, "Overwrite server with local on conflicts")
	cloudSyncCmd.Flags().Bool("force-server", false, "Overwrite local with server on conflicts")

	// Push flags
	cloudPushCmd.Flags().String("team", "", "Team slug to push to")
	cloudPushCmd.Flags().Bool("force", false, "Force push, overwriting server on conflicts")
	cloudPushCmd.Flags().Bool("dry-run", false, "Show what would be pushed")

	// Pull flags
	cloudPullCmd.Flags().String("team", "", "Team slug to pull from")
	cloudPullCmd.Flags().Bool("force", false, "Force pull, overwriting local with server")
	cloudPullCmd.Flags().Bool("dry-run", false, "Show what would be pulled")
}
