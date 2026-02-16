package pattern

import "testing"

func TestContainsNonASCII(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello world", false},
		{"Hello World 123", false},
		{"API Error Handling", false},
		{"你好世界", true},
		{"Hello 世界", true},
		{"日本語テスト", true},
		{"한국어", true},
		{"café", true}, // Non-ASCII but not CJK
	}

	for _, tt := range tests {
		result := ContainsNonASCII(tt.input)
		if result != tt.expected {
			t.Errorf("ContainsNonASCII(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestContainsCJK(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello world", false},
		{"café résumé", false}, // Non-ASCII but not CJK
		{"中文", true},
		{"日本語", true},
		{"한글", true},
		{"Hello 世界", true},
		{"API 錯誤處理", true},
	}

	for _, tt := range tests {
		result := ContainsCJK(tt.input)
		if result != tt.expected {
			t.Errorf("ContainsCJK(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestNeedsTranslation(t *testing.T) {
	tests := []struct {
		name     string
		pattern  *Pattern
		expected bool
	}{
		{
			name: "English pattern",
			pattern: &Pattern{
				Name:        "api-error-handling",
				Description: "Handle API errors gracefully",
				Content:     "# API Error Handling\n\nUse try-catch...",
			},
			expected: false,
		},
		{
			name: "Chinese name",
			pattern: &Pattern{
				Name:        "api-錯誤處理",
				Description: "Handle API errors",
				Content:     "# Error Handling",
			},
			expected: true,
		},
		{
			name: "Chinese description",
			pattern: &Pattern{
				Name:        "api-error-handling",
				Description: "處理 API 錯誤",
				Content:     "# Error Handling",
			},
			expected: true,
		},
		{
			name: "Chinese content",
			pattern: &Pattern{
				Name:        "api-error-handling",
				Description: "Handle API errors",
				Content:     "# 錯誤處理\n\n使用 try-catch...",
			},
			expected: true,
		},
		{
			name: "Japanese content",
			pattern: &Pattern{
				Name:        "async-handling",
				Description: "Async patterns",
				Content:     "# 非同期処理\n\nAsync/await を使う",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NeedsTranslation(tt.pattern)
			if result != tt.expected {
				t.Errorf("NeedsTranslation() = %v, want %v", result, tt.expected)
			}
		})
	}
}
