package metadata

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	PageCount int `xml:"PageCount,omitempty"`

	// Additional
	SeriesGroup     string `xml:"SeriesGroup,omitempty"`
	StoryArc        string `xml:"StoryArc,omitempty"`
	ScanInformation string `xml:"ScanInformation,omitempty"`
}

// MangaDexTitleMetadata represents MangaDex API response
type MangaDexTitleMetadata struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Attributes    map[string]any `json:"attributes"`
	Relationships []any          `json:"relationships"`
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

// ---------- Public Functions for CBZ ----------

// ReadComicInfo extracts ComicInfo.xml from a CBZ file
func ReadComicInfo(cbzPath string) (*ComicInfo, error) {
	r, err := zip.OpenReader(cbzPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CBZ: %w", err)
	}
	defer r.Close()

	// Look for ComicInfo.xml
	comicInfoContent, err := readFileFromZipCBZ(r, "ComicInfo.xml")
	if err != nil {
		return nil, fmt.Errorf("ComicInfo.xml not found in CBZ: %w", err)
	}

	var comicInfo ComicInfo
	if err := xml.Unmarshal(comicInfoContent, &comicInfo); err != nil {
		return nil, fmt.Errorf("failed to parse ComicInfo.xml: %w", err)
	}

	return &comicInfo, nil
}

// UpdateComicInfo modifies ComicInfo.xml in a CBZ file
func UpdateComicInfo(cbzPath, outputPath string, updates map[string]string) error {
	r, err := zip.OpenReader(cbzPath)
	if err != nil {
		return fmt.Errorf("failed to open CBZ: %w", err)
	}
	defer r.Close()

	// Try to read existing ComicInfo.xml
	var comicInfo *ComicInfo
	existingContent, err := readFileFromZipCBZ(r, "ComicInfo.xml")
	if err == nil {
		// Parse existing
		comicInfo = &ComicInfo{}
		if err := xml.Unmarshal(existingContent, comicInfo); err != nil {
			return fmt.Errorf("failed to parse existing ComicInfo.xml: %w", err)
		}
	} else {
		// Create new
		comicInfo = &ComicInfo{}
	}

	// Apply updates
	applyComicInfoUpdates(comicInfo, updates)

	// Serialize to XML
	modifiedXML, err := xml.MarshalIndent(comicInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal ComicInfo: %w", err)
	}
	fullXML := []byte(xml.Header + string(modifiedXML))

	// Write modified CBZ
	if err := writeModifiedCBZ(r, fullXML, outputPath); err != nil {
		return fmt.Errorf("failed to write modified CBZ: %w", err)
	}

	return nil
}

// CompareComicInfo compares ComicInfo.xml from two CBZ files
func CompareComicInfo(originalPath, modifiedPath string) (string, error) {
	origInfo, err := ReadComicInfo(originalPath)
	if err != nil {
		return "", fmt.Errorf("reading original: %w", err)
	}

	modInfo, err := ReadComicInfo(modifiedPath)
	if err != nil {
		return "", fmt.Errorf("reading modified: %w", err)
	}

	return generateComicInfoDiff(origInfo, modInfo), nil
}

// WriteComicInfoToFile writes a ComicInfo struct to an XML file
func WriteComicInfoToFile(comic *ComicInfo, filename string) (string, error) {
	if comic == nil {
		return "", fmt.Errorf("nil ComicInfo provided")
	}

	if filename == "" {
		return "", fmt.Errorf("filename must be specified")
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		return "", fmt.Errorf("failed to resolve file path: %w", err)
	}

	xmlData, err := xml.MarshalIndent(comic, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal ComicInfo to XML: %w", err)
	}

	fullXML := xml.Header + string(xmlData)

	err = os.WriteFile(absPath, []byte(fullXML), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write ComicInfo.xml: %w", err)
	}

	return absPath, nil
}

// ---------- MangaDex Integration ----------

