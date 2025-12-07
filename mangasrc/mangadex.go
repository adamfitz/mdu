// mangadex.go is for interacting with managadex api to retrieve metadata about specific manga titles
package mangasrc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	if mangadexId == "" {
		return nil, fmt.Errorf("manga ID is empty")
	}

	url := fmt.Sprintf("%s/manga/%s", mangadexApiBaseUri, mangadexId)

	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}

	var wrapper struct {
		Data MangadexTitleMetadata `json:"data"`
	}

	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w\nbody: %s", err, string(body))
	}

	return &wrapper.Data, nil
}

// Search manga titles from mangadex, should be passed to another func to determine the best match
// (remove the punctuation and build string match etc)
func MangadexTitleSearch(name string) ([]MangadexTitleSearchResponse, error) {
	baseURL := mangadexApiBaseUri + "/manga"
	params := url.Values{}
	params.Add("title", name)

	mangaURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// DEBUG: log final request URL
	log.Printf("[Mangadex] Request URL: %s", mangaURL)

	res, err := http.Get(mangaURL)
	if err != nil {
		log.Printf("[Mangadex] HTTP request failed: %v", err)
		return nil, fmt.Errorf("error requesting manga info: %w", err)
	}
	defer res.Body.Close()

	// DEBUG: log HTTP status
	log.Printf("[Mangadex] Status: %d %s", res.StatusCode, res.Status)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("[Mangadex] Error reading body: %v", err)
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// DEBUG: log raw body if status is not 200
	if res.StatusCode != http.StatusOK {
		log.Printf("[Mangadex] NON-200 response body:\n%s", string(body))
		return nil, fmt.Errorf("bad status %d from mangadex", res.StatusCode)
	}

	// DEBUG: detect HTML (Cloudflare or error page)
	if len(body) > 0 && body[0] == '<' {
		log.Printf("[Mangadex] WARNING: Body starts with '<' (HTML returned instead of JSON)\n%s", string(body))
		return nil, fmt.Errorf("mangadex returned HTML, not JSON")
	}

	var response MangadexSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("[Mangadex] JSON decode error: %v", err)
		log.Printf("[Mangadex] RAW BODY for debugging:\n%s", string(body))
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	if len(response.Results) == 0 {
		log.Printf("[Mangadex] Empty results for title: %s", name)
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

// converts mangadex response to the conicinfo sype struct
func MangadexToComicInfo(md *MangadexTitleMetadata) map[string]string {
	updates := make(map[string]string)

	// Extract title
	if title, ok := md.Attributes["title"].(map[string]any); ok {
		if enTitle, ok := title["en"].(string); ok {
			updates["Series"] = enTitle
		}
	}

	// Extract description
	if desc, ok := md.Attributes["description"].(map[string]any); ok {
		if enDesc, ok := desc["en"].(string); ok {
			updates["Summary"] = enDesc
		}
	}

	// Extract tags/genres, etc.

	return updates
}
