package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/cloud"
	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/session"
	"github.com/mur-run/mur-core/internal/workflow"
)

var workflowsCmd = &cobra.Command{
	Use:     "workflows",
	Aliases: []string{"wf"},
	Short:   "Manage workflows",
	Long: `Create, view, run, and export reusable workflows.

Workflows are editable, versionable action sequences extracted from
recorded sessions. Unlike patterns (knowledge snippets), workflows
are complete SOPs that can be replayed and shared.

Commands:
  mur workflows list                          List local workflows
  mur workflows show <id>                     Show workflow details
  mur workflows create --from-session <id>    Create from a session
  mur workflows run <id>                      Execute workflow locally
  mur workflows export <id>                   Export as skill/yaml/md
  mur workflows delete <id>                   Delete a workflow
  mur workflows publish <id>                  Bump published version`,
}

var workflowsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List local workflows",
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := workflow.List()
		if err != nil {
			return fmt.Errorf("list workflows: %w", err)
		}

		if len(entries) == 0 {
			fmt.Println("No workflows found.")
			fmt.Println("\nCreate one with: mur workflows create --from-session <session-id>")
			return nil
		}

		fmt.Println("Workflows")
		fmt.Println("=========")
		fmt.Println()

		for _, e := range entries {
			version := "draft"
			if e.PublishedVersion > 0 {
				version = fmt.Sprintf("v%d", e.PublishedVersion)
			}
			shortID := e.ID
			if len(shortID) > 8 {
				shortID = shortID[:8]
			}
			tags := ""
			if len(e.Tags) > 0 {
				tags = fmt.Sprintf("  [%s]", strings.Join(e.Tags, ", "))
			}
			fmt.Printf("  %s  %-6s  %s%s\n", shortID, version, e.Name, tags)
		}

		fmt.Printf("\nTotal: %d workflows\n", len(entries))
		return nil
	},
}

var workflowsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show workflow details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wf, meta, err := workflow.Get(args[0])
		if err != nil {
			return err
		}

		version := "draft"
		if meta.PublishedVersion > 0 {
			version = fmt.Sprintf("v%d", meta.PublishedVersion)
		}

		fmt.Printf("Workflow: %s\n", wf.Name)
		fmt.Printf("ID:       %s\n", wf.ID)
		fmt.Printf("Version:  %s (revision %d)\n", version, meta.RevisionCount)
		fmt.Printf("Created:  %s\n", meta.CreatedAt.Format("2006-01-02 15:04"))
		fmt.Printf("Updated:  %s\n", meta.UpdatedAt.Format("2006-01-02 15:04"))

		if wf.Trigger != "" {
			fmt.Printf("\nTrigger: %s\n", wf.Trigger)
		}
		if wf.Description != "" {
			fmt.Printf("\n%s\n", wf.Description)
		}

		if len(wf.Variables) > 0 {
			fmt.Printf("\nVariables:\n")
			for _, v := range wf.Variables {
				req := "optional"
				if v.Required {
					req = "required"
				}
				fmt.Printf("  $%s (%s, %s): %s\n", v.Name, v.Type, req, v.Description)
			}
		}

		if len(wf.Steps) > 0 {
			fmt.Printf("\nSteps:\n")
			for _, s := range wf.Steps {
				approval := ""
				if s.NeedsApproval {
					approval = " [approval required]"
				}
				fmt.Printf("  %d. %s%s\n", s.Order, s.Description, approval)
				if s.Command != "" {
					fmt.Printf("     $ %s\n", s.Command)
				}
			}
		}

		if len(wf.Tools) > 0 {
			fmt.Printf("\nTools: %s\n", strings.Join(wf.Tools, ", "))
		}
		if len(wf.Tags) > 0 {
			fmt.Printf("Tags:  %s\n", strings.Join(wf.Tags, ", "))
		}

		if len(wf.SourceSessions) > 0 {
			fmt.Printf("\nSource sessions:\n")
			for _, s := range wf.SourceSessions {
				shortID := s.SessionID
				if len(shortID) > 8 {
					shortID = shortID[:8]
				}
				if s.StartEvent > 0 || s.EndEvent > 0 {
					fmt.Printf("  %s (events %d-%d)\n", shortID, s.StartEvent, s.EndEvent)
				} else {
					fmt.Printf("  %s\n", shortID)
				}
			}
		}

		return nil
	},
}

var workflowsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a workflow from a session",
	Long: `Create a new workflow from an analyzed session recording.

The session must have been analyzed first (mur session analyze <id>).
Use --start and --end to extract only a subset of steps.

Examples:
  mur workflows create --from-session abc123
  mur workflows create --from-session abc123 --start 2 --end 5`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID, _ := cmd.Flags().GetString("from-session")
		if sessionID == "" {
			return fmt.Errorf("--from-session is required")
		}

		// Resolve short session ID prefix
		fullID, err := session.ResolveSessionID(sessionID)
		if err != nil {
			return err
		}

		start, _ := cmd.Flags().GetInt("start")
		end, _ := cmd.Flags().GetInt("end")

		wf, err := workflow.CreateFromSession(fullID, start, end)
		if err != nil {
			return err
		}

		shortID := wf.ID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		fmt.Fprintf(os.Stderr, "Workflow created: %s (%s)\n", wf.Name, shortID)
		fmt.Fprintf(os.Stderr, "  Steps:     %d\n", len(wf.Steps))
		fmt.Fprintf(os.Stderr, "  Variables: %d\n", len(wf.Variables))
		fmt.Println(wf.ID)

		return nil
	},
}

var workflowsRunCmd = &cobra.Command{
	Use:   "run <id>",
	Short: "Execute a workflow locally",
	Long: `Run a workflow by executing its steps sequentially.

Steps with commands are executed in a shell. Steps requiring approval
will prompt before proceeding. Steps without commands print the
description for manual execution.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		wf, _, err := workflow.Get(args[0])
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Running workflow: %s\n\n", wf.Name)

		for _, step := range wf.Steps {
			fmt.Fprintf(os.Stderr, "Step %d: %s\n", step.Order, step.Description)

			if step.NeedsApproval && !dryRun {
				fmt.Fprintf(os.Stderr, "  Requires approval. Proceed? [y/N] ")
				var answer string
				fmt.Scanln(&answer)
				if answer != "y" && answer != "Y" {
					fmt.Fprintf(os.Stderr, "  Skipped.\n\n")
					continue
				}
			}

			if step.Command != "" {
				if dryRun {
					fmt.Fprintf(os.Stderr, "  [dry-run] $ %s\n\n", step.Command)
					continue
				}

				fmt.Fprintf(os.Stderr, "  $ %s\n", step.Command)
				c := exec.Command("sh", "-c", step.Command)
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				c.Stdin = os.Stdin

				if err := c.Run(); err != nil {
					switch step.OnFailure {
					case "skip":
						fmt.Fprintf(os.Stderr, "  Failed (skipping): %v\n\n", err)
						continue
					case "retry":
						fmt.Fprintf(os.Stderr, "  Failed: %v\n", err)
						fmt.Fprintf(os.Stderr, "  Retry? [y/N] ")
						var answer string
						fmt.Scanln(&answer)
						if answer == "y" || answer == "Y" {
							c2 := exec.Command("sh", "-c", step.Command)
							c2.Stdout = os.Stdout
							c2.Stderr = os.Stderr
							c2.Stdin = os.Stdin
							if err := c2.Run(); err != nil {
								return fmt.Errorf("step %d failed on retry: %w", step.Order, err)
							}
						} else {
							return fmt.Errorf("step %d failed: %w", step.Order, err)
						}
					default: // abort
						return fmt.Errorf("step %d failed: %w", step.Order, err)
					}
				}
			} else {
				if step.Tool != "" {
					fmt.Fprintf(os.Stderr, "  (manual step, tool: %s)\n", step.Tool)
				} else {
					fmt.Fprintf(os.Stderr, "  (manual step)\n")
				}
			}
			fmt.Fprintln(os.Stderr)
		}

		fmt.Fprintf(os.Stderr, "Workflow complete.\n")
		return nil
	},
}

var workflowsExportCmd = &cobra.Command{
	Use:   "export <id>",
	Short: "Export a workflow as skill, YAML, or markdown",
	Long: `Export a workflow to a reusable format.

Formats:
  skill     Full skill directory: SKILL.md + workflow.yaml + run.sh + steps/
  yaml      Standalone workflow YAML file
  md        Human-readable markdown documentation

Examples:
  mur workflows export abc123 --format skill
  mur workflows export abc123 --format yaml --output my-workflow.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wf, _, err := workflow.Get(args[0])
		if err != nil {
			return err
		}

		// Convert Workflow back to AnalysisResult for reusing session export functions
		result := &session.AnalysisResult{
			Name:        wf.Name,
			Trigger:     wf.Trigger,
			Description: wf.Description,
			Variables:   wf.Variables,
			Steps:       wf.Steps,
			Tools:       wf.Tools,
			Tags:        wf.Tags,
		}

		sessionID := ""
		if len(wf.SourceSessions) > 0 {
			sessionID = wf.SourceSessions[0].SessionID
		}

		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")

		switch format {
		case "skill":
			outputDir := output
			if outputDir == "" {
				outputDir, err = session.DefaultSkillsOutputDir()
				if err != nil {
					return err
				}
			}
			skillPath, err := session.ExportAsSkill(result, sessionID, outputDir)
			if err != nil {
				return fmt.Errorf("export skill: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Skill exported to %s\n", skillPath)
			fmt.Println(skillPath)

		case "yaml":
			if output == "" {
				output = wf.Name + ".yaml"
			}
			if err := session.ExportAsYAML(result, sessionID, output); err != nil {
				return fmt.Errorf("export YAML: %w", err)
			}
			fmt.Fprintf(os.Stderr, "YAML exported to %s\n", output)

		case "markdown", "md":
			if output == "" {
				output = wf.Name + ".md"
			}
			if err := session.ExportAsMarkdown(result, output); err != nil {
				return fmt.Errorf("export markdown: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Markdown exported to %s\n", output)

		default:
			return fmt.Errorf("unknown format %q (use: skill, yaml, md)", format)
		}

		return nil
	},
}

var workflowsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a workflow",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")

		// Show what we're about to delete
		wf, _, err := workflow.Get(args[0])
		if err != nil {
			return err
		}

		if !force {
			fmt.Fprintf(os.Stderr, "Delete workflow %q (%s)? [y/N] ", wf.Name, args[0][:8])
			var answer string
			fmt.Scanln(&answer)
			if answer != "y" && answer != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := workflow.Delete(args[0]); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Deleted workflow: %s\n", wf.Name)
		return nil
	},
}

