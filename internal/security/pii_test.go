package security

import (
	"strings"
	"testing"

	"github.com/mur-run/mur-core/internal/config"
)

func defaultPIIConfig() config.PrivacyConfig {
	return config.DefaultPrivacyConfig()
}

func TestPIIScannerEmailDetection(t *testing.T) {
	scanner := NewPIIScanner(defaultPIIConfig())

	tests := []struct {
		name    string
		content string
		want    bool // true = should detect
	}{
		{"simple email", "contact user@example.com for help", true},
		{"email with plus", "send to user+tag@company.io", true},
		{"email with dots", "first.last@sub.domain.org", true},
		{"no email", "this is just normal text", false},
		{"already redacted", "contact <REDACTED_EMAIL> for help", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, findings := scanner.ScanAndRedact(tt.content)
			found := false
			for _, f := range findings {
				if f.Type == "email" {
					found = true
				}
			}
			if found != tt.want {
				t.Errorf("email detection = %v, want %v", found, tt.want)
			}
		})
	}
}

func TestPIIScannerEmailRedaction(t *testing.T) {
	scanner := NewPIIScanner(defaultPIIConfig())

	content := "Contact admin@company.com for support"
	cleaned, findings := scanner.ScanAndRedact(content)

	if !strings.Contains(cleaned, "<REDACTED_EMAIL>") {
		t.Errorf("expected email to be redacted, got: %s", cleaned)
	}
	if strings.Contains(cleaned, "admin@company.com") {
		t.Errorf("expected original email to be removed, got: %s", cleaned)
	}
	if len(findings) == 0 {
		t.Error("expected at least one finding")
	}
	if findings[0].Type != "email" {
		t.Errorf("expected email finding, got: %s", findings[0].Type)
	}
	if findings[0].Original != "admin@company.com" {
		t.Errorf("expected original to be 'admin@company.com', got: %s", findings[0].Original)
	}
}

func TestPIIScannerInternalIPDetection(t *testing.T) {
	scanner := NewPIIScanner(defaultPIIConfig())

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"10.x.x.x", "server at 10.0.1.5 is down", true},
		{"172.16.x.x", "connect to 172.16.0.1", true},
		{"172.31.x.x", "proxy at 172.31.255.255", true},
		{"192.168.x.x", "gateway is 192.168.1.1", true},
		{"public IP", "server at 8.8.8.8 is up", false},
		{"172.15 not private", "host 172.15.0.1 ok", false},
		{"172.32 not private", "host 172.32.0.1 ok", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, findings := scanner.ScanAndRedact(tt.content)
			found := false
			for _, f := range findings {
				if f.Type == "internal-ip" {
					found = true
				}
			}
			if found != tt.want {
				t.Errorf("IP detection = %v, want %v (content: %s)", found, tt.want, tt.content)
			}
		})
	}
}

func TestPIIScannerIPRedaction(t *testing.T) {
	scanner := NewPIIScanner(defaultPIIConfig())

	content := "Connect to 192.168.1.100 on port 8080"
	cleaned, _ := scanner.ScanAndRedact(content)

	if !strings.Contains(cleaned, "<REDACTED_IP>") {
		t.Errorf("expected IP to be redacted, got: %s", cleaned)
	}
	if strings.Contains(cleaned, "192.168.1.100") {
		t.Errorf("expected original IP to be removed, got: %s", cleaned)
	}
}

func TestPIIScannerUserPathDetection(t *testing.T) {
	scanner := NewPIIScanner(defaultPIIConfig())

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"macOS path", "file at /Users/john/project/main.go", true},
		{"Linux path", "config in /home/alice/.config", true},
		{"Windows path", `located at C:\Users\bob\Documents`, true},
		{"system path", "binary at /usr/local/bin/tool", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, findings := scanner.ScanAndRedact(tt.content)
			found := false
			for _, f := range findings {
				if f.Type == "user-path" {
					found = true
				}
			}
			if found != tt.want {
				t.Errorf("path detection = %v, want %v (content: %s)", found, tt.want, tt.content)
			}
		})
	}
}

func TestPIIScannerPathRedaction(t *testing.T) {
	scanner := NewPIIScanner(defaultPIIConfig())

	content := "Error in /Users/david/projects/app/main.go"
	cleaned, _ := scanner.ScanAndRedact(content)

	if !strings.Contains(cleaned, "<REDACTED_PATH>") {
		t.Errorf("expected path to be redacted, got: %s", cleaned)
	}
}

