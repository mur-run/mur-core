package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// WorkflowYAML is the structured YAML format for exported workflows.
type WorkflowYAML struct {
	Kind        string     `yaml:"kind"`
	Version     string     `yaml:"version"`
	Name        string     `yaml:"name"`
	Trigger     string     `yaml:"trigger"`
	Description string     `yaml:"description"`
	Variables   []Variable `yaml:"variables,omitempty"`
	Steps       []Step     `yaml:"steps"`
	Tools       []string   `yaml:"tools,omitempty"`
	Tags        []string   `yaml:"tags,omitempty"`
	Source      *Source    `yaml:"source,omitempty"`
}

// Source records provenance metadata for an exported workflow.
type Source struct {
	RecordedAt string `yaml:"recorded_at"`
	SessionID  string `yaml:"session_id"`
}

// ExportAsSkill creates a complete skill directory structure:
//
//	<outputDir>/<name>/
//	  ├── SKILL.md        # OpenClaw-compatible skill description
//	  ├── workflow.yaml   # Structured workflow definition
//	  ├── run.sh          # Entry point script
//	  └── steps/
//	      ├── 01-<step>.sh
//	      └── 02-<step>.sh
func ExportAsSkill(result *AnalysisResult, sessionID, outputDir string) (string, error) {
	if result.Name == "" {
		return "", fmt.Errorf("workflow name is required for skill export")
	}

	skillDir := filepath.Join(outputDir, result.Name)
	stepsDir := filepath.Join(skillDir, "steps")

	if err := os.MkdirAll(stepsDir, 0755); err != nil {
		return "", fmt.Errorf("create skill directory: %w", err)
	}

	// Write SKILL.md
	if err := writeSkillMD(result, skillDir); err != nil {
		return "", fmt.Errorf("write SKILL.md: %w", err)
	}

	// Write workflow.yaml
	if err := writeWorkflowYAML(result, sessionID, skillDir); err != nil {
		return "", fmt.Errorf("write workflow.yaml: %w", err)
	}

	// Write run.sh
	if err := writeRunSH(result, skillDir); err != nil {
		return "", fmt.Errorf("write run.sh: %w", err)
	}

	// Write individual step scripts
	if err := writeStepScripts(result, stepsDir); err != nil {
		return "", fmt.Errorf("write step scripts: %w", err)
	}

	return skillDir, nil
}

