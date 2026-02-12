package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mur-run/mur-core/internal/cloud"
)

// ConflictResolution represents how to resolve a conflict
type ConflictResolution int

const (
	ResolutionSkip ConflictResolution = iota
	ResolutionKeepServer
	ResolutionKeepLocal
)

// ResolveConflictsInteractive presents an interactive UI to resolve conflicts
func ResolveConflictsInteractive(conflicts []cloud.Conflict) (map[string]ConflictResolution, error) {
	if len(conflicts) == 0 {
		return nil, nil
	}

	resolutions := make(map[string]ConflictResolution)

	fmt.Println()
	fmt.Printf("⚠️  %d conflict(s) detected\n", len(conflicts))
	fmt.Println()

	// First, ask for global resolution strategy
	globalOptions := []string{
		"[i] Interactive - choose for each pattern",
		"[s] Accept all from server",
		"[l] Accept all from local",
		"[x] Skip all (no changes)",
	}

	var globalChoice string
	prompt := &survey.Select{
		Message: "How do you want to resolve conflicts?",
		Options: globalOptions,
	}
	if err := survey.AskOne(prompt, &globalChoice); err != nil {
		return nil, err
	}

	switch {
	case strings.HasPrefix(globalChoice, "[s]"):
		// Accept all from server
		for _, c := range conflicts {
			resolutions[c.PatternName] = ResolutionKeepServer
		}
		fmt.Println("✓ Will accept all server versions")
		return resolutions, nil

	case strings.HasPrefix(globalChoice, "[l]"):
		// Accept all from local
		for _, c := range conflicts {
			resolutions[c.PatternName] = ResolutionKeepLocal
		}
		fmt.Println("✓ Will keep all local versions")
		return resolutions, nil

	case strings.HasPrefix(globalChoice, "[x]"):
		// Skip all
		for _, c := range conflicts {
			resolutions[c.PatternName] = ResolutionSkip
		}
		fmt.Println("✓ Skipping all conflicts")
		return resolutions, nil
	}

	// Interactive mode - resolve each conflict
	fmt.Println()
	fmt.Println("Resolving conflicts interactively...")
	fmt.Println()

	for i, c := range conflicts {
		resolution, err := resolveConflictInteractive(c, i+1, len(conflicts))
		if err != nil {
			return nil, err
		}
		resolutions[c.PatternName] = resolution
	}

	return resolutions, nil
}

func resolveConflictInteractive(c cloud.Conflict, index, total int) (ConflictResolution, error) {
	fmt.Printf("┌─────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ [%d/%d] %s\n", index, total, c.PatternName)
	fmt.Printf("├─────────────────────────────────────────────────────────────┤\n")

	// Show info about both versions
	if c.ServerVersion != nil {
		serverInfo := formatPatternInfo(c.ServerVersion, "Server")
		fmt.Print(serverInfo)
	}
	if c.ClientVersion != nil {
		localInfo := formatPatternInfo(c.ClientVersion, "Local")
		fmt.Print(localInfo)
	}

	fmt.Printf("└─────────────────────────────────────────────────────────────┘\n")

	options := []string{
		"[s] Keep server version",
		"[l] Keep local version",
		"[d] View diff",
		"[x] Skip (no change)",
	}

	for {
		var choice string
		prompt := &survey.Select{
			Message: "Your choice:",
			Options: options,
		}
		if err := survey.AskOne(prompt, &choice); err != nil {
			return ResolutionSkip, err
		}

		switch {
		case strings.HasPrefix(choice, "[s]"):
			fmt.Printf("  ✓ Keeping server version\n\n")
			return ResolutionKeepServer, nil

		case strings.HasPrefix(choice, "[l]"):
			fmt.Printf("  ✓ Keeping local version\n\n")
			return ResolutionKeepLocal, nil

		case strings.HasPrefix(choice, "[d]"):
			showDiff(c)
			// Continue loop to ask again

		case strings.HasPrefix(choice, "[x]"):
			fmt.Printf("  → Skipped\n\n")
			return ResolutionSkip, nil
		}
	}
}

func formatPatternInfo(p *cloud.Pattern, label string) string {
	var sb strings.Builder

	// Get content preview (first 100 chars)
	preview := p.Content
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}
	preview = strings.ReplaceAll(preview, "\n", " ")

	lines := strings.Count(p.Content, "\n") + 1

	sb.WriteString(fmt.Sprintf("│ %s: %d lines\n", label, lines))
	if p.Description != "" {
		desc := p.Description
		if len(desc) > 50 {
			desc = desc[:50] + "..."
		}
		sb.WriteString(fmt.Sprintf("│   %s\n", desc))
	}

	return sb.String()
}

func showDiff(c cloud.Conflict) {
	fmt.Println()
	fmt.Println("─── Diff ───────────────────────────────────────────────────")

	serverContent := ""
	localContent := ""

	if c.ServerVersion != nil {
		serverContent = c.ServerVersion.Content
	}
	if c.ClientVersion != nil {
		localContent = c.ClientVersion.Content
	}

	// Simple line-by-line comparison
	serverLines := strings.Split(serverContent, "\n")
	localLines := strings.Split(localContent, "\n")

	fmt.Printf("Server: %d lines | Local: %d lines\n", len(serverLines), len(localLines))
	fmt.Println()

	// Show first differences (up to 20 lines)
	maxLines := 20
	shown := 0

	// Find different lines
	maxLen := len(serverLines)
	if len(localLines) > maxLen {
		maxLen = len(localLines)
	}

	for i := 0; i < maxLen && shown < maxLines; i++ {
		serverLine := ""
		localLine := ""
		if i < len(serverLines) {
			serverLine = serverLines[i]
		}
		if i < len(localLines) {
			localLine = localLines[i]
		}

		if serverLine != localLine {
			if serverLine != "" {
				fmt.Printf("  - %s\n", conflictTruncate(serverLine, 60))
			}
			if localLine != "" {
				fmt.Printf("  + %s\n", conflictTruncate(localLine, 60))
			}
			shown++
		}
	}

	if shown == 0 {
		fmt.Println("  (no visible differences in first 20 lines)")
	} else if maxLen > maxLines {
		fmt.Printf("  ... and %d more lines\n", maxLen-maxLines)
	}

	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Println()
}

func conflictTruncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// ApplyResolutions applies the conflict resolutions
func ApplyResolutions(resolutions map[string]ConflictResolution) (keepServer, keepLocal, skipped int) {
	for _, r := range resolutions {
		switch r {
		case ResolutionKeepServer:
			keepServer++
		case ResolutionKeepLocal:
			keepLocal++
		case ResolutionSkip:
			skipped++
		}
	}
	return
}

// FormatTimeSince formats a time as "Xh ago" or "Xm ago"
func FormatTimeSince(t time.Time) string {
	d := time.Since(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}
