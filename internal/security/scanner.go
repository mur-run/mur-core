// Package security provides secret scanning to prevent accidental leakage.
package security

import (
	"regexp"
	"strings"
)

// Finding represents a detected secret.
type Finding struct {
	Type    string `json:"type"`
	Line    int    `json:"line"`
	Match   string `json:"match"`   // redacted
	Message string `json:"message"`
}

// ScanResult contains the results of a secret scan.
type ScanResult struct {
	Safe     bool      `json:"safe"`
	Findings []Finding `json:"findings,omitempty"`
}

// Scanner detects secrets in content.
type Scanner struct {
	rules []secretRule
}

type secretRule struct {
	id          string
	description string
	pattern     *regexp.Regexp
}

// NewScanner creates a new secret scanner with common patterns.
// Uses pure Go regex patterns (no CGO required).
func NewScanner() *Scanner {
	rules := []secretRule{
		// AWS
		{
			id:          "aws-access-key-id",
			description: "AWS Access Key ID",
			pattern:     regexp.MustCompile(`(?i)(AKIA|A3T|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}`),
		},
		{
			id:          "aws-secret-access-key",
			description: "AWS Secret Access Key",
			pattern:     regexp.MustCompile(`(?i)aws_secret_access_key\s*[=:]\s*['"]?([A-Za-z0-9/+=]{40})['"]?`),
		},
		// GitHub
		{
			id:          "github-token",
			description: "GitHub Personal Access Token",
			pattern:     regexp.MustCompile(`ghp_[A-Za-z0-9]{36}`),
		},
		{
			id:          "github-fine-grained-token",
			description: "GitHub Fine-Grained PAT",
			pattern:     regexp.MustCompile(`github_pat_[A-Za-z0-9]{22}_[A-Za-z0-9]{59}`),
		},
		{
			id:          "github-oauth",
			description: "GitHub OAuth Token",
			pattern:     regexp.MustCompile(`gho_[A-Za-z0-9]{36}`),
		},
		{
			id:          "github-app-token",
			description: "GitHub App Token",
			pattern:     regexp.MustCompile(`(ghu|ghs)_[A-Za-z0-9]{36}`),
		},
		// GitLab
		{
			id:          "gitlab-token",
			description: "GitLab Personal Access Token",
			pattern:     regexp.MustCompile(`glpat-[A-Za-z0-9\-]{20}`),
		},
		// Slack
		{
			id:          "slack-token",
			description: "Slack Token",
			pattern:     regexp.MustCompile(`xox[baprs]-[A-Za-z0-9\-]{10,}`),
		},
		{
			id:          "slack-webhook",
			description: "Slack Webhook URL",
			pattern:     regexp.MustCompile(`https://hooks\.slack\.com/services/T[A-Z0-9]+/B[A-Z0-9]+/[A-Za-z0-9]+`),
		},
		// Discord
		{
			id:          "discord-token",
			description: "Discord Bot Token",
			pattern:     regexp.MustCompile(`[MN][A-Za-z0-9]{23,}\.[A-Za-z0-9-_]{6}\.[A-Za-z0-9-_]{27}`),
		},
		{
			id:          "discord-webhook",
			description: "Discord Webhook URL",
			pattern:     regexp.MustCompile(`https://discord(app)?\.com/api/webhooks/[0-9]+/[A-Za-z0-9_-]+`),
		},
		// OpenAI
		{
			id:          "openai-api-key",
			description: "OpenAI API Key",
			pattern:     regexp.MustCompile(`sk-[A-Za-z0-9]{20}T3BlbkFJ[A-Za-z0-9]{20}`),
		},
		{
			id:          "openai-api-key-new",
			description: "OpenAI API Key (new format)",
			pattern:     regexp.MustCompile(`sk-proj-[A-Za-z0-9_-]{80,}`),
		},
		// Anthropic
		{
			id:          "anthropic-api-key",
			description: "Anthropic API Key",
			pattern:     regexp.MustCompile(`sk-ant-[A-Za-z0-9\-]{90,}`),
		},
		// Google
		{
			id:          "google-api-key",
			description: "Google API Key",
			pattern:     regexp.MustCompile(`AIza[A-Za-z0-9_\-]{35}`),
		},
		// Stripe
		{
			id:          "stripe-api-key",
			description: "Stripe API Key",
			pattern:     regexp.MustCompile(`(sk|pk)_(test|live)_[A-Za-z0-9]{24,}`),
		},
		// SendGrid
		{
			id:          "sendgrid-api-key",
			description: "SendGrid API Key",
			pattern:     regexp.MustCompile(`SG\.[A-Za-z0-9_\-]{22}\.[A-Za-z0-9_\-]{43}`),
		},
		// Twilio
		{
			id:          "twilio-api-key",
			description: "Twilio API Key",
			pattern:     regexp.MustCompile(`SK[A-Za-z0-9]{32}`),
		},
		// npm
		{
			id:          "npm-token",
			description: "npm Access Token",
			pattern:     regexp.MustCompile(`npm_[A-Za-z0-9]{36}`),
		},
		// PyPI
		{
			id:          "pypi-token",
			description: "PyPI API Token",
			pattern:     regexp.MustCompile(`pypi-AgEIcHlwaS5vcmc[A-Za-z0-9\-_]{50,}`),
		},
		// Private Keys
		{
			id:          "private-key",
			description: "Private Key",
			pattern:     regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH |PGP )?PRIVATE KEY( BLOCK)?-----`),
		},
		// Generic patterns
		{
			id:          "generic-api-key",
			description: "Generic API Key",
			pattern:     regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[=:]\s*['"]?([A-Za-z0-9_\-]{20,})['"]?`),
		},
		{
			id:          "generic-secret",
			description: "Generic Secret",
			pattern:     regexp.MustCompile(`(?i)(secret|password|passwd|pwd)\s*[=:]\s*['"]?([A-Za-z0-9_\-!@#$%^&*]{8,})['"]?`),
		},
		{
			id:          "generic-token",
			description: "Generic Token",
			pattern:     regexp.MustCompile(`(?i)(auth[_-]?token|access[_-]?token|bearer)\s*[=:]\s*['"]?([A-Za-z0-9_\-]{20,})['"]?`),
		},
		// Connection strings
		{
			id:          "connection-string",
			description: "Database Connection String",
			pattern:     regexp.MustCompile(`(?i)(mongodb|postgres|mysql|redis|amqp)://[A-Za-z0-9_\-]+:[A-Za-z0-9_\-!@#$%^&*]+@`),
		},
		// JWT
		{
			id:          "jwt-token",
			description: "JWT Token",
			pattern:     regexp.MustCompile(`eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}`),
		},
	}

	return &Scanner{rules: rules}
}

// ScanContent scans content for secrets and returns findings.
func (s *Scanner) ScanContent(content string) *ScanResult {
	lines := strings.Split(content, "\n")
	var findings []Finding

	for lineNum, line := range lines {
		for _, rule := range s.rules {
			if matches := rule.pattern.FindStringSubmatch(line); len(matches) > 0 {
				// Get the full match or first captured group
				matchStr := matches[0]
				if len(matches) > 1 && matches[1] != "" {
					// For patterns with capture groups, use the captured secret
					for i := len(matches) - 1; i >= 1; i-- {
						if matches[i] != "" && len(matches[i]) > 8 {
							matchStr = matches[i]
							break
						}
					}
				}

				findings = append(findings, Finding{
					Type:    rule.id,
					Line:    lineNum + 1, // 1-indexed
					Match:   redact(matchStr),
					Message: rule.description,
				})
			}
		}
	}

	if len(findings) == 0 {
		return &ScanResult{Safe: true}
	}

	return &ScanResult{
		Safe:     false,
		Findings: findings,
	}
}

// redact masks a secret, showing only first and last 4 chars.
func redact(s string) string {
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "..." + s[len(s)-4:]
}
