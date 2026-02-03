# 06: Smart Cost Routing

**Status:** Done  
**Priority:** High  
**Effort:** Medium (3-4 hours)

## Problem

目前 `mur run` 使用固定的 default tool，不考慮任務複雜度或成本。實際上：
- 簡單問題（"what is git?"）用免費 CLI 就夠了
- 複雜任務（architecture, refactoring）需要 Claude 這種強力模型
- 開發者浪費時間手動選擇工具
- 沒有考慮成本效益

需要智慧路由：根據 prompt 複雜度自動選擇最適合的工具。

## Solution

建立 `internal/router` package，用 heuristics 分析 prompt 複雜度，自動選擇工具：
1. **AnalyzePrompt()** — 分析 prompt 複雜度（0-1）
2. **SelectTool()** — 根據複雜度和 config 選擇最佳工具
3. **Config 擴充** — 加入 routing settings 和 tool tiers
4. **CLI 整合** — `mur run` 自動路由，`--explain` 顯示決策

## Implementation

### Files to Create/Modify

```
internal/
  router/
    router.go       # Core routing logic
    analyzer.go     # Prompt complexity analysis
    router_test.go  # Tests
internal/
  config/
    config.go       # Add routing config
cmd/
  mur/
    cmd/
      run.go        # Integrate router
      config.go     # Add routing subcommand
```

### Config Extensions

```go
// internal/config/config.go

type Config struct {
    DefaultTool string          `yaml:"default_tool"`
    Tools       map[string]Tool `yaml:"tools"`
    Routing     RoutingConfig   `yaml:"routing"`  // NEW
    Learning    LearningConfig  `yaml:"learning"`
    MCP         MCPConfig       `yaml:"mcp"`
    Hooks       HooksConfig     `yaml:"hooks"`
}

// RoutingConfig controls automatic tool selection.
type RoutingConfig struct {
    Mode                string  `yaml:"mode"`                 // auto | manual | cost-first | quality-first
    ComplexityThreshold float64 `yaml:"complexity_threshold"` // 0-1, default 0.5
}

// Tool represents configuration for an AI tool.
type Tool struct {
    Enabled      bool     `yaml:"enabled"`
    Binary       string   `yaml:"binary"`
    Flags        []string `yaml:"flags"`
    Tier         string   `yaml:"tier"`         // free | paid
    Capabilities []string `yaml:"capabilities"` // coding, analysis, simple-qa, tool-use
}
```

### Router Package

```go
// internal/router/router.go

package router

import "github.com/karajanchang/murmur-ai/internal/config"

// PromptAnalysis contains the result of analyzing a prompt.
type PromptAnalysis struct {
    Complexity   float64  // 0-1 score
    Length       int      // Character count
    Keywords     []string // Detected complexity keywords
    NeedsToolUse bool     // Appears to need file/code operations
    Category     string   // simple-qa, coding, analysis, architecture
}

// ToolSelection represents the routing decision.
type ToolSelection struct {
    Tool     string   // Selected tool name
    Reason   string   // Human-readable explanation
    Analysis PromptAnalysis
    Fallback string   // Alternative if selected unavailable
}

// SelectTool chooses the best tool for the given prompt.
func SelectTool(prompt string, cfg *config.Config) (*ToolSelection, error)

// GetAvailableTools returns enabled tools sorted by preference for the routing mode.
func GetAvailableTools(cfg *config.Config) []string
```

```go
// internal/router/analyzer.go

package router

// AnalyzePrompt returns a complexity analysis of the prompt.
func AnalyzePrompt(prompt string) PromptAnalysis

// Complexity keywords and their weights
var ComplexityKeywords = map[string]float64{
    // High complexity (0.3 each)
    "refactor":     0.3,
    "architecture": 0.3,
    "design":       0.25,
    "optimize":     0.25,
    "debug":        0.25,
    "fix bug":      0.25,
    
    // Medium complexity (0.15 each)
    "implement":    0.15,
    "create":       0.15,
    "build":        0.15,
    "write":        0.1,
    "code":         0.1,
    
    // Low complexity indicators (reduce score)
    "what is":      -0.1,
    "explain":      -0.05,
    "help":         -0.05,
}

// ToolUseKeywords indicate file/code operations needed
var ToolUseKeywords = []string{
    "file", "read", "write", "edit", "create file",
    "run", "execute", "test", "build",
    "in this project", "in this repo", "this code",
}
```

### Routing Logic

