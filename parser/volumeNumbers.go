package parser

import (
	"regexp"
	"strings"
)

var VolumePatterns = []string{
	`(?i)v(?:ol(?:ume)?)?[-_\s]*(\d+)`,
	`(?i)volume[-_\s]*(\d+)`,
	`(?i)vol[-_\s]*(\d+)`,
	`[-_\s]v(\d+)`,
}

// ExtractVolumeNumber attempts to extract a volume number from a filename
func ExtractVolumeNumber(filename string) string {
	name := strings.TrimSuffix(filename, ".cbz")
	name = strings.TrimSuffix(name, ".zip")

	for _, pattern := range VolumePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(name)
		if len(matches) > 1 {
			vol := strings.TrimLeft(matches[1], "0")
			if vol == "" {
				vol = "0"
			}
			return vol
		}
	}
	return ""
}
