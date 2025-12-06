package parser

import (
	"mdu/mangasrc"
)

// isLikelyEnglish returns true if the string contains no CJK or Hangul runes.
func isLikelyEnglish(s string) bool {
	for _, r := range s {
		// Hiragana
		if r >= 0x3040 && r <= 0x309F {
			return false
		}
		// Katakana
		if r >= 0x30A0 && r <= 0x30FF {
			return false
		}
		// CJK Unified Ideographs
		if r >= 0x4E00 && r <= 0x9FFF {
			return false
		}
		// Hangul
		if r >= 0xAC00 && r <= 0xD7A3 {
			return false
		}
	}
	return true
}

// ExtractEnglishTitles returns all English titles (main + alt) for use in scoring or matching.
func ExtractEnglishTitles(results []mangasrc.MangadexTitleSearchResponse) []string {
	var titles []string

	for _, r := range results {
		// Main title (already selected as English if available in your search func)
		if r.MainTitle != "" {
			titles = append(titles, r.MainTitle)
		}

		// Alt titles: keep only those that appear to be English
		for _, alt := range r.AltTitles {
			if isLikelyEnglish(alt) {
				titles = append(titles, alt)
			}
		}
	}

	return titles
}
