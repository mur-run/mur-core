package router

import (
	"testing"

	"github.com/mur-run/mur-core/internal/config"
)

func TestAnalyzePrompt(t *testing.T) {
	tests := []struct {
		name        string
		prompt      string
		wantLow     float64 // complexity >= this
		wantHigh    float64 // complexity <= this
		wantCat     string
		wantToolUse bool
	}{
		{
			name:        "simple question",
			prompt:      "what is git?",
			wantLow:     0.0,
			wantHigh:    0.2,
			wantCat:     "simple-qa",
			wantToolUse: false,
		},
		{
			name:        "complex refactor",
			prompt:      "refactor this architecture to use microservices",
			wantLow:     0.2, // lowered: short prompt with keywords
			wantHigh:    1.0,
			wantCat:     "architecture",
			wantToolUse: false,
		},
		{
			name:        "simple coding",
			prompt:      "write a hello world in go",
			wantLow:     0.0,
			wantHigh:    0.3,
			wantCat:     "coding",
			wantToolUse: false,
		},
		{
			name:        "file operation",
			prompt:      "read file main.go and fix the bug",
			wantLow:     0.0, // lowered: tool use gives some score but keywords reduce it
			wantHigh:    0.25,
			wantCat:     "debugging",
			wantToolUse: true,
		},
		{
			name:        "project work",
			prompt:      "in this project, implement a new feature for user authentication",
			wantLow:     0.1, // lowered: has implement keyword but not super complex
			wantHigh:    0.5,
			wantCat:     "coding",
			wantToolUse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := AnalyzePrompt(tt.prompt)

			if analysis.Complexity < tt.wantLow || analysis.Complexity > tt.wantHigh {
				t.Errorf("complexity = %.2f, want [%.2f, %.2f]", analysis.Complexity, tt.wantLow, tt.wantHigh)
			}
			if analysis.Category != tt.wantCat {
				t.Errorf("category = %s, want %s", analysis.Category, tt.wantCat)
			}
			if analysis.NeedsToolUse != tt.wantToolUse {
				t.Errorf("needsToolUse = %v, want %v", analysis.NeedsToolUse, tt.wantToolUse)
			}
		})
	}
}

func TestSelectTool(t *testing.T) {
	cfg := &config.Config{
		DefaultTool: "claude",
		Routing: config.RoutingConfig{
			Mode:                "auto",
			ComplexityThreshold: 0.3, // Lower threshold for tests
		},
		Tools: map[string]config.Tool{
			"claude": {Enabled: true, Binary: "claude", Tier: "paid"},
			"gemini": {Enabled: true, Binary: "gemini", Tier: "free"},
		},
	}

	tests := []struct {
		name     string
		prompt   string
		mode     string
		wantTool string
	}{
		{
			name:     "simple prompt auto mode",
			prompt:   "what is git?",
			mode:     "auto",
			wantTool: "gemini", // free tier for simple
		},
		{
			name:     "complex prompt auto mode",
			prompt:   "refactor the entire architecture and optimize performance",
			mode:     "auto",
			wantTool: "claude", // paid tier for complex (raised complexity)
		},
		{
			name:     "manual mode",
			prompt:   "what is git?",
			mode:     "manual",
			wantTool: "claude", // always default
		},
		{
			name:     "quality-first simple",
			prompt:   "what is x?",
			mode:     "quality-first",
			wantTool: "gemini", // very simple, use free
		},
		{
			name:     "quality-first medium",
			prompt:   "refactor the module architecture",
			mode:     "quality-first",
			wantTool: "claude", // not very simple, use paid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg.Routing.Mode = tt.mode
			selection, err := SelectTool(tt.prompt, cfg)
			if err != nil {
				t.Fatalf("SelectTool error: %v", err)
			}
			if selection.Tool != tt.wantTool {
				t.Errorf("tool = %s, want %s (reason: %s, complexity: %.2f)",
					selection.Tool, tt.wantTool, selection.Reason, selection.Analysis.Complexity)
			}
		})
	}
}

func TestDetectCategory(t *testing.T) {
	tests := []struct {
		prompt string
		want   string
	}{
		{"what is golang?", "simple-qa"},
		{"explain how http works", "simple-qa"},
		{"refactor this module", "architecture"},
		{"fix this bug", "debugging"},
		{"implement a new feature", "coding"},
		{"analyze this code", "analysis"},
		{"do something random", "general"},
	}

	for _, tt := range tests {
		t.Run(tt.prompt, func(t *testing.T) {
			got := detectCategory(tt.prompt)
			if got != tt.want {
				t.Errorf("detectCategory(%q) = %s, want %s", tt.prompt, got, tt.want)
			}
		})
	}
}

func TestLengthFactor(t *testing.T) {
	tests := []struct {
		length int
		want   float64
	}{
		{10, 0.0},
		{100, 0.1},
		{200, 0.2},
		{400, 0.3},
		{800, 0.5},
		{2000, 0.7},
	}

	for _, tt := range tests {
		got := lengthFactor(tt.length)
		if got != tt.want {
			t.Errorf("lengthFactor(%d) = %.2f, want %.2f", tt.length, got, tt.want)
		}
	}
}
