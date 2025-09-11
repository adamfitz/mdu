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

// Contains the metadata fields this tool can read/update
var SupportedFields = []string{
	"calibre:series",
	"calibre:series_index",
	"summary",
	"isbn",   // represents dc:identifier with opf:scheme="isbn"
	"author", // maps to dc:creator
}

// ---------- EPUB Structures ----------

type Container struct {
	XMLName   xml.Name `xml:"container"`
	Rootfiles []struct {
		FullPath string `xml:"full-path,attr"`
	} `xml:"rootfiles>rootfile"`
}

type Package struct {
	XMLName  xml.Name `xml:"package"`
	Metadata Metadata `xml:"metadata"`
}

type Metadata struct {
	Title       string    `xml:"title"`
	Creator     string    `xml:"creator"`
	Identifier  []DCIdent `xml:"identifier"`
	Description string    `xml:"description"`
	Subject     []string  `xml:"subject"`
	Meta        []Meta    `xml:"meta"`
}

type DCIdent struct {
	ID     string `xml:"id,attr,omitempty"`
	Scheme string `xml:"scheme,attr,omitempty"`
	Value  string `xml:",chardata"`
}

type Meta struct {
	Name    string `xml:"name,attr"`
	Content string `xml:"content,attr"`
}

// ---------- Public Functions ----------

// Read extracts metadata fields from an EPUB file and returns them in a map.
func Read(epubPath string, all bool) (map[string]string, error) {
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// Locate container.xml
	var containerFile *zip.File
	for _, f := range r.File {
		if f.Name == "META-INF/container.xml" {
			containerFile = f
			break
		}
	}
	if containerFile == nil {
		return nil, fmt.Errorf("container.xml not found")
	}

	rc, _ := containerFile.Open()
	defer rc.Close()

	var container Container
	if err := xml.NewDecoder(rc).Decode(&container); err != nil {
		return nil, err
	}
	opfPath := container.Rootfiles[0].FullPath

	// Locate OPF file
	var opfFile *zip.File
	for _, f := range r.File {
		if f.Name == opfPath {
			opfFile = f
			break
		}
	}
	if opfFile == nil {
		return nil, fmt.Errorf("OPF file not found")
	}

	opfReader, _ := opfFile.Open()
	defer opfReader.Close()
	var pkg Package
	if err := xml.NewDecoder(opfReader).Decode(&pkg); err != nil {
		return nil, err
	}

	// Collect all metadata
	result := make(map[string]string)

	// Standard fields
	if pkg.Metadata.Title != "" {
		result["title"] = pkg.Metadata.Title
	}
	if pkg.Metadata.Creator != "" {
		result["author"] = pkg.Metadata.Creator
	}
	if pkg.Metadata.Description != "" {
		result["summary"] = pkg.Metadata.Description
	}

	// Identifiers
	for _, id := range pkg.Metadata.Identifier {
		key := id.Value
		if strings.EqualFold(id.Scheme, "isbn") {
			key = "isbn"
		} else if id.Scheme != "" {
			key = "identifier:" + id.Scheme
		}
		result[key] = id.Value
	}

	// Meta tags
	for _, m := range pkg.Metadata.Meta {
		if m.Name != "" {
			result[m.Name] = m.Content
		}
	}

	// If not --all, filter to only supported fields
	if !all {
		filtered := make(map[string]string)
		for _, f := range SupportedFields {
			if val, ok := result[f]; ok {
				filtered[f] = val
			}
		}
		result = filtered
	}

	// Sort keys alphabetically
	sortedResult := make(map[string]string)
	keys := make([]string, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sortedResult[k] = result[k]
	}

	return sortedResult, nil
}

// Update applies metadata changes and writes a new EPUB file.
func Update(epubPath, updatedPath string, updates map[string]string) error {
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return err
	}
	defer r.Close()

	// Locate container.xml
	var containerFile *zip.File
	for _, f := range r.File {
		if f.Name == "META-INF/container.xml" {
			containerFile = f
			break
		}
	}
	if containerFile == nil {
		return fmt.Errorf("container.xml not found")
	}

	rc, _ := containerFile.Open()
	defer rc.Close()

	var container Container
	if err := xml.NewDecoder(rc).Decode(&container); err != nil {
		return err
	}
	opfPath := container.Rootfiles[0].FullPath

	// Locate OPF file
	var opfFile *zip.File
	for _, f := range r.File {
		if f.Name == opfPath {
			opfFile = f
			break
		}
	}
	if opfFile == nil {
		return fmt.Errorf("OPF file not found")
	}

	opfReader, _ := opfFile.Open()
	defer opfReader.Close()
	var pkg Package
	if err := xml.NewDecoder(opfReader).Decode(&pkg); err != nil {
		return err
	}

	// Apply updates
	updateMetadata(&pkg.Metadata, updates)

	// Write back new EPUB
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	for _, f := range r.File {
		w, _ := zw.Create(f.Name)
		if f.Name == opfPath {
			newData, _ := xml.MarshalIndent(pkg, "", "  ")
			w.Write([]byte(xml.Header))
			w.Write(newData)
		} else {
			rc, _ := f.Open()
			io.Copy(w, rc)
			rc.Close()
		}
	}
	zw.Close()

	return os.WriteFile(updatedPath, buf.Bytes(), 0644)
}

// ---------- Internal Metadata Update Logic ----------

func updateMetadata(md *Metadata, updates map[string]string) {
	// Helper to check if a key is supported
	isSupported := func(key string) bool {
		for _, f := range SupportedFields {
			if strings.EqualFold(f, key) {
				return true
			}
		}
		return false
	}

	for key, value := range updates {
		if !isSupported(key) {
			continue // skip unsupported keys
		}

		switch strings.ToLower(key) {

		case "calibre:series", "calibre:series_index":
			found := false
			for i := range md.Meta {
				if strings.EqualFold(md.Meta[i].Name, key) {
					md.Meta[i].Content = value
					found = true
					break
				}
			}
			if !found {
				md.Meta = append(md.Meta, Meta{Name: key, Content: value})
			}

		case "summary":
			md.Description = value

		case "isbn":
			updated := false
			for i := range md.Identifier {
				if strings.EqualFold(md.Identifier[i].Scheme, "isbn") {
					md.Identifier[i].Value = value
					updated = true
					break
				}
			}
			if !updated {
				md.Identifier = append(md.Identifier, DCIdent{
					Scheme: "isbn",
					Value:  value,
				})
			}

		case "author":
			md.Creator = value
		}
	}
}
