package session

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// mockLLMProvider returns a canned response for testing.
type mockLLMProvider struct {
	response string
	err      error
}

func (m *mockLLMProvider) Complete(prompt string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

// fixtureTranscript returns a realistic session transcript as EventRecords.
func fixtureTranscript() []EventRecord {
	return []EventRecord{
		{Timestamp: 1000, Type: "user", Content: "The nginx container is returning 502 errors after our latest deploy"},
		{Timestamp: 1001, Type: "assistant", Content: "Let me check the Docker container logs to diagnose the 502 error."},
		{Timestamp: 1002, Type: "tool_call", Tool: "shell", Content: "docker compose logs web --tail 50"},
		{Timestamp: 1003, Type: "tool_result", Tool: "shell", Content: "[error] upstream prematurely closed connection while reading response header from upstream"},
		{Timestamp: 1004, Type: "assistant", Content: "The upstream connection is being closed. Let me check the nginx configuration."},
		{Timestamp: 1005, Type: "tool_call", Tool: "file-read", Content: "cat /etc/nginx/conf.d/default.conf"},
		{Timestamp: 1006, Type: "tool_result", Tool: "file-read", Content: "proxy_pass http://app:3000;\nproxy_read_timeout 30s;"},
		{Timestamp: 1007, Type: "assistant", Content: "The proxy_read_timeout is too low. Let me increase it and add proper headers."},
		{Timestamp: 1008, Type: "tool_call", Tool: "file-edit", Content: "Updated proxy_read_timeout to 120s and added proxy_set_header"},
		{Timestamp: 1009, Type: "tool_result", Tool: "file-edit", Content: "File saved"},
		{Timestamp: 1010, Type: "tool_call", Tool: "shell", Content: "docker compose restart nginx"},
		{Timestamp: 1011, Type: "tool_result", Tool: "shell", Content: "nginx restarted"},
		{Timestamp: 1012, Type: "tool_call", Tool: "shell", Content: "curl -f http://localhost/health"},
		{Timestamp: 1013, Type: "tool_result", Tool: "shell", Content: "HTTP/1.1 200 OK"},
		{Timestamp: 1014, Type: "assistant", Content: "The 502 error is resolved. The issue was a too-short proxy timeout."},
	}
}

// sampleLLMResponse is a realistic JSON response from the LLM.
const sampleLLMResponse = `
Q1: The user's goal was to fix nginx 502 errors after a deployment.
Q2: The root cause was a proxy_read_timeout that was too short (30s).
Q3: Checked Docker logs (succeeded), read nginx config (succeeded), edited config (succeeded), restarted nginx (succeeded), verified health (succeeded).
Q4: Check logs → read nginx config → fix timeout → restart → verify.
Q5: shell for logs/restart/health, file-read for config, file-edit for fix.
Q6: The server host and service name are environment-specific.
Q7: No conditional branches.
Q8: Restarting nginx should need approval.
Q9: Name: fix-nginx-502-timeout, Trigger: nginx returns 502 after deploy.
Q10: nginx, docker, deploy, timeout, 502.

{
  "name": "fix-nginx-502-timeout",
  "trigger": "Nginx returns 502 error after deployment",
  "description": "Diagnose and fix Nginx 502 errors caused by upstream timeout misconfiguration",
  "variables": [
    {"name": "service_name", "type": "string", "required": false, "default": "web", "description": "Docker Compose service name for the app"},
    {"name": "timeout_value", "type": "string", "required": false, "default": "120s", "description": "Proxy read timeout value"}
  ],
  "steps": [
    {"order": 1, "description": "Check Docker container logs for error details", "command": "docker compose logs $service_name --tail 50", "tool": "shell", "needs_approval": false, "on_failure": "abort"},
    {"order": 2, "description": "Read nginx configuration file", "tool": "file-read", "needs_approval": false, "on_failure": "abort"},
    {"order": 3, "description": "Update proxy_read_timeout to $timeout_value", "tool": "file-edit", "needs_approval": false, "on_failure": "abort"},
    {"order": 4, "description": "Restart nginx container", "command": "docker compose restart nginx", "tool": "shell", "needs_approval": true, "on_failure": "abort"},
    {"order": 5, "description": "Verify health check passes", "command": "curl -f http://localhost/health", "tool": "shell", "needs_approval": false, "on_failure": "retry"}
  ],
  "tools": ["shell", "file-read", "file-edit"],
  "tags": ["nginx", "docker", "deploy", "timeout", "502"]
}
`

func TestFormatTranscript(t *testing.T) {
	events := fixtureTranscript()
	result := formatTranscript(events)

	// Check that all event types are represented
	checks := []string{
		"[USER]",
		"[ASSISTANT]",
		"[TOOL_CALL: shell]",
		"[TOOL_RESULT: shell]",
		"[TOOL_CALL: file-read]",
		"[TOOL_RESULT: file-edit]",
	}
	for _, check := range checks {
		if !contains(result, check) {
			t.Errorf("transcript missing %q", check)
		}
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain JSON",
			input: `{"name": "test"}`,
			want:  `{"name": "test"}`,
		},
		{
			name:  "JSON with preamble",
			input: "Here is the analysis:\n\n" + `{"name": "test"}`,
			want:  `{"name": "test"}`,
		},
		{
			name:  "JSON in markdown fence",
			input: "```json\n" + `{"name": "test"}` + "\n```",
			want:  `{"name": "test"}`,
		},
		{
			name:  "JSON in plain fence",
			input: "```\n" + `{"name": "test"}` + "\n```",
			want:  `{"name": "test"}`,
		},
		{
			name:  "nested braces",
			input: `{"steps": [{"order": 1}]}`,
			want:  `{"steps": [{"order": 1}]}`,
		},
		{
			name:  "braces in strings",
			input: `{"cmd": "echo {hello}"}`,
			want:  `{"cmd": "echo {hello}"}`,
		},
		{
			name:  "no JSON",
			input: "just some text without JSON",
			want:  "",
		},
		{
			name:  "escaped quotes in strings",
			input: `{"desc": "say \"hello\""}`,
			want:  `{"desc": "say \"hello\""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSON(tt.input)
			if got != tt.want {
				t.Errorf("extractJSON() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseAnalysisResponse(t *testing.T) {
	result, err := parseAnalysisResponse(sampleLLMResponse)
	if err != nil {
		t.Fatalf("parseAnalysisResponse() error: %v", err)
	}

	if result.Name != "fix-nginx-502-timeout" {
		t.Errorf("Name = %q, want %q", result.Name, "fix-nginx-502-timeout")
	}
	if result.Trigger != "Nginx returns 502 error after deployment" {
		t.Errorf("Trigger = %q", result.Trigger)
	}
	if len(result.Variables) != 2 {
		t.Errorf("len(Variables) = %d, want 2", len(result.Variables))
	}
	if len(result.Steps) != 5 {
		t.Errorf("len(Steps) = %d, want 5", len(result.Steps))
	}
	if len(result.Tools) != 3 {
		t.Errorf("len(Tools) = %d, want 3", len(result.Tools))
	}
	if len(result.Tags) != 5 {
		t.Errorf("len(Tags) = %d, want 5", len(result.Tags))
	}

	// Check step details
	if result.Steps[3].NeedsApproval != true {
		t.Error("Step 4 should need approval")
	}
	if result.Steps[4].OnFailure != "retry" {
		t.Errorf("Step 5 OnFailure = %q, want %q", result.Steps[4].OnFailure, "retry")
	}
}

func TestParseAnalysisResponse_NoJSON(t *testing.T) {
	_, err := parseAnalysisResponse("just some text without JSON")
	if err == nil {
		t.Error("expected error for response with no JSON")
	}
}

func TestAnalyze_WithMockProvider(t *testing.T) {
	// Create a temp session with fixture data
	tmpDir := t.TempDir()
	sessionID := "test-analyze-session"

	// Write fixture JSONL
	events := fixtureTranscript()
	var lines []byte
	for _, e := range events {
		data, _ := json.Marshal(e)
		lines = append(lines, data...)
		lines = append(lines, '\n')
	}

	recDir := tmpDir + "/recordings"
	if err := writeTestFile(recDir, sessionID+".jsonl", lines); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// Override recordingsDir for test
	origFunc := recordingsDirFunc
	recordingsDirFunc = func() (string, error) { return recDir, nil }
	defer func() { recordingsDirFunc = origFunc }()

	mock := &mockLLMProvider{response: sampleLLMResponse}
	result, err := Analyze(sessionID, mock)
	if err != nil {
		t.Fatalf("Analyze() error: %v", err)
	}

	if result.Name != "fix-nginx-502-timeout" {
		t.Errorf("Name = %q, want %q", result.Name, "fix-nginx-502-timeout")
	}
	if len(result.Steps) != 5 {
		t.Errorf("len(Steps) = %d, want 5", len(result.Steps))
	}
}

func TestAnalyze_EmptySession(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "empty-session"
	recDir := tmpDir + "/recordings"

	if err := writeTestFile(recDir, sessionID+".jsonl", []byte{}); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	origFunc := recordingsDirFunc
	recordingsDirFunc = func() (string, error) { return recDir, nil }
	defer func() { recordingsDirFunc = origFunc }()

	mock := &mockLLMProvider{response: "{}"}
	_, err := Analyze(sessionID, mock)
	if err == nil {
		t.Error("expected error for empty session")
	}
}

func TestAnalyze_LLMError(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "error-session"
	recDir := tmpDir + "/recordings"

	events := fixtureTranscript()
	var lines []byte
	for _, e := range events {
		data, _ := json.Marshal(e)
		lines = append(lines, data...)
		lines = append(lines, '\n')
	}
	if err := writeTestFile(recDir, sessionID+".jsonl", lines); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	origFunc := recordingsDirFunc
	recordingsDirFunc = func() (string, error) { return recDir, nil }
	defer func() { recordingsDirFunc = origFunc }()

	mock := &mockLLMProvider{err: fmt.Errorf("API rate limited")}
	_, err := Analyze(sessionID, mock)
	if err == nil {
		t.Error("expected error when LLM fails")
	}
}

func TestStepOrderNormalization(t *testing.T) {
	raw := `{"name":"test","steps":[{"description":"a"},{"description":"b"},{"description":"c"}]}`
	result, err := parseAnalysisResponse(raw)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	for i, step := range result.Steps {
		if step.Order != i+1 {
			t.Errorf("step %d Order = %d, want %d", i, step.Order, i+1)
		}
		if step.OnFailure != "abort" {
			t.Errorf("step %d OnFailure = %q, want %q", i, step.OnFailure, "abort")
		}
	}
}

// helpers

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func writeTestFile(dir, name string, data []byte) error {
	if err := mkdirAll(dir); err != nil {
		return err
	}
	return writeFile(dir+"/"+name, data)
}

func mkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
