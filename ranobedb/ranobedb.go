package ranobedb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// =====================
// Structs (Search)
// =====================

type SearchResponse struct {
	Series []Series    `json:"series"`
	Count  json.Number `json:"count"`
}

type Series struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// =====================
// Structs (Details)
// =====================

type NovelAPIResponse struct {
	Series SeriesDetail `json:"series"`
}

type SeriesDetail struct {
	ID          int               `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Books       []Book            `json:"books"`
	Staff       []StaffMember     `json:"staff"`
	Publishers  []SeriesPublisher `json:"publishers"`
	Tags        []Tag             `json:"tags"`
}

type Book struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type StaffMember struct {
	Name     string  `json:"name"`
	Romaji   *string `json:"romaji"`
	RoleType string  `json:"role_type"`
	Lang     *string `json:"lang"` // nullable — authors have null lang
}

type SeriesPublisher struct {
	Name          string  `json:"name"`
	Romaji        *string `json:"romaji"`
	PublisherType string  `json:"publisher_type"`
	Lang          string  `json:"lang"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// =====================
// Final Clean Struct (used by CLI)
// =====================

type NovelInfo struct {
	ID          int
	Title       string
	Authors     []string
	Description string
	Publisher   string
	Genres      []string
	Volumes     int
}

// =====================
// Client
// =====================

type RanobeDB struct{}

// =====================
// Search
// =====================

func (s *RanobeDB) SearchNovel(name string) (SearchResponse, error) {
	novelQuery := url.QueryEscape(name)
	url := fmt.Sprintf("https://ranobedb.org/api/v0/series?q=%s", novelQuery)

	resp, err := http.Get(url)
	if err != nil {
		return SearchResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SearchResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return SearchResponse{}, fmt.Errorf("ranobedb search failed: %s\n%s", resp.Status, string(body))
	}

	if len(body) == 0 {
		return SearchResponse{}, fmt.Errorf("empty response from RanobeDB")
	}

	var result SearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return SearchResponse{}, err
	}

	return result, nil
}

// =====================
// Get Full Info
// =====================

func (s *RanobeDB) GetNovelInfo(id string) (NovelInfo, error) {
	url := fmt.Sprintf("https://ranobedb.org/api/v0/series/%s", id)

	resp, err := http.Get(url)
	if err != nil {
		return NovelInfo{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NovelInfo{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return NovelInfo{}, fmt.Errorf("ranobedb fetch failed: %s\n%s", resp.Status, string(body))
	}

	var apiResp NovelAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return NovelInfo{}, err
	}

	series := apiResp.Series

	// Authors have null lang in RanobeDB — filter by role_type only
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

	// Prefer ja publisher, fallback to any
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

	var genres []string
	for _, t := range series.Tags {
		genres = append(genres, t.Name)
	}

	return NovelInfo{
		ID:          series.ID,
		Title:       series.Title,
		Authors:     authors,
		Description: series.Description,
		Publisher:   publisher,
		Genres:      genres,
		Volumes:     len(series.Books),
	}, nil
}
