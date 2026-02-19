package consolidate

import (
	"fmt"
	"strings"
)

// FormatReport renders a ConsolidationReport for CLI output.
func FormatReport(r *ConsolidationReport, patternNames map[string]string) string {
	var b strings.Builder

	b.WriteString("Pattern Consolidation Report\n")
	b.WriteString("============================\n\n")
	b.WriteString(fmt.Sprintf("Mode:           %s\n", r.Mode))
	b.WriteString(fmt.Sprintf("Total patterns: %d\n", r.TotalPatterns))
	b.WriteString(fmt.Sprintf("Duration:       %s\n\n", r.Duration.Round(1e6)))

	// Summary
	b.WriteString("Summary\n")
	b.WriteString("-------\n")
	b.WriteString(fmt.Sprintf("  Keep:    %d\n", r.PatternsKept))
	b.WriteString(fmt.Sprintf("  Archive: %d\n", r.PatternsArchived))
	b.WriteString(fmt.Sprintf("  Merge:   %d\n", r.PatternsMerged))
	b.WriteString(fmt.Sprintf("  Update:  %d\n", r.PatternsUpdated))
	if r.Mode == ModeAuto {
		b.WriteString(fmt.Sprintf("  Actions applied: %d\n", r.ActionsApplied))
	}
	b.WriteString("\n")

	// Health scores for non-keep actions
	actionPatterns := make([]HealthScore, 0)
	for _, hs := range r.HealthScores {
		if hs.Action != ActionKeep {
			actionPatterns = append(actionPatterns, hs)
		}
	}

	if len(actionPatterns) > 0 {
		b.WriteString("Action Items\n")
		b.WriteString("------------\n")
		for _, hs := range actionPatterns {
			name := patternNames[hs.PatternID]
			if name == "" {
				name = hs.PatternID
			}
			b.WriteString(fmt.Sprintf("  [%s] %s (score: %.2f) — %s\n",
				actionLabel(hs.Action), name, hs.Overall, hs.Reason))
		}
		b.WriteString("\n")
	}

	// Merge proposals
	if len(r.MergeProposals) > 0 {
		b.WriteString("Merge Proposals\n")
		b.WriteString("---------------\n")
		for i, mp := range r.MergeProposals {
			b.WriteString(fmt.Sprintf("  %d. Similarity: %.1f%% | Strategy: %s\n",
				i+1, mp.Similarity*100, mp.Strategy))
			for _, p := range mp.Patterns {
				marker := "  "
				if p.ID == mp.KeepID {
					marker = "* "
				}
				b.WriteString(fmt.Sprintf("     %s%s (%s)\n", marker, p.Name, p.ID))
			}
			if mp.KeepID != "" {
				keepName := patternNames[mp.KeepID]
				if keepName == "" {
					keepName = mp.KeepID
				}
				b.WriteString(fmt.Sprintf("     → Keep: %s\n", keepName))
			}
			b.WriteString("\n")
		}
	}

	// Conflicts
	if len(r.Conflicts) > 0 {
		b.WriteString("Conflicts\n")
		b.WriteString("---------\n")
		for _, c := range r.Conflicts {
			b.WriteString(fmt.Sprintf("  [%s] %s <-> %s\n",
				c.Type, c.PatternA.Name, c.PatternB.Name))
			b.WriteString(fmt.Sprintf("    %s\n", c.Reason))
		}
		b.WriteString("\n")
	}

	// Dry run notice
	if r.Mode == ModeDryRun {
		b.WriteString("(dry-run mode — no changes applied)\n")
		b.WriteString("Run with --auto to apply safe actions, or --interactive to review each.\n")
	}

	return b.String()
}

func actionLabel(a Action) string {
	switch a {
	case ActionArchive:
		return "ARCHIVE"
	case ActionMerge:
		return "MERGE  "
	case ActionUpdate:
		return "UPDATE "
	case ActionDelete:
		return "DELETE "
	default:
		return "KEEP   "
	}
}
