package security

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// InjectionRisk represents the risk level of an injection finding.
type InjectionRisk string

const (
	InjectionRiskLow    InjectionRisk = "low"
	InjectionRiskMedium InjectionRisk = "medium"
	InjectionRiskHigh   InjectionRisk = "high"
)

// InjectionFinding represents a detected injection pattern.
type InjectionFinding struct {
	Rule    string        `json:"rule"`
	Risk    InjectionRisk `json:"risk"`
	Line    int           `json:"line"`
	Match   string        `json:"match"`
	Message string        `json:"message"`
}

// InjectionResult contains the results of an injection scan.
type InjectionResult struct {
	Risk     InjectionRisk      `json:"risk"`
	Findings []InjectionFinding `json:"findings,omitempty"`
}

// InjectionScanner detects prompt injection patterns in content.
type InjectionScanner struct {
	rules []injectionRule
}

type injectionRule struct {
	id      string
	risk    InjectionRisk
	message string
	pattern *regexp.Regexp
	// custom check function for non-regex detection
	check func(line string) (bool, string)
}

// NewInjectionScanner creates a new injection scanner with common patterns.
func NewInjectionScanner() *InjectionScanner {
	rules := []injectionRule{
		// High risk: direct instruction override
		{
			id:      "ignore-instructions",
			risk:    InjectionRiskHigh,
			message: "Instruction override attempt",
			pattern: regexp.MustCompile(`(?i)(ignore|disregard|forget|override)\s+(all\s+)?(previous|prior|above|earlier|existing)\s+(instructions|rules|guidelines|constraints|prompts)`),
		},
		{
			id:      "new-instructions",
			risk:    InjectionRiskHigh,
			message: "New instruction injection",
			pattern: regexp.MustCompile(`(?i)(new|updated|revised|real)\s+(instructions|rules|guidelines|system\s+prompt)\s*[:=]`),
		},

		// High risk: role markers that could hijack context
		{
			id:      "role-marker-system",
			risk:    InjectionRiskHigh,
			message: "System role marker detected",
			pattern: regexp.MustCompile(`(?m)^(system|SYSTEM)\s*:\s*`),
		},
		{
			id:      "role-marker-assistant",
			risk:    InjectionRiskMedium,
			message: "Assistant role marker detected",
			pattern: regexp.MustCompile(`(?m)^(assistant|ASSISTANT)\s*:\s*`),
		},
		{
			id:      "role-marker-user",
			risk:    InjectionRiskMedium,
			message: "User role marker detected",
			pattern: regexp.MustCompile(`(?m)^(user|USER|human|HUMAN)\s*:\s*`),
		},

		// High risk: identity/role hijacking
		{
			id:      "role-hijack",
			risk:    InjectionRiskHigh,
			message: "Role hijacking attempt",
			pattern: regexp.MustCompile(`(?i)(you\s+are\s+now|act\s+as|pretend\s+(to\s+be|you\s+are)|from\s+now\s+on\s+you\s+are|your\s+new\s+role\s+is)`),
		},

		// Medium risk: base64 blocks (could hide instructions)
		{
			id:      "base64-block",
			risk:    InjectionRiskMedium,
			message: "Base64-encoded block detected (may hide instructions)",
			check:   checkBase64Block,
		},

		// Medium risk: markdown/HTML comment injection
		{
			id:      "html-comment",
			risk:    InjectionRiskMedium,
			message: "HTML comment with instructions detected",
			pattern: regexp.MustCompile(`<!--\s*(?i)(system|instruction|ignore|override|prompt|role).*?-->`),
		},
		{
			id:      "markdown-comment",
			risk:    InjectionRiskMedium,
			message: "Hidden markdown content detected",
			pattern: regexp.MustCompile(`\[//\]:\s*#\s*\(.*(?i)(ignore|system|instruction|override).*\)`),
		},

		// Low risk: unicode homoglyph obfuscation
		{
			id:      "unicode-homoglyph",
			risk:    InjectionRiskLow,
			message: "Unicode homoglyph obfuscation detected",
			check:   checkHomoglyphs,
		},

		// Medium risk: delimiter injection
		{
			id:      "delimiter-injection",
			risk:    InjectionRiskMedium,
			message: "Context delimiter injection",
			pattern: regexp.MustCompile(`(?i)(</?(system|context|instructions|prompt)>|---\s*(system|end\s+of\s+instructions)\s*---)`),
		},

		// Low risk: suspicious jailbreak keywords
		{
			id:      "jailbreak-keywords",
			risk:    InjectionRiskLow,
			message: "Potential jailbreak keywords detected",
			pattern: regexp.MustCompile(`(?i)(DAN\s+mode|jailbreak|bypass\s+(safety|filter|restriction)|do\s+anything\s+now)`),
		},
	}

	return &InjectionScanner{rules: rules}
}

