package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mdu/manga_src"
)

// Formats metadata as a two-column table and returns it as a string.
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

// Pads a string with spaces on the right to the specified length
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// Returns all .epub files in the given directory, sorted alphabetically.
func ListEPUBFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".epub") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	sort.Strings(files)
	return files, nil
}

// Returns a formatted string with a filename header
func RenderMetadataWithHeader(filename string, md map[string]string) string {
	var sb strings.Builder
	sb.WriteString("File - " + filename + ":\n\n")
	sb.WriteString(RenderMetadataOutput(md))
	sb.WriteString("\n")
	return sb.String()
}

// PrintTitleSearchResults prints the main titles and Mangadex IDs in two columns.
// Accepts the return value from MangadexTitleSearch().
// PrintTitleSearchResults wraps titles to fit a fixed column width while keeping ID on the first line.
func PrintTitleSearchResults(results []mangasrc.MangadexTitleSearchResponse) {
	if len(results) == 0 {
		fmt.Println("No results to display")
		return
	}

	// Column widths
	const titleWidth = 60
	const idWidth = 36

	// Print header
	fmt.Printf("%-*s | %-*s\n", titleWidth, "Title", idWidth, "Mangadex ID")
	fmt.Printf("%s-+-%s\n", strings.Repeat("-", titleWidth), strings.Repeat("-", idWidth))

	for _, r := range results {
		title := r.MainTitle
		if title == "" && len(r.AltTitles) > 0 {
			title = r.AltTitles[0] // fallback to first alt title
		}

		lines := wrapText(title, titleWidth)

		for i, line := range lines {
			if i == 0 {
				// Print first line with ID
				fmt.Printf("%-*s | %-*s\n", titleWidth, line, idWidth, r.ID)
			} else {
				// Subsequent lines: only print the wrapped title
				fmt.Printf("%-*s | %-*s\n", titleWidth, line, idWidth, "")
			}
		}
	}
}

// wrapText splits a string into lines of max width n, respecting word boundaries.
func wrapText(text string, width int) []string {
	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	current := words[0]
	for _, word := range words[1:] {
		if len(current)+1+len(word) > width {
			lines = append(lines, current)
			current = word
		} else {
			current += " " + word
		}
	}
	lines = append(lines, current)
	return lines
}

// Hard-coded language priority (can be updated in the future)
var langPriority = []string{"en", "ja", "zh"} // English > Japanese > Chinese

// PrintMangaDexMetadata prints a pretty formatted view of a MangadexTitleMetadata struct.
// Accepts the return value from TitleMetadata().
func PrintMangaDexMetadata(md *mangasrc.MangadexTitleMetadata) {
	if md == nil {
		fmt.Println("No metadata provided (nil pointer)")
		return
	}

	fmt.Println("────────── MangaDex Metadata ──────────")
	fmt.Printf("ID:   %s\n", md.ID)
	fmt.Printf("Type: %s\n", md.Type)
	fmt.Println()

	// Extract main title
	if title := extractMainTitle(md.Attributes); title != "" {
		fmt.Printf("Title: %s\n", title)
	}

	// Extract alternative titles
	if alt := extractAltTitles(md.Attributes); len(alt) > 0 {
		fmt.Println("Alt Titles:")
		for _, t := range alt {
			fmt.Printf("  • %s\n", t)
		}
	}

	// Dump all attributes (pretty JSON)
	fmt.Println("\nAttributes:")
	if attrJSON, err := json.MarshalIndent(md.Attributes, "", "  "); err == nil {
		fmt.Println(string(attrJSON))
	} else {
		fmt.Printf("Error printing attributes: %v\n", err)
	}

	// Dump relationships if any
	if len(md.Relationships) > 0 {
		fmt.Println("\nRelationships:")
		if relJSON, err := json.MarshalIndent(md.Relationships, "", "  "); err == nil {
			fmt.Println(string(relJSON))
		} else {
			fmt.Printf("Error printing relationships: %v\n", err)
		}
	}

	fmt.Println("────────────────────────────────────────")
}

// ------------------ Helper functions ------------------

// extractMainTitle extracts the main title using the langPriority slice
func extractMainTitle(attrs map[string]any) string {
	if attrs == nil {
		return ""
	}

	if titles, ok := attrs["title"].(map[string]any); ok {
		for _, lang := range langPriority {
			if t, ok := titles[lang].(string); ok && t != "" {
				return t
			}
		}
		// fallback to first available title
		for _, v := range titles {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

// extractAltTitles extracts all alternative titles, ordered by langPriority first
func extractAltTitles(attrs map[string]any) []string {
	var altList []string
	if attrs == nil {
		return altList
	}

	if altArr, ok := attrs["altTitles"].([]any); ok {
		for _, at := range altArr {
			if m, ok := at.(map[string]any); ok {
				// Try priority languages first
				for _, lang := range langPriority {
					if t, ok := m[lang].(string); ok && t != "" {
						altList = append(altList, t)
					}
				}
				// Then append any other remaining languages
				for k, v := range m {
					if !contains(langPriority, k) {
						if s, ok := v.(string); ok {
							altList = append(altList, s)
						}
					}
				}
			}
		}
	}

	return altList
}

// ------------------ Utility ------------------

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
