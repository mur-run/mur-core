// Package stats provides usage tracking and statistics for murmur-ai.
package stats

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// UsageRecord represents a single tool usage event.
type UsageRecord struct {
	Tool         string    `json:"tool"`
	Timestamp    time.Time `json:"timestamp"`
	PromptLength int       `json:"prompt_length"`
	DurationMs   int64     `json:"duration_ms"`
	CostEstimate float64   `json:"cost_estimate"`
	Tier         string    `json:"tier"`
	RoutingMode  string    `json:"routing_mode"`
	AutoRouted   bool      `json:"auto_routed"`
	Complexity   float64   `json:"complexity"`
	Success      bool      `json:"success"`
}

// QueryFilter specifies criteria for filtering records.
type QueryFilter struct {
	Tool      string
	StartTime time.Time
	EndTime   time.Time
	Tier      string
}

// ToolStats aggregates statistics for a single tool.
type ToolStats struct {
	Count       int     `json:"count"`
	TotalTimeMs int64   `json:"total_time_ms"`
	AvgTimeMs   int64   `json:"avg_time_ms"`
	TotalCost   float64 `json:"total_cost"`
	SuccessRate float64 `json:"success_rate"`
}

// AutoRouteStats tracks automatic routing decisions.
type AutoRouteStats struct {
	Total     int     `json:"total"`
	ToFree    int     `json:"to_free"`
	ToPaid    int     `json:"to_paid"`
	FreeRatio float64 `json:"free_ratio"`
}

// DailyStats tracks usage per day.
type DailyStats struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// Summary aggregates overall usage data.
type Summary struct {
	TotalRuns      int                `json:"total_runs"`
	ByTool         map[string]ToolStats `json:"by_tool"`
	EstimatedCost  float64            `json:"estimated_cost"`
	EstimatedSaved float64            `json:"estimated_saved"`
	AutoRouteStats AutoRouteStats     `json:"auto_route_stats"`
	DailyTrend     []DailyStats       `json:"daily_trend"`
	Period         string             `json:"period"`
}

// Cost per 1K characters (rough estimates)
var costPerKChars = map[string]float64{
	"claude": 0.003, // ~$3/M input tokens, rough estimate
	"gemini": 0.0,   // free
	"auggie": 0.0,   // free
}

// EstimateCost calculates the cost estimate for a tool usage.
func EstimateCost(tool string, promptLength int) float64 {
	rate, ok := costPerKChars[tool]
	if !ok {
		return 0.0
	}
	return rate * float64(promptLength) / 1000.0
}

// StatsPath returns the path to the stats file (~/.murmur/stats.jsonl).
func StatsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".murmur", "stats.jsonl"), nil
}

// Record appends a usage record to the stats file.
func Record(record UsageRecord) error {
	path, err := StatsPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create stats directory: %w", err)
	}

	// Open file for append
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot open stats file: %w", err)
	}
	defer f.Close()

	// Write JSON line
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("cannot serialize record: %w", err)
	}

	if _, err := f.WriteString(string(data) + "\n"); err != nil {
		return fmt.Errorf("cannot write record: %w", err)
	}

	return nil
}

// Query reads and filters usage records.
func Query(filter QueryFilter) ([]UsageRecord, error) {
	path, err := StatsPath()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []UsageRecord{}, nil
		}
		return nil, fmt.Errorf("cannot open stats file: %w", err)
	}
	defer f.Close()

	var records []UsageRecord
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var record UsageRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			// Skip malformed lines
			continue
		}

		// Apply filters
		if filter.Tool != "" && record.Tool != filter.Tool {
			continue
		}
		if filter.Tier != "" && record.Tier != filter.Tier {
			continue
		}
		if !filter.StartTime.IsZero() && record.Timestamp.Before(filter.StartTime) {
			continue
		}
		if !filter.EndTime.IsZero() && record.Timestamp.After(filter.EndTime) {
			continue
		}

		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading stats file: %w", err)
	}

	return records, nil
}

