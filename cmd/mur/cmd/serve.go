package cmd

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/mur-run/mur-core/internal/stats"
	"github.com/spf13/cobra"
)

var (
	servePort int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start local dashboard server",
	Long: `Start a local web dashboard for viewing patterns and analytics.

The dashboard runs on localhost and provides:
  - Pattern browser with search and filters
  - Usage analytics and charts
  - Tool usage breakdown
  - Cost tracking and savings
  - Effectiveness metrics
  - Sync status for all targets
  - Quick actions

Examples:
  mur serve              # Start on default port 8742
  mur serve --port 3000  # Start on custom port`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8742, "Port to run dashboard on")
}

// DashboardData holds data for the dashboard template
type DashboardData struct {
	// Patterns
	Patterns       []PatternView
	TotalPatterns  int
	ActivePatterns int
	TopPatterns    []PatternView
	RecentPatterns []PatternView

	// Usage Stats
	TotalUsage      int
	TotalRuns       int
	AvgEffective    float64
	EstimatedCost   float64
	EstimatedSaved  float64
	DailyTrend      []DailyPoint
	ToolBreakdown   []ToolUsage
	AutoRouteStats  AutoRouteView

	// Sync Status
	SyncTargets []SyncTarget

	// Meta
	LastSync    string
	GeneratedAt string
	Version     string
}

// PatternView is a simplified pattern for display
type PatternView struct {
	Name          string
	Description   string
	Tags          []string
	Domain        string
	Effectiveness float64
	UsageCount    int
	LastUsed      string
	CreatedAt     string
	Status        string
	Source        string
}

// DailyPoint for trend chart
type DailyPoint struct {
	Date  string
	Day   string
	Count int
}

// ToolUsage for breakdown
type ToolUsage struct {
	Name       string
	Count      int
	Percentage float64
	Cost       float64
	AvgTimeMs  int64
	Tier       string
}

// AutoRouteView for auto-routing stats
type AutoRouteView struct {
	Total     int
	ToFree    int
	ToPaid    int
	FreeRatio float64
}

// SyncTarget for sync status
type SyncTarget struct {
	Name      string
	Type      string // "cli" or "ide"
	Path      string
	Exists    bool
	FileCount int
	LastMod   string
}

func runServe(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)

	// Set up HTTP handlers
	mux := http.NewServeMux()

	// Main dashboard
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		serveDashboard(w, r, store)
	})

	// API endpoints
	mux.HandleFunc("/api/patterns", func(w http.ResponseWriter, r *http.Request) {
		servePatterns(w, r, store)
	})

	mux.HandleFunc("/api/pattern/", func(w http.ResponseWriter, r *http.Request) {
		servePatternDetail(w, r, store)
	})

	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		serveStats(w, r, store)
	})

	mux.HandleFunc("/api/sync", func(w http.ResponseWriter, r *http.Request) {
		handleSyncAction(w, r)
	})

	addr := fmt.Sprintf("localhost:%d", servePort)
	url := fmt.Sprintf("http://%s", addr)

	fmt.Println()
	fmt.Println("🌐 mur Dashboard")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("   Running at: %s\n", url)
	fmt.Println("   Press Ctrl+C to stop")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Try to open browser
	openBrowser(url)

	return http.ListenAndServe(addr, mux)
}

func serveDashboard(w http.ResponseWriter, r *http.Request, store *pattern.Store) {
	patterns, err := store.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := buildDashboardData(patterns)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
		"sub": func(a, b float64) float64 { return a - b },
		"div": func(a, b int) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b) * 100
		},
		"printf": fmt.Sprintf,
	}

	tmpl := template.Must(template.New("dashboard").Funcs(funcMap).Parse(dashboardHTML))
	tmpl.Execute(w, data)
}

