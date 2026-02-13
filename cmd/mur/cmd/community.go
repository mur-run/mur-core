package cmd

import (
	"fmt"
	"strings"

	"github.com/mur-run/mur-core/internal/cloud"
	"github.com/mur-run/mur-core/internal/config"
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

var communityFeaturedCmd = &cobra.Command{
	Use:   "featured",
	Short: "Show featured community patterns",
	RunE:  runCommunityFeatured,
}

var communityUserCmd = &cobra.Command{
	Use:   "user <login>",
	Short: "View a user's profile and patterns",
	Args:  cobra.ExactArgs(1),
	RunE:  runCommunityUser,
}

var communityShareCmd = &cobra.Command{
	Use:   "share [pattern-name]",
	Short: "Share a pattern to the community",
	Long: `Share one of your patterns with the community.
	
The pattern will be submitted for review before being published.
You can optionally specify a category and tags.

Examples:
  mur community share "API retry with backoff"
  mur community share my-pattern --category "Error Handling" --tags "api,retry,resilience"`,
	Args: cobra.ExactArgs(1),
	RunE: runCommunityShare,
}

var (
	communityLimit    int
	communityTeamID   string
	shareCategory     string
	shareTags         string
	shareDescription  string
)

func init() {
	rootCmd.AddCommand(communityCmd)
	communityCmd.AddCommand(communitySearchCmd)
	communityCmd.AddCommand(communityCopyCmd)
	communityCmd.AddCommand(communityRecentCmd)
	communityCmd.AddCommand(communityShareCmd)
	communityCmd.AddCommand(communityFeaturedCmd)
	communityCmd.AddCommand(communityUserCmd)

	communityCmd.PersistentFlags().IntVarP(&communityLimit, "limit", "n", 10, "Number of results")
	communityCopyCmd.Flags().StringVarP(&communityTeamID, "team", "t", "", "Target team ID")

	// Share command flags
	communityShareCmd.Flags().StringVarP(&shareCategory, "category", "c", "", "Pattern category (e.g., 'Error Handling', 'Testing')")
	communityShareCmd.Flags().StringVarP(&shareTags, "tags", "t", "", "Comma-separated tags")
	communityShareCmd.Flags().StringVarP(&shareDescription, "description", "d", "", "Override pattern description")
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

func runCommunityFeatured(cmd *cobra.Command, args []string) error {
	client, err := cloud.NewClient("")
	if err != nil {
		return err
	}

	resp, err := client.GetCommunityFeatured(communityLimit)
	if err != nil {
		return fmt.Errorf("failed to get featured patterns: %w", err)
	}

	fmt.Println("⭐ Featured Community Patterns")
	fmt.Println(strings.Repeat("━", 50))
	fmt.Println()

	if len(resp.Patterns) == 0 {
		fmt.Println("  No featured patterns yet.")
		return nil
	}

	for _, p := range resp.Patterns {
		author := p.AuthorName
		if p.AuthorLogin != "" {
			author = "@" + p.AuthorLogin
		}
		fmt.Printf("  ⭐ %s (⬇️ %d) by %s\n", p.Name, p.CopyCount, author)
		if p.Description != "" {
			desc := p.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Printf("     %s\n", desc)
		}
	}

	fmt.Println()
	fmt.Println("Use 'mur community copy <name>' to copy a pattern")

	return nil
}

func runCommunityUser(cmd *cobra.Command, args []string) error {
	login := args[0]

	client, err := cloud.NewClient("")
	if err != nil {
		return err
	}

	profile, err := client.GetUserProfile(login)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	fmt.Printf("👤 %s", profile.Name)
	if profile.Login != "" {
		fmt.Printf(" (@%s)", profile.Login)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("━", 50))

	if profile.Bio != "" {
		fmt.Printf("   %s\n", profile.Bio)
	}

	fmt.Println()
	fmt.Printf("   📊 %d patterns | ⬇️ %d copies | ⭐ %d stars\n",
		profile.PatternCount, profile.TotalCopies, profile.TotalStars)

	if profile.Website != "" || profile.GitHub != "" || profile.Twitter != "" {
		fmt.Println()
		if profile.Website != "" {
			fmt.Printf("   🌐 %s\n", profile.Website)
		}
		if profile.GitHub != "" {
			fmt.Printf("   🐙 github.com/%s\n", profile.GitHub)
		}
		if profile.Twitter != "" {
			fmt.Printf("   🐦 @%s\n", profile.Twitter)
		}
	}

	if len(profile.Patterns) > 0 {
		fmt.Println()
		fmt.Println("Patterns:")
		for _, p := range profile.Patterns {
			fmt.Printf("   • %s (⬇️ %d)\n", p.Name, p.CopyCount)
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

func runCommunityShare(cmd *cobra.Command, args []string) error {
	patternName := args[0]

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client, err := cloud.NewClient(cfg.Server.URL)
	if err != nil {
		return err
	}

	if !client.AuthStore().IsLoggedIn() {
		return fmt.Errorf("not logged in. Run 'mur login' first")
	}

	// Get team from config
	teamSlug := cfg.Server.Team
	if teamSlug == "" {
		return fmt.Errorf("no team configured. Run 'mur cloud select <team>' first")
	}

	// Pull patterns to find the one to share
	pullResp, err := client.Pull(teamSlug, 0)
	if err != nil {
		return fmt.Errorf("failed to get patterns: %w", err)
	}

	var targetPattern *cloud.Pattern
	for i, p := range pullResp.Patterns {
		if p.Name == patternName && !p.Deleted {
			targetPattern = &pullResp.Patterns[i]
			break
		}
	}

	if targetPattern == nil {
		fmt.Printf("Pattern \"%s\" not found in your team. Available patterns:\n\n", patternName)
		for _, p := range pullResp.Patterns {
			if !p.Deleted {
				fmt.Printf("  • %s\n", p.Name)
			}
		}
		return nil
	}

	// Parse tags
	var tags []string
	if shareTags != "" {
		for _, t := range strings.Split(shareTags, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}

	// Submit to community
	req := &cloud.SharePatternRequest{
		PatternID:   targetPattern.ID,
		Category:    shareCategory,
		Tags:        tags,
		Description: shareDescription,
	}

	err = client.SharePattern(req)
	if err != nil {
		return fmt.Errorf("failed to share pattern: %w", err)
	}

	fmt.Printf("✓ Submitted \"%s\" for community review\n", targetPattern.Name)
	fmt.Println()
	fmt.Println("  Your pattern will be visible to everyone once approved.")
	fmt.Println("  You'll be notified when it's published.")

	return nil
}
