package providers

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"torrentsearch-web/internal/models"
)

// BaseProvider provides common functionality for all providers
type BaseProvider struct {
	Name        string
	BaseURLStr  string
	Enabled     bool
	DefaultIcon string
	DisplayName string
	IconURL     string
	UserAgent   string
	Timeout     time.Duration
	Categories  []models.Category
	Client      *http.Client
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(name, baseURL, icon string, enabled bool, userAgent string, timeout time.Duration, categories []models.Category) *BaseProvider {
	return &BaseProvider{
		Name:        name,
		BaseURLStr:  baseURL,
		Enabled:     enabled,
		DefaultIcon: icon,
		DisplayName: name,
		IconURL:     icon,
		UserAgent:   userAgent,
		Timeout:     timeout,
		Categories:  categories,
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetName returns the provider name
func (p *BaseProvider) GetName() string {
	return p.Name
}

// GetBaseURL returns the base URL
func (p *BaseProvider) GetBaseURL() string {
	return p.BaseURLStr
}

// GetEnabled returns whether provider is enabled
func (p *BaseProvider) GetEnabled() bool {
	return p.Enabled
}

// SetEnabled sets the enabled state
func (p *BaseProvider) SetEnabled(enabled bool) {
	p.Enabled = enabled
}

// GetSupportedCategories returns supported categories
func (p *BaseProvider) GetSupportedCategories() []models.Category {
	return p.Categories
}

// SupportsCategory checks if category is supported
func (p *BaseProvider) SupportsCategory(category models.Category) bool {
	for _, c := range p.Categories {
		if c == category {
			return true
		}
	}
	return false
}

// GetDetails gets torrent details (to be implemented by specific providers)
func (p *BaseProvider) GetDetails(infoHash string) (*models.TorrentDetails, error) {
	return nil, fmt.Errorf("GetDetails not implemented for %s", p.Name)
}

// Search searches for torrents (to be implemented by specific providers)
func (p *BaseProvider) Search(query string, category models.Category, page int) ([]models.Torrent, error) {
	return nil, fmt.Errorf("Search not implemented for %s", p.Name)
}

// BuildSearchURL builds the search URL
func (p *BaseProvider) BuildSearchURL(query string, category models.Category, page int) string {
	return p.BaseURLStr
}

// FetchHTML fetches HTML content from a URL
func (p *BaseProvider) FetchHTML(ctx context.Context, urlStr string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", p.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

// FetchJSON fetches JSON content from a URL
func (p *BaseProvider) FetchJSON(ctx context.Context, urlStr string, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", p.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := p.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return nil // JSON parsing would go here
}

// ExtractMagnet extracts magnet link from text/HTML
func (p *BaseProvider) ExtractMagnet(text string) string {
	magnetRegex := regexp.MustCompile(`magnet:\?xt=urn:btih:[a-zA-Z0-9]{32,40}[^"\s<>]*`)
	matches := magnetRegex.FindStringSubmatch(text)
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

// ExtractInfoHash extracts info hash from magnet or URL
func (p *BaseProvider) ExtractInfoHash(text string) string {
	// From magnet
	magnetRegex := regexp.MustCompile(`magnet:\?xt=urn:btih:([a-zA-Z0-9]{32,40})`)
	matches := magnetRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.ToUpper(matches[1])
	}

	// From URL patterns
	urlRegex := regexp.MustCompile(`[/=]([a-fA-F0-9]{40})[/=&]`)
	matches = urlRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.ToUpper(matches[1])
	}

	// 32 char base32
	base32Regex := regexp.MustCompile(`[/=]([A-Z2-7]{32})[/=&]`)
	matches = base32Regex.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// ParseSize parses size string to bytes
func (p *BaseProvider) ParseSize(sizeStr string) int64 {
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))
	if sizeStr == "" || sizeStr == "-" {
		return 0
	}

	re := regexp.MustCompile(`^([\d.]+)\s*([KMGT]?B?)`)
	matches := re.FindStringSubmatch(sizeStr)
	if len(matches) < 3 {
		return 0
	}

	size := 0.0
	fmt.Sscanf(matches[1], "%f", &size)
	unit := matches[2]

	multipliers := map[string]int64{
		"B":   1,
		"KB":  1024,
		"MB":  1024 * 1024,
		"GB":  1024 * 1024 * 1024,
		"TB":  1024 * 1024 * 1024 * 1024,
		"K":   1024,
		"M":   1024 * 1024,
		"G":   1024 * 1024 * 1024,
		"T":   1024 * 1024 * 1024 * 1024,
		"KIB": 1024,
		"MIB": 1024 * 1024,
		"GIB": 1024 * 1024 * 1024,
		"TIB": 1024 * 1024 * 1024 * 1024,
	}

	if mult, ok := multipliers[unit]; ok {
		return int64(size * float64(mult))
	}

	return int64(size)
}

// FormatSize formats bytes to human readable string
func (p *BaseProvider) FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CleanHTML cleans HTML text
func (p *BaseProvider) CleanHTML(htmlStr string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(htmlStr, "")
	// Decode HTML entities using standard library
	text = html.UnescapeString(text)
	// Clean whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// GetCategoryMapping returns category mapping for provider
func (p *BaseProvider) GetCategoryMapping() map[models.Category]string {
	return map[models.Category]string{
		models.CategoryMovies:   "movies",
		models.CategoryTV:       "tv",
		models.CategoryMusic:    "music",
		models.CategoryGames:    "games",
		models.CategorySoftware: "software",
		models.CategoryAnime:    "anime",
		models.CategoryBooks:    "books",
		models.CategoryXXX:      "xxx",
	}
}

// ResolveURL resolves relative URL to absolute
func (p *BaseProvider) ResolveURL(relativeURL string) string {
	if strings.HasPrefix(relativeURL, "http") {
		return relativeURL
	}
	base, _ := url.Parse(p.BaseURLStr)
	rel, _ := url.Parse(relativeURL)
	return base.ResolveReference(rel).String()
}

// GetDisplayName returns the display name of the provider
func (p *BaseProvider) GetDisplayName() string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	return p.Name
}

// GetIconURL returns the icon URL of the provider
func (p *BaseProvider) GetIconURL() string {
	if p.IconURL != "" {
		return p.IconURL
	}
	return p.DefaultIcon
}

// GetCategories returns the categories (alias for GetSupportedCategories)
func (p *BaseProvider) GetCategories() []models.Category {
	return p.Categories
}

// parseInt safely parses an integer from a string
func parseInt(s string) int {
	s = regexp.MustCompile(`[^\d]`).ReplaceAllString(s, "")
	if s == "" {
		return 0
	}
	var result int
	for _, c := range s {
		result = result*10 + int(c-'0')
	}
	return result
}

