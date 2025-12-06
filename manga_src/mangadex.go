// mangadex.go is for interacting with managadex api to retrieve metadata about specific manga titles
package mangasrc

import (
	"encoding/json"
	"fmt"
	"io"
	//"log"
	"net/http"
	"net/url"
)

var mangadexApiBaseUri string = "https://api.mangadex.org"

//var mangadexBaseUri string = "https://mangadex.org"

// Generic MangaDex response
type MangadexSearchResponse struct {
	Results []MangadexTitleMetadata `json:"data"`
}

// One manga entry (VERY flexible – accepts all metadata)
type MangadexTitleMetadata struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Attributes    map[string]any `json:"attributes"`
	Relationships []any          `json:"relationships"`
}

type MangadexTitleSearchResponse struct {
	ID        string   `json:"id"`
	MainTitle string   `json:"mainTitle"`
	AltTitles []string `json:"altTitles"`
}

// Function to search for a manga by name (title) and return the full metadata
func TitleMetadata(mangadexId string) (*MangadexTitleMetadata, error) {
	// Build URL using the ID, not the title
	mangaURL := fmt.Sprintf("%s/manga/%s", mangadexApiBaseUri, mangadexId)

	res, err := http.Get(mangaURL)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// The response for /manga/{id} is a single object, NOT an array
	// so we use MangadexTitleMetadata directly, not MangadexSearchResponse
	var result struct {
		Data MangadexTitleMetadata `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	return &result.Data, nil
}

// Search manga titles from mangadex, should be passed to another func to determine the best match
// (remove the punctuation and build string amtch etc)
func MangadexTitleSearch(name string) ([]MangadexTitleSearchResponse, error) {
	baseURL := mangadexApiBaseUri + "/manga"
	params := url.Values{}
	params.Add("title", name)

	mangaURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	res, err := http.Get(mangaURL)
	if err != nil {
		return nil, fmt.Errorf("error requesting manga info: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var response MangadexSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	if len(response.Results) == 0 {
		return nil, fmt.Errorf("no manga found for title: %s", name)
	}

	results := make([]MangadexTitleSearchResponse, 0, len(response.Results))

	for _, item := range response.Results {
		attrs := item.Attributes

		// ------ MAIN TITLE ------
		mainTitle := ""
		if titles, ok := attrs["title"].(map[string]any); ok {
			if en, ok := titles["en"].(string); ok {
				mainTitle = en
			} else {
				for _, v := range titles {
					if s, ok := v.(string); ok {
						mainTitle = s
						break
					}
				}
			}
		}

		// ------ ALT TITLES ------
		altTitles := []string{}
		if ats, ok := attrs["altTitles"].([]any); ok {
			for _, at := range ats {
				if m, ok := at.(map[string]any); ok {
					for _, v := range m {
						if s, ok := v.(string); ok {
							altTitles = append(altTitles, s)
						}
					}
				}
			}
		}

		results = append(results, MangadexTitleSearchResponse{
			ID:        item.ID,
			MainTitle: mainTitle,
			AltTitles: altTitles,
		})
	}

	return results, nil
}