```go
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
        // Prefer free tools unless complexity is very high
        selected = selectByTier(available, cfg, "free")
        if analysis.Complexity > 0.8 {
            selected = selectByTier(available, cfg, "paid")
            reason = fmt.Sprintf("cost-first: complexity %.2f > 0.8, using paid tool", analysis.Complexity)
        } else {
            reason = fmt.Sprintf("cost-first: complexity %.2f, using free tool", analysis.Complexity)
        }
        
    case "quality-first":
        // Prefer paid/powerful tools unless very simple
        selected = selectByTier(available, cfg, "paid")
        if analysis.Complexity < 0.2 && !analysis.NeedsToolUse {
            selected = selectByTier(available, cfg, "free")
            reason = fmt.Sprintf("quality-first: complexity %.2f < 0.2, using free tool", analysis.Complexity)
        } else {
            reason = fmt.Sprintf("quality-first: complexity %.2f, using paid tool", analysis.Complexity)
        }
        
    default: // "auto"
        if analysis.Complexity >= threshold || analysis.NeedsToolUse {
            selected = selectByTier(available, cfg, "paid")
            reason = fmt.Sprintf("auto: complexity %.2f >= %.2f or needs tool use", analysis.Complexity, threshold)
        } else {
            selected = selectByTier(available, cfg, "free")
            reason = fmt.Sprintf("auto: complexity %.2f < %.2f, using free tool", analysis.Complexity, threshold)
        }
    }
    
    // Fallback if selected tool unavailable
    if selected == "" {
        selected = available[0]
        reason += " (fallback)"
    }
    
    return &ToolSelection{
        Tool:     selected,
        Reason:   reason,
        Analysis: analysis,
        Fallback: findFallback(selected, available),
    }, nil
}

func selectByTier(available []string, cfg *config.Config, tier string) string {
    for _, name := range available {
        if tool, ok := cfg.GetTool(name); ok {
            if tool.Tier == tier {
                return name
            }
        }
    }
    // Fallback to any available
    if len(available) > 0 {
        return available[0]
    }
    return ""
}
```

### CLI Integration

```go
// cmd/mur/cmd/run.go

var runCmd = &cobra.Command{
    Use:   "run",
    Short: "Run a prompt with an AI tool",
    Long: `Run a prompt using the configured AI tool.

With routing.mode=auto (default), murmur selects the best tool based on
prompt complexity. Simple questions use free tools; complex tasks use paid.

Use -t to override automatic selection.

Examples:
  mur run -p "what is git?"              # Auto-routes to free tool
  mur run -p "refactor this module"      # Auto-routes to paid tool
  mur run -p "explain x" -t claude       # Force specific tool
  mur run -p "test" --explain            # Show routing decision only`,
    RunE: runExecute,
}

func init() {
    rootCmd.AddCommand(runCmd)
    runCmd.Flags().StringP("prompt", "p", "", "The prompt to run")
    runCmd.Flags().StringP("tool", "t", "", "Force specific tool (overrides routing)")
    runCmd.Flags().Bool("explain", false, "Show routing decision without executing")
}

func runExecute(cmd *cobra.Command, args []string) error {
    prompt, _ := cmd.Flags().GetString("prompt")
    forceTool, _ := cmd.Flags().GetString("tool")
    explain, _ := cmd.Flags().GetBool("explain")
    
    cfg, err := config.Load()
    if err != nil {
        return err
    }
    
    var tool string
    var reason string
    
    if forceTool != "" {
        // User explicitly chose a tool
        tool = forceTool
        reason = "user specified with -t flag"
    } else {
        // Use router
        selection, err := router.SelectTool(prompt, cfg)
        if err != nil {
            return err
        }
        tool = selection.Tool
        reason = selection.Reason
        
        if explain {
            // Show decision and exit
            fmt.Printf("Routing Decision\n")
            fmt.Printf("================\n")
            fmt.Printf("Prompt:     %s\n", truncate(prompt, 50))
            fmt.Printf("Complexity: %.2f\n", selection.Analysis.Complexity)
            fmt.Printf("Category:   %s\n", selection.Analysis.Category)
            fmt.Printf("Tool Use:   %v\n", selection.Analysis.NeedsToolUse)
            fmt.Printf("Keywords:   %v\n", selection.Analysis.Keywords)
            fmt.Printf("\nSelected:   %s\n", selection.Tool)
            fmt.Printf("Reason:     %s\n", selection.Reason)
            if selection.Fallback != "" {
                fmt.Printf("Fallback:   %s\n", selection.Fallback)
            }
            return nil
        }
    }
    
    // Validate tool
    if err := cfg.EnsureTool(tool); err != nil {
        return err
    }
    
    toolCfg, _ := cfg.GetTool(tool)
    
    fmt.Printf("→ %s (%s)\n\n", tool, reason)
    
    // Execute...
    return executeWithTool(toolCfg, prompt)
}
```

### Routing Config Command

```go
// cmd/mur/cmd/config.go (additions)

