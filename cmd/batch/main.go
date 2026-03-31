package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"

	"foodie-go/youtube"
	"foodie-go/youtube/storage"
)

// Config represents the keywords configuration file
type Config struct {
	Keywords []string `json:"keywords"`
}

// Default keywords file path
const defaultKeywordsFile = "keywords.json"

func main() {
	godotenv.Load()

	// CLI flags
	count := flag.Int("count", 5, "Number of random keywords to search")
	maxResults := flag.Int("max", 10, "Max results per keyword")
	keywordsFile := flag.String("keywords", defaultKeywordsFile, "Path to keywords JSON file")
	storageType := flag.String("storage", "", "Storage adapter: csv, sqlite, postgres")
	output := flag.String("output", "", "Output file for CSV storage")
	dbPath := flag.String("db", "", "Database path for SQLite storage")
	apiKey := flag.String("api-key", "", "YouTube API key (or set YOUTUBE_API_KEY env)")
	verbose := flag.Bool("v", false, "Verbose output")
	delay := flag.Int("delay", 2, "Delay between searches in seconds")
	flag.Parse()

	// Get API key (flag > env)
	key := *apiKey
	if key == "" {
		key = os.Getenv("YOUTUBE_API_KEY")
	}
	if key == "" {
		log.Fatal("YOUTUBE_API_KEY is required (use -api-key flag or set env variable)")
	}

	// Load keywords from JSON file
	keywords, err := loadKeywords(*keywordsFile)
	if err != nil {
		log.Fatalf("Failed to load keywords: %v", err)
	}
	fmt.Printf("Loaded %d keywords from %s\n", len(keywords), *keywordsFile)

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

	// Pick storage
	store, err := pickStorage(storeType, outputPath, databasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	// Get random keywords
	selectedKeywords := getRandomKeywords(keywords, *count)
	fmt.Printf("Selected %d keywords: %v\n", len(selectedKeywords), selectedKeywords)

	// Collect all videos
	allVideos := make([]youtube.Video, 0)

	for i, keyword := range selectedKeywords {
		fmt.Printf("\n[%d/%d] Searching: %q\n", i+1, len(selectedKeywords), keyword)

		searchResp, err := client.SearchVideos(keyword, *maxResults)
		if err != nil {
			log.Printf("Error searching %q: %v", keyword, err)
			continue
		}

		videoIDs := youtube.ParseSearchResults(searchResp)
		videoIDs = youtube.DeduplicateIDs(videoIDs)

		if len(videoIDs) == 0 {
			log.Printf("No videos found for %q", keyword)
			continue
		}

		if *verbose {
			log.Printf("  Found %d video IDs", len(videoIDs))
		}

		// Get video details
		apiResp, err := client.GetVideos(videoIDs)
		if err != nil {
			log.Printf("Error getting details for %q: %v", keyword, err)
			continue
		}

		videos := youtube.ParseVideos(apiResp)
		allVideos = append(allVideos, videos...)

		if *verbose {
			for _, v := range videos {
				fmt.Printf("  - [%s] %s\n", v.VideoID, v.Title)
			}
		}

		// Delay between requests to avoid rate limiting
		if i < len(selectedKeywords)-1 {
			time.Sleep(time.Duration(*delay) * time.Second)
		}
	}

	// Deduplicate all videos
	allVideos = youtube.DeduplicateVideos(allVideos)

	// Save all videos
	if len(allVideos) > 0 {
		if err := store.Save(allVideos); err != nil {
			log.Fatalf("Save failed: %v", err)
		}
		fmt.Printf("\nSaved %d unique videos to %s\n", len(allVideos), storeType)
	} else {
		fmt.Println("\nNo videos to save")
	}
}

// loadKeywords reads keywords from a JSON file
func loadKeywords(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open keywords file: %w", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("parse keywords file: %w", err)
	}

	if len(config.Keywords) == 0 {
		return nil, fmt.Errorf("no keywords found in %s", path)
	}

	return config.Keywords, nil
}

// getRandomKeywords returns a random subset of keywords
func getRandomKeywords(keywords []string, count int) []string {
	if count >= len(keywords) {
		return keywords
	}

	shuffled := make([]string, len(keywords))
	copy(shuffled, keywords)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}

// pickStorage creates a storage adapter based on type
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
