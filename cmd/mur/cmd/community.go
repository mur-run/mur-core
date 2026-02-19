package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mur-run/mur-core/internal/cloud"
	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/mur-run/mur-core/internal/security"
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
	communityLimit     int
	communityTeamID    string
	shareCategory      string
	shareTags          string
	shareDescription   string
	shareAutoTranslate bool
	shareDryRun        bool
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
	communityShareCmd.Flags().BoolVar(&shareAutoTranslate, "translate", true, "Auto-translate non-English patterns to English")
	communityShareCmd.Flags().BoolVar(&shareDryRun, "dry-run", false, "Preview PII redactions without sharing")
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

	// Resolve team slug to ID
	teamID, err := client.ResolveTeamID(teamSlug)
	if err != nil {
		return fmt.Errorf("failed to resolve team: %w", err)
	}

	// Pull patterns to find the one to share
	pullResp, err := client.Pull(teamID, 0)
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

	// PII scanning and redaction
	piiScanner := security.NewPIIScanner(cfg.Privacy)
	contentToScan := targetPattern.Name + "\n" + targetPattern.Description + "\n" + targetPattern.Content
	cleaned, findings := piiScanner.ScanAndRedact(contentToScan)

	if len(findings) > 0 {
		fmt.Println("🔒 PII detected and redacted:")
		fmt.Print(security.FormatFindings(findings))
		fmt.Println()

		// Reconstruct the cleaned parts
		parts := strings.SplitN(cleaned, "\n", 3)
		if len(parts) >= 1 {
			targetPattern.Name = parts[0]
		}
		if len(parts) >= 2 {
			targetPattern.Description = parts[1]
		}
		if len(parts) >= 3 {
			targetPattern.Content = parts[2]
		}
	}

	// LLM semantic anonymization (after regex PII, before secret scan)
	var anonChanges []security.AnonymizationChange
	if cfg.Privacy.SemanticAnonymization.Enabled {
		sa := cfg.Privacy.SemanticAnonymization
		llmClient, err := security.NewLLMClient(sa.Provider, sa.Model, sa.OllamaURL)
		if err != nil {
			fmt.Printf("⚠️  Semantic anonymization unavailable: %v\n", err)
		} else {
			cacheDir := ""
			if sa.CacheResults {
				home, _ := os.UserHomeDir()
				if home != "" {
					cacheDir = filepath.Join(home, ".mur", "cache", "anonymization")
				}
			}
			anonymizer := security.NewSemanticAnonymizer(llmClient, cacheDir)

			anonContent := targetPattern.Name + "\n" + targetPattern.Description + "\n" + targetPattern.Content
			anonCleaned, changes, anonErr := anonymizer.Anonymize(anonContent)
			if anonErr != nil {
				fmt.Printf("⚠️  Semantic anonymization failed: %v\n", anonErr)
			} else if len(changes) > 0 {
				anonChanges = changes
				fmt.Println("🧠 LLM semantic anonymization applied:")
				fmt.Print(security.FormatAnonymizationChanges(changes))
				fmt.Println()

				parts := strings.SplitN(anonCleaned, "\n", 3)
				if len(parts) >= 1 {
					targetPattern.Name = parts[0]
				}
				if len(parts) >= 2 {
					targetPattern.Description = parts[1]
				}
				if len(parts) >= 3 {
					targetPattern.Content = parts[2]
				}
			}
		}
	}

	if shareDryRun {
		fmt.Println("🔍 Dry run — content after redaction:")
		fmt.Println(strings.Repeat("─", 50))
		fmt.Printf("Name: %s\n", targetPattern.Name)
		fmt.Printf("Description: %s\n", targetPattern.Description)
		fmt.Println(strings.Repeat("─", 50))
		fmt.Println(targetPattern.Content)
		fmt.Println(strings.Repeat("─", 50))
		if len(anonChanges) > 0 {
			fmt.Printf("\n🧠 LLM detected %d semantic identifiers.\n", len(anonChanges))
		}
		fmt.Println("No changes were made. Remove --dry-run to share.")
		return nil
	}

	// Interactive preview (non-quiet mode)
	if !cmd.Flags().Changed("quiet") {
		if len(findings) > 0 {
			fmt.Println("Content will be shared with the above redactions applied.")
		}
		fmt.Print("Proceed with sharing? [Y/n] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Println("Share cancelled.")
			return nil
		}
	}

	// Check if translation is needed
	if shareAutoTranslate {
		localPattern := &pattern.Pattern{
			Name:        targetPattern.Name,
			Description: targetPattern.Description,
			Content:     targetPattern.Content,
		}
		if pattern.NeedsTranslation(localPattern) {
			fmt.Println("🌐 Detected non-English content, translating...")
			
			translateReq := &cloud.TranslatePatternRequest{
				Name:        targetPattern.Name,
				Description: targetPattern.Description,
				Content:     targetPattern.Content,
			}
			
			translated, err := client.TranslatePattern(translateReq)
			if err != nil {
				fmt.Printf("⚠️  Translation failed: %v\n", err)
				fmt.Println("   Sharing original content instead.")
			} else {
				// Update pattern with translated content
				targetPattern.Name = translated.Name
				targetPattern.Description = translated.Description
				targetPattern.Content = translated.Content
				fmt.Println("✓ Translated to English")
			}
		}
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
