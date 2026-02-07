package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mur-run/mur-core/internal/sync"
	"github.com/spf13/cobra"
)

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Manage skills and methodologies",
	Long: `Manage skills and methodologies that are synced to AI CLI tools.

Skills are stored in ~/.murmur/skills/ and synced to:
  • Claude Code: ~/.claude/skills/
  • Gemini CLI: ~/.gemini/skills/
  • Auggie: ~/.augment/skills/`,
}

var skillsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		skills, err := sync.ListSkills()
		if err != nil {
			return fmt.Errorf("failed to list skills: %w", err)
		}

		if len(skills) == 0 {
			fmt.Println("No skills found.")
			fmt.Println()
			fmt.Println("Add skills to ~/.murmur/skills/ or import with:")
			fmt.Println("  mur skills import <path>")
			fmt.Println("  mur skills import --superpowers")
			return nil
		}

		fmt.Println("Available skills:")
		fmt.Println()
		for _, s := range skills {
			desc := s.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			if desc == "" {
				desc = "(no description)"
			}
			fmt.Printf("  • %s — %s\n", s.Name, desc)
		}

		fmt.Println()
		fmt.Printf("Total: %d skills\n", len(skills))
		fmt.Println()
		fmt.Println("Run 'mur sync skills' to sync to all CLI tools.")
		return nil
	},
}

var superpowersFlag bool

var skillsImportCmd = &cobra.Command{
	Use:   "import [path]",
	Short: "Import a skill from a file or Superpowers plugin",
	Long: `Import a skill from a file path or from the Superpowers plugin.

Examples:
  mur skills import ~/my-skill.md      Import a single skill file
  mur skills import --superpowers      Import all skills from Superpowers plugin`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Import from Superpowers
		if superpowersFlag {
			fmt.Println("Importing from Superpowers plugin...")
			count, err := sync.ImportFromSuperpowers()
			if err != nil {
				return fmt.Errorf("import failed: %w", err)
			}
			fmt.Printf("Imported %d skills from Superpowers.\n", count)
			fmt.Println()
			fmt.Println("Run 'mur sync skills' to sync to all CLI tools.")
			return nil
		}

		// Import from path
		if len(args) == 0 {
			return fmt.Errorf("please provide a file path or use --superpowers")
		}

		path := args[0]
		// Expand ~ if present
		if len(path) > 0 && path[0] == '~' {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("cannot determine home directory: %w", err)
			}
			path = filepath.Join(home, path[1:])
		}

		fmt.Printf("Importing skill from %s...\n", path)
		if err := sync.ImportSkill(path); err != nil {
			return fmt.Errorf("import failed: %w", err)
		}

		skillName := filepath.Base(path)
		if ext := filepath.Ext(skillName); ext != "" {
			skillName = skillName[:len(skillName)-len(ext)]
		}
		fmt.Printf("Imported skill: %s\n", skillName)
		fmt.Println()
		fmt.Println("Run 'mur sync skills' to sync to all CLI tools.")
		return nil
	},
}

var skillsShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show details of a skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		skillName := args[0]

		skills, err := sync.ListSkills()
		if err != nil {
			return fmt.Errorf("failed to list skills: %w", err)
		}

		for _, s := range skills {
			if s.Name == skillName || filepath.Base(s.SourcePath) == skillName+".md" {
				fmt.Printf("# %s\n\n", s.Name)
				if s.Description != "" {
					fmt.Printf("## Description\n\n%s\n\n", s.Description)
				}
				if s.Instructions != "" {
					fmt.Printf("## Instructions\n\n%s\n", s.Instructions)
				}
				fmt.Printf("\nSource: %s\n", s.SourcePath)
				return nil
			}
		}

		return fmt.Errorf("skill not found: %s", skillName)
	},
}

func init() {
	rootCmd.AddCommand(skillsCmd)
	skillsCmd.AddCommand(skillsListCmd)
	skillsCmd.AddCommand(skillsImportCmd)
	skillsCmd.AddCommand(skillsShowCmd)

	skillsImportCmd.Flags().BoolVar(&superpowersFlag, "superpowers", false, "Import from Superpowers plugin")
}