// ExportAsYAML exports the workflow as a single YAML file.
func ExportAsYAML(result *AnalysisResult, sessionID, path string) error {
	wf := buildWorkflowYAML(result, sessionID)

	data, err := yaml.Marshal(wf)
	if err != nil {
		return fmt.Errorf("marshal YAML: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// ExportAsMarkdown exports the workflow as a readable Markdown document.
func ExportAsMarkdown(result *AnalysisResult, path string) error {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", result.Name)

	if result.Trigger != "" {
		fmt.Fprintf(&b, "> **Trigger:** %s\n\n", result.Trigger)
	}

	if result.Description != "" {
		fmt.Fprintf(&b, "%s\n\n", result.Description)
	}

	// Variables
	if len(result.Variables) > 0 {
		fmt.Fprintf(&b, "## Variables\n\n")
		fmt.Fprintf(&b, "| Name | Type | Required | Default | Description |\n")
		fmt.Fprintf(&b, "|------|------|----------|---------|-------------|\n")
		for _, v := range result.Variables {
			req := "No"
			if v.Required {
				req = "Yes"
			}
			def := v.Default
			if def == "" {
				def = "-"
			}
			fmt.Fprintf(&b, "| `%s` | %s | %s | %s | %s |\n",
				v.Name, v.Type, req, def, v.Description)
		}
		b.WriteString("\n")
	}

	// Steps
	if len(result.Steps) > 0 {
		fmt.Fprintf(&b, "## Steps\n\n")
		for _, step := range result.Steps {
			approval := ""
			if step.NeedsApproval {
				approval = " [requires approval]"
			}
			fmt.Fprintf(&b, "%d. **%s**%s\n", step.Order, step.Description, approval)
			if step.Command != "" {
				fmt.Fprintf(&b, "   ```bash\n   %s\n   ```\n", step.Command)
			}
			if step.Tool != "" {
				fmt.Fprintf(&b, "   - Tool: `%s`\n", step.Tool)
			}
			if step.OnFailure != "" && step.OnFailure != "abort" {
				fmt.Fprintf(&b, "   - On failure: %s\n", step.OnFailure)
			}
			b.WriteString("\n")
		}
	}

	// Tools
	if len(result.Tools) > 0 {
		fmt.Fprintf(&b, "## Tools\n\n")
		for _, tool := range result.Tools {
			fmt.Fprintf(&b, "- `%s`\n", tool)
		}
		b.WriteString("\n")
	}

	// Tags
	if len(result.Tags) > 0 {
		fmt.Fprintf(&b, "## Tags\n\n")
		fmt.Fprintf(&b, "%s\n", strings.Join(result.Tags, ", "))
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}

// DefaultSkillsOutputDir returns ~/.mur/skills/.
func DefaultSkillsOutputDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".mur", "skills"), nil
}

// --- internal helpers ---

func buildWorkflowYAML(result *AnalysisResult, sessionID string) WorkflowYAML {
	wf := WorkflowYAML{
		Kind:        "workflow",
		Version:     "1",
		Name:        result.Name,
		Trigger:     result.Trigger,
		Description: result.Description,
		Variables:   result.Variables,
		Steps:       result.Steps,
		Tools:       result.Tools,
		Tags:        result.Tags,
	}

	if sessionID != "" {
		wf.Source = &Source{
			RecordedAt: time.Now().Format(time.RFC3339),
			SessionID:  sessionID,
		}
	}

	return wf
}

func writeSkillMD(result *AnalysisResult, skillDir string) error {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", result.Name)

	if result.Description != "" {
		fmt.Fprintf(&b, "## Description\n\n%s\n\n", result.Description)
	}

	if result.Trigger != "" {
		fmt.Fprintf(&b, "## When to Use\n\n%s\n\n", result.Trigger)
	}

	fmt.Fprintf(&b, "## Instructions\n\n")
	fmt.Fprintf(&b, "This is an automated workflow extracted from a recorded session.\n\n")

	fmt.Fprintf(&b, "### Steps\n\n")
	for _, step := range result.Steps {
		approval := ""
		if step.NeedsApproval {
			approval = " *(requires approval)*"
		}
		fmt.Fprintf(&b, "%d. %s%s\n", step.Order, step.Description, approval)
		if step.Command != "" {
			fmt.Fprintf(&b, "   ```bash\n   %s\n   ```\n", step.Command)
		}
	}
	b.WriteString("\n")

	if len(result.Variables) > 0 {
		fmt.Fprintf(&b, "### Variables\n\n")
		for _, v := range result.Variables {
			req := "optional"
			if v.Required {
				req = "required"
			}
			fmt.Fprintf(&b, "- `$%s` (%s, %s): %s\n", v.Name, v.Type, req, v.Description)
			if v.Default != "" {
				fmt.Fprintf(&b, "  Default: `%s`\n", v.Default)
			}
		}
		b.WriteString("\n")
	}

	if len(result.Tools) > 0 {
		fmt.Fprintf(&b, "### Tools Required\n\n")
		for _, tool := range result.Tools {
			fmt.Fprintf(&b, "- %s\n", tool)
		}
		b.WriteString("\n")
	}

	if len(result.Tags) > 0 {
		fmt.Fprintf(&b, "### Tags\n\n")
		fmt.Fprintf(&b, "%s\n", strings.Join(result.Tags, ", "))
	}

	return os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(b.String()), 0644)
}

func writeWorkflowYAML(result *AnalysisResult, sessionID, skillDir string) error {
	wf := buildWorkflowYAML(result, sessionID)

	data, err := yaml.Marshal(wf)
	if err != nil {
		return fmt.Errorf("marshal workflow: %w", err)
	}

	return os.WriteFile(filepath.Join(skillDir, "workflow.yaml"), data, 0644)
}

func writeRunSH(result *AnalysisResult, skillDir string) error {
	var b strings.Builder

	b.WriteString("#!/bin/bash\n")
	b.WriteString("# mur-managed-workflow v1\n")
	fmt.Fprintf(&b, "# %s\n", result.Name)
	if result.Description != "" {
		fmt.Fprintf(&b, "# %s\n", result.Description)
	}
	b.WriteString("set -euo pipefail\n\n")

	b.WriteString("SCRIPT_DIR=\"$(cd \"$(dirname \"${BASH_SOURCE[0]}\")\" && pwd)\"\n\n")

	// Variable declarations with defaults
	for _, v := range result.Variables {
		envName := strings.ToUpper(v.Name)
		if v.Default != "" {
			fmt.Fprintf(&b, "%s=\"${%s:-%s}\"\n", envName, envName, v.Default)
		} else if v.Required {
			fmt.Fprintf(&b, "if [ -z \"${%s:-}\" ]; then\n", envName)
			fmt.Fprintf(&b, "  echo \"Error: %s is required\" >&2\n", envName)
			fmt.Fprintf(&b, "  exit 1\n")
			fmt.Fprintf(&b, "fi\n")
		}
	}

	if len(result.Variables) > 0 {
		b.WriteString("\n")
	}

	b.WriteString("echo \"Running workflow: " + result.Name + "\"\n\n")

	// Execute step scripts
	for i, step := range result.Steps {
		scriptName := fmt.Sprintf("%02d-%s.sh", i+1, slugify(step.Description))
		if step.NeedsApproval {
			fmt.Fprintf(&b, "echo \"Step %d: %s [requires approval]\"\n", i+1, step.Description)
			fmt.Fprintf(&b, "read -p \"Proceed? [y/N] \" confirm\n")
			fmt.Fprintf(&b, "if [ \"$confirm\" = \"y\" ] || [ \"$confirm\" = \"Y\" ]; then\n")
			fmt.Fprintf(&b, "  bash \"$SCRIPT_DIR/steps/%s\"\n", scriptName)
			fmt.Fprintf(&b, "else\n")
			fmt.Fprintf(&b, "  echo \"Skipped step %d\"\n", i+1)
			fmt.Fprintf(&b, "fi\n\n")
		} else {
			fmt.Fprintf(&b, "echo \"Step %d: %s\"\n", i+1, step.Description)
			fmt.Fprintf(&b, "bash \"$SCRIPT_DIR/steps/%s\"\n\n", scriptName)
		}
	}

	b.WriteString("echo \"Workflow complete.\"\n")

	return os.WriteFile(filepath.Join(skillDir, "run.sh"), []byte(b.String()), 0755)
}

func writeStepScripts(result *AnalysisResult, stepsDir string) error {
	for i, step := range result.Steps {
		scriptName := fmt.Sprintf("%02d-%s.sh", i+1, slugify(step.Description))

		var b strings.Builder
		b.WriteString("#!/bin/bash\n")
		fmt.Fprintf(&b, "# Step %d: %s\n", i+1, step.Description)
		b.WriteString("set -euo pipefail\n\n")

		if step.Command != "" {
			fmt.Fprintf(&b, "%s\n", step.Command)
		} else if step.Tool != "" {
			fmt.Fprintf(&b, "# Tool: %s\n", step.Tool)
			fmt.Fprintf(&b, "echo \"This step requires manual execution with tool: %s\"\n", step.Tool)
		} else {
			fmt.Fprintf(&b, "echo \"TODO: %s\"\n", step.Description)
		}

		if err := os.WriteFile(filepath.Join(stepsDir, scriptName), []byte(b.String()), 0755); err != nil {
			return fmt.Errorf("write step %d: %w", i+1, err)
		}
	}

	return nil
}

// slugify converts a description to a filesystem-safe slug.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return '-'
	}, s)

	// Collapse multiple dashes
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")

	// Limit length
	if len(s) > 40 {
		s = s[:40]
		s = strings.TrimRight(s, "-")
	}

	return s
}