// Scan scans content for injection patterns and returns the overall risk and findings.
func (s *InjectionScanner) Scan(content string) (InjectionRisk, []InjectionFinding) {
	lines := strings.Split(content, "\n")
	var findings []InjectionFinding

	for lineNum, line := range lines {
		for _, rule := range s.rules {
			if rule.pattern != nil {
				if matches := rule.pattern.FindString(line); matches != "" {
					findings = append(findings, InjectionFinding{
						Rule:    rule.id,
						Risk:    rule.risk,
						Line:    lineNum + 1,
						Match:   truncateMatch(matches),
						Message: rule.message,
					})
				}
			} else if rule.check != nil {
				if found, match := rule.check(line); found {
					findings = append(findings, InjectionFinding{
						Rule:    rule.id,
						Risk:    rule.risk,
						Line:    lineNum + 1,
						Match:   truncateMatch(match),
						Message: rule.message,
					})
				}
			}
		}
	}

	// Also check multi-line patterns against full content
	for _, rule := range s.rules {
		if rule.pattern != nil {
			// Check for multi-line HTML comments that span lines
			if rule.id == "html-comment" {
				if matches := rule.pattern.FindAllString(content, -1); len(matches) > 0 {
					// Only add if not already found line-by-line
					alreadyFound := false
					for _, f := range findings {
						if f.Rule == "html-comment" {
							alreadyFound = true
							break
						}
					}
					if !alreadyFound {
						for _, m := range matches {
							findings = append(findings, InjectionFinding{
								Rule:    rule.id,
								Risk:    rule.risk,
								Line:    0, // multi-line, no specific line
								Match:   truncateMatch(m),
								Message: rule.message,
							})
						}
					}
				}
			}
		}
	}

	// Determine overall risk
	overallRisk := determineOverallRisk(findings)
	return overallRisk, findings
}

// checkBase64Block checks if a line contains a suspicious base64-encoded block.
func checkBase64Block(line string) (bool, string) {
	// Look for long base64 strings (at least 40 chars to avoid false positives)
	b64Pattern := regexp.MustCompile(`[A-Za-z0-9+/]{40,}={0,2}`)
	matches := b64Pattern.FindAllString(line, -1)

	for _, match := range matches {
		decoded, err := base64.StdEncoding.DecodeString(match)
		if err != nil {
			// Try with padding
			padded := match
			for len(padded)%4 != 0 {
				padded += "="
			}
			decoded, err = base64.StdEncoding.DecodeString(padded)
			if err != nil {
				continue
			}
		}

		// Check if decoded content looks like text with suspicious keywords
		decodedStr := strings.ToLower(string(decoded))
		if isPrintableText(decoded) {
			suspicious := []string{"ignore", "system", "instruction", "override", "pretend", "act as", "you are"}
			for _, keyword := range suspicious {
				if strings.Contains(decodedStr, keyword) {
					return true, match
				}
			}
		}
	}

	return false, ""
}

// isPrintableText checks if bytes look like readable text.
func isPrintableText(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	printable := 0
	for _, b := range data {
		if b >= 32 && b < 127 || b == '\n' || b == '\r' || b == '\t' {
			printable++
		}
	}
	return float64(printable)/float64(len(data)) > 0.8
}

// Common homoglyph mappings (Cyrillic/Greek letters that look like Latin).
var homoglyphMap = map[rune]rune{
	'\u0410': 'A', // Cyrillic А
	'\u0412': 'B', // Cyrillic В
	'\u0421': 'C', // Cyrillic С
	'\u0415': 'E', // Cyrillic Е
	'\u041D': 'H', // Cyrillic Н
	'\u041A': 'K', // Cyrillic К
	'\u041C': 'M', // Cyrillic М
	'\u041E': 'O', // Cyrillic О
	'\u0420': 'P', // Cyrillic Р
	'\u0422': 'T', // Cyrillic Т
	'\u0425': 'X', // Cyrillic Х
	'\u0430': 'a', // Cyrillic а
	'\u0435': 'e', // Cyrillic е
	'\u043E': 'o', // Cyrillic о
	'\u0440': 'p', // Cyrillic р
	'\u0441': 'c', // Cyrillic с
	'\u0443': 'y', // Cyrillic у
	'\u0445': 'x', // Cyrillic х
	'\u0456': 'i', // Cyrillic і
	'\u0458': 'j', // Cyrillic ј
	'\u0405': 'S', // Cyrillic Ѕ
	'\u0455': 's', // Cyrillic ѕ
}

// checkHomoglyphs detects text containing mixed scripts that could be homoglyph obfuscation.
func checkHomoglyphs(line string) (bool, string) {
	hasLatin := false
	hasHomoglyph := false

	for _, r := range line {
		if unicode.Is(unicode.Latin, r) {
			hasLatin = true
		}
		if _, ok := homoglyphMap[r]; ok {
			hasHomoglyph = true
		}
	}

	if hasLatin && hasHomoglyph {
		return true, "mixed Latin/Cyrillic characters"
	}

	return false, ""
}

// determineOverallRisk returns the highest risk level from findings.
func determineOverallRisk(findings []InjectionFinding) InjectionRisk {
	if len(findings) == 0 {
		return InjectionRiskLow
	}

	highest := InjectionRiskLow
	for _, f := range findings {
		if f.Risk == InjectionRiskHigh {
			return InjectionRiskHigh
		}
		if f.Risk == InjectionRiskMedium {
			highest = InjectionRiskMedium
		}
	}
	return highest
}

// truncateMatch truncates a match string for display.
func truncateMatch(s string) string {
	if len(s) > 80 {
		return s[:80] + "..."
	}
	return s
}

// FormatInjectionFindings returns a human-readable summary of injection findings.
func FormatInjectionFindings(findings []InjectionFinding) string {
	if len(findings) == 0 {
		return "  No injection patterns detected."
	}

	var b strings.Builder
	for _, f := range findings {
		fmt.Fprintf(&b, "  Line %d: [%s] %s (%s)\n", f.Line, f.Risk, f.Message, f.Match)
	}
	return b.String()
}
