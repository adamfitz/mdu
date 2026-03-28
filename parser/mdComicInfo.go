package parser

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mdu/mangasrc"
)

// ComicInfo represents the ComicInfo.xml structure for Kavita compatibility
type ComicInfo struct {
	XMLName xml.Name `xml:"ComicInfo"`

	// Basic Information
	Title           string `xml:"Title,omitempty"`
	Series          string `xml:"Series,omitempty"`
	Number          string `xml:"Number,omitempty"`
	Count           int    `xml:"Count,omitempty"`
	Volume          string `xml:"Volume,omitempty"`
	AlternateSeries string `xml:"AlternateSeries,omitempty"`
	AlternateNumber string `xml:"AlternateNumber,omitempty"`
	AlternateCount  int    `xml:"AlternateCount,omitempty"`

	// Summary and Description
	Summary string `xml:"Summary,omitempty"`
	Notes   string `xml:"Notes,omitempty"`

	// People
	Writer      string `xml:"Writer,omitempty"`
	Penciller   string `xml:"Penciller,omitempty"`
	Inker       string `xml:"Inker,omitempty"`
	Colorist    string `xml:"Colorist,omitempty"`
	Letterer    string `xml:"Letterer,omitempty"`
	CoverArtist string `xml:"CoverArtist,omitempty"`
	Editor      string `xml:"Editor,omitempty"`
	Publisher   string `xml:"Publisher,omitempty"`

	// Dates
	Year  int `xml:"Year,omitempty"`
	Month int `xml:"Month,omitempty"`
	Day   int `xml:"Day,omitempty"`

	// Classification
	Genre         string `xml:"Genre,omitempty"`
	Tags          string `xml:"Tags,omitempty"`
	Web           string `xml:"Web,omitempty"`
	LanguageISO   string `xml:"LanguageISO,omitempty"`
	Format        string `xml:"Format,omitempty"`
	AgeRating     string `xml:"AgeRating,omitempty"`
	Manga         string `xml:"Manga,omitempty"`
	BlackAndWhite string `xml:"BlackAndWhite,omitempty"`

	// Pages
	PageCount int            `xml:"PageCount,omitempty"`
	Pages     *ComicPageList `xml:"Pages,omitempty"`

	// Additional
	SeriesGroup     string `xml:"SeriesGroup,omitempty"`
	StoryArc        string `xml:"StoryArc,omitempty"`
	ScanInformation string `xml:"ScanInformation,omitempty"`
}

// ComicPageInfo represents a single page entry in the ComicInfo Pages block.
// Kavita uses Type="FrontCover" on page 0 to identify the cover image.
type ComicPageInfo struct {
	Image int    `xml:"Image,attr"`
	Type  string `xml:"Type,attr,omitempty"`
}

// ComicPageList is the container element for the Pages block in ComicInfo.xml.
type ComicPageList struct {
	Pages []ComicPageInfo `xml:"Page"`
}

// MangaDexToComicInfo converts MangaDex metadata to ComicInfo struct
func MangaDexToComicInfo(md *mangasrc.MangadexTitleMetadata) (*ComicInfo, error) {
	if md == nil {
		return nil, fmt.Errorf("nil metadata provided")
	}

	comic := &ComicInfo{
		Manga: "Yes", // MangaDex is for manga
	}

	// Extract main title
	if titleMap, ok := md.Attributes["title"].(map[string]interface{}); ok {
		if enTitle, ok := titleMap["en"].(string); ok && enTitle != "" {
			comic.Title = enTitle
			comic.Series = enTitle
		}
	}

	// Extract English description/summary
	if descMap, ok := md.Attributes["description"].(map[string]interface{}); ok {
		if enDesc, ok := descMap["en"].(string); ok && enDesc != "" {
			// Clean up description (remove links section if present)
			lines := strings.Split(enDesc, "\n---\n")
			comic.Summary = strings.TrimSpace(lines[0])
		}
	}

	// Extract author/artist from relationships
	authors := extractAuthorsFromRelationships(md.Relationships)
	if len(authors) > 0 {
		// Use first author as Writer and Penciller
		comic.Writer = strings.Join(authors, ", ")
		comic.Penciller = strings.Join(authors, ", ")
	}

	// Extract tags/genres
	var genres []string
	var tags []string

	if tagsList, ok := md.Attributes["tags"].([]interface{}); ok {
		for _, tagInterface := range tagsList {
			if tagMap, ok := tagInterface.(map[string]interface{}); ok {
				if attrs, ok := tagMap["attributes"].(map[string]interface{}); ok {
					// Get tag name
					if nameMap, ok := attrs["name"].(map[string]interface{}); ok {
						if enName, ok := nameMap["en"].(string); ok && enName != "" {
							// Get tag group
							group := ""
							if g, ok := attrs["group"].(string); ok {
								group = g
							}

							// Separate by group
							if group == "genre" {
								genres = append(genres, enName)
							} else {
								tags = append(tags, enName)
							}
						}
					}
				}
			}
		}
	}

	if len(genres) > 0 {
		comic.Genre = strings.Join(genres, ", ")
	}
	if len(tags) > 0 {
		comic.Tags = strings.Join(tags, ", ")
	}

	// Extract publication demographic
	if demographic, ok := md.Attributes["publicationDemographic"].(string); ok && demographic != "" {
		if comic.Tags != "" {
			comic.Tags += ", " + demographic
		} else {
			comic.Tags = demographic
		}
	}

	// Extract year
	if year, ok := md.Attributes["year"].(float64); ok {
		comic.Year = int(year)
	}

	// Extract status and add to notes
	if status, ok := md.Attributes["status"].(string); ok && status != "" {
		comic.Notes = fmt.Sprintf("Status: %s", status)
	}

	// Extract content rating for age rating
	if contentRating, ok := md.Attributes["contentRating"].(string); ok {
		switch contentRating {
		case "safe":
			comic.AgeRating = "Everyone"
		case "suggestive":
			comic.AgeRating = "Teen"
		case "erotica":
			comic.AgeRating = "Mature 17+"
		case "pornographic":
			comic.AgeRating = "Adults Only 18+"
		}
	}

	// Extract last chapter/volume info
	if lastVol, ok := md.Attributes["lastVolume"].(string); ok && lastVol != "" {
		comic.Volume = lastVol
	}
	if lastChapter, ok := md.Attributes["lastChapter"].(string); ok && lastChapter != "" {
		// Store total chapters in Count
		// Note: This is the total number of chapters, not the current chapter
		if comic.Notes != "" {
			comic.Notes += fmt.Sprintf(" | Total Chapters: %s", lastChapter)
		} else {
			comic.Notes = fmt.Sprintf("Total Chapters: %s", lastChapter)
		}
	}

	// Extract language
	if lang, ok := md.Attributes["originalLanguage"].(string); ok && lang != "" {
		comic.LanguageISO = lang
	}

	// Set black and white based on tags
	comic.BlackAndWhite = "Yes"
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), "colored") || strings.Contains(strings.ToLower(tag), "official colored") {
			comic.BlackAndWhite = "No"
			break
		}
	}

	// Add MangaDex ID to web field
	comic.Web = fmt.Sprintf("https://mangadex.org/title/%s", md.ID)

	// Extract creation/update dates for additional info
	if createdAt, ok := md.Attributes["createdAt"].(string); ok && createdAt != "" {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			if comic.Year == 0 {
				comic.Year = t.Year()
				comic.Month = int(t.Month())
				comic.Day = t.Day()
			}
		}
	}

	return comic, nil
}

