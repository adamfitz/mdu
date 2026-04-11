package metadata

import (
	"strings"
	"testing"
)

// --- applyComicInfoUpdates ---

func TestApplyComicInfoUpdates_BasicFields(t *testing.T) {
	ci := &ComicInfo{}
	updates := map[string]string{
		"title":     "My Title",
		"series":    "My Series",
		"number":    "5",
		"volume":    "2",
		"summary":   "A great story",
		"publisher": "Test Publisher",
		"genre":     "Action, Drama",
		"tags":      "tag1, tag2",
		"web":       "https://example.com",
	}
	applyComicInfoUpdates(ci, updates)

	if ci.Title != "My Title" {
		t.Errorf("Title: got %q, want 'My Title'", ci.Title)
	}
	if ci.Series != "My Series" {
		t.Errorf("Series: got %q, want 'My Series'", ci.Series)
	}
	if ci.Number != "5" {
		t.Errorf("Number: got %q, want '5'", ci.Number)
	}
	if ci.Volume != "2" {
		t.Errorf("Volume: got %q, want '2'", ci.Volume)
	}
	if ci.Summary != "A great story" {
		t.Errorf("Summary: got %q, want 'A great story'", ci.Summary)
	}
	if ci.Publisher != "Test Publisher" {
		t.Errorf("Publisher: got %q, want 'Test Publisher'", ci.Publisher)
	}
	if ci.Genre != "Action, Drama" {
		t.Errorf("Genre: got %q, want 'Action, Drama'", ci.Genre)
	}
	if ci.Tags != "tag1, tag2" {
		t.Errorf("Tags: got %q, want 'tag1, tag2'", ci.Tags)
	}
	if ci.Web != "https://example.com" {
		t.Errorf("Web: got %q, want 'https://example.com'", ci.Web)
	}
}

func TestApplyComicInfoUpdates_AuthorAlias(t *testing.T) {
	ci := &ComicInfo{}
	applyComicInfoUpdates(ci, map[string]string{"author": "Some Author"})
	if ci.Writer != "Some Author" {
		t.Errorf("expected Writer to be set via 'author' alias, got %q", ci.Writer)
	}
}

func TestApplyComicInfoUpdates_WriterKey(t *testing.T) {
	ci := &ComicInfo{}
	applyComicInfoUpdates(ci, map[string]string{"writer": "Some Writer"})
	if ci.Writer != "Some Writer" {
		t.Errorf("expected Writer 'Some Writer', got %q", ci.Writer)
	}
}

func TestApplyComicInfoUpdates_CaseInsensitiveKeys(t *testing.T) {
	ci := &ComicInfo{}
	applyComicInfoUpdates(ci, map[string]string{
		"TITLE":     "Upper Title",
		"Publisher": "Mixed Publisher",
	})
	if ci.Title != "Upper Title" {
		t.Errorf("expected case-insensitive title update, got %q", ci.Title)
	}
	if ci.Publisher != "Mixed Publisher" {
		t.Errorf("expected case-insensitive publisher update, got %q", ci.Publisher)
	}
}

func TestApplyComicInfoUpdates_LanguageAlias(t *testing.T) {
	ci := &ComicInfo{}
	applyComicInfoUpdates(ci, map[string]string{"language": "en"})
	if ci.LanguageISO != "en" {
		t.Errorf("expected LanguageISO 'en' via 'language' alias, got %q", ci.LanguageISO)
	}
}

func TestApplyComicInfoUpdates_UnknownKeyIgnored(t *testing.T) {
	ci := &ComicInfo{Title: "Original"}
	applyComicInfoUpdates(ci, map[string]string{"nonexistent": "value"})
	if ci.Title != "Original" {
		t.Errorf("expected title unchanged, got %q", ci.Title)
	}
}

func TestApplyComicInfoUpdates_EmptyUpdates(t *testing.T) {
	ci := &ComicInfo{Title: "Original"}
	applyComicInfoUpdates(ci, map[string]string{})
	if ci.Title != "Original" {
		t.Errorf("expected title unchanged after empty updates, got %q", ci.Title)
	}
}

// --- extractAuthorIDsFromRelationships ---

func TestExtractAuthorIDsFromRelationships_AuthorAndArtist(t *testing.T) {
	relationships := []interface{}{
		map[string]interface{}{"type": "author", "id": "author-uuid-1"},
		map[string]interface{}{"type": "artist", "id": "artist-uuid-2"},
		map[string]interface{}{"type": "cover_art", "id": "cover-uuid-3"},
	}
	ids := extractAuthorIDsFromRelationships(relationships)
	if len(ids) != 2 {
		t.Fatalf("expected 2 IDs (author + artist), got %d", len(ids))
	}
	if ids[0] != "author-uuid-1" {
		t.Errorf("expected first ID 'author-uuid-1', got %q", ids[0])
	}
	if ids[1] != "artist-uuid-2" {
		t.Errorf("expected second ID 'artist-uuid-2', got %q", ids[1])
	}
}

func TestExtractAuthorIDsFromRelationships_NoAuthors(t *testing.T) {
	relationships := []interface{}{
		map[string]interface{}{"type": "cover_art", "id": "cover-uuid-1"},
	}
	ids := extractAuthorIDsFromRelationships(relationships)
	if len(ids) != 0 {
		t.Errorf("expected 0 IDs, got %d", len(ids))
	}
}

