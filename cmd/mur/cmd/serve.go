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
	"time"

	"github.com/mur-run/mur-core/internal/core/pattern"
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
  - Pattern browser with search
  - Usage analytics and charts
  - Effectiveness metrics

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
	Patterns       []PatternView
	TotalPatterns  int
	TotalUsage     int
	AvgEffective   float64
	TopPatterns    []PatternView
	RecentPatterns []PatternView
	LastSync       string
	GeneratedAt    string
}

// PatternView is a simplified pattern for display
type PatternView struct {
	Name          string
	Description   string
	Tags          []string
	Effectiveness float64
	UsageCount    int
	LastUsed      string
	CreatedAt     string
	Status        string
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

	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		serveStats(w, r, store)
	})

	// Static assets
	mux.HandleFunc("/static/", serveStatic)

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

func serveStats(w http.ResponseWriter, r *http.Request, store *pattern.Store) {
	patterns, err := store.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := buildDashboardData(patterns)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"totalPatterns": data.TotalPatterns,
		"totalUsage":    data.TotalUsage,
		"avgEffective":  data.AvgEffective,
	})
}

func serveStatic(w http.ResponseWriter, r *http.Request) {
	// For now, inline CSS/JS in the HTML template
	http.NotFound(w, r)
}

func buildDashboardData(patterns []pattern.Pattern) DashboardData {
	data := DashboardData{
		Patterns:    make([]PatternView, 0, len(patterns)),
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
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

	return data
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

	return PatternView{
		Name:          p.Name,
		Description:   p.Description,
		Tags:          tags,
		Effectiveness: p.Learning.Effectiveness,
		UsageCount:    p.Learning.UsageCount,
		LastUsed:      lastUsed,
		CreatedAt:     createdAt,
		Status:        string(p.Lifecycle.Status),
	}
}

func openBrowser(url string) {
	// Try to open browser (best effort)
	var cmd string
	var args []string

	switch {
	case fileExists("/usr/bin/open"): // macOS
		cmd = "open"
		args = []string{url}
	case fileExists("/usr/bin/xdg-open"): // Linux
		cmd = "xdg-open"
		args = []string{url}
	default:
		return
	}

	// Run in background, ignore errors
	go func() {
		execCmd := exec.Command(cmd, args...)
		execCmd.Run()
	}()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Dashboard HTML template
const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>mur Dashboard</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0f172a;
            color: #e2e8f0;
            min-height: 100vh;
        }
        .container { max-width: 1200px; margin: 0 auto; padding: 2rem; }
        
        header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 2rem;
            padding-bottom: 1rem;
            border-bottom: 1px solid #334155;
        }
        .logo {
            font-size: 1.5rem;
            font-weight: bold;
            color: #38bdf8;
        }
        .logo span { color: #e2e8f0; }
        .generated { color: #64748b; font-size: 0.875rem; }
        
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1rem;
            margin-bottom: 2rem;
        }
        .stat-card {
            background: #1e293b;
            border-radius: 0.75rem;
            padding: 1.5rem;
            border: 1px solid #334155;
        }
        .stat-label { color: #94a3b8; font-size: 0.875rem; margin-bottom: 0.5rem; }
        .stat-value { font-size: 2rem; font-weight: bold; color: #38bdf8; }
        .stat-value.green { color: #4ade80; }
        
        .section { margin-bottom: 2rem; }
        .section-title {
            font-size: 1.25rem;
            margin-bottom: 1rem;
            color: #f1f5f9;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        
        .patterns-grid {
            display: grid;
            gap: 1rem;
        }
        .pattern-card {
            background: #1e293b;
            border-radius: 0.75rem;
            padding: 1.25rem;
            border: 1px solid #334155;
            transition: border-color 0.2s;
        }
        .pattern-card:hover { border-color: #38bdf8; }
        .pattern-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 0.75rem;
        }
        .pattern-name {
            font-weight: 600;
            color: #f1f5f9;
            font-size: 1rem;
        }
        .pattern-effectiveness {
            background: #065f46;
            color: #34d399;
            padding: 0.25rem 0.5rem;
            border-radius: 0.375rem;
            font-size: 0.75rem;
            font-weight: 600;
        }
        .pattern-description {
            color: #94a3b8;
            font-size: 0.875rem;
            margin-bottom: 0.75rem;
        }
        .pattern-tags { display: flex; gap: 0.5rem; flex-wrap: wrap; }
        .tag {
            background: #334155;
            color: #94a3b8;
            padding: 0.25rem 0.5rem;
            border-radius: 0.25rem;
            font-size: 0.75rem;
        }
        .pattern-meta {
            margin-top: 0.75rem;
            display: flex;
            gap: 1rem;
            color: #64748b;
            font-size: 0.75rem;
        }
        
        .search-box {
            background: #1e293b;
            border: 1px solid #334155;
            border-radius: 0.5rem;
            padding: 0.75rem 1rem;
            color: #e2e8f0;
            width: 100%;
            font-size: 1rem;
            margin-bottom: 1rem;
        }
        .search-box:focus {
            outline: none;
            border-color: #38bdf8;
        }
        .search-box::placeholder { color: #64748b; }
        
        .empty-state {
            text-align: center;
            padding: 3rem;
            color: #64748b;
        }
        .empty-state-icon { font-size: 3rem; margin-bottom: 1rem; }
        
        footer {
            text-align: center;
            padding: 2rem;
            color: #64748b;
            font-size: 0.875rem;
        }
        footer a { color: #38bdf8; text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <div class="logo">mur<span>.dashboard</span></div>
            <div class="generated">Generated: {{.GeneratedAt}}</div>
        </header>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-label">Total Patterns</div>
                <div class="stat-value">{{.TotalPatterns}}</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Total Usage</div>
                <div class="stat-value">{{.TotalUsage}}</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Avg Effectiveness</div>
                <div class="stat-value green">{{printf "%.0f" .AvgEffective}}%</div>
            </div>
        </div>
        
        {{if .TopPatterns}}
        <div class="section">
            <h2 class="section-title">⭐ Top Patterns</h2>
            <div class="patterns-grid">
                {{range .TopPatterns}}
                <div class="pattern-card">
                    <div class="pattern-header">
                        <span class="pattern-name">{{.Name}}</span>
                        <span class="pattern-effectiveness">{{printf "%.0f" (mul .Effectiveness 100)}}%</span>
                    </div>
                    {{if .Description}}
                    <div class="pattern-description">{{.Description}}</div>
                    {{end}}
                    <div class="pattern-tags">
                        {{range .Tags}}
                        <span class="tag">{{.}}</span>
                        {{end}}
                    </div>
                    <div class="pattern-meta">
                        <span>Used {{.UsageCount}} times</span>
                        <span>Last: {{.LastUsed}}</span>
                    </div>
                </div>
                {{end}}
            </div>
        </div>
        {{end}}
        
        <div class="section">
            <h2 class="section-title">📚 All Patterns</h2>
            <input type="text" class="search-box" placeholder="Search patterns..." id="search">
            
            {{if .Patterns}}
            <div class="patterns-grid" id="patterns-list">
                {{range .Patterns}}
                <div class="pattern-card" data-name="{{.Name}}" data-tags="{{range .Tags}}{{.}} {{end}}">
                    <div class="pattern-header">
                        <span class="pattern-name">{{.Name}}</span>
                        {{if gt .Effectiveness 0}}
                        <span class="pattern-effectiveness">{{printf "%.0f" (mul .Effectiveness 100)}}%</span>
                        {{end}}
                    </div>
                    {{if .Description}}
                    <div class="pattern-description">{{.Description}}</div>
                    {{end}}
                    <div class="pattern-tags">
                        {{range .Tags}}
                        <span class="tag">{{.}}</span>
                        {{end}}
                    </div>
                    <div class="pattern-meta">
                        <span>Used {{.UsageCount}} times</span>
                        {{if .CreatedAt}}<span>Created: {{.CreatedAt}}</span>{{end}}
                    </div>
                </div>
                {{end}}
            </div>
            {{else}}
            <div class="empty-state">
                <div class="empty-state-icon">📭</div>
                <p>No patterns yet</p>
                <p>Run <code>mur learn add</code> to create your first pattern</p>
            </div>
            {{end}}
        </div>
        
        <footer>
            <p>mur — Continuous learning for AI assistants</p>
            <p><a href="https://github.com/mur-run/mur-core">GitHub</a></p>
        </footer>
    </div>
    
    <script>
        // Simple search functionality
        const search = document.getElementById('search');
        const patterns = document.querySelectorAll('.pattern-card');
        
        search?.addEventListener('input', (e) => {
            const query = e.target.value.toLowerCase();
            patterns.forEach(card => {
                const name = card.dataset.name?.toLowerCase() || '';
                const tags = card.dataset.tags?.toLowerCase() || '';
                const visible = name.includes(query) || tags.includes(query);
                card.style.display = visible ? 'block' : 'none';
            });
        });
    </script>
</body>
</html>
`