// extractAuthorsFromRelationships extracts author names from relationships
// Note: This is a simplified version - in practice you'd need to fetch author details
func extractAuthorsFromRelationships(relationships []interface{}) []string {
	var authors []string

	for _, rel := range relationships {
		if relMap, ok := rel.(map[string]interface{}); ok {
			if relType, ok := relMap["type"].(string); ok {
				if relType == "author" || relType == "artist" {
					// In practice, you'd need to fetch the author name from the API
					// using the ID. For now, we'll just note the ID exists
					if id, ok := relMap["id"].(string); ok {
						// This would need another API call to get the actual name
						authors = append(authors, fmt.Sprintf("Author ID: %s", id))
					}
				}
			}
		}
	}

	return authors
}

// ToXML converts ComicInfo to XML string
func (c *ComicInfo) ToXML() (string, error) {
	output, err := xml.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(output), nil
}

// Example usage function
func ConvertAndPrintComicInfo(md *mangasrc.MangadexTitleMetadata) {
	comicInfo, err := MangaDexToComicInfo(md)
	if err != nil {
		fmt.Printf("Error converting metadata: %v\n", err)
		return
	}

	xmlStr, err := comicInfo.ToXML()
	if err != nil {
		fmt.Printf("Error generating XML: %v\n", err)
		return
	}

	fmt.Println("────────── ComicInfo.xml ──────────")
	fmt.Println(xmlStr)
	fmt.Println("───────────────────────────────────")
}

// MangaDexAuthorResponse represents the API response for author lookup
type MangaDexAuthorResponse struct {
	Result string `json:"result"`
	Data   struct {
		Attributes struct {
			Name string `json:"name"`
		} `json:"attributes"`
	} `json:"data"`
}

// WriteComicInfo writes a ComicInfo struct to an XML file at the specified path
// Returns the absolute path to the written file and any error
func WriteComicInfo(comic *ComicInfo, filename string) (string, error) {
	if comic == nil {
		return "", fmt.Errorf("nil ComicInfo provided")
	}

	if filename == "" {
		return "", fmt.Errorf("filename must be specified")
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return "", fmt.Errorf("failed to resolve file path: %w", err)
	}

	// Generate XML
	xmlData, err := xml.MarshalIndent(comic, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal ComicInfo to XML: %w", err)
	}

	// Add XML header
	fullXML := xml.Header + string(xmlData)

	// Write the file
	err = os.WriteFile(absPath, []byte(fullXML), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write ComicInfo.xml: %w", err)
	}

	return absPath, nil
}

// MangaAuthorName fetches an author's name from MangaDex by their ID
func MangaAuthorName(authorID string) (string, error) {
	if authorID == "" {
		return "", fmt.Errorf("authorID cannot be empty")
	}

	url := fmt.Sprintf("https://api.mangadex.org/author/%s", authorID)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch author: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var authorResp MangaDexAuthorResponse
	if err := json.Unmarshal(body, &authorResp); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	return authorResp.Data.Attributes.Name, nil
}
