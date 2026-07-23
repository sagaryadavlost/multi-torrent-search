package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"torrentsearch-web/internal/models"
)

// YTSProvider implements YTS.mx torrent search (Movies only)
type YTSProvider struct {
	*BaseProvider
	apiBaseURL string
}

// YTS API response structures
type YTSMovie struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Year        int      `json:"year"`
	Rating      float64  `json:"rating"`
	Runtime     int      `json:"runtime"`
	Genres      []string `json:"genres"`
	Summary     string   `json:"summary"`
	Description string   `json:"description_full"`
	Language    string   `json:"language"`
	MpaRating   string   `json:"mpa_rating"`
	Background  string   `json:"background_image"`
	SmallCover  string   `json:"small_cover_image"`
	MediumCover string   `json:"medium_cover_image"`
	LargeCover  string   `json:"large_cover_image"`
	State       string   `json:"state"`
	Torrents    []YTSTorrent `json:"torrents"`
}

type YTSTorrent struct {
	URL         string `json:"url"`
	Hash        string `json:"hash"`
	Quality     string `json:"quality"`
	Type        string `json:"type"`
	Seeds       int    `json:"seeds"`
	Peers       int    `json:"peers"`
	Size        string `json:"size"`
	SizeBytes   int64  `json:"size_bytes"`
	DateUploaded string `json:"date_uploaded"`
	DateUploadedUnix int64 `json:"date_uploaded_unix"`
}

type YTSResponse struct {
	Status      string     `json:"status"`
	StatusMessage string   `json:"status_message"`
	Data        YTSData    `json:"data"`
}

type YTSData struct {
	MovieCount int        `json:"movie_count"`
	Limit      int        `json:"limit"`
	PageNumber int        `json:"page_number"`
	Movies     []YTSMovie `json:"movies"`
}

type YTSMovieDetailsResponse struct {
	Status      string     `json:"status"`
	StatusMessage string   `json:"status_message"`
	Data        YTSMovieData `json:"data"`
}

type YTSMovieData struct {
	Movie YTSMovie `json:"movie"`
}

// NewYTSProvider creates a new YTS provider
func NewYTSProvider(userAgent string, timeout int, enabled bool) *YTSProvider {
	base := NewBaseProvider(
		"yts",
		"https://yts.mx",
		"https://yts.mx/favicon.ico",
		enabled,
		userAgent,
		time.Duration(timeout)*time.Second,
		[]models.Category{models.CategoryMovies},
	)

	return &YTSProvider{
		BaseProvider: base,
		apiBaseURL:   "https://yts.mx/api/v2",
	}
}

// Search searches for movies on YTS
func (p *YTSProvider) Search(ctx context.Context, query string, category models.Category, page int) ([]models.Torrent, error) {
	if !p.Enabled {
		return nil, nil
	}

	if category != models.CategoryAll && category != models.CategoryMovies {
		return nil, nil
	}

	searchURL := fmt.Sprintf("%s/list_movies.json?query_term=%s&limit=50&page=%d",
		p.apiBaseURL, url.QueryEscape(query), page)

	return p.searchJSON(ctx, searchURL)
}

// searchJSON searches using JSON API
func (p *YTSProvider) searchJSON(ctx context.Context, searchURL string) ([]models.Torrent, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", p.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var response YTSResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return p.parseMovies(response.Data.Movies), nil
}

// parseMovies parses YTS movies into torrents
func (p *YTSProvider) parseMovies(movies []YTSMovie) []models.Torrent {
	var torrents []models.Torrent

	for _, movie := range movies {
		for _, torrent := range movie.Torrents {
			infoHash := strings.ToUpper(torrent.Hash)
			magnetURI := fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s", infoHash, url.QueryEscape(movie.Title))

			name := fmt.Sprintf("%s (%d) [%s] [%s]", movie.Title, movie.Year, torrent.Quality, torrent.Type)

			t := models.Torrent{
				InfoHash:           infoHash,
				Name:               name,
				Size:               torrent.Size,
				SizeBytes:          torrent.SizeBytes,
				Seeders:            torrent.Seeds,
				Peers:              torrent.Peers,
				Provider:           p.Name,
				ProviderName:       p.Name,
				MagnetURI:          magnetURI,
				DescriptionPageURL: fmt.Sprintf("%s/movie/%d", p.BaseURLStr, movie.ID),
				UploadDate:         torrent.DateUploaded,
				Category:           models.CategoryMovies,
				IsNSFW:             false,
			}
			torrents = append(torrents, t)
		}
	}

	return torrents
}

