package models

import (
	"time"
)

// Category represents a torrent category
type Category string

const (
	CategoryAll          Category = "all"
	CategoryMovies       Category = "movies"
	CategoryTV           Category = "tv"
	CategoryMusic        Category = "music"
	CategoryGames        Category = "games"
	CategorySoftware     Category = "software"
	CategoryAnime        Category = "anime"
	CategoryBooks        Category = "books"
	CategoryXXX          Category = "xxx"
	CategoryOther        Category = "other"
)

// IsNSFW returns true if the category is NSFW
func (c Category) IsNSFW() bool {
	return c == CategoryXXX
}

// DisplayName returns the display name of the category
func (c Category) DisplayName() string {
	switch c {
	case CategoryAll:
		return "All"
	case CategoryMovies:
		return "Movies"
	case CategoryTV:
		return "TV Shows"
	case CategoryMusic:
		return "Music"
	case CategoryGames:
		return "Games"
	case CategorySoftware:
		return "Software"
	case CategoryAnime:
		return "Anime"
	case CategoryBooks:
		return "Books"
	case CategoryXXX:
		return "XXX"
	case CategoryOther:
		return "Other"
	default:
		return string(c)
	}
}

// TorrentFile represents a file within a torrent
type TorrentFile struct {
	Name       string `json:"name"`
	Size       string `json:"size"`
	SizeBytes  int64  `json:"size_bytes"`
	Path       string `json:"path,omitempty"`
	Priority   int    `json:"priority,omitempty"`
}

// Torrent represents a torrent search result
type Torrent struct {
	InfoHash           string     `json:"info_hash"`
	Name               string     `json:"name"`
	Size               string     `json:"size,omitempty"`
	SizeBytes          int64      `json:"size_bytes,omitempty"`
	Seeders            int        `json:"seeders,omitempty"`
	Peers              int        `json:"peers,omitempty"`
	Provider           string     `json:"provider"`
	ProviderName       string     `json:"provider_name"`
	UploadDate         string     `json:"upload_date,omitempty"`
	Category           Category   `json:"category,omitempty"`
	DescriptionPageURL string     `json:"description_page_url,omitempty"`
	MagnetURI          string     `json:"magnet_uri,omitempty"`
	FileDownloadLink   string     `json:"file_download_link,omitempty"`

	// Computed fields
	IsNSFW bool `json:"is_nsfw"`
	IsDead bool `json:"is_dead"`
}

// TorrentDetails represents detailed torrent information
type TorrentDetails struct {
	InfoHash           string       `json:"info_hash"`
	Name               string       `json:"name"`
	Size               string       `json:"size,omitempty"`
	SizeBytes          int64        `json:"size_bytes,omitempty"`
	Seeders            int          `json:"seeders,omitempty"`
	Peers              int          `json:"peers,omitempty"`
	UploadDate         string       `json:"upload_date,omitempty"`
	Category           Category     `json:"category,omitempty"`
	Uploader           string       `json:"uploader,omitempty"`
	LastChecked        string       `json:"last_checked,omitempty"`
	MagnetURI          string       `json:"magnet_uri"`
	FileDownloadLink   string       `json:"file_download_link,omitempty"`
	Description        string       `json:"description,omitempty"`
	PosterURL          string       `json:"poster_url,omitempty"`
	ScreenshotURLs     []string     `json:"screenshot_urls,omitempty"`
	Files              []TorrentFile `json:"files,omitempty"`
	Provider           string       `json:"provider,omitempty"`
	DescriptionPageURL string       `json:"description_page_url,omitempty"`

	IsNSFW bool `json:"is_nsfw"`
}

// SearchResults represents search results from providers
type SearchResults struct {
	Torrents []Torrent `json:"torrents"`
	Errors   []string  `json:"errors,omitempty"`
}

// SearchParams represents search parameters
type SearchParams struct {
	Query    string   `json:"query" form:"query"`
	Category Category `json:"category" form:"category"`
	Page     int      `json:"page" form:"page"`
	Limit    int      `json:"limit" form:"limit"`
}

// SortCriteria represents sort criteria
type SortCriteria string

const (
	SortCriteriaRelevance SortCriteria = "relevance"
	SortCriteriaSeeders   SortCriteria = "seeders"
	SortCriteriaPeers     SortCriteria = "peers"
	SortCriteriaSize      SortCriteria = "size"
	SortCriteriaDate      SortCriteria = "date"
)

// SortOrder represents sort order
type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

// SortOptions represents sort options
type SortOptions struct {
	Criteria SortCriteria `json:"criteria" form:"criteria"`
	Order    SortOrder    `json:"order" form:"order"`
}

// TorrentFilter represents filter options for torrents
type TorrentFilter struct {
	Providers       []ProviderFilterOption `json:"providers"`
	ShowDeadTorrents bool                  `json:"show_dead_torrents"`
	Category        Category              `json:"category"`
	HideViewed      bool                  `json:"hide_viewed"`
	Query           string                `json:"query"`
}

