package parser

import (
	"regexp"
	"strings"
)

var nonAlnum = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// CleanTitle strips punctuation and normalizes to lowercase tokens.
func CleanTitle(s string) string {
	s = strings.ToLower(s)
	s = nonAlnum.ReplaceAllString(s, " ")
	s = strings.Join(strings.Fields(s), " ")
	return s
}
