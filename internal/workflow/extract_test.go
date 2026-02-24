package workflow

import (
	"testing"

	"github.com/mur-run/mur-core/internal/session"
)

func sampleAnalysisResult() *session.AnalysisResult {
	return &session.AnalysisResult{
		Name:        "fix-nginx-502-timeout",
		Trigger:     "Nginx returns 502 error after deployment",
		Description: "Diagnose and fix Nginx 502 errors caused by upstream timeout misconfiguration",
		Variables: []session.Variable{
			{Name: "service_name", Type: "string", Required: false, Default: "web", Description: "Docker Compose service name"},
			{Name: "timeout_value", Type: "string", Required: true, Default: "", Description: "Proxy read timeout value"},
		},
		Steps: []session.Step{
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

func TestExtractFromAnalysis(t *testing.T) {
	result := sampleAnalysisResult()

	wf, err := ExtractFromAnalysis(result, ExtractOptions{
		SessionID: "test-session-123",
	})
	if err != nil {
		t.Fatalf("ExtractFromAnalysis() error: %v", err)
	}

	if wf.ID == "" {
		t.Error("workflow ID should be generated")
	}
	if wf.Name != "fix-nginx-502-timeout" {
		t.Errorf("Name = %q, want %q", wf.Name, "fix-nginx-502-timeout")
	}
	if wf.Description != "Diagnose and fix Nginx 502 errors caused by upstream timeout misconfiguration" {
		t.Errorf("Description = %q", wf.Description)
	}
	if wf.Trigger != "Nginx returns 502 error after deployment" {
		t.Errorf("Trigger = %q", wf.Trigger)
	}
	if len(wf.Variables) != 2 {
		t.Errorf("len(Variables) = %d, want 2", len(wf.Variables))
	}
	if len(wf.Steps) != 5 {
		t.Errorf("len(Steps) = %d, want 5", len(wf.Steps))
	}
	if len(wf.Tools) != 3 {
		t.Errorf("len(Tools) = %d, want 3", len(wf.Tools))
	}
	if len(wf.Tags) != 4 {
		t.Errorf("len(Tags) = %d, want 4", len(wf.Tags))
	}

	// Verify source session reference
	if len(wf.SourceSessions) != 1 {
		t.Fatalf("len(SourceSessions) = %d, want 1", len(wf.SourceSessions))
	}
	if wf.SourceSessions[0].SessionID != "test-session-123" {
		t.Errorf("SourceSession.SessionID = %q, want %q", wf.SourceSessions[0].SessionID, "test-session-123")
	}
}

func TestExtractFromAnalysis_NilResult(t *testing.T) {
	_, err := ExtractFromAnalysis(nil, ExtractOptions{})
	if err == nil {
		t.Error("expected error for nil analysis result")
	}
}

func TestExtractFromAnalysis_NoSessionID(t *testing.T) {
	result := sampleAnalysisResult()

	wf, err := ExtractFromAnalysis(result, ExtractOptions{})
	if err != nil {
		t.Fatalf("ExtractFromAnalysis() error: %v", err)
	}

	if len(wf.SourceSessions) != 0 {
		t.Errorf("len(SourceSessions) = %d, want 0 when no session ID provided", len(wf.SourceSessions))
	}
}

func TestExtractFromAnalysis_WithStartEnd(t *testing.T) {
	result := sampleAnalysisResult()

	wf, err := ExtractFromAnalysis(result, ExtractOptions{
		SessionID: "slice-session",
		Start:     1,
		End:       4,
	})
	if err != nil {
		t.Fatalf("ExtractFromAnalysis() error: %v", err)
	}

	// Should include steps[1:4] = steps 2, 3, 4 from original
	if len(wf.Steps) != 3 {
		t.Fatalf("len(Steps) = %d, want 3", len(wf.Steps))
	}

	// Steps should be renumbered sequentially
	if wf.Steps[0].Order != 1 {
		t.Errorf("Steps[0].Order = %d, want 1", wf.Steps[0].Order)
	}
	if wf.Steps[1].Order != 2 {
		t.Errorf("Steps[1].Order = %d, want 2", wf.Steps[1].Order)
	}
	if wf.Steps[2].Order != 3 {
		t.Errorf("Steps[2].Order = %d, want 3", wf.Steps[2].Order)
	}

	// Verify the content of the sliced steps
	if wf.Steps[0].Description != "Read nginx configuration file" {
		t.Errorf("Steps[0].Description = %q, want %q", wf.Steps[0].Description, "Read nginx configuration file")
	}
	if wf.Steps[2].NeedsApproval != true {
		t.Error("Steps[2] (originally step 4) should need approval")
	}

	// Verify source ref has start/end
	if len(wf.SourceSessions) != 1 {
		t.Fatalf("len(SourceSessions) = %d, want 1", len(wf.SourceSessions))
	}
	if wf.SourceSessions[0].StartEvent != 1 {
		t.Errorf("SourceSession.StartEvent = %d, want 1", wf.SourceSessions[0].StartEvent)
	}
	if wf.SourceSessions[0].EndEvent != 4 {
		t.Errorf("SourceSession.EndEvent = %d, want 4", wf.SourceSessions[0].EndEvent)
	}
}

func TestExtractFromAnalysis_StartBeyondSteps(t *testing.T) {
	result := sampleAnalysisResult()

	_, err := ExtractFromAnalysis(result, ExtractOptions{
		Start: 10,
		End:   15,
	})
	if err == nil {
		t.Error("expected error when start exceeds step count")
	}
}

func TestExtractFromAnalysis_StartGteEnd(t *testing.T) {
	result := sampleAnalysisResult()

	_, err := ExtractFromAnalysis(result, ExtractOptions{
		Start: 3,
		End:   3,
	})
	if err == nil {
		t.Error("expected error when start >= end")
	}
}

func TestExtractFromAnalysis_EndBeyondSteps(t *testing.T) {
	result := sampleAnalysisResult()

	// End beyond step count should be clamped to len(steps)
	wf, err := ExtractFromAnalysis(result, ExtractOptions{
		Start: 3,
		End:   100,
	})
	if err != nil {
		t.Fatalf("ExtractFromAnalysis() error: %v", err)
	}

	// steps[3:5] = steps 4, 5 from original
	if len(wf.Steps) != 2 {
		t.Errorf("len(Steps) = %d, want 2", len(wf.Steps))
	}
}

func TestExtractFromAnalysis_UniqueIDs(t *testing.T) {
	result := sampleAnalysisResult()

	wf1, err := ExtractFromAnalysis(result, ExtractOptions{})
	if err != nil {
		t.Fatalf("first ExtractFromAnalysis() error: %v", err)
	}

	wf2, err := ExtractFromAnalysis(result, ExtractOptions{})
	if err != nil {
		t.Fatalf("second ExtractFromAnalysis() error: %v", err)
	}

	if wf1.ID == wf2.ID {
		t.Errorf("two extractions should produce different IDs, got %q", wf1.ID)
	}
}
