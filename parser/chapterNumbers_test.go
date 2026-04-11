package parser

import "testing"

func TestExtractChapterNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// ch prefix variants
		{
			name:     "ch prefix padded zeros",
			input:    "ch0001.cbz",
			expected: "1",
		},
		{
			name:     "ch prefix no padding",
			input:    "ch1.cbz",
			expected: "1",
		},
		{
			name:     "ch dash prefix",
			input:    "ch-01.cbz",
			expected: "1",
		},
		{
			name:     "ch underscore prefix",
			input:    "ch_1.cbz",
			expected: "1",
		},
		{
			name:     "ch space prefix",
			input:    "ch 1.cbz",
			expected: "1",
		},
		// chapter prefix variants
		{
			name:     "chapter word prefix",
			input:    "chapter 1.cbz",
			expected: "1",
		},
		{
			name:     "chapter dash prefix",
			input:    "chapter-01.cbz",
			expected: "1",
		},
		{
			name:     "chapter underscore prefix",
			input:    "chapter_001.cbz",
			expected: "1",
		},
		// c prefix variants
		{
			name:     "c prefix",
			input:    "c1.cbz",
			expected: "1",
		},
		{
			name:     "c dash prefix",
			input:    "c-01.cbz",
			expected: "1",
		},
		// inline variants
		{
			name:     "inline ch",
			input:    "seriesname-ch01.cbz",
			expected: "1",
		},
		{
			name:     "inline chapter",
			input:    "seriesname-chapter-01.cbz",
			expected: "1",
		},
		{
			name:     "inline c",
			input:    "seriesname-c01.cbz",
			expected: "1",
		},
		// zero handling
		{
			name:     "chapter zero",
			input:    "ch00.cbz",
			expected: "0",
		},
		{
			name:     "chapter zero single digit",
			input:    "ch0.cbz",
			expected: "0",
		},
		// large numbers
		{
			name:     "three digit chapter",
			input:    "ch100.cbz",
			expected: "100",
		},
		{
			name:     "four digit chapter",
			input:    "ch1000.cbz",
			expected: "1000",
		},
		// no match
		{
			name:     "no chapter number",
			input:    "somefile.cbz",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "zip extension",
			input:    "ch01.zip",
			expected: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractChapterNumber(tt.input)
			if got != tt.expected {
				t.Errorf("ExtractChapterNumber(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
