# FoodieAI — YouTube Video Fetcher

A Go CLI tool to search YouTube videos and export metadata to CSV, SQLite, or PostgreSQL.

## Installation

```bash
# Build all commands
go build -o bin/youtube ./cmd/youtube
go build -o bin/batch ./cmd/batch

# Or build both at once
go build -o bin/youtube ./cmd/youtube && go build -o bin/batch ./cmd/batch
```

## Configuration

Create a `.env` file with your YouTube API key:

```bash
cp .env.example .env
```

```env
YOUTUBE_API_KEY=your_api_key_here
```

Get your API key from [Google Cloud Console](https://console.cloud.google.com/apis/credentials).

## Usage

### Single Search

Search for videos with a specific query:

```bash
# Basic search (outputs to videos.csv)
./bin/youtube -query "vietnamese cooking"

# With verbose output
./bin/youtube -query "cach nau pho" -max 20 -v

# Save to SQLite
./bin/youtube -query "thit kho trung" -storage sqlite -db recipes.db

# Save to PostgreSQL
DATABASE_URL="postgres://user:pass@localhost/foodie" ./bin/youtube -query "banh mi" -storage postgres
```

### Batch Search

Search multiple random food keywords automatically:

```bash
# Search 5 random keywords (default)
./bin/batch

# Search 10 keywords, 20 results each, save to SQLite
./bin/batch -count 10 -max 20 -storage sqlite -db recipes.db

# Verbose output with custom delay between searches
./bin/batch -count 3 -max 15 -delay 3 -v
```

## CLI Flags

### Single Search (`youtube`)

| Flag | Default | Description |
|------|---------|-------------|
| `-query` | (required) | Search query |
| `-max` | 50 | Maximum results |
| `-storage` | csv | Storage type: csv, sqlite, postgres |
| `-output` | videos.csv | CSV output path |
| `-db` | videos.db | SQLite database path |
| `-api-key` | (env) | YouTube API key |
| `-v` | false | Verbose output |

### Batch Search (`batch`)

| Flag | Default | Description |
|------|---------|-------------|
| `-count` | 5 | Number of random keywords |
| `-max` | 10 | Max results per keyword |
| `-storage` | csv | Storage type: csv, sqlite, postgres |
| `-output` | videos.csv | CSV output path |
| `-db` | videos.db | SQLite database path |
| `-delay` | 2 | Delay between searches (seconds) |
| `-api-key` | (env) | YouTube API key |
| `-v` | false | Verbose output |

## API Quota

YouTube Data API has a daily quota of **10,000 units** (free tier).

| Operation | Cost |
|-----------|------|
| `search.list` | 100 units |
| `videos.list` | 1 unit (batch up to 50) |
| `playlistItems.list` | 1 unit |

**Example costs:**
- Single search with 50 results: 100 + 1 = 101 units
- Batch search with 5 keywords × 10 results: 5 × (100 + 1) = 505 units

Quota resets at **8:00 AM Vietnam time** (0:00 PST).

## Output Schema

### CSV Columns
```
video_id, title, channel_name, channel_id, thumbnail_url,
duration_sec, view_count, like_count, published_at, is_embeddable
```

### SQLite / PostgreSQL Table
```sql
CREATE TABLE videos (
    video_id      TEXT PRIMARY KEY,
    title         TEXT,
    channel_name  TEXT,
    channel_id    TEXT,
    thumbnail_url TEXT,
    duration_sec  INTEGER,
    view_count    BIGINT,
    like_count    BIGINT,
    published_at  TIMESTAMPTZ,
    is_embeddable BOOLEAN
);
```

## Project Structure

```
cmd/
├── youtube/main.go       # Single search CLI
└── batch/main.go         # Batch search CLI
youtube/
├── client.go             # YouTube API client
├── models.go             # Domain models + API structs
├── parser.go             # Parse API responses + deduplication
└── storage/
    ├── storage.go        # Storage interface
    ├── csv.go            # CSV adapter
    ├── sqlite.go         # SQLite adapter
    └── postgres.go       # PostgreSQL adapter
```

## Features

- **Multiple storage backends**: CSV, SQLite, PostgreSQL
- **Deduplication**: Automatic removal of duplicate videos by ID
- **Batch processing**: Random keyword selection for bulk collection
- **Rate limiting**: Configurable delay between API calls
- **Error handling**: Clear messages for quota exceeded, rate limits, etc.

## License

MIT