func servePatterns(w http.ResponseWriter, r *http.Request, store *pattern.Store) {
	patterns, err := store.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	views := make([]PatternView, 0, len(patterns))
	for _, p := range patterns {
		views = append(views, patternToView(&p))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(views)
}

func servePatternDetail(w http.ResponseWriter, r *http.Request, store *pattern.Store) {
	name := strings.TrimPrefix(r.URL.Path, "/api/pattern/")
	if name == "" {
		http.Error(w, "pattern name required", http.StatusBadRequest)
		return
	}

	p, err := store.Get(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func serveStats(w http.ResponseWriter, r *http.Request, store *pattern.Store) {
	patterns, err := store.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := buildDashboardData(patterns)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func handleSyncAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cmd := exec.Command("mur", "sync", "--quiet")
	output, err := cmd.CombinedOutput()

	result := map[string]interface{}{
		"success": err == nil,
		"output":  string(output),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func buildDashboardData(patterns []pattern.Pattern) DashboardData {
	data := DashboardData{
		Patterns:    make([]PatternView, 0, len(patterns)),
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Version:     "0.7.0",
	}

	var totalEffectiveness float64
	effectiveCount := 0

	for _, p := range patterns {
		view := patternToView(&p)
		data.Patterns = append(data.Patterns, view)
		data.TotalUsage += view.UsageCount

		if view.Effectiveness > 0 {
			totalEffectiveness += view.Effectiveness
			effectiveCount++
		}

		if view.Status == "active" || view.Status == "" {
			data.ActivePatterns++
		}
	}

	data.TotalPatterns = len(patterns)
	if effectiveCount > 0 {
		data.AvgEffective = totalEffectiveness / float64(effectiveCount) * 100
	}

	// Sort for top patterns (by effectiveness)
	sorted := make([]PatternView, len(data.Patterns))
	copy(sorted, data.Patterns)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Effectiveness > sorted[j].Effectiveness
	})
	if len(sorted) > 5 {
		sorted = sorted[:5]
	}
	data.TopPatterns = sorted

	// Sort for recent patterns (by created date)
	recent := make([]PatternView, len(data.Patterns))
	copy(recent, data.Patterns)
	sort.Slice(recent, func(i, j int) bool {
		return recent[i].CreatedAt > recent[j].CreatedAt
	})
	if len(recent) > 5 {
		recent = recent[:5]
	}
	data.RecentPatterns = recent

	// Load usage stats
	records, _ := stats.Query(stats.QueryFilter{})
	if len(records) > 0 {
		summary := stats.Summarize(records)
		data.TotalRuns = summary.TotalRuns
		data.EstimatedCost = summary.EstimatedCost
		data.EstimatedSaved = summary.EstimatedSaved

		data.AutoRouteStats = AutoRouteView{
			Total:     summary.AutoRouteStats.Total,
			ToFree:    summary.AutoRouteStats.ToFree,
			ToPaid:    summary.AutoRouteStats.ToPaid,
			FreeRatio: summary.AutoRouteStats.FreeRatio,
		}

		// Daily trend
		for _, d := range summary.DailyTrend {
			date, _ := time.Parse("2006-01-02", d.Date)
			data.DailyTrend = append(data.DailyTrend, DailyPoint{
				Date:  d.Date,
				Day:   date.Format("Mon"),
				Count: d.Count,
			})
		}

		// Tool breakdown
		for name, ts := range summary.ByTool {
			tier := "paid"
			if name == "gemini" || name == "auggie" {
				tier = "free"
			}
			data.ToolBreakdown = append(data.ToolBreakdown, ToolUsage{
				Name:       name,
				Count:      ts.Count,
				Percentage: float64(ts.Count) / float64(summary.TotalRuns) * 100,
				Cost:       ts.TotalCost,
				AvgTimeMs:  ts.AvgTimeMs,
				Tier:       tier,
			})
		}
		// Sort by count descending
		sort.Slice(data.ToolBreakdown, func(i, j int) bool {
			return data.ToolBreakdown[i].Count > data.ToolBreakdown[j].Count
		})
	}

	// Sync targets
	data.SyncTargets = getSyncTargets()

	return data
}

func getSyncTargets() []SyncTarget {
	home, _ := os.UserHomeDir()
	targets := []SyncTarget{
		// CLIs
		{Name: "Claude Code", Type: "cli", Path: filepath.Join(home, ".claude", "skills", "mur")},
		{Name: "Gemini CLI", Type: "cli", Path: filepath.Join(home, ".gemini", "skills", "mur")},
		{Name: "Codex CLI", Type: "cli", Path: filepath.Join(home, ".codex", "instructions.md")},
		{Name: "Auggie", Type: "cli", Path: filepath.Join(home, ".augment", "skills", "mur")},
		{Name: "Aider", Type: "cli", Path: filepath.Join(home, ".aider", "mur-patterns.md")},
		// IDEs
		{Name: "Continue", Type: "ide", Path: filepath.Join(home, ".continue", "rules", "mur")},
		{Name: "Cursor", Type: "ide", Path: filepath.Join(home, ".cursor", "rules", "mur")},
		{Name: "Windsurf", Type: "ide", Path: filepath.Join(home, ".windsurf", "rules", "mur")},
	}

	for i := range targets {
		info, err := os.Stat(targets[i].Path)
		if err == nil {
			targets[i].Exists = true
			if info.IsDir() {
				files, _ := os.ReadDir(targets[i].Path)
				targets[i].FileCount = len(files)
			} else {
				targets[i].FileCount = 1
			}
			targets[i].LastMod = info.ModTime().Format("Jan 2 15:04")
		}
	}

	return targets
}

func patternToView(p *pattern.Pattern) PatternView {
	var tags []string
	tags = append(tags, p.Tags.Confirmed...)
	for _, t := range p.Tags.Inferred {
		if t.Confidence >= 0.7 {
			tags = append(tags, t.Tag)
		}
	}

	lastUsed := "Never"
	if p.Learning.LastUsed != nil {
		lastUsed = p.Learning.LastUsed.Format("2006-01-02")
	}

	createdAt := ""
	if !p.Lifecycle.Created.IsZero() {
		createdAt = p.Lifecycle.Created.Format("2006-01-02")
	}

	// Extract domain from tags if available
	domain := ""
	for _, t := range p.Tags.Confirmed {
		if t == "go" || t == "swift" || t == "python" || t == "node" || t == "rust" {
			domain = t
			break
		}
	}
	if domain == "" {
		for _, t := range p.Tags.Inferred {
			if t.Confidence >= 0.7 && (t.Tag == "go" || t.Tag == "swift" || t.Tag == "python" || t.Tag == "node" || t.Tag == "rust") {
				domain = t.Tag
				break
			}
		}
	}

	return PatternView{
		Name:          p.Name,
		Description:   p.Description,
		Tags:          tags,
		Domain:        domain,
		Effectiveness: p.Learning.Effectiveness,
		UsageCount:    p.Learning.UsageCount,
		LastUsed:      lastUsed,
		CreatedAt:     createdAt,
		Status:        string(p.Lifecycle.Status),
		Source:        "",
	}
}

func openBrowser(url string) {
	var cmd string
	var args []string

	switch {
	case fileExists("/usr/bin/open"):
		cmd = "open"
		args = []string{url}
	case fileExists("/usr/bin/xdg-open"):
		cmd = "xdg-open"
		args = []string{url}
	default:
		return
	}

	go func() {
		execCmd := exec.Command(cmd, args...)
		execCmd.Run()
	}()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Dashboard HTML template with enhanced features
const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>mur Dashboard</title>
    <style>
        :root {
            --bg-primary: #0f172a;
            --bg-secondary: #1e293b;
            --bg-tertiary: #334155;
            --text-primary: #f1f5f9;
            --text-secondary: #94a3b8;
            --text-muted: #64748b;
            --accent: #38bdf8;
            --accent-hover: #0ea5e9;
            --success: #4ade80;
            --success-bg: #065f46;
            --warning: #fbbf24;
            --warning-bg: #78350f;
            --error: #f87171;
            --border: #334155;
        }
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--bg-primary);
            color: var(--text-primary);
            min-height: 100vh;
            line-height: 1.5;
        }
        .container { max-width: 1400px; margin: 0 auto; padding: 2rem; }
        
        /* Header */
        header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 2rem;
            padding-bottom: 1rem;
            border-bottom: 1px solid var(--border);
        }
        .logo {
            font-size: 1.5rem;
            font-weight: bold;
            color: var(--accent);
        }
        .logo span { color: var(--text-primary); }
        .header-right { display: flex; align-items: center; gap: 1rem; }
        .version {
            background: var(--bg-tertiary);
            padding: 0.25rem 0.5rem;
            border-radius: 0.25rem;
            font-size: 0.75rem;
            color: var(--text-secondary);
        }
        .generated { color: var(--text-muted); font-size: 0.875rem; }
        
        /* Grid Layout */
        .grid { display: grid; gap: 1.5rem; }
        .grid-2 { grid-template-columns: repeat(2, 1fr); }
        .grid-3 { grid-template-columns: repeat(3, 1fr); }
        .grid-4 { grid-template-columns: repeat(4, 1fr); }
        @media (max-width: 1200px) { .grid-4 { grid-template-columns: repeat(2, 1fr); } }
        @media (max-width: 768px) { 
            .grid-2, .grid-3, .grid-4 { grid-template-columns: 1fr; }
        }
        
        /* Cards */
        .card {
            background: var(--bg-secondary);
            border-radius: 0.75rem;
            padding: 1.5rem;
            border: 1px solid var(--border);
        }
        .card-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1rem;
        }
        .card-title {
            font-size: 0.875rem;
            color: var(--text-secondary);
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
        
        /* Stats */
        .stat-value {
            font-size: 2.5rem;
            font-weight: bold;
            color: var(--accent);
            line-height: 1;
        }
        .stat-value.green { color: var(--success); }
        .stat-value.yellow { color: var(--warning); }
        .stat-label {
            font-size: 0.875rem;
            color: var(--text-secondary);
            margin-top: 0.5rem;
        }
        .stat-change {
            font-size: 0.75rem;
            margin-top: 0.25rem;
        }
        .stat-change.up { color: var(--success); }
        .stat-change.down { color: var(--error); }
        
        /* Section */
        .section { margin-bottom: 2rem; }
        .section-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1rem;
        }
        .section-title {
            font-size: 1.25rem;
            color: var(--text-primary);
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        
        /* Patterns */
        .patterns-grid { display: grid; gap: 1rem; }
        .pattern-card {
            background: var(--bg-secondary);
            border-radius: 0.75rem;
            padding: 1.25rem;
            border: 1px solid var(--border);
            transition: all 0.2s;
            cursor: pointer;
        }
        .pattern-card:hover { 
            border-color: var(--accent);
            transform: translateY(-2px);
        }
        .pattern-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 0.75rem;
        }
        .pattern-name {
            font-weight: 600;
            color: var(--text-primary);
            font-size: 1rem;
        }
        .pattern-effectiveness {
            background: var(--success-bg);
            color: var(--success);
            padding: 0.25rem 0.5rem;
            border-radius: 0.375rem;
            font-size: 0.75rem;
            font-weight: 600;
        }
        .pattern-effectiveness.low {
            background: var(--warning-bg);
            color: var(--warning);
        }
        .pattern-description {
            color: var(--text-secondary);
            font-size: 0.875rem;
            margin-bottom: 0.75rem;
            display: -webkit-box;
            -webkit-line-clamp: 2;
            -webkit-box-orient: vertical;
            overflow: hidden;
        }
        .pattern-tags { display: flex; gap: 0.5rem; flex-wrap: wrap; }
        .tag {
            background: var(--bg-tertiary);
            color: var(--text-secondary);
            padding: 0.25rem 0.5rem;
            border-radius: 0.25rem;
            font-size: 0.75rem;
        }
        .tag.domain {
            background: rgba(56, 189, 248, 0.2);
            color: var(--accent);
        }
        .pattern-meta {
            margin-top: 0.75rem;
            display: flex;
            gap: 1rem;
            color: var(--text-muted);
            font-size: 0.75rem;
        }
        
        /* Search */
        .search-container { position: relative; margin-bottom: 1rem; }
        .search-box {
            background: var(--bg-secondary);
            border: 1px solid var(--border);
            border-radius: 0.5rem;
            padding: 0.75rem 1rem 0.75rem 2.5rem;
            color: var(--text-primary);
            width: 100%;
            font-size: 1rem;
        }
        .search-box:focus {
            outline: none;
            border-color: var(--accent);
        }
        .search-box::placeholder { color: var(--text-muted); }
        .search-icon {
            position: absolute;
            left: 0.75rem;
            top: 50%;
            transform: translateY(-50%);
            color: var(--text-muted);
        }
        
        /* Filters */
        .filters {
            display: flex;
            gap: 0.5rem;
            margin-bottom: 1rem;
            flex-wrap: wrap;
        }
        .filter-btn {
            background: var(--bg-tertiary);
            border: 1px solid var(--border);
            border-radius: 0.375rem;
            padding: 0.5rem 1rem;
            color: var(--text-secondary);
            font-size: 0.875rem;
            cursor: pointer;
            transition: all 0.2s;
        }
        .filter-btn:hover, .filter-btn.active {
            background: var(--accent);
            color: white;
            border-color: var(--accent);
        }
        
        /* Bar Chart */
        .bar-chart { display: flex; flex-direction: column; gap: 0.75rem; }
        .bar-item { display: flex; align-items: center; gap: 1rem; }
        .bar-label {
            width: 80px;
            font-size: 0.875rem;
            color: var(--text-secondary);
        }
        .bar-container {
            flex: 1;
            height: 24px;
            background: var(--bg-tertiary);
            border-radius: 0.25rem;
            overflow: hidden;
        }
        .bar-fill {
            height: 100%;
            background: linear-gradient(90deg, var(--accent), #818cf8);
            border-radius: 0.25rem;
            transition: width 0.5s ease;
        }
        .bar-fill.free { background: linear-gradient(90deg, var(--success), #22d3ee); }
        .bar-value {
            width: 60px;
            text-align: right;
            font-size: 0.875rem;
            color: var(--text-secondary);
        }
        
        /* Sparkline */
        .sparkline {
            display: flex;
            align-items: flex-end;
            gap: 4px;
            height: 60px;
            padding: 0.5rem 0;
        }
        .spark-bar {
            flex: 1;
            background: var(--accent);
            border-radius: 2px;
            transition: height 0.3s ease;
            min-height: 4px;
        }
        .spark-bar:hover { background: var(--accent-hover); }
        .spark-labels {
            display: flex;
            justify-content: space-between;
            font-size: 0.75rem;
            color: var(--text-muted);
            margin-top: 0.5rem;
        }
        
        /* Sync Status */
        .sync-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 0.75rem; }
        @media (max-width: 768px) { .sync-grid { grid-template-columns: 1fr; } }
        .sync-item {
            display: flex;
            align-items: center;
            gap: 0.75rem;
            padding: 0.75rem;
            background: var(--bg-tertiary);
            border-radius: 0.5rem;
        }
        .sync-icon {
            width: 32px;
            height: 32px;
            border-radius: 0.375rem;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 1rem;
        }
        .sync-icon.cli { background: rgba(56, 189, 248, 0.2); }
        .sync-icon.ide { background: rgba(168, 85, 247, 0.2); }
        .sync-info { flex: 1; }
        .sync-name { font-size: 0.875rem; color: var(--text-primary); }
        .sync-detail { font-size: 0.75rem; color: var(--text-muted); }
        .sync-status {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            background: var(--success);
        }
        .sync-status.inactive { background: var(--text-muted); }
        
        /* Buttons */
        .btn {
            background: var(--accent);
            color: white;
            border: none;
            border-radius: 0.5rem;
            padding: 0.75rem 1.5rem;
            font-size: 0.875rem;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.2s;
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
        }
        .btn:hover { background: var(--accent-hover); }
        .btn:disabled { opacity: 0.5; cursor: not-allowed; }
        .btn-secondary {
            background: var(--bg-tertiary);
            color: var(--text-secondary);
        }
        .btn-secondary:hover { background: var(--border); }
        
        /* Quick Actions */
        .quick-actions {
            display: flex;
            gap: 0.75rem;
            flex-wrap: wrap;
        }
        
        /* Empty State */
        .empty-state {
            text-align: center;
            padding: 3rem;
            color: var(--text-muted);
        }
        .empty-state-icon { font-size: 3rem; margin-bottom: 1rem; }
        
        /* Modal */
        .modal-overlay {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: rgba(0, 0, 0, 0.7);
            z-index: 1000;
            align-items: center;
            justify-content: center;
        }
        .modal-overlay.active { display: flex; }
        .modal {
            background: var(--bg-secondary);
            border-radius: 1rem;
            padding: 2rem;
            max-width: 600px;
            width: 90%;
            max-height: 80vh;
            overflow-y: auto;
            border: 1px solid var(--border);
        }
        .modal-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1.5rem;
        }
        .modal-title { font-size: 1.25rem; font-weight: 600; }
        .modal-close {
            background: none;
            border: none;
            color: var(--text-muted);
            font-size: 1.5rem;
            cursor: pointer;
        }
        .modal-close:hover { color: var(--text-primary); }
        
        /* Tabs */
        .tabs {
            display: flex;
            gap: 0;
            border-bottom: 1px solid var(--border);
            margin-bottom: 1rem;
        }
        .tab {
            padding: 0.75rem 1.5rem;
            color: var(--text-secondary);
            cursor: pointer;
            border-bottom: 2px solid transparent;
            transition: all 0.2s;
        }
        .tab:hover { color: var(--text-primary); }
        .tab.active {
            color: var(--accent);
            border-bottom-color: var(--accent);
        }
        
        /* Footer */
        footer {
            text-align: center;
            padding: 2rem;
            color: var(--text-muted);
            font-size: 0.875rem;
        }
        footer a { color: var(--accent); text-decoration: none; }
        footer a:hover { text-decoration: underline; }
        
        /* Toast */
        .toast {
            position: fixed;
            bottom: 2rem;
            right: 2rem;
            background: var(--bg-secondary);
            border: 1px solid var(--border);
            border-radius: 0.5rem;
            padding: 1rem 1.5rem;
            display: flex;
            align-items: center;
            gap: 0.75rem;
            box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
            transform: translateY(100px);
            opacity: 0;
            transition: all 0.3s ease;
            z-index: 1001;
        }
        .toast.show { transform: translateY(0); opacity: 1; }
        .toast.success { border-color: var(--success); }
        .toast.error { border-color: var(--error); }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <div class="logo">mur<span>.dashboard</span></div>
            <div class="header-right">
                <span class="version">v{{.Version}}</span>
                <span class="generated">{{.GeneratedAt}}</span>
            </div>
        </header>
        
        <!-- Stats Overview -->
        <div class="section">
            <div class="grid grid-4">
                <div class="card">
                    <div class="card-title">Total Patterns</div>
                    <div class="stat-value">{{.TotalPatterns}}</div>
                    <div class="stat-label">{{.ActivePatterns}} active</div>
                </div>
                <div class="card">
                    <div class="card-title">Total Usage</div>
                    <div class="stat-value">{{.TotalUsage}}</div>
                    <div class="stat-label">pattern injections</div>
                </div>
                <div class="card">
                    <div class="card-title">Avg Effectiveness</div>
                    <div class="stat-value green">{{printf "%.0f" .AvgEffective}}%</div>
                    <div class="stat-label">across all patterns</div>
                </div>
                <div class="card">
                    <div class="card-title">Estimated Saved</div>
                    <div class="stat-value yellow">${{printf "%.2f" .EstimatedSaved}}</div>
                    <div class="stat-label">by using free tools</div>
                </div>
            </div>
        </div>
        
        <!-- Charts Row -->
        <div class="section">
            <div class="grid grid-2">
                <!-- Daily Trend -->
                <div class="card">
                    <div class="card-header">
                        <span class="card-title">📈 Usage Trend (7 Days)</span>
                    </div>
                    {{if .DailyTrend}}
                    <div class="sparkline" id="sparkline">
                        {{range .DailyTrend}}
                        <div class="spark-bar" style="height: 4px;" data-count="{{.Count}}" title="{{.Day}}: {{.Count}}"></div>
                        {{end}}
                    </div>
                    <div class="spark-labels">
                        {{range .DailyTrend}}<span>{{.Day}}</span>{{end}}
                    </div>
                    {{else}}
                    <div class="empty-state" style="padding: 1rem;">
                        <p>No usage data yet</p>
                    </div>
                    {{end}}
                </div>
                
                <!-- Tool Breakdown -->
                <div class="card">
                    <div class="card-header">
                        <span class="card-title">🔧 Tool Usage</span>
                    </div>
                    {{if .ToolBreakdown}}
                    <div class="bar-chart">
                        {{range .ToolBreakdown}}
                        <div class="bar-item">
                            <span class="bar-label">{{.Name}}</span>
                            <div class="bar-container">
                                <div class="bar-fill {{if eq .Tier "free"}}free{{end}}" style="width: {{printf "%.0f" .Percentage}}%;"></div>
                            </div>
                            <span class="bar-value">{{.Count}}</span>
                        </div>
                        {{end}}
                    </div>
                    {{else}}
                    <div class="empty-state" style="padding: 1rem;">
                        <p>No tool usage recorded</p>
                    </div>
                    {{end}}
                </div>
            </div>
        </div>
        
        <!-- Sync Status & Quick Actions -->
        <div class="section">
            <div class="grid grid-2">
                <!-- Sync Status -->
                <div class="card">
                    <div class="card-header">
                        <span class="card-title">🔄 Sync Status</span>
                        <button class="btn btn-secondary" onclick="triggerSync()" id="syncBtn">
                            Sync Now
                        </button>
                    </div>
                    <div class="sync-grid">
                        {{range .SyncTargets}}
                        <div class="sync-item">
                            <div class="sync-icon {{.Type}}">
                                {{if eq .Type "cli"}}⌨️{{else}}🖥️{{end}}
                            </div>
                            <div class="sync-info">
                                <div class="sync-name">{{.Name}}</div>
                                <div class="sync-detail">
                                    {{if .Exists}}{{.FileCount}} files • {{.LastMod}}{{else}}Not synced{{end}}
                                </div>
                            </div>
                            <div class="sync-status {{if not .Exists}}inactive{{end}}"></div>
                        </div>
                        {{end}}
                    </div>
                </div>
                
                <!-- Auto-Routing Stats -->
                <div class="card">
                    <div class="card-header">
                        <span class="card-title">🔀 Auto-Routing</span>
                    </div>
                    {{if gt .AutoRouteStats.Total 0}}
                    <div style="display: flex; gap: 2rem; margin-bottom: 1rem;">
                        <div>
                            <div class="stat-value" style="font-size: 1.5rem;">{{.AutoRouteStats.Total}}</div>
                            <div class="stat-label">Total Routed</div>
                        </div>
                        <div>
                            <div class="stat-value green" style="font-size: 1.5rem;">{{printf "%.0f" .AutoRouteStats.FreeRatio}}%</div>
                            <div class="stat-label">To Free Tools</div>
                        </div>
                    </div>
                    <div class="bar-chart">
                        <div class="bar-item">
                            <span class="bar-label">Free</span>
                            <div class="bar-container">
                                <div class="bar-fill free" style="width: {{printf "%.0f" .AutoRouteStats.FreeRatio}}%;"></div>
                            </div>
                            <span class="bar-value">{{.AutoRouteStats.ToFree}}</span>
                        </div>
                        <div class="bar-item">
                            <span class="bar-label">Paid</span>
                            <div class="bar-container">
                                <div class="bar-fill" style="width: {{printf "%.0f" (sub 100 .AutoRouteStats.FreeRatio)}}%;"></div>
                            </div>
                            <span class="bar-value">{{.AutoRouteStats.ToPaid}}</span>
                        </div>
                    </div>
                    {{else}}
                    <div class="empty-state" style="padding: 1rem;">
                        <p>No auto-routing data yet</p>
                        <p style="font-size: 0.75rem; margin-top: 0.5rem;">Use <code>mur run</code> to start</p>
                    </div>
                    {{end}}
                </div>
            </div>
        </div>
        
        {{if .TopPatterns}}
        <!-- Top Patterns -->
        <div class="section">
            <div class="section-header">
                <h2 class="section-title">⭐ Top Patterns</h2>
            </div>
            <div class="patterns-grid" style="grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));">
                {{range .TopPatterns}}
                <div class="pattern-card" onclick="showPattern('{{.Name}}')">
                    <div class="pattern-header">
                        <span class="pattern-name">{{.Name}}</span>
                        {{if gt .Effectiveness 0}}
                        <span class="pattern-effectiveness {{if lt .Effectiveness 0.5}}low{{end}}">{{printf "%.0f" (mul .Effectiveness 100)}}%</span>
                        {{end}}
                    </div>
                    {{if .Description}}
                    <div class="pattern-description">{{.Description}}</div>
                    {{end}}
                    <div class="pattern-tags">
                        {{if .Domain}}<span class="tag domain">{{.Domain}}</span>{{end}}
                        {{range .Tags}}
                        <span class="tag">{{.}}</span>
                        {{end}}
                    </div>
                    <div class="pattern-meta">
                        <span>📊 {{.UsageCount}} uses</span>
                        <span>🕐 {{.LastUsed}}</span>
                    </div>
                </div>
                {{end}}
            </div>
        </div>
        {{end}}
        
        <!-- All Patterns -->
        <div class="section">
            <div class="section-header">
                <h2 class="section-title">📚 All Patterns</h2>
            </div>
            
            <div class="search-container">
                <span class="search-icon">🔍</span>
                <input type="text" class="search-box" placeholder="Search patterns by name, tag, or domain..." id="search">
            </div>
            
            <div class="filters" id="filters">
                <button class="filter-btn active" data-filter="all">All</button>
                <button class="filter-btn" data-filter="active">Active</button>
                <button class="filter-btn" data-filter="deprecated">Deprecated</button>
                <button class="filter-btn" data-filter="go">Go</button>
                <button class="filter-btn" data-filter="swift">Swift</button>
                <button class="filter-btn" data-filter="general">General</button>
            </div>
            
            {{if .Patterns}}
            <div class="patterns-grid" id="patterns-list" style="grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));">
                {{range .Patterns}}
                <div class="pattern-card" 
                     data-name="{{.Name}}" 
                     data-tags="{{range .Tags}}{{.}} {{end}}"
                     data-domain="{{.Domain}}"
                     data-status="{{.Status}}"
                     onclick="showPattern('{{.Name}}')">
                    <div class="pattern-header">
                        <span class="pattern-name">{{.Name}}</span>
                        {{if gt .Effectiveness 0}}
                        <span class="pattern-effectiveness {{if lt .Effectiveness 0.5}}low{{end}}">{{printf "%.0f" (mul .Effectiveness 100)}}%</span>
                        {{end}}
                    </div>
                    {{if .Description}}
                    <div class="pattern-description">{{.Description}}</div>
                    {{end}}
                    <div class="pattern-tags">
                        {{if .Domain}}<span class="tag domain">{{.Domain}}</span>{{end}}
                        {{range .Tags}}
                        <span class="tag">{{.}}</span>
                        {{end}}
                    </div>
                    <div class="pattern-meta">
                        <span>📊 {{.UsageCount}} uses</span>
                        {{if .CreatedAt}}<span>📅 {{.CreatedAt}}</span>{{end}}
                    </div>
                </div>
                {{end}}
            </div>
            {{else}}
            <div class="empty-state">
                <div class="empty-state-icon">📭</div>
                <p>No patterns yet</p>
                <p style="margin-top: 0.5rem; font-size: 0.875rem;">Run <code>mur learn add</code> to create your first pattern</p>
            </div>
            {{end}}
        </div>
        
        <footer>
            <p>mur — Continuous learning for AI assistants</p>
            <p style="margin-top: 0.5rem;">
                <a href="https://github.com/mur-run/mur-core">GitHub</a> · 
                <a href="https://mur.run">Documentation</a>
            </p>
        </footer>
    </div>
    
    <!-- Pattern Detail Modal -->
    <div class="modal-overlay" id="patternModal">
        <div class="modal">
            <div class="modal-header">
                <h3 class="modal-title" id="modalTitle">Pattern Details</h3>
                <button class="modal-close" onclick="closeModal()">&times;</button>
            </div>
            <div id="modalContent">Loading...</div>
        </div>
    </div>
    
    <!-- Toast -->
    <div class="toast" id="toast">
        <span id="toastIcon">✓</span>
        <span id="toastMessage">Action completed</span>
    </div>
    
    <script>
        // Sparkline animation
        document.addEventListener('DOMContentLoaded', () => {
            const bars = document.querySelectorAll('.spark-bar');
            const maxCount = Math.max(...Array.from(bars).map(b => parseInt(b.dataset.count) || 0), 1);
            
            setTimeout(() => {
                bars.forEach(bar => {
                    const count = parseInt(bar.dataset.count) || 0;
                    const height = Math.max((count / maxCount) * 100, 8);
                    bar.style.height = height + '%';
                });
            }, 100);
        });
        
        // Search
        const search = document.getElementById('search');
        const patterns = document.querySelectorAll('#patterns-list .pattern-card');
        
        search?.addEventListener('input', (e) => {
            const query = e.target.value.toLowerCase();
            filterPatterns(query, getCurrentFilter());
        });
        
        // Filters
        document.querySelectorAll('.filter-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
                filterPatterns(search?.value?.toLowerCase() || '', btn.dataset.filter);
            });
        });
        
        function getCurrentFilter() {
            return document.querySelector('.filter-btn.active')?.dataset.filter || 'all';
        }
        
        function filterPatterns(query, filter) {
            patterns.forEach(card => {
                const name = card.dataset.name?.toLowerCase() || '';
                const tags = card.dataset.tags?.toLowerCase() || '';
                const domain = card.dataset.domain?.toLowerCase() || '';
                const status = card.dataset.status?.toLowerCase() || 'active';
                
                let matchesQuery = !query || name.includes(query) || tags.includes(query) || domain.includes(query);
                let matchesFilter = filter === 'all' ||
                    (filter === 'active' && (status === 'active' || !status)) ||
                    (filter === 'deprecated' && status === 'deprecated') ||
                    domain.includes(filter);
                
                card.style.display = (matchesQuery && matchesFilter) ? 'block' : 'none';
            });
        }
        
        // Modal
        async function showPattern(name) {
            const modal = document.getElementById('patternModal');
            const title = document.getElementById('modalTitle');
            const content = document.getElementById('modalContent');
            
            modal.classList.add('active');
            title.textContent = name;
            content.innerHTML = 'Loading...';
            
            try {
                const res = await fetch('/api/pattern/' + encodeURIComponent(name));
                const pattern = await res.json();
                
                content.innerHTML = ` + "`" + `
                    <div style="margin-bottom: 1rem;">
                        <strong>Description:</strong><br>
                        ${pattern.description || 'No description'}
                    </div>
                    <div style="margin-bottom: 1rem;">
                        <strong>Domain:</strong> ${pattern.domain || 'general'}<br>
                        <strong>Status:</strong> ${pattern.lifecycle?.status || 'active'}<br>
                        <strong>Effectiveness:</strong> ${((pattern.learning?.effectiveness || 0) * 100).toFixed(0)}%<br>
                        <strong>Usage Count:</strong> ${pattern.learning?.usage_count || 0}
                    </div>
                    <div style="margin-bottom: 1rem;">
                        <strong>Content:</strong>
                        <pre style="background: var(--bg-tertiary); padding: 1rem; border-radius: 0.5rem; overflow-x: auto; margin-top: 0.5rem; font-size: 0.875rem; white-space: pre-wrap;">${escapeHtml(pattern.content || 'No content')}</pre>
                    </div>
                ` + "`" + `;
            } catch (err) {
                content.innerHTML = 'Error loading pattern: ' + err.message;
            }
        }
        
        function closeModal() {
            document.getElementById('patternModal').classList.remove('active');
        }
        
        document.getElementById('patternModal').addEventListener('click', (e) => {
            if (e.target.classList.contains('modal-overlay')) closeModal();
        });
        
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') closeModal();
        });
        
        // Sync
        async function triggerSync() {
            const btn = document.getElementById('syncBtn');
            btn.disabled = true;
            btn.textContent = 'Syncing...';
            
            try {
                const res = await fetch('/api/sync', { method: 'POST' });
                const result = await res.json();
                showToast(result.success ? 'Sync completed!' : 'Sync failed', result.success ? 'success' : 'error');
            } catch (err) {
                showToast('Sync error: ' + err.message, 'error');
            } finally {
                btn.disabled = false;
                btn.textContent = 'Sync Now';
            }
        }
        
        // Toast
        function showToast(message, type = 'success') {
            const toast = document.getElementById('toast');
            const icon = document.getElementById('toastIcon');
            const msg = document.getElementById('toastMessage');
            
            icon.textContent = type === 'success' ? '✓' : '✗';
            msg.textContent = message;
            toast.className = 'toast ' + type + ' show';
            
            setTimeout(() => { toast.classList.remove('show'); }, 3000);
        }
        
        // Utils
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
    </script>
</body>
</html>
`
