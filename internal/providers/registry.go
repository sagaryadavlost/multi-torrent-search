package providers

import (
	"context"
	"sort"
	"sync"
	"time"

	"torrentsearch-web/internal/models"
)

// ProviderRegistry manages all torrent providers
type ProviderRegistry struct {
	providers []Provider
	mu        sync.RWMutex
	config    *ProviderConfig
}

// ProviderConfig holds configuration for all providers
type ProviderConfig struct {
	UserAgent string
	Timeout   int
	Enabled   map[string]bool
	ProxyURL  string
}

// NewProviderRegistry creates a new provider registry with all built-in providers
func NewProviderRegistry(config *ProviderConfig) *ProviderRegistry {
	if config == nil {
		config = &ProviderConfig{
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			Timeout:   30,
			Enabled: map[string]bool{
				"nyaa":          true,
				"yts":           true,
				"1337x":         true,
				"thepiratebay":  true,
				"limetorrents":  true,
				"eztv":          true,
				"btdigg":        true,
				"torrent9":      true,
				"rutor":         true,
				"tokyotoshokan": true,
				"anirena":       true,
				"animetosho":    true,
				"bangumimoe":    true,
				"mikan":         true,
				"animelibria":   true,
				"sukebei":       true,
				"nyaasi":        true,
				"bittorrentdb":  true,
				"bitsearch":     true,
				"kickasstorrents": true,
				"torrentkitty":  true,
				"torrentdownloads": true,
				"rarbg":         true,
				"megaserch":     true,
			},
		}
	}

	registry := &ProviderRegistry{
		providers: make([]Provider, 0),
		config:    config,
	}

	// Register all providers
	registry.registerBuiltinProviders()

	return registry
}

// registerBuiltinProviders registers all built-in providers
func (r *ProviderRegistry) registerBuiltinProviders() {
	providers := map[string]Provider{
		"nyaa":           NewNyaaProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["nyaa"]),
		"yts":            NewYTSProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["yts"]),
		"1337x":          NewProvider1337x(r.config.UserAgent, r.config.Timeout, r.config.Enabled["1337x"]),
		"thepiratebay":   NewThePirateBayProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["thepiratebay"]),
		"limetorrents":   NewLimeTorrentsProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["limetorrents"]),
		"eztv":           NewEZTVProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["eztv"]),
		"btdigg":         NewBTDiggProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["btdigg"]),
		"torrent9":       NewTorrent9Provider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["torrent9"]),
		"rutor":          NewRutorProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["rutor"]),
		"tokyotoshokan":  NewTokyoToshokanProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["tokyotoshokan"]),
		"anirena":        NewAniRenaProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["anirena"]),
		"animetosho":     NewAnimeToshoProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["animetosho"]),
		"bangumimoe":     NewBangumiMoeProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["bangumimoe"]),
		"mikan":          NewMikanProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["mikan"]),
		"animelibria":    NewAniLibriaProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["animelibria"]),
		"sukebei":        NewSukebeiProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["sukebei"]),
		"nyaasi":         NewNyaaSiProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["nyaasi"]),
		"bittorrentdb":   NewBitTorrentDBProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["bittorrentdb"]),
		"bitsearch":      NewBitSearchProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["bitsearch"]),
		"kickasstorrents": NewKickassTorrentsProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["kickasstorrents"]),
		"torrentkitty":   NewTorrentKittyProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["torrentkitty"]),
		"torrentdownloads": NewTorrentDownloadsProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["torrentdownloads"]),
		"rarbg":          NewRARBGProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["rarbg"]),
		"megaserch":      NewMegaSearchProvider(r.config.UserAgent, r.config.Timeout, r.config.Enabled["megaserch"]),
	}

	// Add enabled providers to registry
	for name, provider := range providers {
		if r.config.Enabled[name] {
			r.providers = append(r.providers, provider)
		}
	}

	// Sort by name for consistent ordering
	sort.Slice(r.providers, func(i, j int) bool {
		return r.providers[i].GetName() < r.providers[j].GetName()
	})
}

// GetProviders returns all registered providers
func (r *ProviderRegistry) GetProviders() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]Provider, len(r.providers))
	copy(providers, r.providers)
	return providers
}

// GetEnabledProviders returns only enabled providers
func (r *ProviderRegistry) GetEnabledProviders() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var enabled []Provider
	for _, p := range r.providers {
		if p.GetEnabled() {
			enabled = append(enabled, p)
		}
	}
	return enabled
}

// GetProvider returns a provider by name
func (r *ProviderRegistry) GetProvider(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.providers {
		if p.GetName() == name {
			return p, true
		}
	}
	return nil, false
}

// EnableProvider enables a provider
func (r *ProviderRegistry) EnableProvider(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.providers {
		if p.GetName() == name {
			p.SetEnabled(true)
			r.config.Enabled[name] = true
			return true
		}
	}
	return false
}

// DisableProvider disables a provider
func (r *ProviderRegistry) DisableProvider(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.providers {
		if p.GetName() == name {
			p.SetEnabled(false)
			r.config.Enabled[name] = false
			return true
		}
	}
	return false
}