// GetDetails gets detailed torrent information
func (p *YTSProvider) GetDetails(ctx context.Context, infoHash string) (*models.TorrentDetails, error) {
	return nil, fmt.Errorf("GetDetails requires movie ID for YTS")
}

// GetDetailsFromURL gets details from a detail page URL
func (p *YTSProvider) GetDetailsFromURL(ctx context.Context, detailURL string) (*models.TorrentDetails, error) {
	if !p.Enabled {
		return nil, nil
	}

	// Extract movie ID from URL
	// URL format: https://yts.mx/movie/{id}
	movieID := ""
	if strings.Contains(detailURL, "/movie/") {
		parts := strings.Split(detailURL, "/movie/")
		if len(parts) > 1 {
			movieID = strings.Split(parts[1], "/")[0]
			movieID = strings.Split(movieID, "?")[0]
		}
	}

	if movieID == "" {
		return nil, fmt.Errorf("could not extract movie ID from URL")
	}

	apiURL := fmt.Sprintf("%s/movie_details.json?movie_id=%s&with_images=true", p.apiBaseURL, movieID)
	return p.getMovieDetailsFromAPI(ctx, apiURL, detailURL)
}

// getMovieDetailsFromAPI gets movie details from YTS API
func (p *YTSProvider) getMovieDetailsFromAPI(ctx context.Context, apiURL, pageURL string) (*models.TorrentDetails, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", p.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var response YTSMovieDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return p.parseMovieDetails(response.Data.Movie, pageURL), nil
}

// parseMovieDetails parses movie details
func (p *YTSProvider) parseMovieDetails(movie YTSMovie, pageURL string) *models.TorrentDetails {
	var files []models.TorrentFile
	var magnetURI string

	for _, torrent := range movie.Torrents {
		infoHash := strings.ToUpper(torrent.Hash)
		if magnetURI == "" {
			magnetURI = fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s", infoHash, url.QueryEscape(movie.Title))
		}
		files = append(files, models.TorrentFile{
			Name:       fmt.Sprintf("%s (%d) [%s] [%s].%s", movie.Title, movie.Year, torrent.Quality, torrent.Type, strings.ToLower(torrent.Type)),
			Size:       torrent.Size,
			SizeBytes:  torrent.SizeBytes,
			Path:       torrent.URL,
		})
	}

	// Use first torrent's info hash for details
	infoHash := ""
	if len(movie.Torrents) > 0 {
		infoHash = strings.ToUpper(movie.Torrents[0].Hash)
	}

	uploadDate := ""
	if len(movie.Torrents) > 0 {
		uploadDate = movie.Torrents[0].DateUploaded
	}

	description := p.CleanHTML(movie.Description)
	if description == "" {
		description = p.CleanHTML(movie.Summary)
	}

	return &models.TorrentDetails{
		InfoHash:           infoHash,
		Name:               fmt.Sprintf("%s (%d)", movie.Title, movie.Year),
		Category:           models.CategoryMovies,
		Description:        description,
		MagnetURI:          magnetURI,
		Files:              files,
		Provider:           p.Name,
		DescriptionPageURL: pageURL,
		PosterURL:          movie.LargeCover,
		ScreenshotURLs:     []string{movie.MediumCover, movie.Background},
		UploadDate:         uploadDate,
		Seeders:            0,
		Peers:              0,
	}
}

// GetLatestMovies gets latest movies
func (p *YTSProvider) GetLatestMovies(ctx context.Context, page int) ([]models.Torrent, error) {
	searchURL := fmt.Sprintf("%s/list_movies.json?limit=50&page=%d&sort_by=date_added",
		p.apiBaseURL, page)
	return p.searchJSON(ctx, searchURL)
}

// GetTopMovies gets top seeded movies
func (p *YTSProvider) GetTopMovies(ctx context.Context, page int) ([]models.Torrent, error) {
	searchURL := fmt.Sprintf("%s/list_movies.json?limit=50&page=%d&sort_by=seeds",
		p.apiBaseURL, page)
	return p.searchJSON(ctx, searchURL)
}
