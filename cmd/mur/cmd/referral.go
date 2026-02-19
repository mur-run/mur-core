package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/cloud"
)

var referralCmd = &cobra.Command{
	Use:   "referral",
	Short: "View your referral status and share link",
	Long:  `View your referral statistics and get your share link to extend your trial.`,
	RunE:  runReferral,
}

func init() {
	rootCmd.AddCommand(referralCmd)
}

func runReferral(cmd *cobra.Command, args []string) error {
	client, err := cloud.NewClient("")
	if err != nil {
		return err
	}

	stats, err := client.GetReferralStats()
	if err != nil {
		return fmt.Errorf("failed to get referral stats: %w", err)
	}

	fmt.Println("📊 Referral Status")
	fmt.Println(strings.Repeat("━", 50))
	fmt.Println()

	fmt.Println("Your referral link:")
	fmt.Printf("  %s\n", stats.ReferralLink)
	fmt.Println()

	fmt.Println("Stats:")
	fmt.Printf("  Shared:    %d people\n", stats.TotalShared)
	fmt.Printf("  Qualified: %d people (7+ days active)\n", stats.TotalQualified)
	fmt.Printf("  Rewards:   %d/%d used (+%d days total)\n", stats.TotalRewarded, stats.TotalRewarded+stats.RewardsLeft, stats.DaysEarned)
	fmt.Println()

	if stats.RewardsLeft > 0 {
		fmt.Printf("🎁 You can earn %d more referral rewards (+%d days each)\n", stats.RewardsLeft, 30)
	} else {
		fmt.Println("✓ You've earned the maximum referral rewards!")
	}

	return nil
}