// GetProvidersByCategory returns providers that support a specific category
func (r *ProviderRegistry) GetProvidersByCategory(category models.Category) []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Provider
	for _, p := range r.providers {
		if p.GetEnabled() && p.SupportsCategory(category) {
			result = append(result, p)
		}
	}
	return result
}

// Search searches all enabled providers that support the category
func (r *ProviderRegistry) Search(ctx context.Context, query string, category models.Category, page int) (map[string][]models.Torrent, error) {
	providers := r.GetProvidersByCategory(category)
	if len(providers) == 0 {
		return nil, nil
	}

	// Create context with timeout
	searchCtx, cancel := context.WithTimeout(ctx, time.Duration(r.config.Timeout)*time.Second)
	defer cancel()

	results := make(map[string][]models.Torrent)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Limit concurrent searches
	sem := make(chan struct{}, 5)

	for _, provider := range providers {
		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			torrents, err := p.Search(searchCtx, query, category, page)
			if err != nil {
				// Log error but continue with other providers
				return
			}

			if len(torrents) > 0 {
				mu.Lock()
				results[p.GetName()] = torrents
				mu.Unlock()
			}
		}(provider)
	}

	wg.Wait()
	return results, nil
}

// SearchSingleProvider searches a single provider
func (r *ProviderRegistry) SearchSingleProvider(ctx context.Context, providerName, query string, category models.Category, page int) ([]models.Torrent, error) {
	provider, ok := r.GetProvider(providerName)
	if !ok {
		return nil, ErrProviderNotFound
	}

	if !provider.GetEnabled() {
		return nil, ErrProviderDisabled
	}

	if !provider.SupportsCategory(category) && category != models.CategoryAll {
		return nil, ErrCategoryNotSupported
	}

	return provider.Search(ctx, query, category, page)
}

// GetDetails gets details from a provider
func (r *ProviderRegistry) GetDetails(ctx context.Context, providerName, detailURL string) (*models.TorrentDetails, error) {
	provider, ok := r.GetProvider(providerName)
	if !ok {
		return nil, ErrProviderNotFound
	}

	if !provider.GetEnabled() {
		return nil, ErrProviderDisabled
	}

	// Check if provider supports details fetching
	if detailsProvider, ok := provider.(interface{ GetDetailsFromURL(context.Context, string) (*models.TorrentDetails, error) }); ok {
		return detailsProvider.GetDetailsFromURL(ctx, detailURL)
	}

	return nil, ErrDetailsNotSupported
}

// GetLatestTorrents gets latest torrents from all providers
func (r *ProviderRegistry) GetLatestTorrents(ctx context.Context, category models.Category, page int) (map[string][]models.Torrent, error) {
	providers := r.GetEnabledProviders()
	if len(providers) == 0 {
		return nil, nil
	}

	results := make(map[string][]models.Torrent)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, provider := range providers {
		// Check if provider supports latest torrents
		if latestProvider, ok := provider.(interface{ GetLatestTorrents(context.Context, models.Category, int) ([]models.Torrent, error) }); ok {
			wg.Add(1)
			go func(p Provider) {
				defer wg.Done()
				torrents, err := latestProvider.GetLatestTorrents(ctx, category, page)
				if err != nil || len(torrents) == 0 {
					return
				}
				mu.Lock()
				results[p.GetName()] = torrents
				mu.Unlock()
			}(provider)
		}
	}

	wg.Wait()
	return results, nil
}

// GetTopTorrents gets top torrents from all providers
func (r *ProviderRegistry) GetTopTorrents(ctx context.Context, category models.Category, page int) (map[string][]models.Torrent, error) {
	providers := r.GetEnabledProviders()
	if len(providers) == 0 {
		return nil, nil
	}

	results := make(map[string][]models.Torrent)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, provider := range providers {
		if topProvider, ok := provider.(interface{ GetTopTorrents(context.Context, models.Category, int) ([]models.Torrent, error) }); ok {
			wg.Add(1)
			go func(p Provider) {
				defer wg.Done()
				torrents, err := topProvider.GetTopTorrents(ctx, category, page)
				if err != nil || len(torrents) == 0 {
					return
				}
				mu.Lock()
				results[p.GetName()] = torrents
				mu.Unlock()
			}(provider)
		}
	}

	wg.Wait()
	return results, nil
}

// Provider error types
var (
	ErrProviderNotFound     = &ProviderError{Code: "PROVIDER_NOT_FOUND", Message: "Provider not found"}
	ErrProviderDisabled     = &ProviderError{Code: "PROVIDER_DISABLED", Message: "Provider is disabled"}
	ErrCategoryNotSupported = &ProviderError{Code: "CATEGORY_NOT_SUPPORTED", Message: "Category not supported by provider"}
	ErrDetailsNotSupported  = &ProviderError{Code: "DETAILS_NOT_SUPPORTED", Message: "Provider does not support details fetching"}
)

type ProviderError struct {
	Code    string
	Message string
}

func (e *ProviderError) Error() string {
	return e.Message
}