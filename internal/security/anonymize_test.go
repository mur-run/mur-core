package security

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mockLLMClient implements LLMClient for testing.
type mockLLMClient struct {
	response string
	err      error
	calls    int
}

func (m *mockLLMClient) Generate(prompt string) (string, error) {
	m.calls++
	return m.response, m.err
}

func TestSemanticAnonymizerBasic(t *testing.T) {
	mock := &mockLLMClient{
		response: `This pattern at <COMPANY> uses error handling.
Contact <PERSON> for details.
---CHANGES---
[{"original":"Acme Corp","replaced":"<COMPANY>","category":"company"},{"original":"John Smith","replaced":"<PERSON>","category":"person"}]`,
	}

	anonymizer := NewSemanticAnonymizer(mock, "")
	content := "This pattern at Acme Corp uses error handling.\nContact John Smith for details."

	cleaned, changes, err := anonymizer.Anonymize(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(cleaned, "<COMPANY>") {
		t.Errorf("expected <COMPANY> in cleaned content, got: %s", cleaned)
	}
	if !strings.Contains(cleaned, "<PERSON>") {
		t.Errorf("expected <PERSON> in cleaned content, got: %s", cleaned)
	}
	if strings.Contains(cleaned, "Acme Corp") {
		t.Errorf("expected 'Acme Corp' to be replaced, got: %s", cleaned)
	}

	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}
	if changes[0].Category != "company" {
		t.Errorf("expected category 'company', got %s", changes[0].Category)
	}
	if changes[0].Original != "Acme Corp" {
		t.Errorf("expected original 'Acme Corp', got %s", changes[0].Original)
	}
	if changes[1].Category != "person" {
		t.Errorf("expected category 'person', got %s", changes[1].Category)
	}
}

func TestSemanticAnonymizerNoChanges(t *testing.T) {
	mock := &mockLLMClient{
		response: `Use errors.Is and errors.As for error checking.
---CHANGES---
[]`,
	}

	anonymizer := NewSemanticAnonymizer(mock, "")
	content := "Use errors.Is and errors.As for error checking."

	cleaned, changes, err := anonymizer.Anonymize(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cleaned != content {
		t.Errorf("expected content unchanged, got: %s", cleaned)
	}
	if len(changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(changes))
	}
}

func TestSemanticAnonymizerLLMError(t *testing.T) {
	mock := &mockLLMClient{
		err: fmt.Errorf("connection refused"),
	}

	anonymizer := NewSemanticAnonymizer(mock, "")
	content := "Some content about Acme Corp"

	cleaned, changes, err := anonymizer.Anonymize(content)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("expected connection refused error, got: %v", err)
	}
	// On error, original content should be returned
	if cleaned != content {
		t.Errorf("expected original content on error, got: %s", cleaned)
	}
	if changes != nil {
		t.Errorf("expected nil changes on error, got: %v", changes)
	}
}

