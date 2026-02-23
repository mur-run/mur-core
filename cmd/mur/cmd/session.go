package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/mur-run/mur-core/internal/config"
	"github.com/mur-run/mur-core/internal/session"
	"github.com/mur-run/mur-core/internal/session/ui"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage conversation recording sessions",
	Long: `Record conversation segments for workflow extraction.

Use /mur:in and /mur:out in Claude Code (or other AI tools) to mark
conversation segments for recording. MUR captures all events between
markers, analyzes them with an LLM, and lets you edit and export
the extracted workflow as a reusable skill.

Commands:
  mur session start          Start recording events
  mur session stop           Stop recording (optionally analyze)
  mur session status         Show recording indicator
  mur session list           List past recordings
  mur session analyze <id>   Run LLM analysis on a recording
  mur session ui <id>        Open interactive workflow editor
  mur session export <id>    Export workflow as skill/YAML/markdown

Typical flow:
  1. mur session start --source claude-code
  2. (conversation happens, events are captured)
  3. mur session stop --analyze --open
  4. (edit workflow in web UI, click Save)

Or manually:
  1. mur session analyze <session-id>
  2. mur session export <session-id> --format skill`,
}

var sessionStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start recording a conversation segment",
	Long: `Begin recording conversation events for later workflow extraction.

If a recording session is already active, it will be stopped automatically
before starting a new one. Events are stored as JSONL in ~/.mur/session/recordings/.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		source, _ := cmd.Flags().GetString("source")
		marker, _ := cmd.Flags().GetString("marker")

		if source == "" {
			source = "cli"
		}

		state, err := session.StartRecording(source, marker)
		if err != nil {
			return fmt.Errorf("failed to start recording: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Recording started (session: %s)\n", state.SessionID[:8])
		return nil
	},
}

var sessionStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop recording and optionally analyze",
	Long: `Stop the active recording session.

With --analyze, runs LLM analysis on the captured transcript to extract
a structured workflow. With --open, also launches the interactive web
editor to refine the workflow before saving.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		analyze, _ := cmd.Flags().GetBool("analyze")
		openUI, _ := cmd.Flags().GetBool("open")

		state, err := session.StopRecording()
		if err != nil {
			return err
		}

		events, _ := session.ReadEvents(state.SessionID)
		duration := time.Since(time.Unix(state.StartedAt, 0)).Truncate(time.Second)

		fmt.Fprintf(os.Stderr, "Recording stopped (session: %s)\n", state.SessionID[:8])
		fmt.Fprintf(os.Stderr, "  Duration: %s\n", duration)
		fmt.Fprintf(os.Stderr, "  Events:   %d\n", len(events))

		if analyze {
			llmProvider, _ := cmd.Flags().GetString("provider")
			llmModel, _ := cmd.Flags().GetString("model")
			llmOllamaURL, _ := cmd.Flags().GetString("ollama-url")

			result, err := runAnalysis(state.SessionID, llmProvider, llmModel, llmOllamaURL)
			if err != nil {
				return err
			}

			if openUI && result != nil {
				srv, err := ui.NewServer(result, state.SessionID)
				if err != nil {
					return err
				}
				return srv.Serve(3939)
			}
		}

		return nil
	},
}

var sessionStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current recording status",
	Long: `Display the current recording state.

With --quiet, exits 0 if recording is active, 1 if not.
This is useful in hook scripts to check state programmatically.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		quiet, _ := cmd.Flags().GetBool("quiet")

		state, err := session.GetState()
		if err != nil {
			if quiet {
				os.Exit(1)
			}
			return fmt.Errorf("cannot read state: %w", err)
		}

		if state == nil || !state.Active {
			if quiet {
				os.Exit(1)
			}
			fmt.Println("No active recording session.")
			return nil
		}

		if quiet {
			// Exit 0 = recording active (for use in hook scripts)
			os.Exit(0)
		}

		duration := time.Since(time.Unix(state.StartedAt, 0)).Truncate(time.Second)
		events, _ := session.ReadEvents(state.SessionID)

		shortID := state.SessionID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		fmt.Printf("REC  Recording active\n")
		fmt.Printf("  Session:  %s\n", shortID)
		fmt.Printf("  Source:   %s\n", state.Source)
		fmt.Printf("  Duration: %s\n", duration)
		fmt.Printf("  Events:   %d captured\n", len(events))
		fmt.Printf("\nUse 'mur session stop' to end recording.\n")

		return nil
	},
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List past recordings",
	RunE: func(cmd *cobra.Command, args []string) error {
		recordings, err := session.ListRecordings()
		if err != nil {
			return fmt.Errorf("failed to list recordings: %w", err)
		}

		if len(recordings) == 0 {
			fmt.Println("No recordings found.")
			return nil
		}

		fmt.Println("Session Recordings")
		fmt.Println("==================")
		fmt.Println()

		for _, r := range recordings {
			source := r.Source
			if source == "" {
				source = "unknown"
			}
			fmt.Printf("  %s  %s  %d events  (%s)\n",
				r.SessionID[:8],
				r.StartedAt.Format("2006-01-02 15:04"),
				r.EventCount,
				source,
			)
		}

		fmt.Printf("\nTotal: %d recordings\n", len(recordings))
		return nil
	},
}

var sessionAnalyzeCmd = &cobra.Command{
	Use:   "analyze <session-id>",
	Short: "Analyze a recorded session and extract a workflow",
	Long: `Analyze a recorded session and extract a workflow.

