package parser

import (
	"sort"
	"strings"
)

// RenderMetadata formats metadata as a two-column table and returns it as a string.
// Both columns are left-aligned. The first column width is automatically adjusted
// to the longest metadata key (minimum 20 characters).
func RenderMetadataOutput(md map[string]string) string {
	if len(md) == 0 {
		return ""
	}

	// Sort keys alphabetically
	keys := make([]string, 0, len(md))
	for k := range md {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder

	// Header words
	header1 := "Metadata field"
	header2 := "Metadata value"

	// Write header
	sb.WriteString(header1 + "\t\t" + header2 + "\n")
	sb.WriteString(strings.Repeat("-", len(header1)) + "\t\t" + strings.Repeat("-", len(header2)) + "\n")

	// Rows
	for _, k := range keys {
		sb.WriteString(padRight(k, len(header1)) + "\t\t" + md[k] + "\n")
	}

	return sb.String()
}

// padRight pads a string with spaces on the right to the specified length
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
