package parser

import "testing"

func TestCleanTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase conversion",
			input:    "Hello World",
			expected: "hello world",
		},
		{
			name:     "strips punctuation",
			input:    "Hello, World!",
			expected: "hello world",
		},
		{
			name:     "strips special characters",
			input:    "86—EIGHTY-SIX",
			expected: "86 eighty six",
		},
		{
			name:     "collapses multiple spaces",
			input:    "hello   world",
			expected: "hello world",
		},
		{
			name:     "trims leading and trailing spaces",
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only punctuation",
			input:    "!!!---",
			expected: "",
		},
		{
			name:     "numbers preserved",
			input:    "Vol. 1",
			expected: "vol 1",
		},
		{
			name:     "mixed alphanumeric",
			input:    "Attack on Titan Vol.1",
			expected: "attack on titan vol 1",
		},
		{
			name:     "already clean",
			input:    "hello world",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanTitle(tt.input)
			if got != tt.expected {
				t.Errorf("CleanTitle(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
