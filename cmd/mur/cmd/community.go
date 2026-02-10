package cmd

import (
	"fmt"
	"strings"

	"github.com/mur-run/mur-core/internal/cloud"
	"github.com/spf13/cobra"
)

var communityCmd = &cobra.Command{
	Use:   "community",
	Short: "Browse and copy community patterns",
	Long:  `Browse popular patterns from the community and copy them to your local collection.`,
	RunE:  runCommunity,
}

var communitySearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search community patterns",
	Args:  cobra.ExactArgs(1),
	RunE:  runCommunitySearch,
}

var communityCopyCmd = &cobra.Command{
	Use:   "copy [pattern-name]",
	Short: "Copy a community pattern to your team",
	Args:  cobra.ExactArgs(1),
	RunE:  runCommunityCopy,
}

var communityRecentCmd = &cobra.Command{
	Use:   "recent",
	Short: "Show recent community patterns",
	RunE:  runCommunityRecent,
}

var (
	communityLimit  int
	communityTeamID string
)

func init() {
	rootCmd.AddCommand(communityCmd)
	communityCmd.AddCommand(communitySearchCmd)
	communityCmd.AddCommand(communityCopyCmd)
	communityCmd.AddCommand(communityRecentCmd)

	communityCmd.PersistentFlags().IntVarP(&communityLimit, "limit", "n", 10, "Number of results")
	communityCopyCmd.Flags().StringVarP(&communityTeamID, "team", "t", "", "Target team ID")
}

func runCommunity(cmd *cobra.Command, args []string) error {
	client, err := cloud.NewClient("")
	if err != nil {
		return err
	}

	resp, err := client.GetCommunityPopular(communityLimit)
	if err != nil {
		return fmt.Errorf("failed to get community patterns: %w", err)
	}

	fmt.Println("🌍 Community Patterns")
	fmt.Println(strings.Repeat("━", 50))
	fmt.Println()

	if len(resp.Patterns) == 0 {
		fmt.Println("  No community patterns available yet.")
		return nil
	}

	fmt.Println("Popular:")
	for i, p := range resp.Patterns {
		author := p.AuthorName
		if p.AuthorLogin != "" {
			author = "@" + p.AuthorLogin
		}
		fmt.Printf("  %d. %s (⭐ %d) by %s\n", i+1, p.Name, p.CopyCount, author)
		if p.Description != "" {
			desc := p.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Printf("     %s\n", desc)
		}
	}

	fmt.Println()
	fmt.Println("Use 'mur community search <query>' to search")
	fmt.Println("Use 'mur community copy <name>' to copy a pattern")

	return nil
}

func runCommunitySearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	client, err := cloud.NewClient("")
	if err != nil {
		return err
	}

	resp, err := client.SearchCommunity(query, communityLimit)
	if err != nil {
		return fmt.Errorf("failed to search: %w", err)
	}

	fmt.Printf("🔍 Search results for \"%s\"\n", query)
	fmt.Println(strings.Repeat("━", 50))
	fmt.Println()

	if len(resp.Patterns) == 0 {
		fmt.Println("  No patterns found.")
		return nil
	}

	fmt.Printf("Found %d patterns:\n\n", resp.Count)

	for _, p := range resp.Patterns {
		author := p.AuthorName
		if p.AuthorLogin != "" {
			author = "@" + p.AuthorLogin
		}
		fmt.Printf("  • %s (⭐ %d) by %s\n", p.Name, p.CopyCount, author)
		if p.Description != "" {
			desc := p.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Printf("    %s\n", desc)
		}
	}

	return nil
}

func runCommunityRecent(cmd *cobra.Command, args []string) error {
	client, err := cloud.NewClient("")
	if err != nil {
		return err
	}

	resp, err := client.GetCommunityRecent(communityLimit)
	if err != nil {
		return fmt.Errorf("failed to get recent patterns: %w", err)
	}

	fmt.Println("🆕 Recent Community Patterns")
	fmt.Println(strings.Repeat("━", 50))
	fmt.Println()

	if len(resp.Patterns) == 0 {
		fmt.Println("  No recent patterns.")
		return nil
	}

	for _, p := range resp.Patterns {
		author := p.AuthorName
		if p.AuthorLogin != "" {
			author = "@" + p.AuthorLogin
		}
		fmt.Printf("  • %s by %s\n", p.Name, author)
		if p.Description != "" {
			desc := p.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Printf("    %s\n", desc)
		}
	}

	return nil
}

func runCommunityCopy(cmd *cobra.Command, args []string) error {
	patternName := args[0]

	client, err := cloud.NewClient("")
	if err != nil {
		return err
	}

	// Search for the pattern first
	resp, err := client.SearchCommunity(patternName, 10)
	if err != nil {
		return fmt.Errorf("failed to search: %w", err)
	}

	var targetPattern *cloud.CommunityPattern
	for _, p := range resp.Patterns {
		if p.Name == patternName {
			targetPattern = &p
			break
		}
	}

	if targetPattern == nil {
		if len(resp.Patterns) > 0 {
			fmt.Printf("Pattern \"%s\" not found. Did you mean:\n", patternName)
			for _, p := range resp.Patterns {
				fmt.Printf("  • %s\n", p.Name)
			}
			return nil
		}
		return fmt.Errorf("pattern not found: %s", patternName)
	}

	// Get team ID if not provided
	teamID := communityTeamID
	if teamID == "" {
		// Try to get default team
		teams, err := client.ListTeams()
		if err != nil {
			return fmt.Errorf("failed to list teams: %w", err)
		}
		if len(teams) == 0 {
			return fmt.Errorf("no teams found. Create a team first with 'mur team create'")
		}
		teamID = teams[0].ID
	}

	pattern, err := client.CopyPattern(targetPattern.ID, teamID)
	if err != nil {
		return fmt.Errorf("failed to copy pattern: %w", err)
	}

	fmt.Printf("✓ Copied \"%s\" to your patterns\n", pattern.Name)
	fmt.Println("  Run 'mur sync' to download it locally")

	return nil
}