var workflowsPublishCmd = &cobra.Command{
	Use:   "publish <id>",
	Short: "Bump the published version of a workflow",
	Long: `Publish creates a new version milestone for a workflow.

Every edit auto-saves a revision, but only explicit publish creates
a user-visible version number (v1, v2, v3...).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version, err := workflow.Publish(args[0])
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Published v%d\n", version)
		return nil
	},
}

// Phase 2: Sync command
var workflowsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync workflows to/from the cloud",
	Long: `Synchronize local workflows with the mur cloud server.
Requires Pro plan or higher. Uses the same auth as 'mur cloud login'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		client, err := cloud.NewClient(cfg.Server.URL)
		if err != nil {
			return fmt.Errorf("create cloud client: %w", err)
		}

		teamID := cfg.Server.Team
		if teamID == "" {
			return fmt.Errorf("no team configured. Run 'mur cloud login' first")
		}

		// Pull first
		fmt.Fprintf(os.Stderr, "Checking for server updates...\n")
		pullResp, err := client.WorkflowPull(teamID, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Pull skipped: %v\n", err)
		} else if len(pullResp.Workflows) > 0 {
			fmt.Fprintf(os.Stderr, "  Pulled %d workflow(s) from server\n", len(pullResp.Workflows))
			// TODO: apply pulled workflows to local store
		}

		// Build known IDs from pull response
		knownIDs := map[string]bool{}
		if pullResp != nil {
			for _, w := range pullResp.Workflows {
				knownIDs[w.ID] = true
			}
		}

		// Push local changes
		changes, err := workflow.BuildChangesFromLocal(knownIDs)
		if err != nil {
			return fmt.Errorf("build sync changes: %w", err)
		}

		if len(changes) == 0 {
			fmt.Println("No workflows to sync.")
			return nil
		}

		fmt.Fprintf(os.Stderr, "Pushing %d workflow(s) to cloud...\n", len(changes))

		resp, err := client.WorkflowPush(teamID, workflow.WorkflowPushRequest{
			BaseVersion: 0,
			Changes:     changes,
		})
		if err != nil {
			return fmt.Errorf("push workflows: %w", err)
		}

		if resp.OK {
			fmt.Fprintf(os.Stderr, "Synced successfully (server version: %d)\n", resp.Version)
		} else if len(resp.Conflicts) > 0 {
			fmt.Fprintf(os.Stderr, "Sync completed with %d conflict(s)\n", len(resp.Conflicts))
			for _, c := range resp.Conflicts {
				fmt.Fprintf(os.Stderr, "  ⚠ %s (%s)\n", c.WorkflowName, c.WorkflowID)
			}
		}

		return nil
	},
}