func TestPIIScannerInternalURLDetection(t *testing.T) {
	scanner := NewPIIScanner(defaultPIIConfig())

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"localhost", "api at http://localhost:3000/api", true},
		{".local domain", "connect to https://myapp.local/health", true},
		{".internal domain", "use http://service.internal:8080", true},
		{".corp domain", "visit https://jira.corp/browse/PROJ-1", true},
		{".lan domain", "at http://printer.lan", true},
		{"public URL", "visit https://github.com/org/repo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, findings := scanner.ScanAndRedact(tt.content)
			found := false
			for _, f := range findings {
				if f.Type == "internal-url" {
					found = true
				}
			}
			if found != tt.want {
				t.Errorf("URL detection = %v, want %v (content: %s)", found, tt.want, tt.content)
			}
		})
	}
}

func TestPIIScannerPhoneDetection(t *testing.T) {
	scanner := NewPIIScanner(defaultPIIConfig())

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"US format", "call 555-123-4567", true},
		{"with parens", "phone: (555) 123-4567", true},
		{"with country code", "reach +1-555-123-4567", true},
		{"dots", "fax: 555.123.4567", true},
		{"no phone", "version 1.2.3 is released", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, findings := scanner.ScanAndRedact(tt.content)
			found := false
			for _, f := range findings {
				if f.Type == "phone" {
					found = true
				}
			}
			if found != tt.want {
				t.Errorf("phone detection = %v, want %v (content: %s)", found, tt.want, tt.content)
			}
		})
	}
}

func TestPIIScannerBlocklist(t *testing.T) {
	cfg := defaultPIIConfig()
	cfg.RedactTerms = []string{"acme-corp", "ProjectX"}
	scanner := NewPIIScanner(cfg)

	content := "The acme-corp ProjectX deployment uses custom config"
	cleaned, findings := scanner.ScanAndRedact(content)

	// Check findings
	blocklistFindings := 0
	for _, f := range findings {
		if f.Type == "blocklist" {
			blocklistFindings++
		}
	}
	if blocklistFindings != 2 {
		t.Errorf("expected 2 blocklist findings, got %d", blocklistFindings)
	}

	// Check redaction
	if strings.Contains(cleaned, "acme-corp") {
		t.Errorf("expected 'acme-corp' to be redacted, got: %s", cleaned)
	}
	if strings.Contains(cleaned, "ProjectX") {
		t.Errorf("expected 'ProjectX' to be redacted, got: %s", cleaned)
	}
}

func TestPIIScannerBlocklistCaseInsensitive(t *testing.T) {
	cfg := defaultPIIConfig()
	cfg.RedactTerms = []string{"acme"}
	scanner := NewPIIScanner(cfg)

	content := "ACME Corp and Acme Inc and acme ltd"
	cleaned, findings := scanner.ScanAndRedact(content)

	// The cleaned text should contain the redaction tag but NOT the original term outside of tags
	// Remove all redaction tags then check for the original term
	withoutTags := strings.ReplaceAll(cleaned, "<REDACTED_ACME>", "")
	if strings.Contains(strings.ToLower(withoutTags), "acme") {
		t.Errorf("expected all case variants of 'acme' to be redacted, got: %s", cleaned)
	}

	blocklistFindings := 0
	for _, f := range findings {
		if f.Type == "blocklist" {
			blocklistFindings++
		}
	}
	if blocklistFindings == 0 {
		t.Error("expected blocklist findings for case-insensitive match")
	}
}

func TestPIIScannerCustomReplacements(t *testing.T) {
	cfg := defaultPIIConfig()
	cfg.Replacements = map[string]string{
		"company.com": "<COMPANY_DOMAIN>",
		"ProjectX":    "<PROJECT>",
	}
	scanner := NewPIIScanner(cfg)

	content := "Deploy to company.com for ProjectX"
	cleaned, findings := scanner.ScanAndRedact(content)

	if !strings.Contains(cleaned, "<COMPANY_DOMAIN>") {
		t.Errorf("expected 'company.com' replacement, got: %s", cleaned)
	}
	if !strings.Contains(cleaned, "<PROJECT>") {
		t.Errorf("expected 'ProjectX' replacement, got: %s", cleaned)
	}

	replacementFindings := 0
	for _, f := range findings {
		if f.Type == "replacement" {
			replacementFindings++
		}
	}
	if replacementFindings != 2 {
		t.Errorf("expected 2 replacement findings, got %d", replacementFindings)
	}
}

func TestPIIScannerDisabledRules(t *testing.T) {
	cfg := defaultPIIConfig()
	f := false
	cfg.AutoDetect.Emails = &f
	cfg.AutoDetect.PhoneNumbers = &f
	scanner := NewPIIScanner(cfg)

	content := "Contact admin@example.com or call 555-123-4567"
	_, findings := scanner.ScanAndRedact(content)

	for _, finding := range findings {
		if finding.Type == "email" || finding.Type == "phone" {
			t.Errorf("expected disabled rules to not trigger, got: %s", finding.Type)
		}
	}
}