// MangaDexToComicInfo converts MangaDex metadata to ComicInfo struct
func MangaDexToComicInfo(md *MangaDexTitleMetadata) (*ComicInfo, error) {
	if md == nil {
		return nil, fmt.Errorf("nil metadata provided")
	}

	comic := &ComicInfo{
		Manga: "Yes",
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
			lines := strings.Split(enDesc, "\n---\n")
			comic.Summary = strings.TrimSpace(lines[0])
		}
	}

	// Extract author/artist from relationships
	authorIDs := extractAuthorIDsFromRelationships(md.Relationships)
	var authorNames []string
	for _, authorID := range authorIDs {
		name, err := GetMangaDexAuthorName(authorID)
		if err == nil {
			authorNames = append(authorNames, name)
		}
	}
	if len(authorNames) > 0 {
		authorsStr := strings.Join(authorNames, ", ")
		comic.Writer = authorsStr
		comic.Penciller = authorsStr
	}

	// Extract tags/genres
	var genres []string
	var tags []string

	if tagsList, ok := md.Attributes["tags"].([]interface{}); ok {
		for _, tagInterface := range tagsList {
			if tagMap, ok := tagInterface.(map[string]interface{}); ok {
				if attrs, ok := tagMap["attributes"].(map[string]interface{}); ok {
					if nameMap, ok := attrs["name"].(map[string]interface{}); ok {
						if enName, ok := nameMap["en"].(string); ok && enName != "" {
							group := ""
							if g, ok := attrs["group"].(string); ok {
								group = g
							}

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

	// Extract status
	if status, ok := md.Attributes["status"].(string); ok && status != "" {
		comic.Notes = fmt.Sprintf("Status: %s", status)
	}

	// Extract content rating
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

	// Extract last volume
	if lastVol, ok := md.Attributes["lastVolume"].(string); ok && lastVol != "" {
		comic.Volume = lastVol
	}

	// Extract last chapter
	if lastChapter, ok := md.Attributes["lastChapter"].(string); ok && lastChapter != "" {
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
		if strings.Contains(strings.ToLower(tag), "colored") {
			comic.BlackAndWhite = "No"
			break
		}
	}

	// Add MangaDex URL
	comic.Web = fmt.Sprintf("https://mangadex.org/title/%s", md.ID)

	// Extract creation date
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

// GetMangaDexAuthorName fetches an author's name from MangaDex by ID
func GetMangaDexAuthorName(authorID string) (string, error) {
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

// ---------- Internal Helper Functions ----------

func readFileFromZipCBZ(r *zip.ReadCloser, filename string) ([]byte, error) {
	for _, f := range r.File {
		if f.Name == filename {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("file not found: %s", filename)
}

func writeModifiedCBZ(r *zip.ReadCloser, comicInfoXML []byte, outputPath string) error {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Track if ComicInfo.xml already exists
	comicInfoExists := false

	// Copy all files, replacing ComicInfo.xml if it exists
	for _, f := range r.File {
		header := &zip.FileHeader{
			Name:   f.Name,
			Method: f.Method,
		}

		w, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		if f.Name == "ComicInfo.xml" {
			comicInfoExists = true
			_, err = w.Write(comicInfoXML)
			if err != nil {
				return err
			}
		} else {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			_, err = io.Copy(w, rc)
			rc.Close()
			if err != nil {
				return err
			}
		}
	}

	// If ComicInfo.xml didn't exist, add it
	if !comicInfoExists {
		header := &zip.FileHeader{
			Name:   "ComicInfo.xml",
			Method: zip.Deflate,
		}
		w, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = w.Write(comicInfoXML)
		if err != nil {
			return err
		}
	}

	if err := zw.Close(); err != nil {
		return err
	}

	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

func applyComicInfoUpdates(comic *ComicInfo, updates map[string]string) {
	for key, value := range updates {
		switch strings.ToLower(key) {
		case "title":
			comic.Title = value
		case "series":
			comic.Series = value
		case "number":
			comic.Number = value
		case "volume":
			comic.Volume = value
		case "summary":
			comic.Summary = value
		case "writer", "author":
			comic.Writer = value
		case "penciller":
			comic.Penciller = value
		case "publisher":
			comic.Publisher = value
		case "genre":
			comic.Genre = value
		case "tags":
			comic.Tags = value
		case "web":
			comic.Web = value
		case "languageiso", "language":
			comic.LanguageISO = value
		case "agerating":
			comic.AgeRating = value
		}
	}
}

func extractAuthorIDsFromRelationships(relationships []interface{}) []string {
	var authorIDs []string

	for _, rel := range relationships {
		if relMap, ok := rel.(map[string]interface{}); ok {
			if relType, ok := relMap["type"].(string); ok {
				if relType == "author" || relType == "artist" {
					if id, ok := relMap["id"].(string); ok {
						authorIDs = append(authorIDs, id)
					}
				}
			}
		}
	}

	return authorIDs
}

func generateComicInfoDiff(original, modified *ComicInfo) string {
	var report strings.Builder

	report.WriteString("=== COMICINFO.XML COMPARISON ===\n\n")

	type field struct {
		name    string
		origVal string
		modVal  string
		changed bool
	}

	fields := []field{
		{"Title", original.Title, modified.Title, original.Title != modified.Title},
		{"Series", original.Series, modified.Series, original.Series != modified.Series},
		{"Number", original.Number, modified.Number, original.Number != modified.Number},
		{"Volume", original.Volume, modified.Volume, original.Volume != modified.Volume},
		{"Summary", original.Summary, modified.Summary, original.Summary != modified.Summary},
		{"Writer", original.Writer, modified.Writer, original.Writer != modified.Writer},
		{"Penciller", original.Penciller, modified.Penciller, original.Penciller != modified.Penciller},
		{"Publisher", original.Publisher, modified.Publisher, original.Publisher != modified.Publisher},
		{"Genre", original.Genre, modified.Genre, original.Genre != modified.Genre},
		{"Tags", original.Tags, modified.Tags, original.Tags != modified.Tags},
		{"AgeRating", original.AgeRating, modified.AgeRating, original.AgeRating != modified.AgeRating},
		{"Web", original.Web, modified.Web, original.Web != modified.Web},
	}

	unchanged, changed := 0, 0

	for _, f := range fields {
		if f.origVal == "" && f.modVal == "" {
			continue
		}

		if f.changed {
			report.WriteString(fmt.Sprintf("~ CHANGED: %s\n", f.name))
			if f.origVal != "" {
				report.WriteString(fmt.Sprintf("  Old: %s\n", f.origVal))
			}
			if f.modVal != "" {
				report.WriteString(fmt.Sprintf("  New: %s\n", f.modVal))
			}
			report.WriteString("\n")
			changed++
		} else {
			unchanged++
		}
	}

	report.WriteString("=== SUMMARY ===\n")
	report.WriteString(fmt.Sprintf("Unchanged: %d\n", unchanged))
	report.WriteString(fmt.Sprintf("Changed:   %d\n", changed))

	return report.String()
}

// ToXML converts ComicInfo to XML string
func (c *ComicInfo) ToXML() (string, error) {
	output, err := xml.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(output), nil
}

// ToMap converts ComicInfo to a map for easier processing
func (c *ComicInfo) ToMap() map[string]string {
	m := make(map[string]string)

	if c.Title != "" {
		m["title"] = c.Title
	}
	if c.Series != "" {
		m["series"] = c.Series
	}
	if c.Number != "" {
		m["number"] = c.Number
	}
	if c.Volume != "" {
		m["volume"] = c.Volume
	}
	if c.Summary != "" {
		m["summary"] = c.Summary
	}
	if c.Writer != "" {
		m["writer"] = c.Writer
	}
	if c.Penciller != "" {
		m["penciller"] = c.Penciller
	}
	if c.Publisher != "" {
		m["publisher"] = c.Publisher
	}
	if c.Genre != "" {
		m["genre"] = c.Genre
	}
	if c.Tags != "" {
		m["tags"] = c.Tags
	}
	if c.Web != "" {
		m["web"] = c.Web
	}
	if c.LanguageISO != "" {
		m["language"] = c.LanguageISO
	}
	if c.AgeRating != "" {
		m["agerating"] = c.AgeRating
	}

	return sortMap(m)
}

func GenerateComicInfo(cbzPath string, outputPath string, updates map[string]string) error {
	// Create ComicInfo struct from updates map
	ci := &ComicInfo{}
	if series, ok := updates["Series"]; ok {
		ci.Series = series
	}
	if number, ok := updates["Number"]; ok {
		ci.Number = number
	}
	// ... map all fields

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(ci, "", "  ")
	if err != nil {
		return err
	}
	comicInfoXML := []byte(xml.Header + string(xmlData))

	// Open CBZ
	r, err := zip.OpenReader(cbzPath)
	if err != nil {
		return err
	}
	defer r.Close()

	// Write modified CBZ
	return writeModifiedCBZ(r, comicInfoXML, outputPath)
}
