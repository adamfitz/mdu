package parser

import (
	"fmt"
	"strings"

	"mdu/metadata"
)

// RenderComicInfo returns a formatted string of ComicInfo metadata for display
// Optionally accepts a filename as the first variadic parameter to add a header
func RenderComicInfo(ci *metadata.ComicInfo, filename ...string) string {
	var sb strings.Builder

	// Add filename header if provided
	if len(filename) > 0 && filename[0] != "" {
		sb.WriteString("File - " + filename[0] + ":\n\n")
	}

	if ci == nil {
		return sb.String()
	}

	if ci.Series != "" {
		sb.WriteString(fmt.Sprintf("  Series: %s\n", ci.Series))
	}
	if ci.Number != "" {
		sb.WriteString(fmt.Sprintf("  Number: %s\n", ci.Number))
	}
	if ci.Volume != "" {
		sb.WriteString(fmt.Sprintf("  Volume: %s\n", ci.Volume))
	}
	if ci.Title != "" {
		sb.WriteString(fmt.Sprintf("  Title: %s\n", ci.Title))
	}
	if ci.Summary != "" {
		sb.WriteString(fmt.Sprintf("  Summary: %s\n", ci.Summary))
	}
	if ci.Notes != "" {
		sb.WriteString(fmt.Sprintf("  Notes: %s\n", ci.Notes))
	}
	if ci.Writer != "" {
		sb.WriteString(fmt.Sprintf("  Writer: %s\n", ci.Writer))
	}
	if ci.Penciller != "" {
		sb.WriteString(fmt.Sprintf("  Penciller: %s\n", ci.Penciller))
	}
	if ci.Inker != "" {
		sb.WriteString(fmt.Sprintf("  Inker: %s\n", ci.Inker))
	}
	if ci.Colorist != "" {
		sb.WriteString(fmt.Sprintf("  Colorist: %s\n", ci.Colorist))
	}
	if ci.Letterer != "" {
		sb.WriteString(fmt.Sprintf("  Letterer: %s\n", ci.Letterer))
	}
	if ci.CoverArtist != "" {
		sb.WriteString(fmt.Sprintf("  CoverArtist: %s\n", ci.CoverArtist))
	}
	if ci.Editor != "" {
		sb.WriteString(fmt.Sprintf("  Editor: %s\n", ci.Editor))
	}
	if ci.Publisher != "" {
		sb.WriteString(fmt.Sprintf("  Publisher: %s\n", ci.Publisher))
	}
	if ci.Genre != "" {
		sb.WriteString(fmt.Sprintf("  Genre: %s\n", ci.Genre))
	}
	if ci.Tags != "" {
		sb.WriteString(fmt.Sprintf("  Tags: %s\n", ci.Tags))
	}
	if ci.PageCount > 0 {
		sb.WriteString(fmt.Sprintf("  PageCount: %d\n", ci.PageCount))
	}
	if ci.LanguageISO != "" {
		sb.WriteString(fmt.Sprintf("  LanguageISO: %s\n", ci.LanguageISO))
	}
	if ci.Format != "" {
		sb.WriteString(fmt.Sprintf("  Format: %s\n", ci.Format))
	}
	if ci.AgeRating != "" {
		sb.WriteString(fmt.Sprintf("  AgeRating: %s\n", ci.AgeRating))
	}
	if ci.Year > 0 {
		sb.WriteString(fmt.Sprintf("  Year: %d\n", ci.Year))
	}
	if ci.Month > 0 {
		sb.WriteString(fmt.Sprintf("  Month: %d\n", ci.Month))
	}
	if ci.Day > 0 {
		sb.WriteString(fmt.Sprintf("  Day: %d\n", ci.Day))
	}
	if ci.Web != "" {
		sb.WriteString(fmt.Sprintf("  Web: %s\n", ci.Web))
	}

	return sb.String()
}
