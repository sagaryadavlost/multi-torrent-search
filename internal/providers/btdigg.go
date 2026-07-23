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

// BTDiggProvider implements BTDigg torrent search (DHT based)
type BTDiggProvider struct {
	*BaseProvider
	mirrorURLs []string
}

// NewBTDiggProvider creates a new BTDigg provider
func NewBTDiggProvider(userAgent string, timeout int, enabled bool) *BTDiggProvider {
	base := NewBaseProvider(
		"btdigg",
		"https://btdig.com",
		"https://btdig.com/favicon.ico",
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
		"https://btdig.com",
		"https://btdigg.org",
		"https://dht.btdigg.org",
	}

	return &BTDiggProvider{
		BaseProvider: base,
		mirrorURLs:   mirrors,
	}
}

// Search searches for torrents on BTDigg
func (p *BTDiggProvider) Search(ctx context.Context, query string, category models.Category, page int) ([]models.Torrent, error) {
	if !p.Enabled {
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
func (p *BTDiggProvider) searchOnMirror(ctx context.Context, mirror, query string, page int) ([]models.Torrent, error) {
	searchURL := p.BuildSearchURL(mirror, query, page)
	doc, err := p.FetchHTML(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	return p.parseResults(doc, searchURL), nil
}

// BuildSearchURL builds the search URL for BTDigg
func (p *BTDiggProvider) BuildSearchURL(baseURL, query string, page int) string {
	if page > 1 {
		return fmt.Sprintf("%s/search?q=%s&p=%d", baseURL, url.QueryEscape(query), page)
	}
	return fmt.Sprintf("%s/search?q=%s", baseURL, url.QueryEscape(query))
}

// parseResults parses BTDigg search results
func (p *BTDiggProvider) parseResults(doc *goquery.Document, pageURL string) []models.Torrent {
	var torrents []models.Torrent

	doc.Find(".search-result .item").Each(func(i int, s *goquery.Selection) {
		nameLink := s.Find("h3 a, .title a").First()
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
		size := ""
		s.Find(".size, .info").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if strings.Contains(text, "Size") || strings.Contains(text, "size") {
				size = text
			}
		})

		// Extract seeders/leechers (DHT doesn't have real seeders)
		seeders := 0
		leechers := 0

		// Extract date
		dateStr := ""
		s.Find(".date, .time").Each(func(i int, s *goquery.Selection) {
			dateStr = strings.TrimSpace(s.Text())
		})

		// Build detail URL
		detailPageURL := p.ResolveURL(detailURL)

		// Build magnet if not present but we have info hash
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
			DescriptionPageURL:  detailPageURL,
			UploadDate:          p.parseDate(dateStr),
			Category:            models.CategoryOther,
		}

		torrents = append(torrents, torrent)
	})

	return torrents
}

// GetDetails gets detailed torrent information
func (p *BTDiggProvider) GetDetails(ctx context.Context, infoHash string) (*models.TorrentDetails, error) {
	return nil, fmt.Errorf("GetDetails requires detail page URL for BTDigg")
}

// GetDetailsFromURL gets details from a detail page URL
func (p *BTDiggProvider) GetDetailsFromURL(ctx context.Context, detailURL string) (*models.TorrentDetails, error) {
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

// parseDetails parses BTDigg detail page
func (p *BTDiggProvider) parseDetails(doc *goquery.Document, pageURL string) *models.TorrentDetails {
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

	// Extract size
	size := ""
	doc.Find(".info").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Size") {
			re := regexp.MustCompile(`Size:\s*([\d.]+\s*[KMGT]?B)`)
			matches := re.FindStringSubmatch(text)
			if len(matches) > 1 {
				size = matches[1]
			}
		}
	})

	// Extract description
	description := ""
	doc.Find(".description, .content").Each(func(i int, s *goquery.Selection) {
		description = s.Text()
	})

	// Extract files
	var files []models.TorrentFile
	doc.Find(".files li, .filelist tr").Each(func(i int, s *goquery.Selection) {
		fileName := strings.TrimSpace(s.Find(".name, td:first-child").Text())
		fileSize := strings.TrimSpace(s.Find(".size, td:last-child").Text())
		if fileName != "" {
			files = append(files, models.TorrentFile{
				Name:       fileName,
				Size:       fileSize,
				SizeBytes:  p.ParseSize(fileSize),
			})
		}
	})

	return &models.TorrentDetails{
		InfoHash:           strings.ToUpper(infoHash),
		Name:               name,
		Size:               size,
		SizeBytes:          p.ParseSize(size),
		Seeders:            0,
		Peers:              0,
		Category:           models.CategoryOther,
		Description:        p.CleanHTML(description),
		MagnetURI:          magnetLink,
		Files:              files,
		Provider:           p.Name,
		DescriptionPageURL: pageURL,
	}
}

// parseDate parses date string
func (p *BTDiggProvider) parseDate(dateStr string) string {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return ""
	}

	formats := []string{
		"2006-01-02",
		"02.01.2006",
		"Jan 2, 2006",
		"2006/01/02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format(time.RFC3339)
		}
	}

	return dateStr
}