// Summarize computes summary from records.
func Summarize(records []UsageRecord) Summary {
	summary := Summary{
		ByTool: make(map[string]ToolStats),
	}

	if len(records) == 0 {
		return summary
	}

	// Track successes per tool for success rate
	successCount := make(map[string]int)
	dailyCounts := make(map[string]int)

	for _, r := range records {
		summary.TotalRuns++
		summary.EstimatedCost += r.CostEstimate

		// Track what would have been paid if free tools weren't used
		if r.Tier == "free" {
			// Estimate what Claude would have cost
			summary.EstimatedSaved += EstimateCost("claude", r.PromptLength)
		}

		// Tool stats
		ts := summary.ByTool[r.Tool]
		ts.Count++
		ts.TotalTimeMs += r.DurationMs
		ts.TotalCost += r.CostEstimate
		if r.Success {
			successCount[r.Tool]++
		}
		summary.ByTool[r.Tool] = ts

		// Auto-route stats
		if r.AutoRouted {
			summary.AutoRouteStats.Total++
			if r.Tier == "free" {
				summary.AutoRouteStats.ToFree++
			} else {
				summary.AutoRouteStats.ToPaid++
			}
		}

		// Daily counts
		dateKey := r.Timestamp.Format("2006-01-02")
		dailyCounts[dateKey]++
	}

	// Calculate averages and success rates
	for tool, ts := range summary.ByTool {
		if ts.Count > 0 {
			ts.AvgTimeMs = ts.TotalTimeMs / int64(ts.Count)
			ts.SuccessRate = float64(successCount[tool]) / float64(ts.Count) * 100
		}
		summary.ByTool[tool] = ts
	}

	// Calculate free ratio
	if summary.AutoRouteStats.Total > 0 {
		summary.AutoRouteStats.FreeRatio = float64(summary.AutoRouteStats.ToFree) / float64(summary.AutoRouteStats.Total) * 100
	}

	// Build daily trend (last 7 days)
	today := time.Now().Truncate(24 * time.Hour)
	for i := 6; i >= 0; i-- {
		date := today.AddDate(0, 0, -i)
		dateKey := date.Format("2006-01-02")
		summary.DailyTrend = append(summary.DailyTrend, DailyStats{
			Date:  dateKey,
			Count: dailyCounts[dateKey],
		})
	}

	return summary
}

// Reset clears all stats.
func Reset() error {
	path, err := StatsPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot remove stats file: %w", err)
	}

	return nil
}