func TestExtractAuthorIDsFromRelationships_Empty(t *testing.T) {
	ids := extractAuthorIDsFromRelationships([]interface{}{})
	if len(ids) != 0 {
		t.Errorf("expected 0 IDs for empty list, got %d", len(ids))
	}
}

func TestExtractAuthorIDsFromRelationships_Nil(t *testing.T) {
	ids := extractAuthorIDsFromRelationships(nil)
	if len(ids) != 0 {
		t.Errorf("expected 0 IDs for nil, got %d", len(ids))
	}
}

// --- ComicInfo.ToXML ---

func TestComicInfoToXML_ContainsFields(t *testing.T) {
	ci := &ComicInfo{
		Title:     "Test Title",
		Series:    "Test Series",
		Writer:    "Test Author",
		Publisher: "Test Publisher",
	}
	xml, err := ci.ToXML()
	if err != nil {
		t.Fatalf("ToXML() returned error: %v", err)
	}
	if !strings.Contains(xml, "<Title>Test Title</Title>") {
		t.Errorf("XML missing title, got: %s", xml)
	}
	if !strings.Contains(xml, "<Series>Test Series</Series>") {
		t.Errorf("XML missing series, got: %s", xml)
	}
	if !strings.Contains(xml, "<Writer>Test Author</Writer>") {
		t.Errorf("XML missing writer, got: %s", xml)
	}
	if !strings.Contains(xml, "<?xml") {
		t.Errorf("XML missing XML header")
	}
}

func TestComicInfoToXML_OmitsEmptyFields(t *testing.T) {
	ci := &ComicInfo{Title: "Only Title"}
	xml, err := ci.ToXML()
	if err != nil {
		t.Fatalf("ToXML() returned error: %v", err)
	}
	if strings.Contains(xml, "<Series>") {
		t.Errorf("XML should omit empty Series field")
	}
	if strings.Contains(xml, "<Writer>") {
		t.Errorf("XML should omit empty Writer field")
	}
}

// --- ComicInfo.ToMap ---

func TestComicInfoToMap_PopulatedFields(t *testing.T) {
	ci := &ComicInfo{
		Title:       "My Title",
		Series:      "My Series",
		Writer:      "My Author",
		Publisher:   "My Publisher",
		Genre:       "Action",
		Tags:        "tag1",
		Web:         "https://example.com",
		LanguageISO: "en",
		AgeRating:   "Teen",
	}
	m := ci.ToMap()

	checks := map[string]string{
		"title":     "My Title",
		"series":    "My Series",
		"writer":    "My Author",
		"publisher": "My Publisher",
		"genre":     "Action",
		"tags":      "tag1",
		"web":       "https://example.com",
		"language":  "en",
		"agerating": "Teen",
	}
	for k, want := range checks {
		got, ok := m[k]
		if !ok {
			t.Errorf("ToMap() missing key %q", k)
			continue
		}
		if got != want {
			t.Errorf("ToMap()[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestComicInfoToMap_EmptyFieldsOmitted(t *testing.T) {
	ci := &ComicInfo{Title: "Only Title"}
	m := ci.ToMap()
	if _, ok := m["series"]; ok {
		t.Error("ToMap() should not include empty 'series'")
	}
	if _, ok := m["writer"]; ok {
		t.Error("ToMap() should not include empty 'writer'")
	}
	if _, ok := m["title"]; !ok {
		t.Error("ToMap() should include non-empty 'title'")
	}
}

// --- generateComicInfoDiff ---

func TestGenerateComicInfoDiff_DetectsChange(t *testing.T) {
	orig := &ComicInfo{Title: "Old Title", Series: "Series"}
	mod := &ComicInfo{Title: "New Title", Series: "Series"}
	diff := generateComicInfoDiff(orig, mod)

	if !strings.Contains(diff, "CHANGED") {
		t.Error("expected diff to report a CHANGED field")
	}
	if !strings.Contains(diff, "Old Title") {
		t.Error("expected diff to contain old value")
	}
	if !strings.Contains(diff, "New Title") {
		t.Error("expected diff to contain new value")
	}
}

func TestGenerateComicInfoDiff_NoChanges(t *testing.T) {
	ci := &ComicInfo{Title: "Same Title", Series: "Same Series"}
	diff := generateComicInfoDiff(ci, ci)

	if strings.Contains(diff, "CHANGED") {
		t.Error("expected no CHANGED fields when identical")
	}
	if !strings.Contains(diff, "Unchanged: 2") {
		t.Errorf("expected 'Unchanged: 2' in diff, got: %s", diff)
	}
}

func TestGenerateComicInfoDiff_SummaryLine(t *testing.T) {
	orig := &ComicInfo{Title: "A", Writer: "Author1"}
	mod := &ComicInfo{Title: "B", Writer: "Author1"}
	diff := generateComicInfoDiff(orig, mod)

	if !strings.Contains(diff, "=== SUMMARY ===") {
		t.Error("expected summary section in diff")
	}
	if !strings.Contains(diff, "Changed:   1") {
		t.Errorf("expected 'Changed: 1', got: %s", diff)
	}
}