Uses the LLM configured in ~/.mur/config.yaml (learning.llm section).
Supported providers: anthropic, openai, ollama, gemini.

Examples:
  mur session analyze abc123                          # Use config default
  mur session analyze abc123 --provider claude        # Use Claude
  mur session analyze abc123 --provider ollama --model qwen3:8b
  mur session analyze abc123 --provider openai --model gpt-4o

See 'mur config providers' for model recommendations.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		llmProvider, _ := cmd.Flags().GetString("provider")
		llmModel, _ := cmd.Flags().GetString("model")
		llmOllamaURL, _ := cmd.Flags().GetString("ollama-url")

		_, err := runAnalysis(args[0], llmProvider, llmModel, llmOllamaURL)
		return err
	},
}

var sessionUICmd = &cobra.Command{
	Use:   "ui <session-id>",
	Short: "Open the workflow editor web UI for a session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID, err := session.ResolveSessionID(args[0])
		if err != nil {
			return err
		}

		result, err := session.LoadAnalysis(sessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "No saved analysis found. Analyzing session...\n")
			result, err = runAnalysis(sessionID, "", "", "")
			if err != nil {
				return err
			}
		}

		port, _ := cmd.Flags().GetInt("port")
		if port == 0 {
			port = 3939
		}

		srv, err := ui.NewServer(result, sessionID)
		if err != nil {
			return err
		}
		return srv.Serve(port)
	},
}

var sessionRecordCmd = &cobra.Command{
	Use:    "record",
	Short:  "Append an event to the active session",
	Hidden: true, // Used by hook scripts, not end users
	RunE: func(cmd *cobra.Command, args []string) error {
		eventType, _ := cmd.Flags().GetString("type")
		content, _ := cmd.Flags().GetString("content")
		tool, _ := cmd.Flags().GetString("tool")

		if eventType == "" {
			return fmt.Errorf("--type is required")
		}

		event := session.EventRecord{
			Type:    eventType,
			Content: content,
			Tool:    tool,
		}

		return session.RecordEventForActive(event)
	},
}

