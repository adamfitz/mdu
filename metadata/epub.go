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

// ---------- EPUB Structures ----------

type EpubContainer struct {
	XMLName   xml.Name `xml:"container"`
	Rootfiles []struct {
		FullPath string `xml:"full-path,attr"`
	} `xml:"rootfiles>rootfile"`
}

type EpubPackage struct {
	XMLName  xml.Name     `xml:"package"`
	Metadata EpubMetadata `xml:"metadata"`
}

type EpubMetadata struct {
	Title       string        `xml:"title"`
	Creator     string        `xml:"creator"`
	Identifier  []EpubDCIdent `xml:"identifier"`
	Description string        `xml:"description"`
	Subject     []string      `xml:"subject"`
	Meta        []EpubMeta    `xml:"meta"`
}

type EpubDCIdent struct {
	ID     string `xml:"id,attr,omitempty"`
	Scheme string `xml:"scheme,attr,omitempty"`
	Value  string `xml:",chardata"`
}

type EpubMeta struct {
	Name    string `xml:"name,attr"`
	Content string `xml:"content,attr"`
}

// ---------- Public EPUB Functions ----------

// EpubRead extracts metadata fields from an EPUB file and returns them in a map.
func EpubRead(epubPath string, all bool) (map[string]string, error) {
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

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

	var container EpubContainer
	if err := xml.NewDecoder(rc).Decode(&container); err != nil {
		return nil, err
	}
	opfPath := container.Rootfiles[0].FullPath

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
	var pkg EpubPackage
	if err := xml.NewDecoder(opfReader).Decode(&pkg); err != nil {
		return nil, err
	}

	result := make(map[string]string)

	if pkg.Metadata.Title != "" {
		result["title"] = pkg.Metadata.Title
	}
	if pkg.Metadata.Creator != "" {
		result["author"] = pkg.Metadata.Creator
	}
	if pkg.Metadata.Description != "" {
		result["summary"] = pkg.Metadata.Description
	}

	for _, id := range pkg.Metadata.Identifier {
		key := id.Value
		if strings.EqualFold(id.Scheme, "isbn") {
			key = "isbn"
		} else if id.Scheme != "" {
			key = "identifier:" + id.Scheme
		}
		result[key] = id.Value
	}

	for _, m := range pkg.Metadata.Meta {
		if m.Name != "" {
			result[m.Name] = m.Content
		}
	}

	if !all {
		filtered := make(map[string]string)
		for _, f := range SupportedFields {
			if val, ok := result[f]; ok {
				filtered[f] = val
			}
		}
		result = filtered
	}

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

// EpubUpdate applies metadata changes and writes a new EPUB file.
func EpubUpdate(epubPath, updatedPath string, updates map[string]string) error {
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return err
	}
	defer r.Close()

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

	var container EpubContainer
	if err := xml.NewDecoder(rc).Decode(&container); err != nil {
		return err
	}
	opfPath := container.Rootfiles[0].FullPath

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
	var pkg EpubPackage
	if err := xml.NewDecoder(opfReader).Decode(&pkg); err != nil {
		return err
	}

	epubUpdateMetadata(&pkg.Metadata, updates)

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

func epubUpdateMetadata(md *EpubMetadata, updates map[string]string) {
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
			continue
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
				md.Meta = append(md.Meta, EpubMeta{Name: key, Content: value})
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
				md.Identifier = append(md.Identifier, EpubDCIdent{
					Scheme: "isbn",
					Value:  value,
				})
			}

		case "author":
			md.Creator = value
		}
	}
}