func TestSemanticAnonymizerInMemoryCache(t *testing.T) {
	mock := &mockLLMClient{
		response: `Cleaned content
---CHANGES---
[{"original":"Acme","replaced":"<COMPANY>","category":"company"}]`,
	}

	anonymizer := NewSemanticAnonymizer(mock, "")
	content := "Content about Acme"

	// First call
	_, _, err := anonymizer.Anonymize(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.calls != 1 {
		t.Errorf("expected 1 LLM call, got %d", mock.calls)
	}

	// Second call with same content should use cache
	_, _, err = anonymizer.Anonymize(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.calls != 1 {
		t.Errorf("expected still 1 LLM call (cached), got %d", mock.calls)
	}

	// Different content should trigger new call
	_, _, err = anonymizer.Anonymize("Different content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.calls != 2 {
		t.Errorf("expected 2 LLM calls, got %d", mock.calls)
	}
}

func TestSemanticAnonymizerDiskCache(t *testing.T) {
	tmpDir := t.TempDir()

	mock := &mockLLMClient{
		response: `<COMPANY> pattern
---CHANGES---
[{"original":"Acme","replaced":"<COMPANY>","category":"company"}]`,
	}

	// First anonymizer writes to disk
	anon1 := NewSemanticAnonymizer(mock, tmpDir)
	content := "Acme pattern"

	_, _, err := anon1.Anonymize(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.calls != 1 {
		t.Fatalf("expected 1 call, got %d", mock.calls)
	}

	// Verify file was written
	hash := contentHash(content)
	cachePath := filepath.Join(tmpDir, hash+".txt")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("expected cache file to exist")
	}

	// Second anonymizer (fresh instance) should load from disk
	anon2 := NewSemanticAnonymizer(mock, tmpDir)
	cleaned, changes, err := anon2.Anonymize(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.calls != 1 {
		t.Errorf("expected still 1 call (disk cache), got %d", mock.calls)
	}
	if !strings.Contains(cleaned, "<COMPANY>") {
		t.Errorf("expected <COMPANY> from cache, got: %s", cleaned)
	}
	if len(changes) != 1 {
		t.Errorf("expected 1 change from cache, got %d", len(changes))
	}
}

func TestSemanticAnonymizerAllCategories(t *testing.T) {
	changesJSON := `[
		{"original":"Acme Corp","replaced":"<COMPANY>","category":"company"},
		{"original":"Jane Doe","replaced":"<PERSON>","category":"person"},
		{"original":"ProjectX","replaced":"<PROJECT>","category":"project"},
		{"original":"$50M ARR","replaced":"<METRIC>","category":"metric"},
		{"original":"San Francisco office","replaced":"<LOCATION>","category":"location"}
	]`

	mock := &mockLLMClient{
		response: "Cleaned content here\n---CHANGES---\n" + changesJSON,
	}

	anonymizer := NewSemanticAnonymizer(mock, "")
	_, changes, err := anonymizer.Anonymize("Some content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(changes) != 5 {
		t.Fatalf("expected 5 changes, got %d", len(changes))
	}

	categories := map[string]bool{}
	for _, c := range changes {
		categories[c.Category] = true
	}

	for _, expected := range []string{"company", "person", "project", "metric", "location"} {
		if !categories[expected] {
			t.Errorf("expected category %q in changes", expected)
		}
	}
}

func TestSemanticAnonymizerLineNumbers(t *testing.T) {
	mock := &mockLLMClient{
		response: `Line one clean
<COMPANY> line two
Line three
---CHANGES---
[{"original":"Acme","replaced":"<COMPANY>","category":"company"}]`,
	}

	anonymizer := NewSemanticAnonymizer(mock, "")
	content := "Line one clean\nAcme line two\nLine three"

	_, changes, err := anonymizer.Anonymize(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Line != 2 {
		t.Errorf("expected line 2, got %d", changes[0].Line)
	}
}

func TestFormatAnonymizationChanges(t *testing.T) {
	changes := []AnonymizationChange{
		{Original: "Acme Corp", Replaced: "<COMPANY>", Category: "company", Line: 1},
		{Original: "John", Replaced: "<PERSON>", Category: "person", Line: 3},
	}

	output := FormatAnonymizationChanges(changes)
	if !strings.Contains(output, "company") {
		t.Error("expected 'company' in output")
	}
	if !strings.Contains(output, "Line 1") {
		t.Error("expected 'Line 1' in output")
	}
	if !strings.Contains(output, "Acme Corp") {
		t.Error("expected 'Acme Corp' in output")
	}

	// Empty changes
	output = FormatAnonymizationChanges(nil)
	if !strings.Contains(output, "No semantic PII") {
		t.Error("expected 'No semantic PII' for empty changes")
	}
}

func TestFormatAnonymizationChangesNoLine(t *testing.T) {
	changes := []AnonymizationChange{
		{Original: "Widget", Replaced: "<PROJECT>", Category: "project", Line: 0},
	}

	output := FormatAnonymizationChanges(changes)
	// Should not contain "Line 0"
	if strings.Contains(output, "Line 0") {
		t.Error("should not show Line 0 for unknown line")
	}
	if !strings.Contains(output, "[project]") {
		t.Error("expected category in output")
	}
}

func TestOllamaClientGenerate(t *testing.T) {
	// Mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/generate" {
			t.Errorf("expected /api/generate, got %s", r.URL.Path)
		}

		var req ollamaRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Stream != false {
			t.Error("expected stream=false")
		}
		if req.Model == "" {
			t.Error("expected non-empty model")
		}

		resp := ollamaResponse{
			Response: "Cleaned output\n---CHANGES---\n[]",
			Done:     true,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "test-model")
	result, err := client.Generate("test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Cleaned output") {
		t.Errorf("expected 'Cleaned output', got: %s", result)
	}
}

func TestOllamaClientServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("model not found"))
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "nonexistent-model")
	_, err := client.Generate("test prompt")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected status 500 in error, got: %v", err)
	}
}

