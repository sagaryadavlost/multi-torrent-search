package providers

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"torrentsearch-web/internal/models"
)

// TorrentDownloadsProvider implements TorrentDownloads torrent search
type TorrentDownloadsProvider struct {
	*BaseProvider
	mirrorURLs []string
}

// NewTorrentDownloadsProvider creates a new TorrentDownloads provider
func NewTorrentDownloadsProvider(userAgent string, timeout int, enabled bool) *TorrentDownloadsProvider {
	base := NewBaseProvider(
		"torrentdownloads",
		"https://www.torrentdownloads.pro",
		"https://www.torrentdownloads.pro/favicon.ico",
		enabled,
		userAgent,
		time.Duration(timeout)*time.Second,
		[]models.Category{
			models.CategoryMovies,
			models.CategoryTV,
			models.CategoryMusic,
			models.CategoryGames,
			models.CategorySoftware,
			models.CategoryAnime,
			models.CategoryBooks,
			models.CategoryXXX,
		},
	)

	mirrors := []string{
		"https://www.torrentdownloads.pro",
		"https://torrentdownloads.unblocked.lol",
		"https://torrentdownloads.unblocked.si",
	}

	return &TorrentDownloadsProvider{
		BaseProvider: base,
		mirrorURLs:   mirrors,
	}
}

// Search searches for torrents on TorrentDownloads
func (p *TorrentDownloadsProvider) Search(ctx context.Context, query string, category models.Category, page int) ([]models.Torrent, error) {
	if !p.Enabled {
		return nil, nil
	}

	var lastErr error
	for _, mirror := range p.mirrorURLs {
		p.BaseURLStr = mirror
		torrents, err := p.searchOnMirror(ctx, mirror, query, category, page)
		if err == nil && len(torrents) > 0 {
			return torrents, nil
		}
		lastErr = err
	}

	return nil, lastErr
}

// searchOnMirror searches on a specific mirror
func (p *TorrentDownloadsProvider) searchOnMirror(ctx context.Context, mirror, query string, category models.Category, page int) ([]models.Torrent, error) {
	searchURL := p.BuildSearchURL(mirror, query, category, page)
	doc, err := p.FetchHTML(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	return p.parseResults(doc, searchURL), nil
}

// BuildSearchURL builds the search URL for TorrentDownloads
func (p *TorrentDownloadsProvider) BuildSearchURL(baseURL, query string, category models.Category, page int) string {
	categoryMap := map[models.Category]string{
		models.CategoryMovies:   "movies",
		models.CategoryTV:       "tv-shows",
		models.CategoryMusic:    "music",
		models.CategoryGames:    "games",
		models.CategorySoftware: "software",
		models.CategoryAnime:    "anime",
		models.CategoryBooks:    "ebooks",
		models.CategoryXXX:      "xxx",
	}

	cat := ""
	if c, ok := categoryMap[category]; ok {
		cat = "/" + c
	}

	if page > 1 {
		return fmt.Sprintf("%s/search/%s%s/%d/", baseURL, url.PathEscape(query), cat, page)
	}
	return fmt.Sprintf("%s/search/%s%s/", baseURL, url.PathEscape(query), cat)
}

// parseResults parses TorrentDownloads search results
func (p *TorrentDownloadsProvider) parseResults(doc *goquery.Document, pageURL string) []models.Torrent {
	var torrents []models.Torrent

	doc.Find(".grey_bar3_back .grey_bar3").Each(func(i int, s *goquery.Selection) {
		nameLink := s.Find("a[href^='/torrent/']").First()
		name := strings.TrimSpace(nameLink.Text())
		detailURL, _ := nameLink.Attr("href")

		if name == "" || detailURL == "" {
			return
		}

		// Extract magnet link
		magnetLink := ""
		s.Find("a[href^='magnet:']").Each(func(i int, s *goquery.Selection) {
			if href, ok := s.Attr("href"); ok {
				magnetLink = href
			}
		})

		// Extract torrent download link
		torrentLink := ""
		s.Find("a[href$='.torrent']").Each(func(i int, s *goquery.Selection) {
			if href, ok := s.Attr("href"); ok {
				torrentLink = href
			}
		})

		// Extract info hash
		infoHash := p.ExtractInfoHash(magnetLink)

		// Extract size
		size := ""
		s.Find(".size").Each(func(i int, s *goquery.Selection) {
			size = strings.TrimSpace(s.Text())
		})

		// Extract seeders/leechers
		seeders := 0
		leechers := 0
		s.Find(".seeds, .leechs").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if i == 0 {
				seeders = parseInt(text)
			} else {
				leechers = parseInt(text)
			}
		})

		// Extract date
		dateStr := ""
		s.Find(".date").Each(func(i int, s *goquery.Selection) {
			dateStr = strings.TrimSpace(s.Text())
		})

		// Build URLs
		detailPageURL := p.ResolveURL(detailURL)
		torrentDownloadLink := p.ResolveURL(torrentLink)

		// Build magnet if we have hash but no magnet
		if magnetLink == "" && infoHash != "" {
			magnetLink = fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s", infoHash, url.QueryEscape(name))
		}

		// Determine category from URL
		category := p.detectCategory(detailURL)

		torrent := models.Torrent{
			InfoHash:            strings.ToUpper(infoHash),
			Name:                name,
			Size:                size,
			SizeBytes:           p.ParseSize(size),
			Seeders:             seeders,
			Peers:               leechers,
			Provider:            p.Name,
			ProviderName:        p.Name,
			MagnetURI:           magnetLink,
			FileDownloadLink:    torrentDownloadLink,
			DescriptionPageURL:  detailPageURL,
			UploadDate:          p.parseDate(dateStr),
			Category:            category,
		}

		torrents = append(torrents, torrent)
	})

	return torrents
}

