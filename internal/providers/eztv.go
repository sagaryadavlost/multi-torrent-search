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

// EZTVProvider implements EZTV torrent search (TV shows)
type EZTVProvider struct {
	*BaseProvider
	mirrorURLs []string
}

// NewEZTVProvider creates a new EZTV provider
func NewEZTVProvider(userAgent string, timeout int, enabled bool) *EZTVProvider {
	base := NewBaseProvider(
		"eztv",
		"https://eztv.re",
		"https://eztv.re/favicon.ico",
		enabled,
		userAgent,
		time.Duration(timeout)*time.Second,
		[]models.Category{models.CategoryTV},
	)

	mirrors := []string{
		"https://eztv.re",
		"https://eztv.io",
		"https://eztvx.to",
	}

	return &EZTVProvider{
		BaseProvider: base,
		mirrorURLs:   mirrors,
	}
}

// Search searches for torrents on EZTV
func (p *EZTVProvider) Search(ctx context.Context, query string, category models.Category, page int) ([]models.Torrent, error) {
	if !p.Enabled {
		return nil, nil
	}

	if category != models.CategoryAll && category != models.CategoryTV {
		return nil, nil
	}

	var lastErr error
	for _, mirror := range p.mirrorURLs {
		p.BaseURLStr = mirror
		torrents, err := p.searchOnMirror(ctx, mirror, query, page)
		if err == nil && len(torrents) > 0 {
			return torrents, nil
		}
		lastErr = err
	}

	return nil, lastErr
}

// searchOnMirror searches on a specific mirror
func (p *EZTVProvider) searchOnMirror(ctx context.Context, mirror, query string, page int) ([]models.Torrent, error) {
	searchURL := p.BuildSearchURL(mirror, query, page)
	doc, err := p.FetchHTML(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	return p.parseResults(doc, searchURL), nil
}

// BuildSearchURL builds the search URL for EZTV
func (p *EZTVProvider) BuildSearchURL(baseURL, query string, page int) string {
	if page > 1 {
		return fmt.Sprintf("%s/search/%s?page=%d", baseURL, url.PathEscape(query), page)
	}
	return fmt.Sprintf("%s/search/%s", baseURL, url.PathEscape(query))
}

// parseResults parses EZTV search results
func (p *EZTVProvider) parseResults(doc *goquery.Document, pageURL string) []models.Torrent {
	var torrents []models.Torrent

	doc.Find("table.forum_header_border tr[name='hover']").Each(func(i int, s *goquery.Selection) {
		nameLink := s.Find("td:nth-child(2) a.epinfo")
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

		// Extract torrent file link
		torrentLink := ""
		s.Find("a[href$='.torrent']").Each(func(i int, s *goquery.Selection) {
			if href, ok := s.Attr("href"); ok {
				torrentLink = href
			}
		})

		// Extract info hash
		infoHash := p.ExtractInfoHash(magnetLink)

		// Extract size
		size := strings.TrimSpace(s.Find("td:nth-child(4)").Text())

		// Extract seeders/peers
		seeders := 0
		leechers := 0
		seedsText := strings.TrimSpace(s.Find("td:nth-child(5) font").First().Text())
		peersText := strings.TrimSpace(s.Find("td:nth-child(6) font").First().Text())
		seeders = parseInt(seedsText)
		leechers = parseInt(peersText)

		// Extract date
		dateStr := strings.TrimSpace(s.Find("td:nth-child(3)").Text())

		// Build detail URL
		detailPageURL := p.ResolveURL(detailURL)

		// Build magnet if we have hash
		if magnetLink == "" && infoHash != "" {
			magnetLink = fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s", infoHash, url.QueryEscape(name))
		}

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
			FileDownloadLink:    p.ResolveURL(torrentLink),
			DescriptionPageURL:  detailPageURL,
			UploadDate:          p.parseDate(dateStr),
			Category:            models.CategoryTV,
		}

		torrents = append(torrents, torrent)
	})

	return torrents
}

// GetDetails gets detailed torrent information
func (p *EZTVProvider) GetDetails(ctx context.Context, infoHash string) (*models.TorrentDetails, error) {
	return nil, fmt.Errorf("GetDetails requires detail page URL for EZTV")
}

// GetDetailsFromURL gets details from a detail page URL
func (p *EZTVProvider) GetDetailsFromURL(ctx context.Context, detailURL string) (*models.TorrentDetails, error) {
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

// parseDetails parses EZTV detail page
func (p *EZTVProvider) parseDetails(doc *goquery.Document, pageURL string) *models.TorrentDetails {
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

	// Extract torrent file
	torrentLink := ""
	doc.Find("a[href$='.torrent']").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			torrentLink = href
		}
	})

	// Extract size
	size := ""
	doc.Find(".thread_post").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Size:") {
			re := regexp.MustCompile(`Size:\s*([\d.]+\s*[KMGT]?B)`)
			matches := re.FindStringSubmatch(text)
			if len(matches) > 1 {
				size = matches[1]
			}
		}
	})

	// Extract seeders/leechers
	seeders := 0
	leechers := 0
	doc.Find(".thread_post font").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if strings.Contains(text, "Peers:") {
			re := regexp.MustCompile(`Peers:\s*(\d+)\s+Seeds:\s*(\d+)`)
			matches := re.FindStringSubmatch(text)
			if len(matches) > 2 {
				leechers = parseInt(matches[1])
				seeders = parseInt(matches[2])
			}
		}
	})

	// Extract description
	description := ""
	doc.Find(".thread_post").Each(func(i int, s *goquery.Selection) {
		description = s.Text()
	})

	// Extract files (usually just the episode file)
	var files []models.TorrentFile
	if name != "" && size != "" {
		files = append(files, models.TorrentFile{
			Name:       name,
			Size:       size,
			SizeBytes:  p.ParseSize(size),
		})
	}

	return &models.TorrentDetails{
		InfoHash:           strings.ToUpper(infoHash),
		Name:               name,
		Size:               size,
		SizeBytes:          p.ParseSize(size),
		Seeders:            seeders,
		Peers:              leechers,
		Category:           models.CategoryTV,
		Description:        p.CleanHTML(description),
		MagnetURI:          magnetLink,
		FileDownloadLink:   p.ResolveURL(torrentLink),
		Files:              files,
		Provider:           p.Name,
		DescriptionPageURL: pageURL,
	}
}

// parseDate parses EZTV date format
func (p *EZTVProvider) parseDate(dateStr string) string {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return ""
	}

	// EZTV format: "Today", "Yesterday", "MM/DD/YYYY"
	now := time.Now()
	switch strings.ToLower(dateStr) {
	case "today":
		return now.Format(time.RFC3339)
	case "yesterday":
		return now.AddDate(0, 0, -1).Format(time.RFC3339)
	}

	formats := []string{
		"01/02/2006",
		"2006-01-02",
		"Jan 2, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format(time.RFC3339)
		}
	}

	return dateStr
}

