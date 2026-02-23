package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sampleAnalysisResult() *AnalysisResult {
	return &AnalysisResult{
		Name:        "fix-nginx-502-timeout",
		Trigger:     "Nginx returns 502 error after deployment",
		Description: "Diagnose and fix Nginx 502 errors caused by upstream timeout misconfiguration",
		Variables: []Variable{
			{Name: "service_name", Type: "string", Required: false, Default: "web", Description: "Docker Compose service name"},
			{Name: "timeout_value", Type: "string", Required: true, Default: "", Description: "Proxy read timeout value"},
		},
		Steps: []Step{
			{Order: 1, Description: "Check Docker container logs", Command: "docker compose logs $service_name --tail 50", Tool: "shell", OnFailure: "abort"},
			{Order: 2, Description: "Read nginx configuration file", Tool: "file-read", OnFailure: "abort"},
			{Order: 3, Description: "Update proxy_read_timeout", Tool: "file-edit", OnFailure: "abort"},
			{Order: 4, Description: "Restart nginx container", Command: "docker compose restart nginx", Tool: "shell", NeedsApproval: true, OnFailure: "abort"},
			{Order: 5, Description: "Verify health check passes", Command: "curl -f http://localhost/health", Tool: "shell", OnFailure: "retry"},
		},
		Tools: []string{"shell", "file-read", "file-edit"},
		Tags:  []string{"nginx", "docker", "deploy", "timeout"},
	}
}

func TestExportAsSkill(t *testing.T) {
	tmpDir := t.TempDir()
	result := sampleAnalysisResult()

	skillPath, err := ExportAsSkill(result, "test-session-id", tmpDir)
	if err != nil {
		t.Fatalf("ExportAsSkill() error: %v", err)
	}

	expectedDir := filepath.Join(tmpDir, "fix-nginx-502-timeout")
	if skillPath != expectedDir {
		t.Errorf("skillPath = %q, want %q", skillPath, expectedDir)
	}

	// Check SKILL.md exists and has expected content
	skillMD, err := os.ReadFile(filepath.Join(skillPath, "SKILL.md"))
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}
	skillContent := string(skillMD)
	if !strings.Contains(skillContent, "# fix-nginx-502-timeout") {
		t.Error("SKILL.md missing title")
	}
	if !strings.Contains(skillContent, "## Description") {
		t.Error("SKILL.md missing Description section")
	}
	if !strings.Contains(skillContent, "## Instructions") {
		t.Error("SKILL.md missing Instructions section")
	}
	if !strings.Contains(skillContent, "$service_name") {
		t.Error("SKILL.md missing variable reference")
	}

	// Check workflow.yaml exists and has expected content
	wfYAML, err := os.ReadFile(filepath.Join(skillPath, "workflow.yaml"))
	if err != nil {
		t.Fatalf("read workflow.yaml: %v", err)
	}
	wfContent := string(wfYAML)
	if !strings.Contains(wfContent, "kind: workflow") {
		t.Error("workflow.yaml missing kind field")
	}
	if !strings.Contains(wfContent, "version: \"1\"") {
		t.Error("workflow.yaml missing version field")
	}
	if !strings.Contains(wfContent, "name: fix-nginx-502-timeout") {
		t.Error("workflow.yaml missing name")
	}
	if !strings.Contains(wfContent, "session_id: test-session-id") {
		t.Error("workflow.yaml missing session_id in source")
	}

	// Check run.sh exists and is executable
	runSH, err := os.ReadFile(filepath.Join(skillPath, "run.sh"))
	if err != nil {
		t.Fatalf("read run.sh: %v", err)
	}
	runContent := string(runSH)
	if !strings.HasPrefix(runContent, "#!/bin/bash") {
		t.Error("run.sh missing shebang")
	}
	if !strings.Contains(runContent, "set -euo pipefail") {
		t.Error("run.sh missing strict mode")
	}
	if !strings.Contains(runContent, "TIMEOUT_VALUE") {
		t.Error("run.sh missing required variable check")
	}

	info, err := os.Stat(filepath.Join(skillPath, "run.sh"))
	if err != nil {
		t.Fatalf("stat run.sh: %v", err)
	}
	if info.Mode()&0100 == 0 {
		t.Error("run.sh is not executable")
	}

	// Check step scripts exist
	stepsDir := filepath.Join(skillPath, "steps")
	entries, err := os.ReadDir(stepsDir)
	if err != nil {
		t.Fatalf("read steps dir: %v", err)
	}
	if len(entries) != 5 {
		t.Errorf("expected 5 step scripts, got %d", len(entries))
	}

	// Check first step script
	step1, err := os.ReadFile(filepath.Join(stepsDir, entries[0].Name()))
	if err != nil {
		t.Fatalf("read step 1: %v", err)
	}
	if !strings.Contains(string(step1), "docker compose logs") {
		t.Error("step 1 missing command")
	}

	// Check step 4 (approval step) has the approval prompt in run.sh
	if !strings.Contains(runContent, "requires approval") {
		t.Error("run.sh missing approval prompt for step 4")
	}
}

func TestExportAsSkill_EmptyName(t *testing.T) {
	tmpDir := t.TempDir()
	result := &AnalysisResult{}

	_, err := ExportAsSkill(result, "test-session", tmpDir)
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestExportAsYAML(t *testing.T) {
	tmpDir := t.TempDir()
	result := sampleAnalysisResult()
	outputPath := filepath.Join(tmpDir, "output.yaml")

	if err := ExportAsYAML(result, "test-session", outputPath); err != nil {
		t.Fatalf("ExportAsYAML() error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "kind: workflow") {
		t.Error("YAML missing kind field")
	}
	if !strings.Contains(content, "name: fix-nginx-502-timeout") {
		t.Error("YAML missing name")
	}
	if !strings.Contains(content, "needs_approval: true") {
		t.Error("YAML missing needs_approval flag")
	}
	if !strings.Contains(content, "session_id: test-session") {
		t.Error("YAML missing source session_id")
	}
}

func TestExportAsMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	result := sampleAnalysisResult()
	outputPath := filepath.Join(tmpDir, "output.md")

	if err := ExportAsMarkdown(result, outputPath); err != nil {
		t.Fatalf("ExportAsMarkdown() error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# fix-nginx-502-timeout") {
		t.Error("markdown missing title")
	}
	if !strings.Contains(content, "## Variables") {
		t.Error("markdown missing Variables section")
	}
	if !strings.Contains(content, "## Steps") {
		t.Error("markdown missing Steps section")
	}
	if !strings.Contains(content, "[requires approval]") {
		t.Error("markdown missing approval indicator")
	}
	if !strings.Contains(content, "## Tools") {
		t.Error("markdown missing Tools section")
	}
	if !strings.Contains(content, "## Tags") {
		t.Error("markdown missing Tags section")
	}
	// Check table format for variables
	if !strings.Contains(content, "| `service_name`") {
		t.Error("markdown missing variable table row")
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Check Docker logs", "check-docker-logs"},
		{"SSH to $server_host", "ssh-to-server-host"},
		{"simple", "simple"},
		{"", ""},
		{"A very long description that should be truncated to forty characters max", "a-very-long-description-that-should-be-t"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
