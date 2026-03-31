package storage

import (
	"encoding/csv"
	"fmt"
	"os"

	"foodie-go/youtube"
)

type CSVStorage struct {
	filePath string
}

func NewCSVStorage(filePath string) *CSVStorage {
	return &CSVStorage{filePath: filePath}
}

func (c *CSVStorage) Save(videos []youtube.Video) error {
	file, err := os.Create(c.filePath)
	if err != nil {
		return fmt.Errorf("create csv file: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	// Header
	w.Write([]string{
		"video_id", "title", "channel_name", "channel_id",
		"thumbnail_url", "duration_sec", "view_count",
		"like_count", "published_at", "is_embeddable",
	})

	for _, v := range videos {
		w.Write([]string{
			v.VideoID,
			v.Title,
			v.ChannelName,
			v.ChannelID,
			v.ThumbnailURL,
			fmt.Sprintf("%d", v.DurationSec),
			fmt.Sprintf("%d", v.ViewCount),
			fmt.Sprintf("%d", v.LikeCount),
			v.PublishedAt.Format("2006-01-02T15:04:05Z"),
			fmt.Sprintf("%t", v.IsEmbeddable),
		})
	}
	return nil
}

func (c *CSVStorage) Close() error { return nil }