// ProviderFilterOption represents a provider filter option
type ProviderFilterOption struct {
	Provider string `json:"provider"`
	Selected bool   `json:"selected"`
}

// SearchState represents the current search state
type SearchState string

const (
	SearchStateLoading       SearchState = "loading"
	SearchStateInternetError SearchState = "internet_error"
	SearchStateNotFound      SearchState = "not_found"
	SearchStateComplete      SearchState = "complete"
	SearchStateSearching     SearchState = "searching"
	SearchStateRefreshing    SearchState = "refreshing"
)

// SearchUIState represents the UI state for search
type SearchUIState struct {
	SearchParams     SearchParams    `json:"search_params"`
	SearchState      SearchState     `json:"search_state"`
	SearchResults    SearchResults   `json:"search_results"`
	SortOptions      SortOptions     `json:"sort_options"`
	TorrentFilter    TorrentFilter   `json:"torrent_filter"`
	ViewedTorrentHashes []string     `json:"viewed_torrent_hashes"`
}

// SearchProvider represents a search provider interface
type SearchProvider interface {
	Name() string
	Search(query string, category Category, page int) ([]Torrent, error)
	GetDetails(infoHash string) (*TorrentDetails, error)
	SupportsCategory(category Category) bool
	GetSupportedCategories() []Category
	BaseURL() string
	IsEnabled() bool
	SetEnabled(bool)
}

// ProviderStatus represents the status of a provider
type ProviderStatus struct {
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Status      string `json:"status"`
	LastChecked string `json:"last_checked"`
	Error       string `json:"error,omitempty"`
}

// TorrentFileDownloadState represents torrent file download state
type TorrentFileDownloadState string

const (
	TorrentFileDownloadStateIdle       TorrentFileDownloadState = "idle"
	TorrentFileDownloadStateDownloading TorrentFileDownloadState = "downloading"
	TorrentFileDownloadStateCompleted  TorrentFileDownloadState = "completed"
	TorrentFileDownloadStateFailed     TorrentFileDownloadState = "failed"
)

// TorrentFileDownloadEvent represents torrent file download events
type TorrentFileDownloadEvent string

const (
	TorrentFileDownloadEventStarted  TorrentFileDownloadEvent = "started"
	TorrentFileDownloadEventProgress TorrentFileDownloadEvent = "progress"
	TorrentFileDownloadEventCompleted TorrentFileDownloadEvent = "completed"
	TorrentFileDownloadEventFailed    TorrentFileDownloadEvent = "failed"
)

// SearchHistory represents a search history entry
type SearchHistory struct {
	ID        int64     `json:"id"`
	Query     string    `json:"query"`
	Category  Category  `json:"category"`
	Timestamp time.Time `json:"timestamp"`
}

// BookmarkedTorrent represents a bookmarked torrent
type BookmarkedTorrent struct {
	ID           int64     `json:"id"`
	InfoHash     string    `json:"info_hash"`
	Name         string    `json:"name"`
	Size         string    `json:"size"`
	Seeders      int       `json:"seeders"`
	Peers        int       `json:"peers"`
	ProviderName string    `json:"provider_name"`
	UploadDate   *time.Time `json:"upload_date,omitempty"`
	Category     Category  `json:"category"`
	MagnetURI    string    `json:"magnet_uri"`
	Timestamp    time.Time `json:"timestamp"`
}

// ViewedTorrent represents a viewed torrent
type ViewedTorrent struct {
	InfoHash  string    `json:"info_hash"`
	Timestamp time.Time `json:"timestamp"`
}

// TorznabConfig represents a Torznab indexer configuration
type TorznabConfig struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	APIKey      string    `json:"api_key"`
	Categories  []string  `json:"categories"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Settings represents application settings
type Settings struct {
	ID                     int64      `json:"id"`
	SaveSearchHistory      bool       `json:"save_search_history"`
	EnableNSFWMode         bool       `json:"enable_nsfw_mode"`
	DefaultSortOptions     SortOptions `json:"default_sort_options"`
	DefaultCategory        Category   `json:"default_category"`
	MaxResultsPerProvider  int        `json:"max_results_per_provider"`
	SearchTimeout          int        `json:"search_timeout"`
	EnableTorznabProviders bool       `json:"enable_torznab_providers"`
	DoHProvider            string     `json:"doh_provider"`
	Theme                  string     `json:"theme"`
	Language               string     `json:"language"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// DefaultSettings returns default settings
func DefaultSettings() Settings {
	return Settings{
		SaveSearchHistory:      true,
		EnableNSFWMode:         false,
		DefaultSortOptions:     SortOptions{Criteria: SortCriteriaSeeders, Order: SortOrderDesc},
		DefaultCategory:        CategoryAll,
		MaxResultsPerProvider:  50,
		SearchTimeout:          30,
		EnableTorznabProviders: true,
		DoHProvider:            "cloudflare",
		Theme:                  "system",
		Language:               "en",
	}
}