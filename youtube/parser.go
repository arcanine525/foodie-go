package youtube

import (
	"regexp"
	"strconv"
	"time"
)

// ParseVideos converts API response items into domain Video models.
func ParseVideos(resp *APIResponse) []Video {
	videos := make([]Video, 0, len(resp.Items))
	for _, item := range resp.Items {
		v := Video{
			VideoID:      item.ID,
			Title:        item.Snippet.Title,
			ChannelName:  item.Snippet.ChannelTitle,
			ChannelID:    item.Snippet.ChannelID,
			ThumbnailURL: pickThumbnail(item.Snippet.Thumbnails),
			DurationSec:  parseISODuration(item.ContentDetails.Duration),
			ViewCount:    parseInt64(item.Statistics.ViewCount),
			LikeCount:    parseInt64(item.Statistics.LikeCount),
			PublishedAt:  parseTime(item.Snippet.PublishedAt),
			IsEmbeddable: item.Status.Embeddable,
		}
		videos = append(videos, v)
	}
	return videos
}

// ParseSearchResults extracts video IDs from search response.
func ParseSearchResults(resp *SearchResponse) []string {
	ids := make([]string, 0, len(resp.Items))
	for _, item := range resp.Items {
		ids = append(ids, item.ID.VideoID)
	}
	return ids
}

// DeduplicateVideos removes duplicate videos by VideoID, keeping the first occurrence.
func DeduplicateVideos(videos []Video) []Video {
	seen := make(map[string]bool)
	result := make([]Video, 0, len(videos))
	for _, v := range videos {
		if !seen[v.VideoID] {
			seen[v.VideoID] = true
			result = append(result, v)
		}
	}
	return result
}

// DeduplicateIDs removes duplicate video IDs, keeping the first occurrence.
func DeduplicateIDs(ids []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}

func pickThumbnail(t APIThumbnails) string {
	if t.MaxRes.URL != "" {
		return t.MaxRes.URL
	}
	if t.High.URL != "" {
		return t.High.URL
	}
	return t.Medium.URL
}

var durationRe = regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)

// parseISODuration converts PT11M34S -> 694 seconds.
func parseISODuration(iso string) int {
	matches := durationRe.FindStringSubmatch(iso)
	if matches == nil {
		return 0
	}
	h, _ := strconv.Atoi(matches[1])
	m, _ := strconv.Atoi(matches[2])
	s, _ := strconv.Atoi(matches[3])
	return h*3600 + m*60 + s
}

func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
