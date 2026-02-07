package cmd

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/mur-run/mur-core/internal/core/pattern"
	"github.com/mur-run/mur-core/internal/stats"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Generate static HTML dashboard",
	Long: `Generate a standalone HTML file with pattern analytics.

The output is a single HTML file that can be viewed in any browser
without needing to run a server.

Examples:
  mur dashboard                    # Output to stdout
  mur dashboard > report.html      # Save to file
  mur dashboard -o report.html     # Save to file
  mur dashboard --open             # Generate and open in browser`,
	RunE: runDashboard,
}

var (
	dashboardOutput string
	dashboardOpen   bool
)

func init() {
	rootCmd.AddCommand(dashboardCmd)
	dashboardCmd.Flags().StringVarP(&dashboardOutput, "output", "o", "", "Output file path")
	dashboardCmd.Flags().BoolVar(&dashboardOpen, "open", false, "Open in browser after generating")
}

func runDashboard(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	patternsDir := filepath.Join(home, ".mur", "patterns")
	store := pattern.NewStore(patternsDir)
	patterns, err := store.List()
	if err != nil {
		return fmt.Errorf("failed to load patterns: %w", err)
	}

	data := buildStaticDashboardData(patterns)

	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
		"sub": func(a, b float64) float64 { return a - b },
		"gt0": func(a float64) bool { return a > 0 },
		"lt":  func(a, b float64) bool { return a < b },
		"printf": fmt.Sprintf,
	}

	tmpl := template.Must(template.New("dashboard").Funcs(funcMap).Parse(staticDashboardHTML))

	// Determine output
	var output *os.File
	if dashboardOutput != "" {
		output, err = os.Create(dashboardOutput)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer output.Close()
	} else if dashboardOpen {
		// Create temp file
		output, err = os.CreateTemp("", "mur-dashboard-*.html")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		dashboardOutput = output.Name()
		defer output.Close()
	} else {
		output = os.Stdout
	}

	if err := tmpl.Execute(output, data); err != nil {
		return fmt.Errorf("failed to generate dashboard: %w", err)
	}

	if dashboardOutput != "" {
		fmt.Fprintf(os.Stderr, "Dashboard saved to: %s\n", dashboardOutput)
	}

	if dashboardOpen && dashboardOutput != "" {
		openBrowser("file://" + dashboardOutput)
	}

	return nil
}

func buildStaticDashboardData(patterns []pattern.Pattern) DashboardData {
	data := DashboardData{
		Patterns:    make([]PatternView, 0, len(patterns)),
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Version:     Version,
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

	// Sort for top patterns
	sorted := make([]PatternView, len(data.Patterns))
	copy(sorted, data.Patterns)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Effectiveness > sorted[j].Effectiveness
	})
	if len(sorted) > 10 {
		sorted = sorted[:10]
	}
	data.TopPatterns = sorted

	// Load usage stats
	records, _ := stats.Query(stats.QueryFilter{})
	if len(records) > 0 {
		summary := stats.Summarize(records)
		data.TotalRuns = summary.TotalRuns
		data.EstimatedCost = summary.EstimatedCost
		data.EstimatedSaved = summary.EstimatedSaved

		for _, d := range summary.DailyTrend {
			date, _ := time.Parse("2006-01-02", d.Date)
			data.DailyTrend = append(data.DailyTrend, DailyPoint{
				Date:  d.Date,
				Day:   date.Format("Mon"),
				Count: d.Count,
			})
		}

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
		sort.Slice(data.ToolBreakdown, func(i, j int) bool {
			return data.ToolBreakdown[i].Count > data.ToolBreakdown[j].Count
		})
	}

	return data
}

