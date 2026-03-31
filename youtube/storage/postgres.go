package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"foodie-go/youtube"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS videos (
			video_id      TEXT PRIMARY KEY,
			title         TEXT,
			channel_name  TEXT,
			channel_id    TEXT,
			thumbnail_url TEXT,
			duration_sec  INTEGER,
			view_count    BIGINT,
			like_count    BIGINT,
			published_at  TIMESTAMPTZ,
			is_embeddable BOOLEAN DEFAULT TRUE
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

func (p *PostgresStorage) Save(videos []youtube.Video) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO videos
		(video_id, title, channel_name, channel_id, thumbnail_url,
		 duration_sec, view_count, like_count, published_at, is_embeddable)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (video_id) DO UPDATE SET
			title = EXCLUDED.title,
			view_count = EXCLUDED.view_count,
			like_count = EXCLUDED.like_count,
			thumbnail_url = EXCLUDED.thumbnail_url
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
			v.PublishedAt, v.IsEmbeddable,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (p *PostgresStorage) Close() error { return p.db.Close() }
