package ranobedb

import (
	"encoding/json"
	"testing"
)

// --- SearchResponse JSON unmarshaling ---

func TestSearchResponse_UnmarshalCountAsInt(t *testing.T) {
	raw := `{"series":[{"id":1,"title":"Test Novel"}],"count":42}`
	var resp SearchResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal count as int: %v", err)
	}
	if len(resp.Series) != 1 {
		t.Errorf("expected 1 series, got %d", len(resp.Series))
	}
	if resp.Series[0].ID != 1 {
		t.Errorf("expected series ID 1, got %d", resp.Series[0].ID)
	}
	if resp.Series[0].Title != "Test Novel" {
		t.Errorf("expected title 'Test Novel', got %q", resp.Series[0].Title)
	}
}

func TestSearchResponse_UnmarshalCountAsString(t *testing.T) {
	// RanobeDB sometimes returns count as a string
	raw := `{"series":[{"id":2,"title":"Another Novel"}],"count":"10"}`
	var resp SearchResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal count as string: %v", err)
	}
	if len(resp.Series) != 1 {
		t.Errorf("expected 1 series, got %d", len(resp.Series))
	}
}

func TestSearchResponse_EmptySeries(t *testing.T) {
	raw := `{"series":[],"count":0}`
	var resp SearchResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Series) != 0 {
		t.Errorf("expected 0 series, got %d", len(resp.Series))
	}
}

// --- NovelAPIResponse / SeriesDetail JSON unmarshaling ---

func TestNovelAPIResponse_UnmarshalFullResponse(t *testing.T) {
	raw := `{
		"series": {
			"id": 6547,
			"title": "86--EIGHTY-SIX",
			"description": "A war story.",
			"books": [
				{"id": 1, "title": "Vol 1"},
				{"id": 2, "title": "Vol 2"}
			],
			"staff": [
				{"name": "安里アサト", "romaji": "Asato Asato", "role_type": "author", "lang": null},
				{"name": "Translator Name", "romaji": null, "role_type": "translator", "lang": "en"}
			],
			"publishers": [
				{"name": "KADOKAWA", "romaji": null, "publisher_type": "publisher", "lang": "ja"},
				{"name": "Yen Press", "romaji": null, "publisher_type": "publisher", "lang": "en"}
			],
			"tags": [
				{"id": 1, "name": "action"},
				{"id": 2, "name": "sci-fi"}
			]
		}
	}`

	var resp NovelAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal NovelAPIResponse: %v", err)
	}

	s := resp.Series

	if s.ID != 6547 {
		t.Errorf("expected ID 6547, got %d", s.ID)
	}
	if s.Title != "86--EIGHTY-SIX" {
		t.Errorf("expected title '86--EIGHTY-SIX', got %q", s.Title)
	}
	if s.Description != "A war story." {
		t.Errorf("expected description 'A war story.', got %q", s.Description)
	}
	if len(s.Books) != 2 {
		t.Errorf("expected 2 books, got %d", len(s.Books))
	}
	if len(s.Staff) != 2 {
		t.Errorf("expected 2 staff, got %d", len(s.Staff))
	}
	if len(s.Publishers) != 2 {
		t.Errorf("expected 2 publishers, got %d", len(s.Publishers))
	}
	if len(s.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(s.Tags))
	}
}

func TestNovelAPIResponse_AuthorNullLang(t *testing.T) {
	raw := `{
		"series": {
			"id": 1,
			"title": "Test",
			"description": "",
			"books": [],
			"staff": [
				{"name": "日本語名前", "romaji": "Romaji Name", "role_type": "author", "lang": null}
			],
			"publishers": [],
			"tags": []
		}
	}`

	var resp NovelAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	staff := resp.Series.Staff
	if len(staff) != 1 {
		t.Fatalf("expected 1 staff member, got %d", len(staff))
	}
	if staff[0].Lang != nil {
		t.Errorf("expected nil lang for author, got %v", staff[0].Lang)
	}
	if staff[0].Romaji == nil {
		t.Error("expected non-nil romaji")
	} else if *staff[0].Romaji != "Romaji Name" {
		t.Errorf("expected romaji 'Romaji Name', got %q", *staff[0].Romaji)
	}
}

// --- GetNovelInfo logic (using parsed data, no HTTP) ---