func TestOllamaClientConnectionError(t *testing.T) {
	client := NewOllamaClient("http://localhost:99999", "test-model")
	_, err := client.Generate("test prompt")
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
	if !strings.Contains(err.Error(), "failed to connect") {
		t.Errorf("expected connection error, got: %v", err)
	}
}

func TestNewLLMClientOllama(t *testing.T) {
	client, err := NewLLMClient("ollama", "llama3.2", "http://localhost:11434")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	ollamaClient, ok := client.(*OllamaClient)
	if !ok {
		t.Fatal("expected OllamaClient type")
	}
	if ollamaClient.Model != "llama3.2" {
		t.Errorf("expected model 'llama3.2', got %s", ollamaClient.Model)
	}
}

func TestNewLLMClientUnsupported(t *testing.T) {
	_, err := NewLLMClient("unsupported", "", "")
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected 'unsupported' in error, got: %v", err)
	}
}

func TestContentHash(t *testing.T) {
	hash1 := contentHash("hello")
	hash2 := contentHash("hello")
	hash3 := contentHash("world")

	if hash1 != hash2 {
		t.Error("same content should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("different content should produce different hash")
	}
	if len(hash1) != 64 {
		t.Errorf("expected 64-char hex hash, got %d chars", len(hash1))
	}
}

func TestParseLLMResponseMalformedJSON(t *testing.T) {
	response := "Cleaned content\n---CHANGES---\nnot valid json"
	cleaned, changes := parseLLMResponse(response, "original")

	if cleaned != "Cleaned content" {
		t.Errorf("expected 'Cleaned content', got: %s", cleaned)
	}
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for malformed JSON, got %d", len(changes))
	}
}

func TestParseLLMResponseNoSeparator(t *testing.T) {
	response := "Just cleaned text without separator"
	cleaned, changes := parseLLMResponse(response, "original")

	if cleaned != response {
		t.Errorf("expected response as-is, got: %s", cleaned)
	}
	if len(changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(changes))
	}
}

func TestSemanticAnonymizerEndToEnd(t *testing.T) {
	// Simulate a full pattern anonymization flow
	mock := &mockLLMClient{
		response: `<COMPANY> Error Handler
Standard error handling pattern
---CHANGES---
[{"original":"Acme Corp","replaced":"<COMPANY>","category":"company"}]`,
	}

	anonymizer := NewSemanticAnonymizer(mock, "")
	content := "Acme Corp Error Handler\nStandard error handling pattern"

	cleaned, changes, err := anonymizer.Anonymize(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it works in the pattern split flow
	parts := strings.SplitN(cleaned, "\n", 3)
	if len(parts) < 2 {
		t.Fatalf("expected at least 2 parts, got %d", len(parts))
	}

	name := parts[0]
	desc := parts[1]

	if strings.Contains(name, "Acme Corp") {
		t.Error("expected Acme Corp to be removed from name")
	}
	if !strings.Contains(name, "<COMPANY>") {
		t.Error("expected <COMPANY> in name")
	}
	if desc != "Standard error handling pattern" {
		t.Errorf("expected description unchanged, got: %s", desc)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Line != 1 {
		t.Errorf("expected change on line 1, got %d", changes[0].Line)
	}
}

func TestFindLineNumber(t *testing.T) {
	content := "line one\nline two has target\nline three"

	if got := findLineNumber(content, "target"); got != 2 {
		t.Errorf("expected line 2, got %d", got)
	}
	if got := findLineNumber(content, "nonexistent"); got != 0 {
		t.Errorf("expected line 0 for missing needle, got %d", got)
	}
	if got := findLineNumber(content, "line one"); got != 1 {
		t.Errorf("expected line 1, got %d", got)
	}
}
