package metadata

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/beevik/etree"
)

// Supported metadata fields based on Kavita's EPUB support
var SupportedFields = []string{
	"calibre:series",       // Kavita: Name
	"calibre:series_index", // Kavita: Volume
	"summary",              // Kavita-specific tag
	"dc:description",       // EPUB standard for description
	"dc:publisher",         // Kavita: Publisher
	"dc:creator",           // Kavita: Writer (author)
	"dc:subject",           // Kavita: Genres
	"dc:identifier",        // ISBN (with opf:scheme="isbn")
}

// ---------- EPUB Structures ----------

type Container struct {
	XMLName   xml.Name `xml:"container"`
	Rootfiles []struct {
		FullPath string `xml:"full-path,attr"`
	} `xml:"rootfiles>rootfile"`
}

// ---------- Public Functions ----------

// Read extracts metadata from an EPUB file
func Read(epubPath string, all bool) (map[string]string, error) {
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB: %w", err)
	}
	defer r.Close()

	opfPath, err := findOPFPath(r)
	if err != nil {
		return nil, fmt.Errorf("failed to locate OPF file: %w", err)
	}

	opfContent, err := readFileFromZip(r, opfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read OPF file at '%s': %w", opfPath, err)
	}

	metadata, err := parseMetadataFromOPF(opfContent, all)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return metadata, nil
}

// Update modifies metadata in an EPUB file
func Update(epubPath, outputPath string, updates map[string]string) error {
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return fmt.Errorf("failed to open EPUB: %w", err)
	}
	defer r.Close()

	opfPath, err := findOPFPath(r)
	if err != nil {
		return fmt.Errorf("failed to locate OPF file: %w (EPUB may be invalid)", err)
	}

	opfContent, err := readFileFromZip(r, opfPath)
	if err != nil {
		return fmt.Errorf("failed to read OPF file at '%s': %w", opfPath, err)
	}

	// Modify the OPF content
	modifiedOPF, err := updateOPFMetadata(opfContent, updates)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	// Create new EPUB with modified OPF
	if err := writeModifiedEPUB(r, opfPath, modifiedOPF, outputPath); err != nil {
		return fmt.Errorf("failed to write modified EPUB: %w", err)
	}

	return nil
}

// CompareOPF compares original and modified OPF files and returns a diff report
func CompareOPF(originalPath, modifiedPath string) (string, error) {
	origContent, err := extractOPFFromEPUB(originalPath)
	if err != nil {
		return "", fmt.Errorf("reading original: %w", err)
	}

	modContent, err := extractOPFFromEPUB(modifiedPath)
	if err != nil {
		return "", fmt.Errorf("reading modified: %w", err)
	}

	origMeta, _ := parseMetadataFromOPF(origContent, true)
	modMeta, _ := parseMetadataFromOPF(modContent, true)

	return generateDiffReport(origMeta, modMeta), nil
}

// ---------- Internal Functions ----------

func findOPFPath(r *zip.ReadCloser) (string, error) {
	containerFile, err := findFile(r, "META-INF/container.xml")
	if err != nil {
		return "", fmt.Errorf("container.xml not found in META-INF")
	}

	rc, err := containerFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	var container Container
	if err := xml.NewDecoder(rc).Decode(&container); err != nil {
		return "", fmt.Errorf("error parsing container.xml: %w", err)
	}

	if len(container.Rootfiles) == 0 {
		return "", fmt.Errorf("no rootfile found in container.xml")
	}

	return container.Rootfiles[0].FullPath, nil
}

func findFile(r *zip.ReadCloser, name string) (*zip.File, error) {
	for _, f := range r.File {
		if f.Name == name {
			return f, nil
		}
	}
	return nil, fmt.Errorf("file not found: %s", name)
}

func readFileFromZip(r *zip.ReadCloser, path string) ([]byte, error) {
	file, err := findFile(r, path)
	if err != nil {
		return nil, err
	}

	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return io.ReadAll(rc)
}

func parseMetadataFromOPF(opfContent []byte, all bool) (map[string]string, error) {
	result := make(map[string]string)

	decoder := xml.NewDecoder(bytes.NewReader(opfContent))

	var inMetadata bool
	var currentElement string
	var currentAttrs []xml.Attr
	var content strings.Builder

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error parsing OPF XML: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "metadata" {
				inMetadata = true
			} else if inMetadata {
				currentElement = t.Name.Local
				currentAttrs = t.Attr
				content.Reset()
			}

		case xml.CharData:
			if inMetadata && currentElement != "" {
				content.Write(t)
			}

		case xml.EndElement:
			if t.Name.Local == "metadata" {
				inMetadata = false
			} else if inMetadata && t.Name.Local == currentElement {
				text := strings.TrimSpace(content.String())
				if text != "" {
					extractMetadataValue(currentElement, currentAttrs, text, result)
				}
				currentElement = ""
				currentAttrs = nil
			}
		}
	}

	if !all {
		result = filterSupportedFields(result)
	}

	return sortMap(result), nil
}

