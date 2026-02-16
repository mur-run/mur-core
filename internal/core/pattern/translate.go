package pattern

import (
	"unicode"
)

// ContainsNonASCII detects if text has non-ASCII characters
// (indicates non-English content that may need translation)
func ContainsNonASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return true
		}
	}
	return false
}

// ContainsCJK detects if text contains CJK (Chinese/Japanese/Korean) characters
func ContainsCJK(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) || // Chinese
			unicode.Is(unicode.Hiragana, r) || // Japanese
			unicode.Is(unicode.Katakana, r) || // Japanese
			unicode.Is(unicode.Hangul, r) { // Korean
			return true
		}
	}
	return false
}

// NeedsTranslation checks if pattern content should be translated before community sharing
// Returns true if name, description, or content contains non-English text
func NeedsTranslation(p *Pattern) bool {
	// Check name first (should be kebab-case English)
	if ContainsCJK(p.Name) {
		return true
	}

	// Check description
	if ContainsCJK(p.Description) {
		return true
	}

	// Check content (allow some non-ASCII like code symbols, but flag CJK)
	if ContainsCJK(p.Content) {
		return true
	}

	return false
}

// TranslationPrompt generates a prompt for translating a pattern to English
func TranslationPrompt(p *Pattern) string {
	return `Translate this pattern to English. Keep these unchanged:
- Code snippets (between ` + "```" + ` blocks)
- Error messages (keep original + add translation)
- Technical terms and library names
- URLs and file paths

Pattern to translate:

Name: ` + p.Name + `
Description: ` + p.Description + `
Content:
` + p.Content + `

Output as YAML:
name: (kebab-case English name)
description: (one-line English description)
content: |
  (translated content in English)
`
}
