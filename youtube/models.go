package youtube

import "time"

// Video is the domain model after parsing API response.
type Video struct {
	VideoID      string
	Title        string
	ChannelName  string
	ChannelID    string
	ThumbnailURL string
	DurationSec  int
	ViewCount    int64
	LikeCount    int64
	PublishedAt  time.Time
	IsEmbeddable bool
}

// APIResponse maps the YouTube Data API v3 JSON response.
type APIResponse struct {
	Items         []APIItem `json:"items"`
	NextPageToken string    `json:"nextPageToken"`
}

type APIItem struct {
	ID             string        `json:"id"`
	Snippet        APISnippet    `json:"snippet"`
	ContentDetails APIContent    `json:"contentDetails"`
	Statistics     APIStatistics `json:"statistics"`
	Status         APIStatus     `json:"status"`
}

type APISnippet struct {
	Title        string        `json:"title"`
	ChannelTitle string        `json:"channelTitle"`
	ChannelID    string        `json:"channelId"`
	PublishedAt  string        `json:"publishedAt"`
	Thumbnails   APIThumbnails `json:"thumbnails"`
}

type APIThumbnails struct {
	Medium APIThumb `json:"medium"`
	High   APIThumb `json:"high"`
	MaxRes APIThumb `json:"maxres"`
}

type APIThumb struct {
	URL string `json:"url"`
}

type APIContent struct {
	Duration string `json:"duration"` // ISO 8601: PT11M34S
}

type APIStatistics struct {
	ViewCount string `json:"viewCount"`
	LikeCount string `json:"likeCount"`
}

type APIStatus struct {
	Embeddable bool `json:"embeddable"`
}

// SearchItem has a different ID structure from videos.list.
type SearchResponse struct {
	Items         []SearchItem `json:"items"`
	NextPageToken string       `json:"nextPageToken"`
}

type SearchItem struct {
	ID      SearchID   `json:"id"`
	Snippet APISnippet `json:"snippet"`
}

type SearchID struct {
	VideoID string `json:"videoId"`
}

// PlaylistResponse for playlistItems.list.
type PlaylistResponse struct {
	Items []struct {
		ContentDetails struct {
			VideoID string `json:"videoId"`
		} `json:"contentDetails"`
	} `json:"items"`
	NextPageToken string `json:"nextPageToken"`
}
