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

// ThePirateBayProvider implements The Pirate Bay torrent search
type ThePirateBayProvider struct {
	*BaseProvider
	mirrorURLs []string
}

// NewThePirateBayProvider creates a new The Pirate Bay provider
func NewThePirateBayProvider(userAgent string, timeout int, enabled bool) *ThePirateBayProvider {
	base := NewBaseProvider(
		"thepiratebay",
		"https://thepiratebay.org",
		"https://thepiratebay.org/favicon.ico",
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

	// Multiple mirror URLs for TPB
	mirrors := []string{
		"https://thepiratebay.org",
		"https://piratebay.party",
		"https://tpb.party",
		"https://thepiratebay10.org",
		"https://pirateproxy.live",
	}

	return &ThePirateBayProvider{
		BaseProvider: base,
		mirrorURLs:   mirrors,
	}
}

// Search searches for torrents on The Pirate Bay
func (p *ThePirateBayProvider) Search(ctx context.Context, query string, category models.Category, page int) ([]models.Torrent, error) {
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

	// If all mirrors fail, return last error
	return nil, lastErr
}

// searchOnMirror searches on a specific mirror
func (p *ThePirateBayProvider) searchOnMirror(ctx context.Context, mirror, query string, category models.Category, page int) ([]models.Torrent, error) {
	searchURL := p.BuildSearchURL(mirror, query, category, page)
	doc, err := p.FetchHTML(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	return p.parseResults(doc, searchURL), nil
}

// BuildSearchURL builds the search URL for TPB
func (p *ThePirateBayProvider) BuildSearchURL(baseURL, query string, category models.Category, page int) string {
	categoryMap := map[models.Category]string{
		models.CategoryMovies:   "201",
		models.CategoryTV:       "205",
		models.CategoryMusic:    "101",
		models.CategoryGames:    "401",
		models.CategorySoftware: "301",
		models.CategoryAnime:    "207",
		models.CategoryBooks:    "601",
		models.CategoryXXX:      "501",
	}

	cat := "0"
	if c, ok := categoryMap[category]; ok {
		cat = c
	}

	if page > 0 {
		return fmt.Sprintf("%s/search/%s/%d/%s/0", baseURL, url.PathEscape(query), page, cat)
	}
	return fmt.Sprintf("%s/search/%s/0/%s/0", baseURL, url.PathEscape(query), cat)
}

// parseResults parses TPB search results
func (p *ThePirateBayProvider) parseResults(doc *goquery.Document, pageURL string) []models.Torrent {
	var torrents []models.Torrent

	doc.Find("#searchResult tr").Each(func(i int, s *goquery.Selection) {
		// Skip header row
		if i == 0 {
			return
		}

		nameLink := s.Find("td:nth-child(2) a.detLink")
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

		// Extract details from description
		descCell := s.Find("td:nth-child(2) font.detDesc")
		descText := strings.TrimSpace(descCell.Text())

		// Parse size, date from description
		var uploadDate, size string
		if descText != "" {
			// Format: "Uploaded 10-15 12:34, Size 1.2 GiB, ULed by user"
			parts := strings.Split(descText, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "Uploaded") {
					uploadDate = strings.TrimSpace(strings.TrimPrefix(part, "Uploaded"))
				} else if strings.HasPrefix(part, "Size") {
					size = strings.TrimSpace(strings.TrimPrefix(part, "Size"))
				}
			}
		}

		// Extract seeders/leechers
		seeders := 0
		leechers := 0
		seedersText := strings.TrimSpace(s.Find("td:nth-child(3)").Text())
		leechersText := strings.TrimSpace(s.Find("td:nth-child(4)").Text())
		seeders = parseInt(seedersText)
		leechers = parseInt(leechersText)

		// Build detail URL
		detailPageURL := p.ResolveURL(detailURL)

		// Build magnet if not present but we have info hash
		if magnetLink == "" && infoHash != "" {
			magnetLink = fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s", infoHash, url.QueryEscape(name))
		}

		// Determine category from URL or description
		category := p.detectCategory(detailURL, descText)

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
			UploadDate:          p.parseDate(uploadDate),
			Category:            category,
		}

		torrents = append(torrents, torrent)
	})

	return torrents
}