// Static dashboard HTML (simplified, no server features)
const staticDashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>mur Dashboard Report</title>
    <style>
        :root {
            --bg: #0f172a;
            --bg2: #1e293b;
            --bg3: #334155;
            --text: #f1f5f9;
            --text2: #94a3b8;
            --muted: #64748b;
            --accent: #38bdf8;
            --green: #4ade80;
            --yellow: #fbbf24;
            --border: #334155;
        }
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--bg);
            color: var(--text);
            padding: 2rem;
            line-height: 1.6;
        }
        .container { max-width: 1200px; margin: 0 auto; }
        header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 2rem;
            padding-bottom: 1rem;
            border-bottom: 1px solid var(--border);
        }
        .logo { font-size: 1.5rem; font-weight: bold; color: var(--accent); }
        .logo span { color: var(--text); }
        .meta { color: var(--muted); font-size: 0.875rem; }
        
        .grid { display: grid; gap: 1.5rem; margin-bottom: 2rem; }
        .grid-4 { grid-template-columns: repeat(4, 1fr); }
        .grid-2 { grid-template-columns: repeat(2, 1fr); }
        @media (max-width: 900px) { .grid-4, .grid-2 { grid-template-columns: 1fr; } }
        
        .card {
            background: var(--bg2);
            border-radius: 0.75rem;
            padding: 1.5rem;
            border: 1px solid var(--border);
        }
        .card-title {
            font-size: 0.75rem;
            color: var(--text2);
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-bottom: 0.5rem;
        }
        .stat { font-size: 2.5rem; font-weight: bold; color: var(--accent); }
        .stat.green { color: var(--green); }
        .stat.yellow { color: var(--yellow); }
        .stat-sub { font-size: 0.875rem; color: var(--text2); margin-top: 0.25rem; }
        
        h2 {
            font-size: 1.25rem;
            margin-bottom: 1rem;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        
        .patterns { display: grid; gap: 1rem; }
        .pattern {
            background: var(--bg2);
            border-radius: 0.75rem;
            padding: 1.25rem;
            border: 1px solid var(--border);
        }
        .pattern-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 0.5rem;
        }
        .pattern-name { font-weight: 600; }
        .pattern-eff {
            background: #065f46;
            color: var(--green);
            padding: 0.25rem 0.5rem;
            border-radius: 0.25rem;
            font-size: 0.75rem;
            font-weight: 600;
        }
        .pattern-eff.low { background: #78350f; color: var(--yellow); }
        .pattern-desc { color: var(--text2); font-size: 0.875rem; margin-bottom: 0.5rem; }
        .pattern-tags { display: flex; gap: 0.5rem; flex-wrap: wrap; }
        .tag {
            background: var(--bg3);
            color: var(--text2);
            padding: 0.25rem 0.5rem;
            border-radius: 0.25rem;
            font-size: 0.75rem;
        }
        .tag.domain { background: rgba(56, 189, 248, 0.2); color: var(--accent); }
        .pattern-meta {
            margin-top: 0.5rem;
            color: var(--muted);
            font-size: 0.75rem;
            display: flex;
            gap: 1rem;
        }
        
        .bar-chart { display: flex; flex-direction: column; gap: 0.75rem; }
        .bar-item { display: flex; align-items: center; gap: 1rem; }
        .bar-label { width: 80px; font-size: 0.875rem; color: var(--text2); }
        .bar-container {
            flex: 1;
            height: 24px;
            background: var(--bg3);
            border-radius: 0.25rem;
            overflow: hidden;
        }
        .bar-fill {
            height: 100%;
            background: linear-gradient(90deg, var(--accent), #818cf8);
            border-radius: 0.25rem;
        }
        .bar-fill.free { background: linear-gradient(90deg, var(--green), #22d3ee); }
        .bar-value { width: 50px; text-align: right; font-size: 0.875rem; color: var(--text2); }
        
        footer {
            text-align: center;
            padding: 2rem;
            color: var(--muted);
            font-size: 0.875rem;
        }
        footer a { color: var(--accent); text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <div class="logo">mur<span>.report</span></div>
            <div class="meta">Generated: {{.GeneratedAt}}</div>
        </header>
        
        <div class="grid grid-4">
            <div class="card">
                <div class="card-title">Patterns</div>
                <div class="stat">{{.TotalPatterns}}</div>
                <div class="stat-sub">{{.ActivePatterns}} active</div>
            </div>
            <div class="card">
                <div class="card-title">Total Usage</div>
                <div class="stat">{{.TotalUsage}}</div>
                <div class="stat-sub">injections</div>
            </div>
            <div class="card">
                <div class="card-title">Effectiveness</div>
                <div class="stat green">{{printf "%.0f" .AvgEffective}}%</div>
                <div class="stat-sub">average</div>
            </div>
            <div class="card">
                <div class="card-title">Saved</div>
                <div class="stat yellow">${{printf "%.2f" .EstimatedSaved}}</div>
                <div class="stat-sub">via free tools</div>
            </div>
        </div>
        
        {{if .ToolBreakdown}}
        <div class="card" style="margin-bottom: 2rem;">
            <h2>🔧 Tool Usage</h2>
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
        </div>
        {{end}}
        
        {{if .TopPatterns}}
        <h2>⭐ Top Patterns</h2>
        <div class="patterns" style="margin-bottom: 2rem;">
            {{range .TopPatterns}}
            <div class="pattern">
                <div class="pattern-header">
                    <span class="pattern-name">{{.Name}}</span>
                    {{if gt0 .Effectiveness}}
                    <span class="pattern-eff {{if lt .Effectiveness 0.5}}low{{end}}">{{printf "%.0f" (mul .Effectiveness 100)}}%</span>
                    {{end}}
                </div>
                {{if .Description}}<div class="pattern-desc">{{.Description}}</div>{{end}}
                <div class="pattern-tags">
                    {{if .Domain}}<span class="tag domain">{{.Domain}}</span>{{end}}
                    {{range .Tags}}<span class="tag">{{.}}</span>{{end}}
                </div>
                <div class="pattern-meta">
                    <span>📊 {{.UsageCount}} uses</span>
                    <span>🕐 {{.LastUsed}}</span>
                </div>
            </div>
            {{end}}
        </div>
        {{end}}
        
        {{if .Patterns}}
        <h2>📚 All Patterns ({{.TotalPatterns}})</h2>
        <div class="patterns">
            {{range .Patterns}}
            <div class="pattern">
                <div class="pattern-header">
                    <span class="pattern-name">{{.Name}}</span>
                    {{if gt0 .Effectiveness}}
                    <span class="pattern-eff {{if lt .Effectiveness 0.5}}low{{end}}">{{printf "%.0f" (mul .Effectiveness 100)}}%</span>
                    {{end}}
                </div>
                {{if .Description}}<div class="pattern-desc">{{.Description}}</div>{{end}}
                <div class="pattern-tags">
                    {{if .Domain}}<span class="tag domain">{{.Domain}}</span>{{end}}
                    {{range .Tags}}<span class="tag">{{.}}</span>{{end}}
                </div>
                <div class="pattern-meta">
                    <span>📊 {{.UsageCount}} uses</span>
                    {{if .CreatedAt}}<span>📅 {{.CreatedAt}}</span>{{end}}
                </div>
            </div>
            {{end}}
        </div>
        {{else}}
        <div class="card" style="text-align: center; padding: 3rem;">
            <p style="font-size: 2rem; margin-bottom: 1rem;">📭</p>
            <p>No patterns yet</p>
            <p style="color: var(--muted); margin-top: 0.5rem;">Run <code>mur learn add</code> to create your first pattern</p>
        </div>
        {{end}}
        
        <footer>
            <p>mur — Continuous learning for AI assistants</p>
            <p style="margin-top: 0.5rem;"><a href="https://github.com/mur-run/mur-core">GitHub</a></p>
        </footer>
    </div>
</body>
</html>
`
