package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"foodie-go/youtube"
	"foodie-go/youtube/storage"
)

func main() {
	godotenv.Load()

	// CLI flags
	channelID := flag.String("channel", "", "YouTube channel ID (required)")
	maxVideos := flag.Int("max", 0, "Maximum videos to fetch (0 = all)")
	storageType := flag.String("storage", "", "Storage adapter: csv, sqlite, postgres")
	output := flag.String("output", "", "Output file for CSV storage")
	dbPath := flag.String("db", "", "Database path for SQLite storage")
	apiKey := flag.String("api-key", "", "YouTube API key (or set YOUTUBE_API_KEY env)")
	verbose := flag.Bool("v", false, "Verbose output")
	flag.Parse()

	// Get API key (flag > env)
	key := *apiKey
	if key == "" {
		key = os.Getenv("YOUTUBE_API_KEY")
	}
	if key == "" {
		log.Fatal("YOUTUBE_API_KEY is required (use -api-key flag or set env variable)")
	}

	// Validate channel ID
	if *channelID == "" {
		log.Fatal("Channel ID is required (use -channel flag)")
	}

	// Determine storage type (flag > env > default)
	storeType := *storageType
	if storeType == "" {
		storeType = os.Getenv("STORAGE")
	}
	if storeType == "" {
		storeType = "csv"
	}

	// Determine output path (flag > env > default)
	outputPath := *output
	if outputPath == "" {
		outputPath = os.Getenv("CSV_OUTPUT_PATH")
	}
	if outputPath == "" {
		outputPath = "videos.csv"
	}

	// Determine db path (flag > env > default)
	databasePath := *dbPath
	if databasePath == "" {
		databasePath = os.Getenv("SQLITE_DB_PATH")
	}
	if databasePath == "" {
		databasePath = "videos.db"
	}

	// Create client
	client := youtube.NewClient(key)

	if *verbose {
		log.Printf("Fetching channel: %s", *channelID)
	}

	// 1. Get channel info
	channelResp, err := client.GetChannel(*channelID)
	if err != nil {
		log.Fatalf("Failed to get channel: %v", err)
	}

	if len(channelResp.Items) == 0 {
		log.Fatal("Channel not found")
	}

	channel := channelResp.Items[0]
	uploadsPlaylistID := channel.ContentDetails.RelatedPlaylists.Uploads

	fmt.Printf("Channel: %s\n", channel.Snippet.Title)
	fmt.Printf("Uploads playlist: %s\n", uploadsPlaylistID)

	// 2. Get all video IDs from uploads playlist
	if *verbose {
		log.Printf("Fetching video IDs from playlist...")
	}

	videoIDs, err := client.GetAllPlaylistItems(uploadsPlaylistID, *maxVideos)
	if err != nil {
		log.Fatalf("Failed to get playlist items: %v", err)
	}

	fmt.Printf("Found %d videos\n", len(videoIDs))

	if len(videoIDs) == 0 {
		fmt.Println("No videos to save")
		return
	}

	// 3. Deduplicate video IDs
	videoIDs = youtube.DeduplicateIDs(videoIDs)

	// 4. Fetch video details in batches
	allVideos := make([]youtube.Video, 0)
	batchSize := 50

	for i := 0; i < len(videoIDs); i += batchSize {
		end := i + batchSize
		if end > len(videoIDs) {
			end = len(videoIDs)
		}

		batch := videoIDs[i:end]
		if *verbose {
			log.Printf("Fetching video details batch %d-%d of %d", i+1, end, len(videoIDs))
		}

		apiResp, err := client.GetVideos(batch)
		if err != nil {
			log.Printf("Error fetching batch %d-%d: %v", i+1, end, err)
			continue
		}

		videos := youtube.ParseVideos(apiResp)
		allVideos = append(allVideos, videos...)
	}

	// 5. Deduplicate videos
	allVideos = youtube.DeduplicateVideos(allVideos)

	if *verbose {
		for _, v := range allVideos {
			fmt.Printf("  - [%s] %s\n", v.VideoID, v.Title)
		}
	}

	// 6. Save to storage
	store, err := pickStorage(storeType, outputPath, databasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	if err := store.Save(allVideos); err != nil {
		log.Fatalf("Save failed: %v", err)
	}

	fmt.Printf("Saved %d videos to %s\n", len(allVideos), storeType)
}

func pickStorage(storageType, outputPath, dbPath string) (storage.Storage, error) {
	switch storageType {
	case "sqlite":
		return storage.NewSQLiteStorage(dbPath)
	case "postgres":
		connStr := os.Getenv("DATABASE_URL")
		if connStr == "" {
			return nil, fmt.Errorf("DATABASE_URL is required for postgres storage")
		}
		return storage.NewPostgresStorage(connStr)
	default:
		return storage.NewCSVStorage(outputPath), nil
	}
}
