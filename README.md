# Torrent Search Web

A web-based torrent search engine that searches across multiple torrent providers simultaneously.

## Features

### Search

- Query all providers simultaneously, with per-provider enable/disable toggles
- Search by category: Movies, TV, Music, Games, Software, Anime, Books, XXX
- Results shown progressively as providers respond
- Sort by name, seeders, peers, file size, or upload date
- Filter out dead or already-viewed torrents
- Filter results by name, provider or category

### Torrent Details

- Torrent name, file size, seeders/peers, upload date, category, NSFW indicator, provider
- Native details view — view torrent details without leaving the app
- Magnet link and .torrent file download
- Copy or share magnet link and details page URL

### Providers

20+ torrent providers supported:

- **Movies/TV/General:** 1337x, The Pirate Bay, LimeTorrents, TorrentDownloads, YTS, BTDigg, EZTV, BitSearch, KickassTorrents, TorrentKitty, MegaSearch, RARBG (mirrors), Torrent9, Rutor
- **Anime:** Nyaa, AniRena, AnimeTosho, Bangumi.moe, Mikan, AniLibria, Sukebei, Nyaa.si
- **DHT/Indexes:** BitTorrentDB, BTDigg

### Web Interface

- Clean, responsive design with Tailwind CSS
- Light/dark mode with system theme detection
- Progressive search results loading
- Provider management UI (enable/disable per provider)

## Architecture

```bash
.
├── cmd/server/          # Go HTTP server entry point
├── internal/
│   ├── handlers/        # HTTP handlers (API routes)
│   ├── models/          # Data models (Torrent, TorrentDetails, Category, Provider)
│   ├── providers/       # Torrent provider implementations
│   └── utils/           # Utilities
├── configs/             # Configuration (config.yaml)
├── static/              # Built React frontend (served by Go)
└── web/                 # React frontend source
    ├── src/
    │   ├── components/  # Reusable UI components
    │   ├── context/     # React Context providers
    │   ├── hooks/       # Custom React hooks
    │   ├── lib/         # API client
    │   └── pages/       # Page components
    └── package.json
```

## Tech Stack

### Backend (Go)

- **Language:** Go 1.21+
- **HTTP:** Standard library `net/http` with chi router
- **HTML Parsing:** goquery (jQuery-like syntax)
- **Configuration:** YAML (config.yaml)
- **Single binary deployment**

### Frontend (React)

- **Framework:** React 18 with Vite
- **Styling:** Tailwind CSS
- **Routing:** React Router v6
- **State:** React Context + Hooks
- **Build:** Vite (outputs to `../static`)

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+ and npm
- Git

### Development

**1. Build the frontend:**

```bash
cd web
npm install
npm run build
# Output goes to ../static
```

**2. Build and run the backend:**

```bash
go build -o torrentsearch ./cmd/server
./torrentsearch
```

**3. Open `http://localhost:8080`**

### Production Build

```bash
cd web
npm run build

cd ..
CGO_ENABLED=0 go build -ldflags="-s -w" -o torrentsearch ./cmd/server

# Run the binary
./torrentsearch
```

### Configuration

Edit `configs/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  static_dir: "./static"

providers:
  timeout_seconds: 10
  user_agent: "TorrentSearch/1.0"
  enabled:
    - "1337x"
    - "nyaa"
    - "yts"
    - "thepiratebay"
    - "eztv"
    - "limetorrents"
    - "btdigg"
    - "torrentdownloads"
    # ... more providers
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/search` | Search torrents (query: `q`, `providers`, `category`, `page`, `sort`) |
| GET | `/api/search/providers` | List all providers with status |
| GET | `/api/torrent/:infoHash` | Get torrent details by info hash |
| GET | `/api/torrent/details` | Get torrent details from provider URL |
| GET | `/api/torrent/magnet` | Generate magnet link from info hash |
| GET | `/api/config` | Get current config |
| GET | `/api/config/providers` | Get provider list |

### Search Example

```bash
curl "http://localhost:8080/api/search?q=ubuntu&providers=nyaa,yts,1337x&category=software&page=1&sort=seeders"
```

## Project Structure Details

### Providers (`internal/providers/`)

Each provider implements the `Provider` interface:

```go
type Provider interface {
    Search(ctx context.Context, query string, category Category, page int) ([]Torrent, error)
    GetDetails(ctx context.Context, infoHash string) (*TorrentDetails, error)
    GetDetailsFromURL(ctx context.Context, detailURL string) (*TorrentDetails, error)
    Name() string
    DisplayName() string
    BaseURL() string
    IconURL() string
    Categories() []Category
    IsEnabled() bool
    SetEnabled(bool)
}
```

Implemented providers:

- `1337x.go` - 1337x.to (full implementation)
- `nyaa.go` - nyaa.si (full implementation)
- `yts.go` - yts.mx (full implementation)
- `thepiratebay.go` - thepiratebay.org (full implementation)
- `eztv.go` - eztv.re (full implementation)
- `limetorrents.go` - limetorrents.lol (full implementation)
- `btdigg.go` - btdig.com (DHT search, full implementation)
- `torrentdownloads.go` - torrentdownloads.pro (full implementation)
- `stub_providers.go` - 15+ stub providers (return empty results, ready for implementation)

### Models (`internal/models/torrent.go`)

```go
type Torrent struct {
    Name        string
    InfoHash    string
    MagnetLink  string
    TorrentURL  string
    Size        string
    Seeders     int
    Peers       int
    UploadDate  string
    Category    Category
    Provider    string
    DetailURL   string
    IsNSFW      bool
}

type Category string // Movies, TV, Music, Games, Software, Anime, Books, XXX
```

## Deployment

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN cd web && npm install && npm run build
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o torrentsearch ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/torrentsearch .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["./torrentsearch"]
```

### Systemd Service

```ini
[Unit]
Description=Torrent Search Web
After=network.target

[Service]
Type=simple
User=torrentsearch
WorkingDirectory=/opt/torrentsearch
ExecStart=/opt/torrentsearch/torrentsearch
Restart=on-failure
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add new providers in `internal/providers/`
4. Update provider registry in `internal/providers/registry.go`
5. Submit a pull request

### Adding a New Provider

1. Create `internal/providers/newprovider.go` implementing the `Provider` interface
2. Register it in `internal/providers/registry.go`
3. Add to `configs/config.yaml` providers list
4. Test with `go run ./cmd/server`

## License

MIT License - see [LICENSE](LICENSE) for details.

## Disclaimer

TorrentSearch Web **does not host, store, or distribute any torrent files or copyrighted content**. It searches publicly accessible third-party sources and displays the results. The developers are not responsible for how those results are accessed or used.

Users are responsible for complying with their local laws and regulations.
