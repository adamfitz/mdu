package cli

import (
	"os"
)

// buildUpdatesMap creates a map of metadata updates from individual flags
func buildUpdatesMap(series, seriesIndex, summary, isbn, author string) map[string]string {
	updates := make(map[string]string)

	if series != "" {
		updates["calibre:series"] = series
	}
	if seriesIndex != "" {
		updates["calibre:series_index"] = seriesIndex
	}
	if summary != "" {
		updates["summary"] = summary
	}
	if isbn != "" {
		updates["isbn"] = isbn
	}
	if author != "" {
		updates["author"] = author
	}

	return updates
}

// fileExists checks if a file exists at the given path
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
