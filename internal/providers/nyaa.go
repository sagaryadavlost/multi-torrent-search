package providers

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"torrentsearch-web/internal/models"
)

// NyaaProvider implements Nyaa.si torrent search
type NyaaProvider struct {
	*BaseProvider
	categoryMap map[models.Category]string
}

// NewNyaaProvider creates a new Nyaa provider
func NewNyaaProvider(userAgent string, timeout int, enabled bool) *NyaaProvider {
	base := NewBaseProvider(
		"nyaa",
		"https://nyaa.si",
		"https://nyaa.si/favicon.ico",
		enabled,
		userAgent,
		time.Duration(timeout)*time.Second,
		[]models.Category{
			models.CategoryAnime,
			models.CategorySoftware,
			models.CategoryGames,
			models.CategoryBooks,
			models.CategoryMusic,
			models.CategoryTV,
		},
	)

	return &NyaaProvider{
		BaseProvider: base,
		categoryMap: map[models.Category]string{
			models.CategoryAll:       "0_0",
			models.CategoryAnime:     "1_0",
			models.CategorySoftware:  "6_1",
			models.CategoryBooks:     "3_0",
			models.CategoryGames:     "6_2",
			models.CategoryMusic:     "2_0",
			models.CategoryTV:        "4_0",
		},
	}
}

// Search searches for torrents on Nyaa
func (p *NyaaProvider) Search(ctx context.Context, query string, category models.Category, page int) ([]models.Torrent, error) {
	if !p.Enabled {
		return nil, nil
	}

	categoryID := p.categoryMap[category]
	if categoryID == "" {
		categoryID = p.categoryMap[models.CategoryAll]
	}

	searchURL := fmt.Sprintf("%s/?f=0&c=%s&q=%s", p.BaseURLStr, categoryID, url.QueryEscape(query))
	if page > 1 {
		searchURL += fmt.Sprintf("&p=%d", page)
	}

	doc, err := p.FetchHTML(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	return p.parseResults(doc, searchURL), nil
}

// GetDetails gets detailed torrent information
func (p *NyaaProvider) GetDetails(ctx context.Context, infoHash string) (*models.TorrentDetails, error) {
	// Nyaa details page requires the full URL, not just info hash
	// This would need the detail page URL to fetch details
	return nil, fmt.Errorf("GetDetails not fully implemented for Nyaa without detail URL")
}

// parseResults parses search results from Nyaa
func (p *NyaaProvider) parseResults(doc *goquery.Document, pageURL string) []models.Torrent {
	var torrents []models.Torrent

	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		nameEl := s.Find("td:nth-child(2) a:not(.comments)")
		if nameEl.Length() == 0 {
			return
		}

		name := strings.TrimSpace(nameEl.Text())
		if name == "" {
			return
		}

		magnetLink, _ := s.Find("td:nth-child(3) a[href^='magnet:']").Attr("href")
		detailsLink, _ := nameEl.Attr("href")
		downloadLink, _ := s.Find("td:nth-child(3) a[href$='.torrent']").Attr("href")

		sizeText := strings.TrimSpace(s.Find("td:nth-child(4)").Text())
		dateText := strings.TrimSpace(s.Find("td:nth-child(5)").Text())
		seedersText := strings.TrimSpace(s.Find("td:nth-child(6)").Text())
		leechersText := strings.TrimSpace(s.Find("td:nth-child(7)").Text())

		// Parse upload date
		var uploadDate *time.Time
		if dateText != "" {
			// Try parsing timestamp from data-timestamp attribute
			if ts, _ := s.Find("td:nth-child(5)").Attr("data-timestamp"); ts != "" {
				if tsInt, err := strconv.ParseInt(ts, 10, 64); err == nil {
					t := time.Unix(tsInt, 0)
					uploadDate = &t
				}
			}
		}

		seeders := parseInt(seedersText)
		peers := parseInt(leechersText)

		infoHash := p.ExtractInfoHash(magnetLink)

		uploadDateStr := ""
		if uploadDate != nil {
			uploadDateStr = uploadDate.Format(time.RFC3339)
		}

		torrent := models.Torrent{
			InfoHash:           infoHash,
			Name:               name,
			Size:               sizeText,
			SizeBytes:          p.ParseSize(sizeText),
			Seeders:            seeders,
			Peers:              peers,
			ProviderName:       p.Name,
			UploadDate:         uploadDateStr,
			DescriptionPageURL: p.ResolveURL(detailsLink),
			MagnetURI:          magnetLink,
			FileDownloadLink:   p.ResolveURL(downloadLink),
			IsNSFW:             false,
			IsDead:             seeders == 0 && peers == 0,
		}

		// Try to determine category from the row
		catLink := s.Find("td:nth-child(1) a").First()
		if catLink.Length() > 0 {
			if href, ok := catLink.Attr("href"); ok {
				if catID := strings.TrimPrefix(href, "/?c="); catID != href {
					torrent.Category = p.categoryFromID(catID)
				}
			}
		}

		torrents = append(torrents, torrent)
	})

	return torrents
}