func extractMetadataValue(elementName string, attrs []xml.Attr, text string, result map[string]string) {
	var key string

	var scheme, name, property string
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "scheme":
			scheme = attr.Value
		case "name":
			name = attr.Value
		case "property":
			property = attr.Value
		}
	}

	switch elementName {
	case "title":
		key = "title"
	case "creator":
		key = "author"
	case "description":
		key = "summary"
	case "language":
		key = "language"
	case "publisher":
		key = "publisher"
	case "date":
		key = "date"
	case "subject":
		existing := result["subject"]
		if existing != "" {
			result["subject"] = existing + ", " + text
		} else {
			result["subject"] = text
		}
		return
	case "identifier":
		if strings.EqualFold(scheme, "isbn") {
			key = "isbn"
		} else if strings.EqualFold(scheme, "uuid") {
			key = "identifier:uuid"
		} else if scheme != "" {
			key = "identifier:" + scheme
		} else {
			key = "identifier"
		}
	case "meta":
		if name != "" {
			key = name
		} else if property != "" {
			key = property
		}
	}

	if key != "" {
		result[key] = text
	}
}

func updateOPFMetadata(opfContent []byte, updates map[string]string) ([]byte, error) {
	doc := string(opfContent)

	metaStart := strings.Index(doc, "<metadata")
	if metaStart == -1 {
		return nil, fmt.Errorf("metadata section not found")
	}

	metaEnd := strings.Index(doc[metaStart:], "</metadata>")
	if metaEnd == -1 {
		return nil, fmt.Errorf("metadata closing tag not found")
	}
	metaEnd += metaStart

	metaEndTag := metaEnd + len("</metadata>")
	beforeMeta := doc[:metaStart]
	metadataSection := doc[metaStart:metaEndTag]
	afterMeta := doc[metaEndTag:]

	// Apply updates to metadata section
	for key, value := range updates {
		metadataSection = updateMetadataField(metadataSection, key, value)
	}

	result := beforeMeta + metadataSection + afterMeta
	return []byte(result), nil
}

func updateMetadataField(metadata, key, value string) string {
	key = strings.ToLower(key)

	switch key {
	case "author":
		// Update dc:creator (EPUB standard for author/writer)
		metadata = findAndUpdateTag(metadata, "creator", value)
		return metadata

	case "summary":
		// Update BOTH dc:description (EPUB standard) AND Summary (Kavita-specific)
		metadata = findAndUpdateTag(metadata, "description", value)
		metadata = updateOrAddKavitaTag(metadata, "Summary", value)
		return metadata

	case "publisher":
		// Update dc:publisher
		metadata = findAndUpdateTag(metadata, "publisher", value)
		return metadata

	case "isbn":
		return updateISBNIdentifier(metadata, value)

	case "calibre:series":
		return updateOrAddCalibreMeta(metadata, "calibre:series", value)

	case "calibre:series_index":
		return updateOrAddCalibreMeta(metadata, "calibre:series_index", value)

	case "subject", "genre", "subjects":
		// Update dc:subject (maps to Genres in Kavita)
		metadata = findAndUpdateTag(metadata, "subject", value)
		return metadata
	}

	return metadata
}

// findAndUpdateTag safely updates or adds an EPUB metadata tag, removes editor <creator> tags,
// and ensures the corresponding Kavita tag is also updated.
func findAndUpdateTag(metadata, localName, value string) string {
	// Parse the XML
	doc := etree.NewDocument()
	if err := doc.ReadFromString(metadata); err != nil {
		// if parsing fails, fallback to original metadata
		fmt.Println("Warning: failed to parse metadata, returning original")
		return metadata
	}

	// Get <metadata> element
	metaEl := doc.FindElement("//metadata")
	if metaEl == nil {
		// fallback
		return metadata
	}

	// Remove any editor creator tags if updating author
	if localName == "creator" {
		for _, el := range metaEl.FindElements(".//creator") {
			role := el.SelectAttrValue("role", "")
			if role == "edt" {
				el.Parent().RemoveChild(el)
			}
		}
	}

	// Try to find an existing element with the given localName (any prefix)
	var existing *etree.Element
	prefixes := []string{"dc", "opf", ""}
	for _, prefix := range prefixes {
		searchName := localName
		if prefix != "" {
			searchName = prefix + ":" + localName
		}
		existing = metaEl.FindElement(searchName)
		if existing != nil {
			break
		}
	}

	if existing != nil {
		// Update existing tag
		existing.SetText(value)
	} else {
		// Add new tag with dc: prefix
		tagName := "dc:" + localName
		newEl := etree.NewElement(tagName)
		newEl.SetText(value)
		metaEl.AddChild(newEl)
	}

	// Update corresponding Kavita tag (plain tag without namespace)
	metadataStr, _ := doc.WriteToString()
	metadataStr = updateOrAddKavitaTag(metadataStr, localName, value)

	return metadataStr
}

