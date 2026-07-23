package providers

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"torrentsearch-web/internal/models"
)

// LimeTorrentsProvider implements LimeTorrents torrent search
type LimeTorrentsProvider struct {
	*BaseProvider
	mirrorURLs []string
}

// NewLimeTorrentsProvider creates a new LimeTorrents provider
func NewLimeTorrentsProvider(userAgent string, timeout int, enabled bool) *LimeTorrentsProvider {
	base := NewBaseProvider(
		"limetorrents",
		"https://www.limetorrents.lol",
		"https://www.limetorrents.lol/favicon.ico",
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
		"https://www.limetorrents.lol",
		"https://limetorrents.asia",
		"https://limetorrents.zone",
		"https://limetorrents.co",
	}

	return &LimeTorrentsProvider{
		BaseProvider: base,
		mirrorURLs:   mirrors,
	}
}

// Search searches for torrents on LimeTorrents
func (p *LimeTorrentsProvider) Search(ctx context.Context, query string, category models.Category, page int) ([]models.Torrent, error) {
	if !p.Enabled {
		return nil, nil
	}

	// Try mirrors until one works
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
func (p *LimeTorrentsProvider) searchOnMirror(ctx context.Context, mirror, query string, category models.Category, page int) ([]models.Torrent, error) {
	searchURL := p.BuildSearchURL(mirror, query, category, page)
	doc, err := p.FetchHTML(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	return p.parseResults(doc, searchURL), nil
}

// BuildSearchURL builds the search URL for LimeTorrents
func (p *LimeTorrentsProvider) BuildSearchURL(baseURL, query string, category models.Category, page int) string {
	categoryMap := map[models.Category]string{
		models.CategoryMovies:   "movies",
		models.CategoryTV:       "tv",
		models.CategoryMusic:    "music",
		models.CategoryGames:    "games",
		models.CategorySoftware: "applications",
		models.CategoryAnime:    "anime",
		models.CategoryBooks:    "ebooks",
		models.CategoryXXX:      "xxx",
	}

	cat := "all"
	if c, ok := categoryMap[category]; ok {
		cat = c
	}

	if page > 1 {
		return fmt.Sprintf("%s/search/%s/%s/%d/", baseURL, cat, url.PathEscape(query), page)
	}
	return fmt.Sprintf("%s/search/%s/%s/1/", baseURL, cat, url.PathEscape(query))
}

// parseResults parses LimeTorrents search results
func (p *LimeTorrentsProvider) parseResults(doc *goquery.Document, pageURL string) []models.Torrent {
	var torrents []models.Torrent

	doc.Find("table.table2 tbody tr").Each(func(i int, s *goquery.Selection) {
		// Skip header row
		if i == 0 {
			return
		}

		nameLink := s.Find("td.tdname a:not([title])").First()
		if nameLink.Length() == 0 {
			nameLink = s.Find("td.tdname a").First()
		}

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

		// Extract info hash from magnet
		infoHash := p.ExtractInfoHash(magnetLink)

		// Extract size
		size := strings.TrimSpace(s.Find("td.tdnormal:nth-child(2)").Text())
		if size == "" {
			size = strings.TrimSpace(s.Find("td:nth-child(3)").Text())
		}

		// Extract seeders/leechers
		seeders := 0
		leechers := 0
		seedsText := strings.TrimSpace(s.Find("td.tdseed").Text())
		leechsText := strings.TrimSpace(s.Find("td.tdleech").Text())
		seeders = parseInt(seedsText)
		leechers = parseInt(leechsText)

		// Extract date
		dateStr := strings.TrimSpace(s.Find("td.tdnormal:last-child").Text())

		// Build detail URL
		detailPageURL := p.ResolveURL(detailURL)

		// Build magnet if not present but we have info hash
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
			DescriptionPageURL:  detailPageURL,
			UploadDate:          p.parseDate(dateStr),
			Category:            category,
		}

		torrents = append(torrents, torrent)
	})

	return torrents
}

// GetDetails gets detailed torrent information
func (p *LimeTorrentsProvider) GetDetails(ctx context.Context, infoHash string) (*models.TorrentDetails, error) {
	return nil, fmt.Errorf("GetDetails requires detail page URL for LimeTorrents")
}

// GetDetailsFromURL gets details from a detail page URL
func (p *LimeTorrentsProvider) GetDetailsFromURL(ctx context.Context, detailURL string) (*models.TorrentDetails, error) {
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

// parseDetails parses LimeTorrents detail page
func (p *LimeTorrentsProvider) parseDetails(doc *goquery.Document, pageURL string) *models.TorrentDetails {
	name := strings.TrimSpace(doc.Find("h1").First().Text())

	// Extract info hash
	infoHash := ""
	doc.Find("div.detDesc").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		hashRegex := regexp.MustCompile(`[a-fA-F0-9]{40}`)
		matches := hashRegex.FindStringSubmatch(text)
		if len(matches) > 0 {
			infoHash = strings.ToUpper(matches[0])
		}
	})

	// Extract magnet link
	magnetLink := ""
	doc.Find("a[href^='magnet:']").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			magnetLink = href
			if infoHash == "" {
				infoHash = p.ExtractInfoHash(href)
			}
		}
	})

	// Extract size
	size := ""
	doc.Find("div.detDesc").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Size:") {
			parts := strings.Split(text, "Size:")
			if len(parts) > 1 {
				size = strings.TrimSpace(strings.Split(parts[1], ",")[0])
			}
		}
	})

	// Extract seeders/leechers
	seeders := 0
	leechers := 0
	doc.Find("span.seeds, span.leechers").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if strings.Contains(s.Parent().Text(), "Seeds") || i == 0 {
			seeders = parseInt(text)
		} else {
			leechers = parseInt(text)
		}
	})

	// Extract description
	description := ""
	doc.Find("#description").Each(func(i int, s *goquery.Selection) {
		description = s.Text()
	})

	// Extract files
	var files []models.TorrentFile
	doc.Find(".filelist tr").Each(func(i int, s *goquery.Selection) {
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
		Files:              files,
		Provider:           p.Name,
		DescriptionPageURL: pageURL,
	}
}

// detectCategory detects category from URL
func (p *LimeTorrentsProvider) detectCategory(url string) models.Category {
	url = strings.ToLower(url)

	categoryMap := map[string]models.Category{
		"/movies/":       models.CategoryMovies,
		"/tv/":           models.CategoryTV,
		"/music/":        models.CategoryMusic,
		"/games/":        models.CategoryGames,
		"/applications/": models.CategorySoftware,
		"/anime/":        models.CategoryAnime,
		"/ebooks/":       models.CategoryBooks,
		"/xxx/":          models.CategoryXXX,
	}

	for keyword, cat := range categoryMap {
		if strings.Contains(url, keyword) {
			return cat
		}
	}

	return models.CategoryOther
}

// parseDate parses date string
func (p *LimeTorrentsProvider) parseDate(dateStr string) string {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return ""
	}

	// Try multiple formats
	formats := []string{
		"02-01-2006",
		"2006-01-02",
		"Jan 2, 2006",
		"January 2, 2006",
		"02/01/2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format(time.RFC3339)
		}
	}

	return dateStr
}

