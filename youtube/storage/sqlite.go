package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"foodie-go/youtube"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS videos (
			video_id      TEXT PRIMARY KEY,
			title         TEXT,
			channel_name  TEXT,
			channel_id    TEXT,
			thumbnail_url TEXT,
			duration_sec  INTEGER,
			view_count    INTEGER,
			like_count    INTEGER,
			published_at  TEXT,
			is_embeddable INTEGER
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}

	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) Save(videos []youtube.Video) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO videos
		(video_id, title, channel_name, channel_id, thumbnail_url,
		 duration_sec, view_count, like_count, published_at, is_embeddable)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, v := range videos {
		_, err := stmt.Exec(
			v.VideoID, v.Title, v.ChannelName, v.ChannelID,
			v.ThumbnailURL, v.DurationSec, v.ViewCount, v.LikeCount,
			v.PublishedAt.Format("2006-01-02T15:04:05Z"),
			v.IsEmbeddable,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLiteStorage) Close() error { return s.db.Close() }
