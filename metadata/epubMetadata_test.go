package metadata

import (
	"strings"
	"testing"

	"mdu/ranobedb"
)

// --- BuildMetadataMap ---

func TestBuildMetadataMap_WithAuthors(t *testing.T) {
	novel := ranobedb.NovelInfo{
		Title:       "86--EIGHTY-SIX",
		Authors:     []string{"Asato Asato"},
		Description: "A war story.",
		Publisher:   "KADOKAWA",
		Genres:      []string{"action", "sci-fi"},
		Volumes:     13,
	}
	m := BuildMetadataMap(novel)

	if m["title"] != "86--EIGHTY-SIX" {
		t.Errorf("title: got %q, want '86--EIGHTY-SIX'", m["title"])
	}
	if m["author"] != "Asato Asato" {
		t.Errorf("author: got %q, want 'Asato Asato'", m["author"])
	}
	if m["summary"] != "A war story." {
		t.Errorf("summary: got %q, want 'A war story.'", m["summary"])
	}
	if m["publisher"] != "KADOKAWA" {
		t.Errorf("publisher: got %q, want 'KADOKAWA'", m["publisher"])
	}
	if m["subject"] != "action, sci-fi" {
		t.Errorf("subject: got %q, want 'action, sci-fi'", m["subject"])
	}
	if m["calibre:series"] != "86--EIGHTY-SIX" {
		t.Errorf("calibre:series: got %q", m["calibre:series"])
	}
	if m["calibre:series_index"] != "13" {
		t.Errorf("calibre:series_index: got %q, want '13'", m["calibre:series_index"])
	}
}

func TestBuildMetadataMap_MultipleAuthors(t *testing.T) {
	novel := ranobedb.NovelInfo{
		Title:   "Test",
		Authors: []string{"Author One", "Author Two"},
	}
	m := BuildMetadataMap(novel)
	if m["author"] != "Author One, Author Two" {
		t.Errorf("expected joined authors, got %q", m["author"])
	}
}

func TestBuildMetadataMap_NoAuthors(t *testing.T) {
	novel := ranobedb.NovelInfo{
		Title:   "Test",
		Authors: []string{},
	}
	m := BuildMetadataMap(novel)
	if m["author"] != "Unknown" {
		t.Errorf("expected 'Unknown' for empty authors, got %q", m["author"])
	}
}

func TestBuildMetadataMap_NoGenres(t *testing.T) {
	novel := ranobedb.NovelInfo{
		Title:  "Test",
		Genres: []string{},
	}
	m := BuildMetadataMap(novel)
	if _, ok := m["subject"]; ok {
		t.Errorf("expected no 'subject' key when genres empty, got %q", m["subject"])
	}
}

func TestBuildMetadataMap_NoPublisher(t *testing.T) {
	novel := ranobedb.NovelInfo{
		Title:     "Test",
		Publisher: "",
	}
	m := BuildMetadataMap(novel)
	if _, ok := m["publisher"]; ok {
		t.Errorf("expected no 'publisher' key when publisher empty, got %q", m["publisher"])
	}
}

// --- parseMetadataFromOPF ---

func TestParseMetadataFromOPF_BasicFields(t *testing.T) {
	opf := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Title</dc:title>
    <dc:creator>Test Author</dc:creator>
    <dc:publisher>Test Publisher</dc:publisher>
    <dc:language>en</dc:language>
    <dc:date>2024-01-01</dc:date>
    <dc:description>Test description</dc:description>
  </metadata>
</package>`

	result, err := parseMetadataFromOPF([]byte(opf), false)
	if err != nil {
		t.Fatalf("parseMetadataFromOPF() error: %v", err)
	}

	checks := map[string]string{
		"title":     "Test Title",
		"author":    "Test Author",
		"publisher": "Test Publisher",
		"language":  "en",
		"date":      "2024-01-01",
		"summary":   "Test description",
	}
	for k, want := range checks {
		got, ok := result[k]
		if !ok {
			t.Errorf("missing key %q in result", k)
			continue
		}
		if got != want {
			t.Errorf("result[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestParseMetadataFromOPF_ISBNIdentifier(t *testing.T) {
	opf := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test</dc:title>
    <dc:identifier opf:scheme="ISBN">978-1234567890</dc:identifier>
  </metadata>
</package>`

	result, err := parseMetadataFromOPF([]byte(opf), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, ok := result["isbn"]; ok {
		if got != "978-1234567890" {
			t.Errorf("isbn: got %q, want '978-1234567890'", got)
		}
	}
}

func TestParseMetadataFromOPF_SubjectsJoined(t *testing.T) {
	opf := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test</dc:title>
    <dc:subject>action</dc:subject>
    <dc:subject>sci-fi</dc:subject>
  </metadata>
</package>`

	result, err := parseMetadataFromOPF([]byte(opf), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	subject, ok := result["subject"]
	if !ok {
		t.Fatal("expected 'subject' in result")
	}
	if !strings.Contains(subject, "action") || !strings.Contains(subject, "sci-fi") {
		t.Errorf("expected joined subjects, got %q", subject)
	}
}

func TestParseMetadataFromOPF_EmptyMetadata(t *testing.T) {
	opf := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
  </metadata>
</package>`

	result, err := parseMetadataFromOPF([]byte(opf), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result for empty metadata, got %v", result)
	}
}

func TestParseMetadataFromOPF_CalibreMetaTag(t *testing.T) {
	opf := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test</dc:title>
    <meta name="calibre:series" content="My Series"/>
    <meta name="calibre:series_index" content="3"/>
  </metadata>
</package>`

	result, err := parseMetadataFromOPF([]byte(opf), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["calibre:series"] != "My Series" {
		t.Errorf("calibre:series: got %q, want 'My Series'", result["calibre:series"])
	}
	if result["calibre:series_index"] != "3" {
		t.Errorf("calibre:series_index: got %q, want '3'", result["calibre:series_index"])
	}
}
