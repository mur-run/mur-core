package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/cloud"
)

var collectionCmd = &cobra.Command{
	Use:     "collection",
	Aliases: []string{"collections"},
	Short:   "Manage pattern collections",
	Long:    `Create and manage curated collections of patterns.`,
	RunE:    runCollectionList,
}

var collectionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List public collections",
	RunE:  runCollectionList,
}

var collectionShowCmd = &cobra.Command{
	Use:   "show <collection-id>",
	Short: "Show a collection's contents",
	Args:  cobra.ExactArgs(1),
	RunE:  runCollectionShow,
}

var collectionCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new collection",
	Args:  cobra.ExactArgs(1),
	RunE:  runCollectionCreate,
}

var (
	collectionDescription string
	collectionVisibility  string
)

func init() {
	rootCmd.AddCommand(collectionCmd)
	collectionCmd.AddCommand(collectionListCmd)
	collectionCmd.AddCommand(collectionShowCmd)
	collectionCmd.AddCommand(collectionCreateCmd)

	collectionCreateCmd.Flags().StringVarP(&collectionDescription, "description", "d", "", "Collection description")
	collectionCreateCmd.Flags().StringVarP(&collectionVisibility, "visibility", "v", "private", "Visibility (private|public)")
}

func runCollectionList(cmd *cobra.Command, args []string) error {
	client, err := cloud.NewClient("")
	if err != nil {
		return err
	}

	collections, err := client.ListCollections(20)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	fmt.Println("📚 Public Collections")
	fmt.Println(strings.Repeat("━", 50))
	fmt.Println()

	if len(collections) == 0 {
		fmt.Println("  No public collections yet.")
		fmt.Println("  Create one with: mur collection create \"My Collection\"")
		return nil
	}

	for _, c := range collections {
		fmt.Printf("  📁 %s (⬇️ %d)\n", c.Name, c.CopyCount)
		if c.Description != "" {
			desc := c.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Printf("     %s\n", desc)
		}
		fmt.Printf("     ID: %s\n", c.ID)
	}

	return nil
}

func runCollectionShow(cmd *cobra.Command, args []string) error {
	collectionID := args[0]

	client, err := cloud.NewClient("")
	if err != nil {
		return err
	}

	collection, patterns, err := client.GetCollection(collectionID)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	fmt.Printf("📁 %s\n", collection.Name)
	fmt.Println(strings.Repeat("━", 50))

	if collection.Description != "" {
		fmt.Printf("   %s\n", collection.Description)
	}
	fmt.Println()

	if len(patterns) == 0 {
		fmt.Println("  (empty collection)")
		return nil
	}

	fmt.Printf("Patterns (%d):\n", len(patterns))
	for _, p := range patterns {
		fmt.Printf("  • %s (⬇️ %d)\n", p.Name, p.CopyCount)
	}

	return nil
}

func runCollectionCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := cloud.NewClient("")
	if err != nil {
		return err
	}

	if !client.AuthStore().IsLoggedIn() {
		return fmt.Errorf("not logged in. Run 'mur login' first")
	}

	collection, err := client.CreateCollection(name, collectionDescription, collectionVisibility)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	fmt.Printf("✓ Created collection \"%s\"\n", collection.Name)
	fmt.Printf("  ID: %s\n", collection.ID)
	fmt.Printf("  Visibility: %s\n", collection.Visibility)
	fmt.Println()
	fmt.Println("Add patterns with: mur collection add <collection-id> <pattern-id>")

	return nil
}
