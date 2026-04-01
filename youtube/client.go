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

// GetChannel calls channels.list. Quota cost: 1 unit per call.
func (c *Client) GetChannel(channelID string) (*ChannelResponse, error) {
	params := url.Values{
		"part": {"snippet,contentDetails"},
		"id":   {channelID},
		"key":  {c.apiKey},
	}

	resp, err := c.httpClient.Get(baseURL + "/channels?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("channel request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("channel API returned status %d", resp.StatusCode)
	}

	var result ChannelResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode channel response: %w", err)
	}
	return &result, nil
}

// GetAllPlaylistItems fetches all video IDs from a playlist with pagination.
// Quota cost: 1 unit per page (50 items per page).
func (c *Client) GetAllPlaylistItems(playlistID string, maxVideos int) ([]string, error) {
	var allIDs []string
	pageToken := ""

	for {
		resp, err := c.GetPlaylistItems(playlistID, pageToken)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.Items {
			allIDs = append(allIDs, item.ContentDetails.VideoID)
			if maxVideos > 0 && len(allIDs) >= maxVideos {
				return allIDs, nil
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return allIDs, nil
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