// FormatSummary returns a human-readable summary string.
func FormatSummary(s Summary) string {
	var sb strings.Builder

	sb.WriteString("📊 Usage Statistics\n")
	sb.WriteString("==================\n\n")

	if s.TotalRuns == 0 {
		sb.WriteString("No usage data recorded yet.\n")
		sb.WriteString("Run `mur run -p \"your prompt\"` to start tracking.\n")
		return sb.String()
	}

	// Overview
	sb.WriteString(fmt.Sprintf("Total Runs: %d\n", s.TotalRuns))
	sb.WriteString(fmt.Sprintf("Estimated Cost: $%.4f\n", s.EstimatedCost))
	sb.WriteString(fmt.Sprintf("Estimated Saved: $%.4f (by using free tools)\n", s.EstimatedSaved))
	sb.WriteString("\n")

	// Tool breakdown
	sb.WriteString("📦 By Tool\n")
	sb.WriteString("----------\n")

	// Sort tools by count
	type toolEntry struct {
		name  string
		stats ToolStats
	}
	var tools []toolEntry
	for name, ts := range s.ByTool {
		tools = append(tools, toolEntry{name, ts})
	}
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].stats.Count > tools[j].stats.Count
	})

	maxCount := 0
	for _, t := range tools {
		if t.stats.Count > maxCount {
			maxCount = t.stats.Count
		}
	}

	for _, t := range tools {
		barLen := 0
		if maxCount > 0 {
			barLen = t.stats.Count * 20 / maxCount
		}
		if barLen == 0 && t.stats.Count > 0 {
			barLen = 1
		}
		bar := strings.Repeat("█", barLen)
		pct := float64(t.stats.Count) / float64(s.TotalRuns) * 100

		sb.WriteString(fmt.Sprintf("%-8s %s %d (%.0f%%)\n", t.name, bar, t.stats.Count, pct))
		sb.WriteString(fmt.Sprintf("         avg: %dms, cost: $%.4f, success: %.0f%%\n",
			t.stats.AvgTimeMs, t.stats.TotalCost, t.stats.SuccessRate))
	}
	sb.WriteString("\n")

	// Auto-routing stats
	if s.AutoRouteStats.Total > 0 {
		sb.WriteString("🔀 Auto-Routing Decisions\n")
		sb.WriteString("-------------------------\n")
		sb.WriteString(fmt.Sprintf("Total auto-routed: %d\n", s.AutoRouteStats.Total))
		sb.WriteString(fmt.Sprintf("→ Free tools: %d (%.0f%%)\n", s.AutoRouteStats.ToFree, s.AutoRouteStats.FreeRatio))
		sb.WriteString(fmt.Sprintf("→ Paid tools: %d (%.0f%%)\n", s.AutoRouteStats.ToPaid, 100-s.AutoRouteStats.FreeRatio))
		sb.WriteString("\n")
	}

	// Daily trend
	sb.WriteString("📈 Last 7 Days\n")
	sb.WriteString("--------------\n")

	maxDaily := 0
	for _, d := range s.DailyTrend {
		if d.Count > maxDaily {
			maxDaily = d.Count
		}
	}

	for _, d := range s.DailyTrend {
		date, _ := time.Parse("2006-01-02", d.Date)
		dayName := date.Format("Mon")

		barLen := 0
		if maxDaily > 0 {
			barLen = d.Count * 20 / maxDaily
		}
		if barLen == 0 && d.Count > 0 {
			barLen = 1
		}
		bar := strings.Repeat("█", barLen)

		sb.WriteString(fmt.Sprintf("%s  %s %d\n", dayName, bar, d.Count))
	}

	return sb.String()
}

// FormatToolStats returns formatted stats for a specific tool.
func FormatToolStats(tool string, records []UsageRecord) string {
	var sb strings.Builder

	// Filter to this tool
	var toolRecords []UsageRecord
	for _, r := range records {
		if r.Tool == tool {
			toolRecords = append(toolRecords, r)
		}
	}

	sb.WriteString(fmt.Sprintf("📊 Statistics for %s\n", tool))
	sb.WriteString(strings.Repeat("=", 20+len(tool)) + "\n\n")

	if len(toolRecords) == 0 {
		sb.WriteString(fmt.Sprintf("No usage data for %s.\n", tool))
		return sb.String()
	}

	// Calculate stats
	var totalTime int64
	var totalCost float64
	var successCount int
	complexities := []float64{}

	for _, r := range toolRecords {
		totalTime += r.DurationMs
		totalCost += r.CostEstimate
		if r.Success {
			successCount++
		}
		complexities = append(complexities, r.Complexity)
	}

	count := len(toolRecords)
	avgTime := totalTime / int64(count)
	successRate := float64(successCount) / float64(count) * 100

	var avgComplexity float64
	for _, c := range complexities {
		avgComplexity += c
	}
	avgComplexity /= float64(len(complexities))

	sb.WriteString(fmt.Sprintf("Total Runs:     %d\n", count))
	sb.WriteString(fmt.Sprintf("Success Rate:   %.0f%%\n", successRate))
	sb.WriteString(fmt.Sprintf("Avg Duration:   %dms\n", avgTime))
	sb.WriteString(fmt.Sprintf("Total Cost:     $%.4f\n", totalCost))
	sb.WriteString(fmt.Sprintf("Avg Complexity: %.2f\n", avgComplexity))

	return sb.String()
}