// extractTagName extracts the tag name from an opening tag like "<dc:creator id='1'>"
func extractTagName(openingTag string) string {
	// Remove < and >
	tag := strings.TrimPrefix(openingTag, "<")
	tag = strings.TrimSuffix(tag, ">")

	// Get the part before any space (attributes)
	if spaceIdx := strings.Index(tag, " "); spaceIdx != -1 {
		tag = tag[:spaceIdx]
	}

	return tag
}

// updateOrAddKavitaTag updates or adds Kavita-specific tags (like Summary)
func updateOrAddKavitaTag(metadata, tagName, value string) string {
	// Kavita tags are just plain tags without namespace
	searchPattern := "<" + tagName
	idx := strings.Index(metadata, searchPattern)

	if idx != -1 {
		// Find end of opening tag
		endOfOpen := strings.Index(metadata[idx:], ">")
		if endOfOpen == -1 {
			goto addNew
		}
		endOfOpen += idx

		// Find closing tag
		closeTag := "</" + tagName + ">"
		closeIdx := strings.Index(metadata[endOfOpen:], closeTag)
		if closeIdx == -1 {
			goto addNew
		}
		closeIdx += endOfOpen

		// Replace content
		before := metadata[:endOfOpen+1]
		after := metadata[closeIdx:]
		return before + value + after
	}

addNew:
	// Add new tag
	closeMetadata := strings.Index(metadata, "</metadata>")
	if closeMetadata == -1 {
		return metadata
	}

	newElement := fmt.Sprintf("    <%s>%s</%s>\n  ", tagName, value, tagName)
	return metadata[:closeMetadata] + newElement + metadata[closeMetadata:]
}

func updateISBNIdentifier(metadata, value string) string {
	// Look for existing ISBN identifier
	lowerMeta := strings.ToLower(metadata)
	idx := strings.Index(lowerMeta, `scheme="isbn"`)

	if idx != -1 {
		// Find the start of this identifier element
		startIdx := strings.LastIndex(metadata[:idx], "<dc:identifier")
		if startIdx == -1 {
			startIdx = strings.LastIndex(metadata[:idx], "<identifier")
		}
		if startIdx == -1 {
			goto addNew
		}

		// Find the end of opening tag
		endOfOpen := strings.Index(metadata[startIdx:], ">")
		if endOfOpen == -1 {
			goto addNew
		}
		endOfOpen += startIdx

		// Extract tag name
		openingTag := metadata[startIdx : endOfOpen+1]
		tagName := extractTagName(openingTag)

		// Find closing tag
		closeTag := "</" + tagName + ">"
		closeIdx := strings.Index(metadata[endOfOpen:], closeTag)
		if closeIdx == -1 {
			goto addNew
		}
		closeIdx += endOfOpen

		// Replace content
		before := metadata[:endOfOpen+1]
		after := metadata[closeIdx:]
		return before + value + after
	}

addNew:
	// Add new ISBN identifier
	closeMetadata := strings.Index(metadata, "</metadata>")
	if closeMetadata == -1 {
		return metadata
	}

	// Determine namespace prefix
	prefix := "opf"
	if strings.Contains(metadata, `xmlns:opf=`) {
		prefix = "opf"
	}

	newElement := fmt.Sprintf(`    <dc:identifier %s:scheme="ISBN">%s</dc:identifier>
  `, prefix, value)
	return metadata[:closeMetadata] + newElement + metadata[closeMetadata:]
}

