package learn

import "testing"

func TestExtractJSONPatterns(t *testing.T) {
	// Test JSON array in code block
	input := `Here are the patterns:

` + "```json" + `
[
  {
    "name": "sparkle-xpc-bootstrap-hang",
    "title": "Sparkle Updater XPC Bootstrap Hang",
    "confidence": "HIGH",
    "score": 0.85,
    "category": "debug",
    "domain": "dev",
    "problem": "App hangs on startup when using Sparkle framework",
    "solution": "Initialize SparkleUpdater with startingUpdater: false"
  },
  {
    "name": "menubarextra-sheet-overlay",
    "title": "MenuBarExtra Cannot Use Sheet",
    "confidence": "MEDIUM",
    "score": 0.7,
    "category": "pattern",
    "domain": "mobile",
    "problem": "SwiftUI .sheet() does not work in MenuBarExtra",
    "solution": "Use ZStack overlay pattern instead"
  }
]
` + "```" + `

That's all.`

	patterns := extractJSONPatterns(input, "test-session")

	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d", len(patterns))
	}

	// Check first pattern
	p1 := patterns[0]
	if p1.Pattern.Name != "sparkle-xpc-bootstrap-hang" {
		t.Errorf("expected name 'sparkle-xpc-bootstrap-hang', got '%s'", p1.Pattern.Name)
	}
	if p1.Confidence != 0.85 {
		t.Errorf("expected confidence 0.85, got %f", p1.Confidence)
	}
	if p1.Pattern.Category != "debug" {
		t.Errorf("expected category 'debug', got '%s'", p1.Pattern.Category)
	}

	// Check second pattern
	p2 := patterns[1]
	if p2.Pattern.Name != "menubarextra-sheet-overlay" {
		t.Errorf("expected name 'menubarextra-sheet-overlay', got '%s'", p2.Pattern.Name)
	}
	// Domain should be normalized from "mobile" to "dev"
	if p2.Pattern.Domain != "dev" {
		t.Errorf("expected domain 'dev', got '%s'", p2.Pattern.Domain)
	}
}

func TestIsValidPatternName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"sparkle-xpc-bootstrap-hang", true},
		{"menubarextra-sheet-overlay", true},
		{"go-cli-detection", true},
		{"now-need-update", false},
		{"wait-changed-line", false},
		{"now-also-need", false},
		{"see-unloadplist-calls", false},
		{"abc", false}, // too short
		{"ab", false},  // too short
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidPatternName(tt.name)
			if got != tt.valid {
				t.Errorf("isValidPatternName(%q) = %v, want %v", tt.name, got, tt.valid)
			}
		})
	}
}
