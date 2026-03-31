package storage

import "foodie-go/youtube"

// Storage is the adapter interface.
// Implement this for CSV, SQLite, PostgreSQL, or any other backend.
type Storage interface {
	Save(videos []youtube.Video) error
	Close() error
}