func updateOrAddCalibreMeta(metadata, name, value string) string {
	// Look for existing meta tag with this name
	searchPattern := fmt.Sprintf(`name="%s"`, name)
	lowerMeta := strings.ToLower(metadata)
	lowerPattern := strings.ToLower(searchPattern)

	idx := strings.Index(lowerMeta, lowerPattern)

	if idx != -1 {
		// Find the start of this meta tag
		startIdx := strings.LastIndex(metadata[:idx], "<meta")
		if startIdx == -1 {
			goto addNew
		}

		// Find end of tag
		endSearch := metadata[startIdx:]
		selfClosingIdx := strings.Index(endSearch, "/>")
		normalCloseIdx := strings.Index(endSearch, "</meta>")

		var endIdx int
		if selfClosingIdx != -1 && (normalCloseIdx == -1 || selfClosingIdx < normalCloseIdx) {
			endIdx = startIdx + selfClosingIdx + 2
		} else if normalCloseIdx != -1 {
			normalCloseStart := startIdx + normalCloseIdx
			normalCloseEnd := strings.Index(metadata[normalCloseStart:], ">")
			if normalCloseEnd != -1 {
				endIdx = normalCloseStart + normalCloseEnd + 1
			} else {
				goto addNew
			}
		} else {
			goto addNew
		}

		// Replace entire tag
		newTag := fmt.Sprintf(`<meta name="%s" content="%s"/>`, name, value)
		before := metadata[:startIdx]
		after := metadata[endIdx:]
		return before + newTag + after
	}

addNew:
	// Add new meta tag
	closeMetadata := strings.Index(metadata, "</metadata>")
	if closeMetadata == -1 {
		return metadata
	}

	newTag := fmt.Sprintf(`    <meta name="%s" content="%s"/>
  `, name, value)
	return metadata[:closeMetadata] + newTag + metadata[closeMetadata:]
}

func writeModifiedEPUB(r *zip.ReadCloser, opfPath string, modifiedOPF []byte, outputPath string) error {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	for _, f := range r.File {
		header := &zip.FileHeader{
			Name:   f.Name,
			Method: f.Method,
		}

		w, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		if f.Name == opfPath {
			_, err = w.Write(modifiedOPF)
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

	if err := zw.Close(); err != nil {
		return err
	}

	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

func extractOPFFromEPUB(epubPath string) ([]byte, error) {
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	opfPath, err := findOPFPath(r)
	if err != nil {
		return nil, err
	}

	return readFileFromZip(r, opfPath)
}

func generateDiffReport(original, modified map[string]string) string {
	var report strings.Builder

	report.WriteString("=== METADATA COMPARISON ===\n\n")

	allKeys := make(map[string]bool)
	for k := range original {
		allKeys[k] = true
	}
	for k := range modified {
		allKeys[k] = true
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	unchanged, changed, added, removed := 0, 0, 0, 0

	for _, key := range keys {
		origVal, origExists := original[key]
		modVal, modExists := modified[key]

		if !origExists && modExists {
			report.WriteString(fmt.Sprintf("+ ADDED: %s\n", key))
			report.WriteString(fmt.Sprintf("  New value: %s\n\n", modVal))
			added++
		} else if origExists && !modExists {
			report.WriteString(fmt.Sprintf("- REMOVED: %s\n", key))
			report.WriteString(fmt.Sprintf("  Old value: %s\n\n", origVal))
			removed++
		} else if origVal != modVal {
			report.WriteString(fmt.Sprintf("~ CHANGED: %s\n", key))
			report.WriteString(fmt.Sprintf("  Old: %s\n", origVal))
			report.WriteString(fmt.Sprintf("  New: %s\n\n", modVal))
			changed++
		} else {
			unchanged++
		}
	}

	report.WriteString("=== SUMMARY ===\n")
	report.WriteString(fmt.Sprintf("Unchanged: %d\n", unchanged))
	report.WriteString(fmt.Sprintf("Changed:   %d\n", changed))
	report.WriteString(fmt.Sprintf("Added:     %d\n", added))
	report.WriteString(fmt.Sprintf("Removed:   %d\n", removed))

	return report.String()
}

func filterSupportedFields(m map[string]string) map[string]string {
	filtered := make(map[string]string)

	// Always include these if present
	commonFields := []string{"title", "author", "summary", "publisher", "language", "date", "isbn"}
	for _, field := range commonFields {
		if val, ok := m[field]; ok {
			filtered[field] = val
		}
	}

	// Include calibre fields
	for key, val := range m {
		if strings.HasPrefix(key, "calibre:") {
			filtered[key] = val
		}
	}

	return filtered
}

func sortMap(m map[string]string) map[string]string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sorted := make(map[string]string, len(m))
	for _, k := range keys {
		sorted[k] = m[k]
	}
	return sorted
}
