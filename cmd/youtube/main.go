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
	// Load .env file (ignore error if file doesn't exist)
	godotenv.Load()

	// CLI flags (empty default = use env var or fallback)
	query := flag.String("query", "", "Search query for YouTube videos (required)")
	maxResults := flag.Int("max", 50, "Maximum number of results (default: 50)")
	storageType := flag.String("storage", "", "Storage adapter: csv, sqlite, postgres (default: from env or csv)")
	output := flag.String("output", "", "Output file path for CSV storage (default: from env or videos.csv)")
	dbPath := flag.String("db", "", "Database path for SQLite storage (default: from env or videos.db)")
	apiKey := flag.String("api-key", "", "YouTube API key (or set YOUTUBE_API_KEY env)")
	verbose := flag.Bool("v", false, "Verbose output")
	flag.Parse()

	// Get API key from flag or env
	key := *apiKey
	if key == "" {
		key = os.Getenv("YOUTUBE_API_KEY")
	}
	if key == "" {
		log.Fatal("YOUTUBE_API_KEY is required (use -api-key flag or set env variable)")
	}

	// Validate query
	if *query == "" {
		log.Fatal("Query is required (use -query flag)")
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

	// 1. Create client
	client := youtube.NewClient(key)

	if *verbose {
		log.Printf("Searching for: %q (max: %d)", *query, *maxResults)
	}

	// 2. Search for videos
	searchResp, err := client.SearchVideos(*query, *maxResults)
	if err != nil {
		handleAPIError(err)
	}
	videoIDs := youtube.ParseSearchResults(searchResp)
	videoIDs = youtube.DeduplicateIDs(videoIDs)
	log.Printf("Extracted %d unique video IDs", len(videoIDs))

	if len(videoIDs) == 0 {
		log.Println("No videos found")
		return
	}

	if *verbose {
		log.Printf("Found %d video IDs", len(videoIDs))
	}

	// 3. Get video details
	apiResp, err := client.GetVideos(videoIDs)
	if err != nil {
		handleAPIError(err)
	}

	// 4. Parse response
	videos := youtube.ParseVideos(apiResp)
	videos = youtube.DeduplicateVideos(videos)

	if *verbose {
		for _, v := range videos {
			fmt.Printf("  [%s] %s (%d views)\n", v.VideoID, v.Title, v.ViewCount)
		}
	}

	// 5. Save to storage
	store, err := pickStorage(storeType, outputPath, databasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	if err := store.Save(videos); err != nil {
		log.Fatalf("Save failed: %v", err)
	}

	fmt.Printf("Saved %d videos to %s\n", len(videos), storeType)
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
	case "csv":
		return storage.NewCSVStorage(outputPath), nil
	default:
		return storage.NewCSVStorage(outputPath), nil
	}
}

func handleAPIError(err error) {
	// Check for common API errors
	errStr := err.Error()
	switch {
	case contains(errStr, "403"):
		log.Fatal("API Error (403): Quota exceeded or invalid API key. Daily quota resets at 8:00 AM Vietnam time.")
	case contains(errStr, "400"):
		log.Fatalf("API Error (400): Bad request - %v", err)
	case contains(errStr, "404"):
		log.Fatal("API Error (404): Video not found")
	case contains(errStr, "429"):
		log.Fatal("API Error (429): Rate limit exceeded. Please wait before retrying.")
	default:
		log.Fatalf("API Error: %v", err)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
