package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/sync"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up old files and caches",
	Long: `Clean up old mur files and caches.

Removes:
  - Old stats data (>30 days)
  - Stale embedding caches
  - Orphaned sync files
  - Temporary files

Examples:
  mur clean           # Preview what would be cleaned
  mur clean --force   # Actually clean files
  mur clean --all     # Clean everything including stats`,
	RunE: runClean,
}

var (
	cleanForce bool
	cleanAll   bool
	cleanDays  int
)

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().BoolVarP(&cleanForce, "force", "f", false, "Actually delete files")
	cleanCmd.Flags().BoolVar(&cleanAll, "all", false, "Clean everything including stats")
	cleanCmd.Flags().IntVar(&cleanDays, "days", 30, "Delete files older than N days")
}

func runClean(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	murDir := filepath.Join(home, ".mur")
	if _, err := os.Stat(murDir); os.IsNotExist(err) {
		fmt.Println("Nothing to clean - ~/.mur doesn't exist")
		return nil
	}

	var totalSize int64
	var fileCount int

	fmt.Println()
	if cleanForce {
		fmt.Println("🧹 Cleaning mur files...")
	} else {
		fmt.Println("🔍 Files that would be cleaned (use --force to delete):")
	}
	fmt.Println()

	// Clean old embeddings cache
	embeddingsDir := filepath.Join(murDir, "embeddings")
	if info, err := os.Stat(embeddingsDir); err == nil && info.IsDir() {
		size, count := cleanDirectory(embeddingsDir, cleanDays, "embeddings cache", cleanForce)
		totalSize += size
		fileCount += count
	}

	// Clean old transcripts
	transcriptsDir := filepath.Join(murDir, "transcripts")
	if info, err := os.Stat(transcriptsDir); err == nil && info.IsDir() {
		size, count := cleanDirectory(transcriptsDir, cleanDays, "transcripts", cleanForce)
		totalSize += size
		fileCount += count
	}

	// Clean temp files
	tempPatterns := []string{
		filepath.Join(murDir, "*.tmp"),
		filepath.Join(murDir, "*.bak"),
		filepath.Join(murDir, ".*.swp"),
	}
	for _, pattern := range tempPatterns {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				continue
			}
			if cleanForce {
				if err := os.Remove(match); err == nil {
					fmt.Printf("   Deleted: %s\n", filepath.Base(match))
					totalSize += info.Size()
					fileCount++
				}
			} else {
				fmt.Printf("   %s (%s)\n", filepath.Base(match), formatSize(info.Size()))
				totalSize += info.Size()
				fileCount++
			}
		}
	}

	// Clean stats if --all
	if cleanAll {
		statsFile := filepath.Join(murDir, "stats.jsonl")
		if info, err := os.Stat(statsFile); err == nil {
			if cleanForce {
				if err := os.Remove(statsFile); err == nil {
					fmt.Printf("   Deleted: stats.jsonl\n")
					totalSize += info.Size()
					fileCount++
				}
			} else {
				fmt.Printf("   stats.jsonl (%s)\n", formatSize(info.Size()))
				totalSize += info.Size()
				fileCount++
			}
		}
	}

	// Clean old pattern directories from skills folders (legacy format)
	fmt.Println("📂 Old pattern directories:")
	if cleanForce {
		cleaned, err := sync.CleanOldPatternDirs()
		if err != nil {
			fmt.Printf("   Error cleaning: %v\n", err)
		} else if cleaned > 0 {
			fmt.Printf("   ✓ Removed %d old pattern directories\n", cleaned)
			fileCount += cleaned
		} else {
			fmt.Println("   (none found)")
		}
	} else {
		// Preview mode - count what would be cleaned
		previewCount := countOldPatternDirs(home)
		if previewCount > 0 {
			fmt.Printf("   Would remove %d old pattern directories from skills folders\n", previewCount)
			fileCount += previewCount
		} else {
			fmt.Println("   (none found)")
		}
	}
	fmt.Println()

	// Clean orphaned sync files
	syncTargets := []struct {
		name string
		path string
	}{
		{"Claude skills", filepath.Join(home, ".claude", "skills", "mur")},
		{"Gemini skills", filepath.Join(home, ".gemini", "skills", "mur")},
		{"Continue rules", filepath.Join(home, ".continue", "rules", "mur")},
		{"Cursor rules", filepath.Join(home, ".cursor", "rules", "mur")},
		{"Windsurf rules", filepath.Join(home, ".windsurf", "rules", "mur")},
	}

	// Check for orphaned patterns in sync targets
	patternsDir := filepath.Join(murDir, "patterns")
	patterns, _ := os.ReadDir(patternsDir)
	patternNames := make(map[string]bool)
	for _, p := range patterns {
		name := strings.TrimSuffix(p.Name(), ".yaml")
		patternNames[name] = true
	}

	for _, target := range syncTargets {
		if _, err := os.Stat(target.path); os.IsNotExist(err) {
			continue
		}

		files, _ := os.ReadDir(target.path)
		for _, f := range files {
			name := strings.TrimSuffix(f.Name(), ".md")
			if !patternNames[name] && name != "README" {
				fullPath := filepath.Join(target.path, f.Name())
				info, _ := os.Stat(fullPath)
				if info == nil {
					continue
				}

				if cleanForce {
					if err := os.Remove(fullPath); err == nil {
						fmt.Printf("   Deleted orphan: %s/%s\n", target.name, f.Name())
						totalSize += info.Size()
						fileCount++
					}
				} else {
					fmt.Printf("   Orphan: %s/%s (%s)\n", target.name, f.Name(), formatSize(info.Size()))
					totalSize += info.Size()
					fileCount++
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if fileCount == 0 {
		fmt.Println("✨ Nothing to clean!")
	} else if cleanForce {
		fmt.Printf("✅ Cleaned %d files (%s)\n", fileCount, formatSize(totalSize))
	} else {
		fmt.Printf("📊 Would clean %d files (%s)\n", fileCount, formatSize(totalSize))
		fmt.Println()
		fmt.Println("Run with --force to delete")
	}

	fmt.Println()

	return nil
}

func cleanDirectory(dir string, days int, label string, force bool) (int64, int) {
	var totalSize int64
	var count int

	cutoff := time.Now().AddDate(0, 0, -days)

	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if info.ModTime().Before(cutoff) {
			if force {
				if err := os.Remove(path); err == nil {
					rel, _ := filepath.Rel(dir, path)
					fmt.Printf("   Deleted: %s/%s\n", label, rel)
					totalSize += info.Size()
					count++
				}
			} else {
				rel, _ := filepath.Rel(dir, path)
				fmt.Printf("   %s/%s (%s, %d days old)\n", label, rel, formatSize(info.Size()), int(time.Since(info.ModTime()).Hours()/24))
				totalSize += info.Size()
				count++
			}
		}
		return nil
	})

	return totalSize, count
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// countOldPatternDirs counts old pattern directories for preview mode.
func countOldPatternDirs(home string) int {
	count := 0
	skillsDirs := []string{
		filepath.Join(home, ".claude", "skills"),
		filepath.Join(home, ".gemini", "skills"),
		filepath.Join(home, ".augment", "skills"),
		filepath.Join(home, ".opencode", "skills"),
		filepath.Join(home, ".continue", "rules"),
		filepath.Join(home, ".cursor", "rules"),
		filepath.Join(home, ".windsurf", "rules"),
	}

	for _, dir := range skillsDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() || entry.Name() == "mur-index" {
				continue
			}
			skillPath := filepath.Join(dir, entry.Name(), "SKILL.md")
			if _, err := os.Stat(skillPath); err == nil {
				content, _ := os.ReadFile(skillPath)
				if looksLikeMurPattern(string(content)) {
					count++
				}
			}
		}
	}
	return count
}

// looksLikeMurPattern checks if SKILL.md content looks like a mur-generated pattern.
func looksLikeMurPattern(content string) bool {
	markers := []string{
		"mur",
		"pattern",
		"Tags:",
		"## Instructions",
		"## Problem",
		"## Solution",
		"## Why This Is Non-Obvious",
		"## Verification",
	}
	for _, marker := range markers {
		if strings.Contains(content, marker) {
			return true
		}
	}
	return false
}
