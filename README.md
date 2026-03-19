# book-keeper

A lightweight sidecar that watches a source folder and copies new files into an ingestion folder. Built for use alongside [Calibre-Web](https://github.com/janeczku/calibre-web), [BookLore](https://github.com/booklore/booklore), and similar tools that consume files from an ingestion directory and delete them after processing.

**book-keeper preserves your originals.** Your source library stays untouched — only copies are sent to the ingestion folder.

## How it works

```
┌─────────────┐              ┌──────────────┐              ┌──────────────┐
│  Source Dir  │── watches ──▶│  book-keeper  │── copies ──▶│ Ingestion Dir│
│  (WATCH_DIR) │              │              │              │(INGESTION_DIR)│
└─────────────┘              └──────┬───────┘              └──────────────┘
                                    │
                                    ▼
                             records.json
```

- Monitors the source directory recursively for new files
- Waits for files to finish writing before copying (write stabilization)
- Deduplicates by content hash (SHA256) — the same file won't be copied twice even under a different name
- Tracks all copied files in a JSON record
- On startup, scans for any files added while it was offline

## Quick start

```yaml
services:
  book-keeper:
    image: kvqn/book-keeper:latest
    environment:
      - WATCH_DIR=/watch
      - INGESTION_DIR=/ingestion
    volumes:
      - /path/to/your/library:/watch:ro
      - /path/to/ingestion:/ingestion
      - /path/to/data:/data
    restart: unless-stopped
```

Mount your source library as **read-only** (`:ro`) — book-keeper never modifies it.

## Configuration

All configuration is done through environment variables.

| Variable | Required | Default | Description |
|---|---|---|---|
| `WATCH_DIR` | Yes | — | Source directory to monitor |
| `INGESTION_DIR` | Yes | — | Destination directory for copies |
| `RECORDS_FILE` | No | `/data/records.json` | Path to the state file |
| `STABILIZATION_SECS` | No | `5` | Seconds to wait for a file to stop changing before copying |
| `SCAN_ON_STARTUP` | No | `true` | Scan for existing files on startup |
| `POLL_INTERVAL_SECS` | No | `0` (disabled) | If set, use polling instead of filesystem events (useful for network drives) |

## Records

book-keeper tracks every file it copies in a JSON file:

```json
{
  "records": [
    {
      "source_path": "Science/physics-intro.epub",
      "hash": "sha256:abc123...",
      "copied_at": "2025-01-15T10:30:00Z",
      "size_bytes": 1048576
    }
  ]
}
```

If book-keeper sees a file with the same content hash (even under a different filename), it skips the copy.

## Building from source

```bash
go build -o book-keeper .
```

```bash
docker build -t kvqn/book-keeper .
```

## License

[MIT](LICENSE)