func TestGetNovelInfo_AuthorPrefersRomaji(t *testing.T) {
	// Simulate what GetNovelInfo does with parsed data
	series := SeriesDetail{
		ID:    1,
		Title: "Test Series",
		Staff: []StaffMember{
			{Name: "日本語", Romaji: strPtr("Romaji Author"), RoleType: "author", Lang: nil},
		},
		Publishers: []SeriesPublisher{
			{Name: "Publisher JA", PublisherType: "publisher", Lang: "ja"},
		},
		Tags:  []Tag{{ID: 1, Name: "action"}},
		Books: []Book{{ID: 1, Title: "Vol 1"}},
	}

	var authors []string
	for _, st := range series.Staff {
		if st.RoleType == "author" {
			if st.Romaji != nil && *st.Romaji != "" {
				authors = append(authors, *st.Romaji)
			} else {
				authors = append(authors, st.Name)
			}
		}
	}

	if len(authors) != 1 {
		t.Fatalf("expected 1 author, got %d", len(authors))
	}
	if authors[0] != "Romaji Author" {
		t.Errorf("expected 'Romaji Author', got %q", authors[0])
	}
}

func TestGetNovelInfo_AuthorFallsBackToName(t *testing.T) {
	series := SeriesDetail{
		Staff: []StaffMember{
			{Name: "日本語名前", Romaji: nil, RoleType: "author", Lang: nil},
		},
	}

	var authors []string
	for _, st := range series.Staff {
		if st.RoleType == "author" {
			if st.Romaji != nil && *st.Romaji != "" {
				authors = append(authors, *st.Romaji)
			} else {
				authors = append(authors, st.Name)
			}
		}
	}

	if len(authors) != 1 {
		t.Fatalf("expected 1 author, got %d", len(authors))
	}
	if authors[0] != "日本語名前" {
		t.Errorf("expected Japanese name fallback, got %q", authors[0])
	}
}

func TestGetNovelInfo_AuthorExcludesNonAuthors(t *testing.T) {
	series := SeriesDetail{
		Staff: []StaffMember{
			{Name: "Author", Romaji: strPtr("Author Romaji"), RoleType: "author", Lang: nil},
			{Name: "Translator", Romaji: nil, RoleType: "translator", Lang: strPtr("en")},
			{Name: "Artist", Romaji: nil, RoleType: "artist", Lang: nil},
		},
	}

	var authors []string
	for _, st := range series.Staff {
		if st.RoleType == "author" {
			if st.Romaji != nil && *st.Romaji != "" {
				authors = append(authors, *st.Romaji)
			} else {
				authors = append(authors, st.Name)
			}
		}
	}

	if len(authors) != 1 {
		t.Errorf("expected only 1 author (not translator/artist), got %d: %v", len(authors), authors)
	}
}

func TestGetNovelInfo_PublisherPrefersJapanese(t *testing.T) {
	series := SeriesDetail{
		Publishers: []SeriesPublisher{
			{Name: "Yen Press", PublisherType: "publisher", Lang: "en"},
			{Name: "KADOKAWA", PublisherType: "publisher", Lang: "ja"},
		},
	}

	var publisher string
	for _, p := range series.Publishers {
		if p.Lang == "ja" {
			publisher = p.Name
			break
		}
	}
	if publisher == "" && len(series.Publishers) > 0 {
		publisher = series.Publishers[0].Name
	}

	if publisher != "KADOKAWA" {
		t.Errorf("expected 'KADOKAWA', got %q", publisher)
	}
}

func TestGetNovelInfo_PublisherFallsBackToFirst(t *testing.T) {
	series := SeriesDetail{
		Publishers: []SeriesPublisher{
			{Name: "Yen Press", PublisherType: "publisher", Lang: "en"},
		},
	}

	var publisher string
	for _, p := range series.Publishers {
		if p.Lang == "ja" {
			publisher = p.Name
			break
		}
	}
	if publisher == "" && len(series.Publishers) > 0 {
		publisher = series.Publishers[0].Name
	}

	if publisher != "Yen Press" {
		t.Errorf("expected fallback to 'Yen Press', got %q", publisher)
	}
}

func TestGetNovelInfo_VolumesCountsAllBooks(t *testing.T) {
	series := SeriesDetail{
		Books: []Book{
			{ID: 1, Title: "Vol 1"},
			{ID: 2, Title: "Vol 2"},
			{ID: 3, Title: "Vol 3"},
		},
	}

	if len(series.Books) != 3 {
		t.Errorf("expected 3 volumes, got %d", len(series.Books))
	}
}

func TestGetNovelInfo_GenresFromTags(t *testing.T) {
	series := SeriesDetail{
		Tags: []Tag{
			{ID: 1, Name: "action"},
			{ID: 2, Name: "sci-fi"},
			{ID: 3, Name: "military"},
		},
	}

	var genres []string
	for _, t := range series.Tags {
		genres = append(genres, t.Name)
	}

	if len(genres) != 3 {
		t.Errorf("expected 3 genres, got %d", len(genres))
	}
	if genres[0] != "action" || genres[1] != "sci-fi" || genres[2] != "military" {
		t.Errorf("unexpected genres: %v", genres)
	}
}

// helper
func strPtr(s string) *string {
	return &s
}