// GetDetails gets detailed torrent information
func (p *TorrentDownloadsProvider) GetDetails(ctx context.Context, infoHash string) (*models.TorrentDetails, error) {
	return nil, fmt.Errorf("GetDetails requires detail page URL for TorrentDownloads")
}

// GetDetailsFromURL gets details from a detail page URL
func (p *TorrentDownloadsProvider) GetDetailsFromURL(ctx context.Context, detailURL string) (*models.TorrentDetails, error) {
	if !p.Enabled {
		return nil, nil
	}

	fullURL := p.ResolveURL(detailURL)
	doc, err := p.FetchHTML(ctx, fullURL)
	if err != nil {
		return nil, err
	}

	return p.parseDetails(doc, fullURL), nil
}

// parseDetails parses TorrentDownloads detail page
func (p *TorrentDownloadsProvider) parseDetails(doc *goquery.Document, pageURL string) *models.TorrentDetails {
	name := strings.TrimSpace(doc.Find("h1").First().Text())

	// Extract info hash
	infoHash := ""
	doc.Find("a[href^='magnet:']").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			if infoHash == "" {
				infoHash = p.ExtractInfoHash(href)
			}
		}
	})

	// Extract magnet link
	magnetLink := ""
	doc.Find("a[href^='magnet:']").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			magnetLink = href
		}
	})

	// Extract torrent download link
	torrentLink := ""
	doc.Find("a[href$='.torrent']").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			torrentLink = href
		}
	})

	// Extract size
	size := ""
	doc.Find(".size").Each(func(i int, s *goquery.Selection) {
		size = strings.TrimSpace(s.Text())
	})

	// Extract seeders/leechers
	seeders := 0
	leechers := 0
	doc.Find(".seeds, .leechs").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if i == 0 {
			seeders = parseInt(text)
		} else {
			leechers = parseInt(text)
		}
	})

	// Extract description
	description := ""
	doc.Find("#description, .description").Each(func(i int, s *goquery.Selection) {
		description = s.Text()
	})

	// Extract files
	var files []models.TorrentFile
	doc.Find("#filelist tr, .filelist tr").Each(func(i int, s *goquery.Selection) {
		fileName := strings.TrimSpace(s.Find("td:first-child").Text())
		fileSize := strings.TrimSpace(s.Find("td:last-child").Text())
		if fileName != "" {
			files = append(files, models.TorrentFile{
				Name:       fileName,
				Size:       fileSize,
				SizeBytes:  p.ParseSize(fileSize),
			})
		}
	})

	category := p.detectCategory(pageURL)

	return &models.TorrentDetails{
		InfoHash:           strings.ToUpper(infoHash),
		Name:               name,
		Size:               size,
		SizeBytes:          p.ParseSize(size),
		Seeders:            seeders,
		Peers:              leechers,
		Category:           category,
		Description:        p.CleanHTML(description),
		MagnetURI:          magnetLink,
		FileDownloadLink:   p.ResolveURL(torrentLink),
		Files:              files,
		Provider:           p.Name,
		DescriptionPageURL: pageURL,
	}
}

// detectCategory detects category from URL
func (p *TorrentDownloadsProvider) detectCategory(url string) models.Category {
	url = strings.ToLower(url)

	categoryMap := map[string]models.Category{
		"/movies/":      models.CategoryMovies,
		"/tv-shows/":    models.CategoryTV,
		"/music/":       models.CategoryMusic,
		"/games/":       models.CategoryGames,
		"/software/":    models.CategorySoftware,
		"/anime/":       models.CategoryAnime,
		"/ebooks/":      models.CategoryBooks,
		"/xxx/":         models.CategoryXXX,
	}

	for keyword, cat := range categoryMap {
		if strings.Contains(url, keyword) {
			return cat
		}
	}

	return models.CategoryOther
}

// parseDate parses date string
func (p *TorrentDownloadsProvider) parseDate(dateStr string) string {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return ""
	}

	formats := []string{
		"2006-01-02",
		"02/01/2006",
		"Jan 2, 2006",
		"January 2, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format(time.RFC3339)
		}
	}

	return dateStr
}

