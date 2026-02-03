package router

import (
	"fmt"

	"github.com/karajanchang/murmur-ai/internal/config"
)

// ToolSelection represents the routing decision.
type ToolSelection struct {
	Tool     string         // Selected tool name
	Reason   string         // Human-readable explanation
	Analysis PromptAnalysis // The prompt analysis that led to this decision
	Fallback string         // Alternative tool if selected unavailable
}

// SelectTool chooses the best tool for the given prompt based on config.
func SelectTool(prompt string, cfg *config.Config) (*ToolSelection, error) {
	analysis := AnalyzePrompt(prompt)

	mode := cfg.Routing.Mode
	if mode == "" {
		mode = "auto"
	}

	threshold := cfg.Routing.ComplexityThreshold
	if threshold == 0 {
		threshold = 0.5
	}

	// Get available tools
	available := GetAvailableTools(cfg)
	if len(available) == 0 {
		return nil, fmt.Errorf("no enabled tools available")
	}

	var selected string
	var reason string

	switch mode {
	case "manual":
		// Use default_tool always
		selected = cfg.GetDefaultTool()
		reason = "manual mode: using default tool"

	case "cost-first":
		// Prefer free tools unless complexity is very high (>0.8)
		if analysis.Complexity > 0.8 || analysis.NeedsToolUse {
			selected = selectByTier(available, cfg, "paid")
			if selected != "" {
				reason = fmt.Sprintf("cost-first: complexity %.2f > 0.8 or needs tool use, using paid tool", analysis.Complexity)
			}
		}
		if selected == "" {
			selected = selectByTier(available, cfg, "free")
			reason = fmt.Sprintf("cost-first: complexity %.2f, using free tool", analysis.Complexity)
		}

	case "quality-first":
		// Prefer paid/powerful tools unless very simple (<0.2)
		if analysis.Complexity < 0.2 && !analysis.NeedsToolUse {
			selected = selectByTier(available, cfg, "free")
			if selected != "" {
				reason = fmt.Sprintf("quality-first: complexity %.2f < 0.2 and no tool use needed, using free tool", analysis.Complexity)
			}
		}
		if selected == "" {
			selected = selectByTier(available, cfg, "paid")
			reason = fmt.Sprintf("quality-first: complexity %.2f, using paid tool", analysis.Complexity)
		}

	default: // "auto"
		if analysis.Complexity >= threshold || analysis.NeedsToolUse {
			selected = selectByTier(available, cfg, "paid")
			if selected != "" {
				if analysis.NeedsToolUse {
					reason = fmt.Sprintf("auto: needs tool use, using paid tool")
				} else {
					reason = fmt.Sprintf("auto: complexity %.2f >= %.2f threshold, using paid tool", analysis.Complexity, threshold)
				}
			}
		}
		if selected == "" {
			selected = selectByTier(available, cfg, "free")
			if selected != "" {
				reason = fmt.Sprintf("auto: complexity %.2f < %.2f threshold, using free tool", analysis.Complexity, threshold)
			}
		}
	}

	// Final fallback to any available tool
	if selected == "" && len(available) > 0 {
		selected = available[0]
		reason = fmt.Sprintf("%s (fallback to first available)", reason)
	}

	if selected == "" {
		return nil, fmt.Errorf("no suitable tool found")
	}

	return &ToolSelection{
		Tool:     selected,
		Reason:   reason,
		Analysis: analysis,
		Fallback: findFallback(selected, available),
	}, nil
}

// GetAvailableTools returns enabled tools from config.
func GetAvailableTools(cfg *config.Config) []string {
	var tools []string
	for name, tool := range cfg.Tools {
		if tool.Enabled {
			tools = append(tools, name)
		}
	}
	return tools
}

// selectByTier returns the first available tool matching the given tier.
func selectByTier(available []string, cfg *config.Config, tier string) string {
	for _, name := range available {
		if tool, ok := cfg.GetTool(name); ok {
			if tool.Tier == tier {
				return name
			}
		}
	}
	return ""
}

// findFallback returns an alternative tool if the selected one fails.
func findFallback(selected string, available []string) string {
	for _, name := range available {
		if name != selected {
			return name
		}
	}
	return ""
}
