package security

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mur-run/mur-core/internal/config"
)

// PIIFinding represents a detected PII match.
type PIIFinding struct {
	Type     string `json:"type"`     // "email", "internal-ip", "user-path", "internal-url", "phone", "blocklist"
	Original string `json:"original"` // the matched text
	Replaced string `json:"replaced"` // what it becomes
	Line     int    `json:"line"`     // 1-indexed line number
}

// PIIScanner detects and redacts PII from content.
type PIIScanner struct {
	rules        []piiRule
	blocklist    []string
	replacements map[string]string
}

type piiRule struct {
	id          string
	label       string
	pattern     *regexp.Regexp
	replacement string
}

// NewPIIScanner creates a PIIScanner from config.
func NewPIIScanner(cfg config.PrivacyConfig) *PIIScanner {
	var rules []piiRule
	ad := cfg.AutoDetect

	if ad.IsEmailsEnabled() {
		rules = append(rules, piiRule{
			id:          "email",
			label:       "Email address",
			pattern:     regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`),
			replacement: "<REDACTED_EMAIL>",
		})
	}

	if ad.IsInternalIPsEnabled() {
		rules = append(rules, piiRule{
			id:          "internal-ip",
			label:       "Internal IP address",
			pattern:     regexp.MustCompile(`\b(10\.\d{1,3}\.\d{1,3}\.\d{1,3}|172\.(1[6-9]|2\d|3[01])\.\d{1,3}\.\d{1,3}|192\.168\.\d{1,3}\.\d{1,3})\b`),
			replacement: "<REDACTED_IP>",
		})
	}

	if ad.IsFilePathsEnabled() {
		rules = append(rules, piiRule{
			id:          "user-path",
			label:       "User file path",
			pattern:     regexp.MustCompile(`(/Users/|/home/|C:\\Users\\)[a-zA-Z0-9._\-]+`),
			replacement: "<REDACTED_PATH>",
		})
	}

	if ad.IsInternalURLsEnabled() {
		rules = append(rules, piiRule{
			id:          "internal-url",
			label:       "Internal URL",
			pattern:     regexp.MustCompile(`https?://[a-z0-9.\-]*(\.local|\.internal|\.corp|\.lan|localhost)[:/]?[^\s]*`),
			replacement: "<REDACTED_URL>",
		})
	}

	if ad.IsPhoneNumbersEnabled() {
		rules = append(rules, piiRule{
			id:          "phone",
			label:       "Phone number",
			pattern:     regexp.MustCompile(`(\+?\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3,4}[-.\s]?\d{4}`),
			replacement: "<REDACTED_PHONE>",
		})
	}

	return &PIIScanner{
		rules:        rules,
		blocklist:    cfg.RedactTerms,
		replacements: cfg.Replacements,
	}
}

// ScanAndRedact scans content for PII, applies redactions, and returns cleaned content with findings.
func (s *PIIScanner) ScanAndRedact(content string) (string, []PIIFinding) {
	var findings []PIIFinding
	lines := strings.Split(content, "\n")

	// Apply custom replacements first (exact string match)
	for original, replacement := range s.replacements {
		for i, line := range lines {
			if strings.Contains(line, original) {
				findings = append(findings, PIIFinding{
					Type:     "replacement",
					Original: original,
					Replaced: replacement,
					Line:     i + 1,
				})
				lines[i] = strings.ReplaceAll(lines[i], original, replacement)
			}
		}
	}

	// Apply blocklist terms (case-insensitive)
	for _, term := range s.blocklist {
		termLower := strings.ToLower(term)
		replacement := fmt.Sprintf("<REDACTED_%s>", strings.ToUpper(strings.ReplaceAll(term, " ", "_")))
		for i, line := range lines {
			lineLower := strings.ToLower(line)
			if strings.Contains(lineLower, termLower) {
				findings = append(findings, PIIFinding{
					Type:     "blocklist",
					Original: term,
					Replaced: replacement,
					Line:     i + 1,
				})
				// Case-insensitive replacement
				lines[i] = caseInsensitiveReplace(lines[i], term, replacement)
			}
		}
	}

	// Apply regex rules
	for _, rule := range s.rules {
		for i, line := range lines {
			matches := rule.pattern.FindAllString(line, -1)
			for _, match := range matches {
				findings = append(findings, PIIFinding{
					Type:     rule.id,
					Original: match,
					Replaced: rule.replacement,
					Line:     i + 1,
				})
			}
			lines[i] = rule.pattern.ReplaceAllString(lines[i], rule.replacement)
		}
	}

	return strings.Join(lines, "\n"), findings
}

// caseInsensitiveReplace replaces all occurrences of old in s, case-insensitively.
func caseInsensitiveReplace(s, old, replacement string) string {
	lower := strings.ToLower(s)
	oldLower := strings.ToLower(old)

	var result strings.Builder
	start := 0
	for {
		idx := strings.Index(lower[start:], oldLower)
		if idx < 0 {
			result.WriteString(s[start:])
			break
		}
		result.WriteString(s[start : start+idx])
		result.WriteString(replacement)
		start += idx + len(old)
	}
	return result.String()
}

// FormatFindings returns a human-readable summary of PII findings.
func FormatFindings(findings []PIIFinding) string {
	if len(findings) == 0 {
		return "  No PII detected."
	}

	var b strings.Builder
	for _, f := range findings {
		fmt.Fprintf(&b, "  Line %d: [%s] %q → %s\n", f.Line, f.Type, f.Original, f.Replaced)
	}
	return b.String()
}