// categoryFromID converts category ID to Category
func (p *NyaaProvider) categoryFromID(id string) models.Category {
	reverseMap := map[string]models.Category{
		"1_0": models.CategoryAnime,
		"1_1": models.CategoryAnime, // Anime Music Video
		"1_2": models.CategoryAnime, // Anime Non-English
		"2_0": models.CategoryMusic,
		"3_0": models.CategoryBooks,
		"4_0": models.CategoryTV,
		"6_1": models.CategorySoftware,
		"6_2": models.CategoryGames,
	}
	return reverseMap[id]
}

// GetDetailsFromURL gets details from a details page URL
func (p *NyaaProvider) GetDetailsFromURL(ctx context.Context, detailURL string) (*models.TorrentDetails, error) {
	doc, err := p.FetchHTML(ctx, detailURL)
	if err != nil {
		return nil, err
	}
	return p.parseDetails(doc, detailURL)
}

// parseDetails parses torrent details from detail page
func (p *NyaaProvider) parseDetails(doc *goquery.Document, pageURL string) (*models.TorrentDetails, error) {
	name := strings.TrimSpace(doc.Find("h3.panel-title").Text())
	if name == "" {
		name = strings.TrimSpace(doc.Find(".panel-heading h3").Text())
	}

	// Extract info hash from magnet link
	magnetLink, _ := doc.Find("a[href^='magnet:']").Attr("href")
	infoHash := p.ExtractInfoHash(magnetLink)

	// Parse details table
	var size string
	var sizeBytes int64
	var uploadDate *time.Time
	var seeders, peers int

	doc.Find(".panel-body .row").Each(func(i int, s *goquery.Selection) {
		label := strings.TrimSpace(s.Find(".col-md-3").Text())
		value := strings.TrimSpace(s.Find(".col-md-9").Text())

		switch label {
		case "File size:":
			size = value
			sizeBytes = p.ParseSize(value)
		case "Date:":
			// Parse date
			if t, err := time.Parse("2006-01-02 15:04", value); err == nil {
				uploadDate = &t
			}
		case "Seeders:":
			seeders = parseInt(value)
		case "Leechers:":
			peers = parseInt(value)
		}
	})

	// Get download link
	downloadLink, _ := doc.Find("a[href$='.torrent']").Attr("href")

	// Get poster/image
	posterURL, _ := doc.Find(".panel-body img").Attr("src")

	uploadDateStr := ""
	if uploadDate != nil {
		uploadDateStr = uploadDate.Format(time.RFC3339)
	}

	return &models.TorrentDetails{
		InfoHash:           infoHash,
		Name:               name,
		Size:               size,
		SizeBytes:          sizeBytes,
		Seeders:            seeders,
		Peers:              peers,
		UploadDate:         uploadDateStr,
		Category:           models.CategoryAnime,
		MagnetURI:          magnetLink,
		FileDownloadLink:   p.ResolveURL(downloadLink),
		PosterURL:          p.ResolveURL(posterURL),
		DescriptionPageURL: pageURL,
	}, nil
}

