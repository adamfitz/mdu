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
)

// Supported metadata fields
var SupportedFields = []string{
	"calibre:series",
	"calibre:series_index",
	"summary",
	"isbn",
	"author",
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
		return nil, err
	}
	defer r.Close()

	opfPath, err := findOPFPath(r)
	if err != nil {
		return nil, err
	}

	opfContent, err := readFileFromZip(r, opfPath)
	if err != nil {
		return nil, err
	}

	return parseMetadataFromOPF(opfContent, all)
}

// Update modifies metadata in an EPUB file
func Update(epubPath, outputPath string, updates map[string]string) error {
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return err
	}
	defer r.Close()

	opfPath, err := findOPFPath(r)
	if err != nil {
		return err
	}

	opfContent, err := readFileFromZip(r, opfPath)
	if err != nil {
		return err
	}

	// Modify the OPF content
	modifiedOPF, err := updateOPFMetadata(opfContent, updates)
	if err != nil {
		return err
	}

	// Create new EPUB with modified OPF
	return writeModifiedEPUB(r, opfPath, modifiedOPF, outputPath)
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

	// Parse the XML and look for metadata section
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
				// We're inside metadata, track the element
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
				// Process the completed element
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

	// Get attributes we care about
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

	// Map element to our field names
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
		// Append to existing subjects
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
			// Old-style meta with name attribute
			key = name
		} else if property != "" {
			// New-style meta with property attribute
			key = property
		}
	}

	if key != "" {
		result[key] = text
	}
}

func updateOPFMetadata(opfContent []byte, updates map[string]string) ([]byte, error) {
	doc := string(opfContent)

	// Find metadata section
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

	// Reconstruct document
	result := beforeMeta + metadataSection + afterMeta
	return []byte(result), nil
}

func updateMetadataField(metadata, key, value string) string {
	key = strings.ToLower(key)

	switch key {
	case "author":
		return updateOrAddElement(metadata, "dc:creator", value, nil)
	case "summary":
		return updateOrAddElement(metadata, "dc:description", value, nil)
	case "isbn":
		// Look for existing ISBN identifier
		return updateISBNIdentifier(metadata, value)
	case "calibre:series":
		return updateOrAddCalibreMeta(metadata, "calibre:series", value)
	case "calibre:series_index":
		return updateOrAddCalibreMeta(metadata, "calibre:series_index", value)
	}

	return metadata
}

func updateISBNIdentifier(metadata, value string) string {
	// Try to find existing ISBN identifier with opf:scheme or scheme attribute
	patterns := []string{
		`<dc:identifier[^>]*opf:scheme="ISBN"[^>]*>`,
		`<dc:identifier[^>]*scheme="ISBN"[^>]*>`,
		`<dc:identifier[^>]*opf:scheme="isbn"[^>]*>`,
		`<dc:identifier[^>]*scheme="isbn"[^>]*>`,
	}

	for _, pattern := range patterns {
		// Simple search for ISBN identifier
		lowerMeta := strings.ToLower(metadata)
		searchStr := strings.ToLower(pattern)
		// Remove regex special chars for simple search
		searchStr = strings.ReplaceAll(searchStr, `[^>]*`, "")
		searchStr = strings.ReplaceAll(searchStr, `>`, "")

		idx := strings.Index(lowerMeta, `scheme="isbn"`)
		if idx != -1 {
			// Find the start of this identifier element
			startIdx := strings.LastIndex(metadata[:idx], "<dc:identifier")
			if startIdx == -1 {
				continue
			}

			// Find the end of opening tag
			endOfOpen := strings.Index(metadata[startIdx:], ">")
			if endOfOpen == -1 {
				continue
			}
			endOfOpen += startIdx

			// Find closing tag
			closeIdx := strings.Index(metadata[endOfOpen:], "</dc:identifier>")
			if closeIdx == -1 {
				continue
			}
			closeIdx += endOfOpen

			// Replace content
			before := metadata[:endOfOpen+1]
			after := metadata[closeIdx:]
			return before + value + after
		}
	}

	// Add new ISBN identifier before </metadata>
	closeMetadata := strings.Index(metadata, "</metadata>")
	if closeMetadata == -1 {
		return metadata
	}

	// Detect namespace prefix (dc vs opf)
	prefix := "opf"
	if strings.Contains(metadata, `xmlns:opf=`) {
		prefix = "opf"
	}

	newElement := fmt.Sprintf(`    <dc:identifier %s:scheme="ISBN">%s</dc:identifier>
  `, prefix, value)
	return metadata[:closeMetadata] + newElement + metadata[closeMetadata:]
}

func updateOrAddElement(metadata, tagName, value string, attrs map[string]string) string {
	// Try to find and update existing element (simple search)
	pattern := fmt.Sprintf("<%s", tagName)
	idx := strings.Index(metadata, pattern)

	if idx != -1 {
		// Find end of opening tag
		endOfOpen := strings.Index(metadata[idx:], ">")
		if endOfOpen == -1 {
			return metadata
		}
		endOfOpen += idx

		// Find closing tag
		closeTag := fmt.Sprintf("</%s>", tagName)
		closeIdx := strings.Index(metadata[endOfOpen:], closeTag)
		if closeIdx == -1 {
			return metadata
		}
		closeIdx += endOfOpen

		// Replace content
		before := metadata[:endOfOpen+1]
		after := metadata[closeIdx:]
		return before + value + after
	}

	// Add new element before </metadata>
	closeMetadata := strings.Index(metadata, "</metadata>")
	if closeMetadata == -1 {
		return metadata
	}

	attrStr := ""
	if attrs != nil {
		for k, v := range attrs {
			attrStr += fmt.Sprintf(` %s="%s"`, k, v)
		}
	}

	newElement := fmt.Sprintf("    <%s%s>%s</%s>\n  ", tagName, attrStr, value, tagName)
	return metadata[:closeMetadata] + newElement + metadata[closeMetadata:]
}

func updateOrAddCalibreMeta(metadata, name, value string) string {
	// Look for existing meta tag with this name
	searchPattern := fmt.Sprintf(`name="%s"`, name)
	idx := strings.Index(metadata, searchPattern)

	if idx != -1 {
		// Find the start of this meta tag
		startIdx := strings.LastIndex(metadata[:idx], "<meta")
		if startIdx == -1 {
			return metadata
		}

		// Find end of tag (could be /> or >)
		selfClosing := strings.Index(metadata[startIdx:], "/>")
		normalClose := strings.Index(metadata[startIdx:], "</meta>")

		var endIdx int
		if selfClosing != -1 && (normalClose == -1 || selfClosing < normalClose) {
			endIdx = startIdx + selfClosing + 2
		} else if normalClose != -1 {
			endIdx = startIdx + normalClose + 7 // length of </meta>
		} else {
			return metadata
		}

		// Replace entire tag
		newTag := fmt.Sprintf(`<meta name="%s" content="%s"/>`, name, value)
		before := metadata[:startIdx]
		after := metadata[endIdx:]
		return before + newTag + after
	}

	// Add new meta tag before </metadata>
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
			// Write modified OPF
			_, err = w.Write(modifiedOPF)
			if err != nil {
				return err
			}
		} else {
			// Copy original file
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

	// Find all keys
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

	// Compare each field
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
	for _, field := range SupportedFields {
		if val, ok := m[field]; ok {
			filtered[field] = val
		}
	}
	// Also include title if present
	if val, ok := m["title"]; ok {
		filtered["title"] = val
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
