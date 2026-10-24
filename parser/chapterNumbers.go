package parser

import (
	"regexp"
	"strings"
)

// ChapterPatterns defines the regex patterns to match chapter numbers in filenames
// These can be easily updated to support additional filename formats
var ChapterPatterns = []string{
	`^ch[-_\s]*(\d+(?:\.\d+)?)`,           // ch0001, ch001.5, ch-01, ch_1, ch 1
	`^chapter[-_\s]*(\d+(?:\.\d+)?)`,      // chapter 1, chapter-01, chapter 001.5
	`^c[-_\s]*(\d+(?:\.\d+)?)`,            // c1, c-01, c001.5
	`[-_\s]ch[-_\s]*(\d+(?:\.\d+)?)`,      // something-ch01, something-ch001.5
	`[-_\s]chapter[-_\s]*(\d+(?:\.\d+)?)`, // something-chapter-01
	`[-_\s]c[-_\s]*(\d+(?:\.\d+)?)`,       // something-c01
}

// ExtractChapterNumber attempts to extract a chapter number from a filename
// Returns the chapter number as a string with leading zeros removed, or empty string if no match found
func ExtractChapterNumber(filename string) string {
	// Remove file extension
	name := strings.TrimSuffix(filename, ".cbz")
	name = strings.TrimSuffix(name, ".zip")

	// Convert to lowercase for case-insensitive matching
	nameLower := strings.ToLower(name)

	// Try each pattern
	for _, pattern := range ChapterPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(nameLower)

		if len(matches) > 1 {
			chapterNum := matches[1]
			// Handle decimal numbers separately
			if strings.Contains(chapterNum, ".") {
				parts := strings.SplitN(chapterNum, ".", 2)
				intPart := strings.TrimLeft(parts[0], "0")
				if intPart == "" {
					intPart = "0"
				}
				chapterNum = intPart + "." + parts[1]
			} else {
				// Remove leading zeros from the captured group
				chapterNum = strings.TrimLeft(chapterNum, "0")
				// Handle edge case where the number is just "0" or "00" etc
				if chapterNum == "" {
					chapterNum = "0"
				}
			}

			return chapterNum
		}
	}

	// No match found
	return ""
}
