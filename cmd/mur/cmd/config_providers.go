package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var configProvidersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Show available LLM providers and recommended models",
	Long: `Show available LLM providers with model recommendations for
mur session analyze and learn extract commands.

Includes quality ratings, speed estimates, and cost per session.
Detects system RAM to highlight which local models will work on your machine.`,
	RunE: runConfigProviders,
}

func init() {
	configCmd.AddCommand(configProvidersCmd)
}

func runConfigProviders(cmd *cobra.Command, args []string) error {
	ramGB := detectSystemRAM()

	fmt.Println("LLM Providers for mur session analyze & learn extract")
	fmt.Println()

	// Paid models
	fmt.Println("PAID MODELS")
	fmt.Println("  Provider    Model                    Quality  Speed   Cost/session  Notes")
	fmt.Println("  ─────────────────────────────────────────────────────────────────────────────")
	fmt.Println("  anthropic   claude-sonnet-4-20250514 ★★★★★  Fast    ~$0.02        Best for pattern extraction")
	fmt.Println("  anthropic   claude-haiku-3.5          ★★★★   V.Fast  ~$0.003       Good quality, 7x cheaper")
	fmt.Println("  openai      gpt-4o                   ★★★★½  Fast    ~$0.02        Reliable")
	fmt.Println("  openai      gpt-4o-mini              ★★★½   V.Fast  ~$0.002       Budget option")
	fmt.Println("  gemini      gemini-2.5-pro           ★★★★½  Fast    Free tier     Generous free quota")
	fmt.Println("  gemini      gemini-2.5-flash         ★★★★   V.Fast  ~Free         Large free tier")
	fmt.Println("  openai*     deepseek-chat            ★★★★   Fast    ~$0.002       Via openai_url: https://api.deepseek.com/v1")
	fmt.Println()

	// Local models with RAM indicators
	fmt.Println("FREE LOCAL MODELS (Ollama)")
	fmt.Println("  Model               Quality  Speed   VRAM    Min RAM   Notes")
	fmt.Println("  ─────────────────────────────────────────────────────────────────────────────")

	type localModel struct {
		name    string
		quality string
		speed   string
		vram    string
		minRAM  int // GB
		notes   string
	}

	models := []localModel{
		{"qwen3:8b", "★★★★  ", "Medium", "~5GB ", 16, "Best free default"},
		{"qwen3:14b", "★★★★½ ", "Slow  ", "~9GB ", 16, "Near paid quality"},
		{"qwen3:32b", "★★★★★ ", "V.Slow", "~20GB", 32, "Best free model"},
		{"gemma3:12b", "★★★★  ", "Medium", "~8GB ", 16, "Good JSON output"},
		{"deepseek-r1:14b", "★★★★  ", "Slow  ", "~9GB ", 16, "Strong reasoning"},
		{"deepseek-r1:32b", "★★★★½ ", "V.Slow", "~20GB", 32, "Deep reasoning"},
		{"llama3.2:3b", "★★½   ", "V.Fast", "~2GB ", 8, "Fast but misses patterns"},
		{"llama3.1:8b", "★★★   ", "Fast  ", "~5GB ", 16, "Decent baseline"},
		{"mistral-small:24b", "★★★★  ", "Slow  ", "~15GB", 32, "Good multilingual"},
		{"phi-4:14b", "★★★½  ", "Medium", "~9GB ", 16, "Good code understanding"},
	}

	for _, m := range models {
		indicator := " "
		if ramGB > 0 {
			if ramGB >= m.minRAM {
				indicator = "✓"
			} else {
				indicator = "✗"
			}
		}
		fmt.Printf("  %s %-19s %s  %s  %s  %dGB+     %s\n",
			indicator, m.name, m.quality, m.speed, m.vram, m.minRAM, m.notes)
	}

	if ramGB > 0 {
		fmt.Printf("\n  System RAM: %dGB (✓ = compatible, ✗ = may not fit)\n", ramGB)
	}

	fmt.Println()
	fmt.Println("CONFIGURATION")
	fmt.Println("  # In ~/.mur/config.yaml:")
	fmt.Println("  learning:")
	fmt.Println("    llm:")
	fmt.Println("      provider: claude              # anthropic | openai | ollama | gemini")
	fmt.Println("      model: claude-sonnet-4-20250514")
	fmt.Println("      # ollama_url: http://localhost:11434")
	fmt.Println("      # openai_url: https://api.deepseek.com/v1  # for OpenAI-compatible APIs")
	fmt.Println()
	fmt.Println("  # Or override per-command:")
	fmt.Println("  mur session analyze <id> --provider ollama --model qwen3:8b")
	fmt.Println()

	return nil
}

// detectSystemRAM returns total system RAM in GB, or 0 if detection fails.
func detectSystemRAM() int {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
		if err != nil {
			return 0
		}
		bytes, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
		if err != nil {
			return 0
		}
		return int(bytes / (1024 * 1024 * 1024))
	case "linux":
		out, err := exec.Command("grep", "MemTotal", "/proc/meminfo").Output()
		if err != nil {
			return 0
		}
		// Format: "MemTotal:       16384000 kB"
		fields := strings.Fields(string(out))
		if len(fields) >= 2 {
			kb, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0
			}
			return int(kb / (1024 * 1024))
		}
		return 0
	default:
		return 0
	}
}