var configRoutingCmd = &cobra.Command{
    Use:   "routing [mode]",
    Short: "Set routing mode",
    Long: `Set the automatic routing mode.

Modes:
  auto          Smart routing based on complexity (default)
  manual        Always use default_tool
  cost-first    Prefer free tools unless very complex
  quality-first Prefer paid tools unless very simple

Examples:
  mur config routing auto
  mur config routing cost-first`,
    Args: cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        mode := args[0]
        validModes := []string{"auto", "manual", "cost-first", "quality-first"}
        
        valid := false
        for _, m := range validModes {
            if mode == m {
                valid = true
                break
            }
        }
        if !valid {
            return fmt.Errorf("invalid mode: %s. Valid: %v", mode, validModes)
        }
        
        cfg, err := config.Load()
        if err != nil {
            return err
        }
        
        cfg.Routing.Mode = mode
        if err := cfg.Save(); err != nil {
            return err
        }
        
        fmt.Printf("✓ Routing mode set to: %s\n", mode)
        return nil
    },
}

func init() {
    configCmd.AddCommand(configRoutingCmd)
}
```

## Complexity Heuristics

### Length Factor
```go
func lengthFactor(prompt string) float64 {
    length := len(prompt)
    switch {
    case length < 50:
        return 0.0   // Very short = simple
    case length < 150:
        return 0.1
    case length < 300:
        return 0.2
    case length < 500:
        return 0.3
    default:
        return 0.4   // Long prompts tend to be complex
    }
}
```

### Category Detection
```go
func detectCategory(prompt string) string {
    lower := strings.ToLower(prompt)
    
    if containsAny(lower, []string{"what is", "explain", "tell me about", "how does"}) {
        return "simple-qa"
    }
    if containsAny(lower, []string{"refactor", "architecture", "design", "optimize"}) {
        return "architecture"
    }
    if containsAny(lower, []string{"debug", "fix", "error", "bug", "issue"}) {
        return "debugging"
    }
    if containsAny(lower, []string{"write", "create", "implement", "build", "code"}) {
        return "coding"
    }
    if containsAny(lower, []string{"analyze", "review", "evaluate", "compare"}) {
        return "analysis"
    }
    
    return "general"
}
```

## Default Config

```yaml
# ~/.murmur/config.yaml

default_tool: claude

routing:
  mode: auto
  complexity_threshold: 0.5

tools:
  claude:
    enabled: true
    binary: claude
    flags: ["-p"]
    tier: paid
    capabilities: [coding, analysis, tool-use, architecture]
  
  gemini:
    enabled: true
    binary: gemini
    flags: ["-p"]
    tier: free
    capabilities: [coding, simple-qa, analysis]
  
  auggie:
    enabled: false
    binary: auggie
    flags: []
    tier: free
    capabilities: [simple-qa]
```

## Tests

```go
// internal/router/router_test.go
func TestAnalyzePrompt(t *testing.T) {
    tests := []struct {
        prompt     string
        wantLow    float64 // complexity >= this
        wantHigh   float64 // complexity <= this
    }{
        {"what is git?", 0.0, 0.3},
        {"refactor this architecture", 0.5, 1.0},
        {"write a hello world", 0.1, 0.4},
    }
    // ...
}

func TestSelectTool(t *testing.T)
func TestSelectByTier(t *testing.T)
func TestDetectCategory(t *testing.T)
```

## Acceptance Criteria

- [x] `go build ./...` 無 warning
- [x] `internal/router` package 完成
- [x] Config 支援 routing.mode, routing.complexity_threshold
- [x] Config 支援 tools[].tier, tools[].capabilities
- [x] `mur run -p "..."` 自動路由（routing.mode=auto）
- [x] `mur run -p "..." -t claude` 強制指定工具
- [x] `mur run -p "..." --explain` 顯示路由決策
- [x] `mur config routing <mode>` 設定路由模式
- [x] 簡單 prompt 路由到 free tier
- [x] 複雜 prompt 路由到 paid tier

## Future Enhancements

- Usage tracking: 記錄每個工具的使用次數和成本
- Feedback loop: 根據結果品質調整路由
- Per-project routing: 不同專案不同設定
- LLM-based analysis: 用 AI 分析複雜度（可選）

## Dependencies

- 無新增依賴

## Related

- `internal/config/config.go` — Config structure
- `cmd/mur/cmd/run.go` — Run command
- `cmd/mur/cmd/config.go` — Config commands
