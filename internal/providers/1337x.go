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

// Provider1337x implements 1337x torrent search
type Provider1337x struct {
	*BaseProvider
}

// NewProvider1337x creates a new 1337x provider
func NewProvider1337x(userAgent string, timeout int, enabled bool) *Provider1337x {
	base := NewBaseProvider(
		"1337x",
		"https://1337x.to",
		"https://1337x.to/favicon.ico",
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

	return &Provider1337x{BaseProvider: base}
}

// Search searches for torrents on 1337x
func (p *Provider1337x) Search(ctx context.Context, query string, category models.Category, page int) ([]models.Torrent, error) {
	if !p.Enabled {
		return nil, nil
	}

	searchURL := p.BuildSearchURL(query, category, page)
	doc, err := p.FetchHTML(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	return p.parseResults(doc, searchURL), nil
}

// BuildSearchURL builds the search URL for 1337x
func (p *Provider1337x) BuildSearchURL(query string, category models.Category, page int) string {
	categoryMap := map[models.Category]string{
		models.CategoryMovies:   "movies",
		models.CategoryTV:       "tv",
		models.CategoryMusic:    "music",
		models.CategoryGames:    "games",
		models.CategorySoftware: "apps",
		models.CategoryAnime:    "anime",
		models.CategoryBooks:    "books",
		models.CategoryXXX:      "xxx",
	}

	cat := ""
	if c, ok := categoryMap[category]; ok {
		cat = "/" + c
	}

	if page > 1 {
		return fmt.Sprintf("%s/search/%s%s/%d/", p.BaseURLStr, url.PathEscape(query), cat, page)
	}
	return fmt.Sprintf("%s/search/%s%s/1/", p.BaseURLStr, url.PathEscape(query), cat)
}

// parseResults parses the search results
func (p *Provider1337x) parseResults(doc *goquery.Document, pageURL string) []models.Torrent {
	var torrents []models.Torrent

	doc.Find("table.torrents tbody tr").Each(func(i int, s *goquery.Selection) {
		nameLink := s.Find("td.name a:not(.icon)")
		name := strings.TrimSpace(nameLink.Text())
		detailURL, _ := nameLink.Attr("href")

		if name == "" || detailURL == "" {
			return
		}

		// Extract info hash from detail URL
		infoHash := ""
		if strings.Contains(detailURL, "/torrent/") {
			parts := strings.Split(detailURL, "/")
			if len(parts) > 2 {
				// The hash might be in the URL or on detail page
				infoHash = parts[len(parts)-2]
			}
		}

		size := strings.TrimSpace(s.Find("td.size").Text())
		seeders := strings.TrimSpace(s.Find("td.seeds").Text())
		leechers := strings.TrimSpace(s.Find("td.leechs").Text())
		dateStr := strings.TrimSpace(s.Find("td.coll-date").Text())

		// Parse seeders/leechers
		seeds := parseInt(seeders)
		leechs := parseInt(leechers)

		// Build magnet if we have info hash
		magnetURI := ""
		if len(infoHash) >= 32 {
			magnetURI = fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s", strings.ToUpper(infoHash), url.QueryEscape(name))
		}

		// Build detail URL
		detailPageURL := p.ResolveURL(detailURL)

		uploadDate := p.parseDate(dateStr)
		torrent := models.Torrent{
			InfoHash:           strings.ToUpper(infoHash),
			Name:               name,
			Size:               size,
			SizeBytes:          p.ParseSize(size),
			Seeders:            seeds,
			Peers:              leechs,
			Provider:           p.Name,
			ProviderName:       p.Name,
			MagnetURI:          magnetURI,
			DescriptionPageURL: detailPageURL,
			UploadDate:         uploadDate,
		}

		torrents = append(torrents, torrent)
	})

	return torrents
}

// GetDetails gets detailed torrent information
func (p *Provider1337x) GetDetails(ctx context.Context, infoHash string) (*models.TorrentDetails, error) {
	if !p.Enabled {
		return nil, nil
	}

	// Need to get detail page URL from search results or construct it
	// 1337x detail pages have format: /torrent/{id}/{name}/
	// We'd need to search first to get the detail URL
	return nil, fmt.Errorf("GetDetails requires detail page URL for 1337x")
}

// GetDetailsFromURL gets details from a detail page URL
func (p *Provider1337x) GetDetailsFromURL(ctx context.Context, detailURL string) (*models.TorrentDetails, error) {
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

// parseDetails parses torrent details page
func (p *Provider1337x) parseDetails(doc *goquery.Document, pageURL string) *models.TorrentDetails {
	name := strings.TrimSpace(doc.Find("div.box-info-heading h1").Text())
	if name == "" {
		name = strings.TrimSpace(doc.Find("h1").First().Text())
	}

	// Extract info hash from page
	infoHash := ""
	doc.Find("div.box-info").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Info Hash") || strings.Contains(text, "infohash") {
			// Extract hash
			hashRegex := regexp.MustCompile(`[a-fA-F0-9]{40}`)
			matches := hashRegex.FindStringSubmatch(text)
			if len(matches) > 0 {
				infoHash = strings.ToUpper(matches[0])
			}
		}
	})

	// Extract magnet link
	magnetLink := ""
	doc.Find("a[href^='magnet:']").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok && strings.Contains(href, "magnet:") {
			magnetLink = href
			// Extract hash from magnet
			if infoHash == "" {
				infoHash = p.ExtractInfoHash(href)
			}
		}
	})

	// Extract size
	size := ""
	doc.Find("div.box-info").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Size") {
			// Find size text
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				if strings.Contains(strings.ToLower(line), "size") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						size = strings.TrimSpace(parts[1])
					}
					break
				}
			}
		}
	})

	// Extract seeders/leechers
	seeders := 0
	leechers := 0
	doc.Find("span.seeds, span.leechers").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if i == 0 {
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

	// Extract files list
	var files []models.TorrentFile
	doc.Find("#filelist tbody tr, .torrent-file-list li").Each(func(i int, s *goquery.Selection) {
		fileName := strings.TrimSpace(s.Find("td:first-child, .file-name").Text())
		fileSize := strings.TrimSpace(s.Find("td:last-child, .file-size").Text())
		if fileName != "" {
			files = append(files, models.TorrentFile{
				Name: fileName,
				Size: fileSize,
				SizeBytes: p.ParseSize(fileSize),
			})
		}
	})

	// Extract category
	category := p.detectCategory(doc)

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

// detectCategory detects category from page
func (p *Provider1337x) detectCategory(doc *goquery.Document) models.Category {
	breadcrumb := doc.Find("nav.breadcrumb, .breadcrumb").Text()
	breadcrumb = strings.ToLower(breadcrumb)

	categoryMap := map[string]models.Category{
		"movies":   models.CategoryMovies,
		"tv":       models.CategoryTV,
		"music":    models.CategoryMusic,
		"games":    models.CategoryGames,
		"apps":     models.CategorySoftware,
		"anime":    models.CategoryAnime,
		"books":    models.CategoryBooks,
		"xxx":      models.CategoryXXX,
	}

	for keyword, cat := range categoryMap {
		if strings.Contains(breadcrumb, keyword) {
			return cat
		}
	}

	// Check URL
	url := doc.Url.String()
	for keyword, cat := range categoryMap {
		if strings.Contains(url, "/"+keyword+"/") {
			return cat
		}
	}

	return models.CategoryOther
}

// parseDate parses date string to time.Time
func (p *Provider1337x) parseDate(dateStr string) string {
	// 1337x uses formats like "Today", "Yesterday", "2023-10-15", "Oct 15, 2023"
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return ""
	}

	now := time.Now()
	switch strings.ToLower(dateStr) {
	case "today":
		return now.Format(time.RFC3339)
	case "yesterday":
		return now.AddDate(0, 0, -1).Format(time.RFC3339)
	}

	// Try parsing various formats
	formats := []string{
		"Jan 2, 2006",
		"January 2, 2006",
		"2006-01-02",
		"02/01/2006",
		"02-01-2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format(time.RFC3339)
		}
	}

	return dateStr // Return as-is if can't parse
}
