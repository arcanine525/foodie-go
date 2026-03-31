package youtube

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const baseURL = "https://www.googleapis.com/youtube/v3"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetVideos calls videos.list with batch IDs (max 50 per call).
// Quota cost: 1 unit per call.
func (c *Client) GetVideos(videoIDs []string) (*APIResponse, error) {
	params := url.Values{
		"part": {"snippet,contentDetails,statistics,status"},
		"id":   {strings.Join(videoIDs, ",")},
		"key":  {c.apiKey},
	}
	return c.get("videos", params)
}

// SearchVideos calls search.list.
// Quota cost: 100 units per call — use sparingly.
func (c *Client) SearchVideos(query string, maxResults int) (*SearchResponse, error) {
	params := url.Values{
		"part":              {"id,snippet"},
		"q":                 {query},
		"type":              {"video"},
		"maxResults":        {fmt.Sprintf("%d", maxResults)},
		"regionCode":        {"VN"},
		"relevanceLanguage": {"vi"},
		"videoEmbeddable":   {"true"},
		"key":               {c.apiKey},
	}

	resp, err := c.httpClient.Get(baseURL + "/search?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned status %d", resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}
	return &result, nil
}

// GetPlaylistItems calls playlistItems.list. Quota cost: 1 unit per call.
func (c *Client) GetPlaylistItems(playlistID, pageToken string) (*PlaylistResponse, error) {
	params := url.Values{
		"part":       {"contentDetails"},
		"playlistId": {playlistID},
		"maxResults": {"50"},
		"key":        {c.apiKey},
	}
	if pageToken != "" {
		params.Set("pageToken", pageToken)
	}

	resp, err := c.httpClient.Get(baseURL + "/playlistItems?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PlaylistResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) get(endpoint string, params url.Values) (*APIResponse, error) {
	resp, err := c.httpClient.Get(baseURL + "/" + endpoint + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}
