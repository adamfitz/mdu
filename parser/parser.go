package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"mdu/mangasrc"

	"github.com/mattn/go-runewidth"
	"golang.org/x/text/width"
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

// padRight pads a string to the specified width, correctly handling display width
// (East Asian characters count as 2 width, most others as 1)
func padRight(s string, width int) string {
	displayWidth := runewidth.StringWidth(s)
	if displayWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-displayWidth)
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

// wrapText splits a string into lines of max display width, respecting word boundaries.
// Uses display width to handle East Asian characters correctly (they display as 2 chars wide).
func wrapText(text string, width int) []string {
	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	current := words[0]
	currentWidth := runewidth.StringWidth(current)

	for _, word := range words[1:] {
		wordWidth := runewidth.StringWidth(word)
		if currentWidth+1+wordWidth > width {
			lines = append(lines, current)
			current = word
			currentWidth = wordWidth
		} else {
			current += " " + word
			currentWidth += 1 + wordWidth
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
	return slices.Contains(slice, val)
}

// FindEntryByTitle returns the entry whose English main title or alt title
// exactly matches the provided target string.
func FindEntryByTitle(results []mangasrc.MangadexTitleSearchResponse, title string) *mangasrc.MangadexTitleSearchResponse {
	title = strings.ToLower(strings.TrimSpace(title))

	for i := range results {
		r := &results[i]

		// Main title
		if strings.ToLower(r.MainTitle) == title {
			return r
		}

		// Alt titles
		for _, alt := range r.AltTitles {
			if strings.ToLower(alt) == title {
				return r
			}
		}
	}

	return nil
}

// LocalChapterList returns a list of all files from the provided rootDir.
// Optionally pass an exclusion list to skip certain file names.
func ListCBZFiles(rootDir string, exclusionList ...string) ([]string, error) {
	// Expand ~ to home directory
	expandedPath, err := ExpandPath(rootDir)
	if err != nil {
		return nil, err
	}

	// Convert exclusionList slice to a map for fast lookup
	exclusions := make(map[string]struct{}, len(exclusionList))
	for _, name := range exclusionList {
		exclusions[name] = struct{}{}
	}

	fileList := make([]string, 0)

	entries, err := os.ReadDir(expandedPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			if _, skip := exclusions[entry.Name()]; !skip {
				fileList = append(fileList, entry.Name())
			}
		}
	}

	filteredFileList := filterCBZFiles(fileList)

	return filteredFileList, nil
}

// expands ~ to the user's home directory, or returns the path as-is
func ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		// Path starts with ~/ so expand it
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, path[2:]), nil
	}
	// Path doesn't start with ~/ so return it unchanged
	return path, nil
}

// filters out any non *.cbz file from the list
func filterCBZFiles(files []string) []string {
	var filtered []string
	for _, f := range files {
		if strings.EqualFold(filepath.Ext(f), ".cbz") {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// visualWidth returns the visual/display width of a string
// accounting for wide characters (like CJK) that take 2 columns
func visualWidth(s string) int {
	w := 0
	for _, r := range s {
		switch width.LookupRune(r).Kind() {
		case width.EastAsianWide, width.EastAsianFullwidth:
			w += 2
		default:
			w += 1
		}
	}
	return w
}

// wrapTextByVisualWidth wraps text based on visual width, not byte/rune count
func wrapTextByVisualWidth(text string, maxWidth int) []string {
	if text == "" {
		return []string{""}
	}

	var lines []string
	var currentLine []rune
	currentWidth := 0

	for _, r := range text {
		runeWidth := 1
		switch width.LookupRune(r).Kind() {
		case width.EastAsianWide, width.EastAsianFullwidth:
			runeWidth = 2
		}

		// If adding this rune would exceed maxWidth, start a new line
		if currentWidth+runeWidth > maxWidth && len(currentLine) > 0 {
			lines = append(lines, string(currentLine))
			currentLine = []rune{r}
			currentWidth = runeWidth
		} else {
			currentLine = append(currentLine, r)
			currentWidth += runeWidth
		}
	}

	// Add the last line
	if len(currentLine) > 0 {
		lines = append(lines, string(currentLine))
	}

	return lines
}

// padRightByVisualWidth pads a string to a fixed visual width
func padRightByVisualWidth(s string, width int) string {
	vw := visualWidth(s)
	if vw >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vw)
}

// PrintTitleSearchResults prints the main titles and Mangadex IDs in two columns.
// Accepts the return value from MangadexTitleSearch().
// PrintTitleSearchResults wraps titles to fit a fixed column width while keeping ID on the first line.
func PrintTitleSearchResults(results []mangasrc.MangadexTitleSearchResponse) {
	if len(results) == 0 {
		fmt.Println("No results to display")
		return
	}

	// Column widths (visual width)
	const colWidth = 40

	// Print header
	fmt.Printf("%s %s %s\n",
		padRightByVisualWidth("Name", colWidth),
		padRightByVisualWidth("Alt Name", colWidth),
		padRightByVisualWidth("Mangadex ID", colWidth))
	fmt.Printf("%s %s %s\n",
		strings.Repeat("-", colWidth),
		strings.Repeat("-", colWidth),
		strings.Repeat("-", colWidth))

	for _, r := range results {
		mainTitle := r.MainTitle
		if mainTitle == "" {
			mainTitle = "-"
		}

		altTitle := ""
		if len(r.AltTitles) > 0 {
			altTitle = r.AltTitles[0]
		} else {
			altTitle = "-"
		}

		// Wrap both title columns based on visual width
		mainLines := wrapTextByVisualWidth(mainTitle, colWidth)
		altLines := wrapTextByVisualWidth(altTitle, colWidth)

		// Determine max lines needed
		maxLines := len(mainLines)
		if len(altLines) > maxLines {
			maxLines = len(altLines)
		}

		// Print each line
		for i := 0; i < maxLines; i++ {
			mainLine := ""
			if i < len(mainLines) {
				mainLine = mainLines[i]
			}

			altLine := ""
			if i < len(altLines) {
				altLine = altLines[i]
			}

			idLine := ""
			if i == 0 {
				idLine = r.ID
			}

			fmt.Printf("%s %s %s\n",
				padRightByVisualWidth(mainLine, colWidth),
				padRightByVisualWidth(altLine, colWidth),
				padRightByVisualWidth(idLine, colWidth))
		}
	}
}