func TestPIIScannerMultiline(t *testing.T) {
	scanner := NewPIIScanner(defaultPIIConfig())

	content := `# Configuration
server: 192.168.1.1
contact: admin@internal.corp
path: /Users/developer/project
phone: 555-867-5309`

	cleaned, findings := scanner.ScanAndRedact(content)

	if len(findings) < 4 {
		t.Errorf("expected at least 4 findings, got %d", len(findings))
	}

	// Verify line numbers
	for _, f := range findings {
		if f.Line < 1 {
			t.Errorf("expected positive line number, got %d", f.Line)
		}
	}

	// Verify all PII is redacted
	if strings.Contains(cleaned, "192.168.1.1") {
		t.Error("IP should be redacted")
	}
	if strings.Contains(cleaned, "admin@internal.corp") {
		t.Error("email should be redacted")
	}
	if strings.Contains(cleaned, "/Users/developer") {
		t.Error("path should be redacted")
	}
	if strings.Contains(cleaned, "555-867-5309") {
		t.Error("phone should be redacted")
	}
}

func TestPIIScannerCleanContent(t *testing.T) {
	scanner := NewPIIScanner(defaultPIIConfig())

	content := `Use errors.Is and errors.As for error checking.
Wrap errors with context using fmt.Errorf with %w verb.
Always return early on error.`

	cleaned, findings := scanner.ScanAndRedact(content)

	if len(findings) != 0 {
		t.Errorf("expected no findings for clean content, got %d: %v", len(findings), findings)
	}
	if cleaned != content {
		t.Errorf("clean content should be unchanged")
	}
}

func TestPIIScannerFullShareFlow(t *testing.T) {
	cfg := defaultPIIConfig()
	cfg.RedactTerms = []string{"acme-corp"}
	cfg.Replacements = map[string]string{
		"internal.acme.com": "<COMPANY_DOMAIN>",
	}
	scanner := NewPIIScanner(cfg)

	// Simulate a pattern with mixed PII
	content := `name: acme-corp error handler
description: Error handling at acme-corp
content: |
  Contact admin@acme-corp.com if errors persist.
  Dashboard: https://monitoring.internal/errors
  Server: 192.168.1.50
  Config: /Users/developer/acme-corp/config.yaml
  Docs: internal.acme.com/docs`

	cleaned, findings := scanner.ScanAndRedact(content)

	if len(findings) == 0 {
		t.Error("expected findings in mixed PII content")
	}

	// Verify redactions
	if strings.Contains(cleaned, "admin@acme-corp.com") {
		t.Error("email should be redacted")
	}
	if strings.Contains(cleaned, "192.168.1.50") {
		t.Error("IP should be redacted")
	}
	if strings.Contains(cleaned, "/Users/developer") {
		t.Error("path should be redacted")
	}
	if strings.Contains(cleaned, "internal.acme.com") {
		t.Error("custom replacement domain should be redacted")
	}

	// Verify replacements are present
	if !strings.Contains(cleaned, "<COMPANY_DOMAIN>") {
		t.Error("expected custom replacement to be present")
	}
}

func TestFormatFindings(t *testing.T) {
	findings := []PIIFinding{
		{Type: "email", Original: "test@example.com", Replaced: "<REDACTED_EMAIL>", Line: 3},
		{Type: "internal-ip", Original: "10.0.0.1", Replaced: "<REDACTED_IP>", Line: 5},
	}

	output := FormatFindings(findings)
	if !strings.Contains(output, "email") {
		t.Error("expected email type in output")
	}
	if !strings.Contains(output, "Line 3") {
		t.Error("expected line number in output")
	}

	// Empty findings
	output = FormatFindings(nil)
	if !strings.Contains(output, "No PII detected") {
		t.Error("expected 'No PII detected' for empty findings")
	}
}

func TestCaseInsensitiveReplace(t *testing.T) {
	tests := []struct {
		input       string
		old         string
		replacement string
		want        string
	}{
		{"Hello WORLD hello", "hello", "HI", "HI WORLD HI"},
		{"FooBarFoo", "foo", "X", "XBarX"},
		{"no match here", "xyz", "Y", "no match here"},
		{"", "test", "X", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := caseInsensitiveReplace(tt.input, tt.old, tt.replacement)
			if got != tt.want {
				t.Errorf("caseInsensitiveReplace(%q, %q, %q) = %q, want %q",
					tt.input, tt.old, tt.replacement, got, tt.want)
			}
		})
	}
}