// GetDetails gets detailed torrent information
func (p *ThePirateBayProvider) GetDetails(ctx context.Context, infoHash string) (*models.TorrentDetails, error) {
	// TPB needs detail page URL, not just info hash
	return nil, fmt.Errorf("GetDetails requires detail page URL for The Pirate Bay")
}

// GetDetailsFromURL gets details from a detail page URL
func (p *ThePirateBayProvider) GetDetailsFromURL(ctx context.Context, detailURL string) (*models.TorrentDetails, error) {
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

// parseDetails parses TPB detail page
func (p *ThePirateBayProvider) parseDetails(doc *goquery.Document, pageURL string) *models.TorrentDetails {
	name := strings.TrimSpace(doc.Find("#title").Text())
	if name == "" {
		name = strings.TrimSpace(doc.Find("h1").First().Text())
	}

	// Extract info hash
	infoHash := ""
	doc.Find("#detailsframe").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		hashRegex := regexp.MustCompile(`[a-fA-F0-9]{40}`)
		matches := hashRegex.FindStringSubmatch(text)
		if len(matches) > 0 {
			infoHash = strings.ToUpper(matches[0])
		}
	})

	// Also check magnet links
	magnetLink := ""
	doc.Find("a[href^='magnet:']").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			magnetLink = href
			if infoHash == "" {
				infoHash = p.ExtractInfoHash(href)
			}
		}
	})

	// Extract description
	description := ""
	doc.Find("#details").Each(func(i int, s *goquery.Selection) {
		description = s.Text()
	})

	// Extract files
	var files []models.TorrentFile
	doc.Find("#filelist tbody tr, .filelist tr").Each(func(i int, s *goquery.Selection) {
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

	// Extract seeders/leechers
	seeders := 0
	leechers := 0
	doc.Find("#seeders, #leechers").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if i == 0 {
			seeders = parseInt(text)
		} else {
			leechers = parseInt(text)
		}
	})

	// Extract size from details
	size := ""
	doc.Find("#details dl dt").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(strings.ToLower(s.Text()), "size") {
			size = strings.TrimSpace(s.Next().Text())
		}
	})

	category := p.detectCategory(pageURL, description)

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

// detectCategory detects category from URL or description
func (p *ThePirateBayProvider) detectCategory(url, desc string) models.Category {
	combined := strings.ToLower(url + " " + desc)

	categoryMap := map[string]models.Category{
		"/video/":       models.CategoryMovies,
		"/movies/":      models.CategoryMovies,
		"/tv/":          models.CategoryTV,
		"/audio/":       models.CategoryMusic,
		"/music/":       models.CategoryMusic,
		"/games/":       models.CategoryGames,
		"/applications/": models.CategorySoftware,
		"/anime/":       models.CategoryAnime,
		"/ebooks/":      models.CategoryBooks,
		"/porn/":        models.CategoryXXX,
		"/other/":       models.CategoryOther,
	}

	for keyword, cat := range categoryMap {
		if strings.Contains(combined, keyword) {
			return cat
		}
	}

	return models.CategoryOther
}

// parseDate parses TPB date format
func (p *ThePirateBayProvider) parseDate(dateStr string) string {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return ""
	}

	// TPB format: "MM-DD HH:MM" (assumes current year)
	now := time.Now()
	currentYear := now.Year()

	// Try parsing with current year
	format := "01-02 15:04"
	if t, err := time.Parse(format, dateStr); err == nil {
		t = t.AddDate(currentYear, 0, 0)
		// If date is in future, assume last year
		if t.After(now) {
			t = t.AddDate(-1, 0, 0)
		}
		return t.Format(time.RFC3339)
	}

	return dateStr
}