var sessionExportCmd = &cobra.Command{
	Use:   "export <session-id>",
	Short: "Export a session workflow as a skill, YAML, or markdown",
	Long: `Export an analyzed session workflow to a reusable format.

Formats:
  skill     Full skill directory: SKILL.md + workflow.yaml + run.sh + steps/
  yaml      Standalone workflow YAML file
  markdown  Human-readable markdown documentation

Examples:
  mur session export abc123 --format skill
  mur session export abc123 --format yaml --output workflow.yaml
  mur session export abc123 --format markdown --output workflow.md`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID, err := session.ResolveSessionID(args[0])
		if err != nil {
			return err
		}

		result, err := session.LoadAnalysis(sessionID)
		if err != nil {
			return fmt.Errorf("no analysis found for session %s (run 'mur session analyze' first): %w", args[0], err)
		}

		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")

		switch format {
		case "skill":
			outputDir := output
			if outputDir == "" {
				outputDir, err = session.DefaultSkillsOutputDir()
				if err != nil {
					return err
				}
			}
			skillPath, err := session.ExportAsSkill(result, sessionID, outputDir)
			if err != nil {
				return fmt.Errorf("export skill: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Skill exported to %s\n", skillPath)
			fmt.Println(skillPath)

		case "yaml":
			if output == "" {
				output = result.Name + ".yaml"
			}
			if err := session.ExportAsYAML(result, sessionID, output); err != nil {
				return fmt.Errorf("export YAML: %w", err)
			}
			fmt.Fprintf(os.Stderr, "YAML exported to %s\n", output)

		case "markdown", "md":
			if output == "" {
				output = result.Name + ".md"
			}
			if err := session.ExportAsMarkdown(result, output); err != nil {
				return fmt.Errorf("export markdown: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Markdown exported to %s\n", output)

		default:
			return fmt.Errorf("unknown format %q (use: skill, yaml, markdown)", format)
		}

		return nil
	},
}

// runAnalysis creates an LLM provider and runs QA-CoT analysis on a session.
func runAnalysis(sessionID, llmProvider, llmModel, llmOllamaURL string) (*session.AnalysisResult, error) {
	shortID := sessionID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	fmt.Fprintf(os.Stderr, "\nAnalyzing session %s...\n", shortID)

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	provider, err := session.NewLLMProviderWithOverrides(cfg, llmProvider, llmModel, llmOllamaURL)
	if err != nil {
		return nil, fmt.Errorf("LLM setup: %w", err)
	}

	result, err := session.Analyze(sessionID, provider)
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	// Save analysis result for later use by the web UI
	if err := session.SaveAnalysis(sessionID, result); err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: could not save analysis: %v\n", err)
	}

	// Print summary to stderr
	fmt.Fprintf(os.Stderr, "  Name:      %s\n", result.Name)
	fmt.Fprintf(os.Stderr, "  Trigger:   %s\n", result.Trigger)
	fmt.Fprintf(os.Stderr, "  Steps:     %d\n", len(result.Steps))
	fmt.Fprintf(os.Stderr, "  Variables: %d\n", len(result.Variables))
	fmt.Fprintf(os.Stderr, "  Tools:     %v\n", result.Tools)
	fmt.Fprintf(os.Stderr, "  Tags:      %v\n", result.Tags)

	// Print full JSON to stdout (for piping)
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	fmt.Println(string(data))

	return result, nil
}

func init() {
	rootCmd.AddCommand(sessionCmd)
	sessionCmd.AddCommand(sessionStartCmd)
	sessionCmd.AddCommand(sessionStopCmd)
	sessionCmd.AddCommand(sessionStatusCmd)
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionAnalyzeCmd)
	sessionCmd.AddCommand(sessionRecordCmd)
	sessionCmd.AddCommand(sessionUICmd)
	sessionCmd.AddCommand(sessionExportCmd)

	sessionStartCmd.Flags().String("source", "", "Recording source (e.g. claude-code, codex)")
	sessionStartCmd.Flags().String("marker", "", "Context marker from /mur:in message")

	sessionStopCmd.Flags().Bool("analyze", false, "Analyze the recording after stopping")
	sessionStopCmd.Flags().Bool("open", false, "Open web UI after analysis")
	sessionStopCmd.Flags().String("provider", "", "LLM provider override (anthropic, openai, ollama, gemini)")
	sessionStopCmd.Flags().String("model", "", "LLM model name override")
	sessionStopCmd.Flags().String("ollama-url", "", "Ollama API URL override")

	sessionStatusCmd.Flags().BoolP("quiet", "q", false, "Exit 0 if recording, 1 if not (for scripts)")

	sessionAnalyzeCmd.Flags().String("provider", "", "LLM provider override (anthropic, openai, ollama, gemini)")
	sessionAnalyzeCmd.Flags().String("model", "", "LLM model name override")
	sessionAnalyzeCmd.Flags().String("ollama-url", "", "Ollama API URL override")

	sessionRecordCmd.Flags().String("type", "", "Event type: user, assistant, tool_call, tool_result")
	sessionRecordCmd.Flags().String("content", "", "Event content")
	sessionRecordCmd.Flags().String("tool", "", "Tool name (for tool_call/tool_result events)")

	sessionUICmd.Flags().Int("port", 3939, "Port for the web UI server")

	sessionExportCmd.Flags().StringP("format", "f", "skill", "Export format: skill, yaml, markdown")
	sessionExportCmd.Flags().StringP("output", "o", "", "Output path (default: ~/.mur/skills/ for skill, ./<name>.yaml/.md for others)")
}