// Phase 2: Share command
var workflowsShareCmd = &cobra.Command{
	Use:   "share <workflow-id>",
	Short: "Share a workflow with a user",
	Long: `Grant a user permission to access a workflow.

Permission levels:
  read          Can view and use the workflow
  write         Can edit the workflow
  execute-only  Can run via Commander but can't see implementation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		userEmail, _ := cmd.Flags().GetString("user")
		perm, _ := cmd.Flags().GetString("permission")

		if userEmail == "" {
			return fmt.Errorf("--user is required")
		}
		if !workflow.ValidPermission(perm) {
			return fmt.Errorf("invalid permission %q (use: read, write, execute-only)", perm)
		}

		// TODO: get current user email from auth
		grantedBy := "owner"

		if err := workflow.SetPermission(args[0], userEmail, workflow.Permission(perm), grantedBy); err != nil {
			return fmt.Errorf("set permission: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Granted %s permission to %s on workflow %s\n", perm, userEmail, args[0])
		return nil
	},
}

// Phase 2: Merge command
var workflowsMergeCmd = &cobra.Command{
	Use:   "merge <id1> <id2> [<id3>...]",
	Short: "Merge multiple workflows into one",
	Long: `Combine multiple workflows into a single new workflow.
Steps are concatenated in order, variables deduplicated, tools and tags unioned.`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		merged, err := workflow.MergeWorkflows(args, name)
		if err != nil {
			return fmt.Errorf("merge workflows: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Created merged workflow: %s\n", merged.ID[:8])
		fmt.Fprintf(os.Stderr, "  Name:  %s\n", merged.Name)
		fmt.Fprintf(os.Stderr, "  Steps: %d\n", len(merged.Steps))
		fmt.Fprintf(os.Stderr, "  Tools: %v\n", merged.Tools)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(workflowsCmd)
	workflowsCmd.AddCommand(workflowsListCmd)
	workflowsCmd.AddCommand(workflowsShowCmd)
	workflowsCmd.AddCommand(workflowsCreateCmd)
	workflowsCmd.AddCommand(workflowsRunCmd)
	workflowsCmd.AddCommand(workflowsExportCmd)
	workflowsCmd.AddCommand(workflowsDeleteCmd)
	workflowsCmd.AddCommand(workflowsPublishCmd)

	// Phase 2 commands
	workflowsCmd.AddCommand(workflowsSyncCmd)
	workflowsCmd.AddCommand(workflowsShareCmd)
	workflowsCmd.AddCommand(workflowsMergeCmd)

	workflowsCreateCmd.Flags().String("from-session", "", "Session ID to create workflow from (required)")
	workflowsCreateCmd.Flags().Int("start", 0, "Start step index for partial extraction")
	workflowsCreateCmd.Flags().Int("end", 0, "End step index for partial extraction")

	workflowsRunCmd.Flags().Bool("dry-run", false, "Print commands without executing")

	workflowsExportCmd.Flags().StringP("format", "f", "skill", "Export format: skill, yaml, md")
	workflowsExportCmd.Flags().StringP("output", "o", "", "Output path")

	workflowsDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	// Phase 2 flags
	workflowsShareCmd.Flags().String("user", "", "User email to share with (required)")
	workflowsShareCmd.Flags().String("permission", "read", "Permission level: read, write, execute-only")

	workflowsMergeCmd.Flags().String("name", "", "Name for the merged workflow")
}
