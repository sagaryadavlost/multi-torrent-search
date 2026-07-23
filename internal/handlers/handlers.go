package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"torrentsearch-web/internal/models"
	"torrentsearch-web/internal/providers"
)

// Handler holds the HTTP handlers
type Handler struct {
	registry *providers.ProviderRegistry
}

// NewHandler creates a new handler
func NewHandler(registry *providers.ProviderRegistry) *Handler {
	return &Handler{registry: registry}
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query    string         `json:"query"`
	Category models.Category `json:"category"`
	Page     int            `json:"page"`
	Provider string         `json:"provider,omitempty"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Results    map[string][]models.Torrent `json:"results"`
	Query      string                     `json:"query"`
	Category   models.Category             `json:"category"`
	Page       int                        `json:"page"`
	TotalPages int                        `json:"total_pages"`
	Providers  []ProviderInfo             `json:"providers"`
	QueryTime  float64                    `json:"query_time_ms"`
}

// ProviderInfo represents provider information
type ProviderInfo struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Icon        string   `json:"icon"`
	Categories  []string `json:"categories"`
	Enabled     bool     `json:"enabled"`
}

// DetailsRequest represents a details request
type DetailsRequest struct {
	Provider   string `json:"provider"`
	DetailURL  string `json:"detail_url"`
	InfoHash   string `json:"info_hash,omitempty"`
}

// Search handles torrent search requests
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse query parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		query = r.URL.Query().Get("query")
	}

	categoryStr := r.URL.Query().Get("category")
	category := models.Category(categoryStr)
	if category == "" {
		category = models.CategoryAll
	}

	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	provider := r.URL.Query().Get("provider")

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var results map[string][]models.Torrent
	var err error

	if provider != "" {
		// Single provider search
		results = make(map[string][]models.Torrent)
		torrents, err := h.registry.SearchSingleProvider(ctx, provider, query, category, page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		results[provider] = torrents
	} else {
		// Multi-provider search
		results, err = h.registry.Search(ctx, query, category, page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	queryTime := float64(time.Since(startTime).Milliseconds())

	// Get provider info
	providerInfos := h.getProviderInfos(category)

	response := SearchResponse{
		Results:    results,
		Query:      query,
		Category:   category,
		Page:       page,
		TotalPages: 50, // Default max pages
		Providers:  providerInfos,
		QueryTime:  queryTime,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SearchPOST handles POST search requests
func (h *Handler) SearchPOST(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	startTime := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var results map[string][]models.Torrent
	var err error

	if req.Provider != "" {
		results = make(map[string][]models.Torrent)
		torrents, err := h.registry.SearchSingleProvider(ctx, req.Provider, req.Query, req.Category, req.Page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		results[req.Provider] = torrents
	} else {
		results, err = h.registry.Search(ctx, req.Query, req.Category, req.Page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	queryTime := float64(time.Since(startTime).Milliseconds())

	providerInfos := h.getProviderInfos(req.Category)

	response := SearchResponse{
		Results:    results,
		Query:      req.Query,
		Category:   req.Category,
		Page:       req.Page,
		TotalPages: 50,
		Providers:  providerInfos,
		QueryTime:  queryTime,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Details handles torrent details requests
func (h *Handler) Details(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	detailURL := r.URL.Query().Get("url")
	infoHash := r.URL.Query().Get("info_hash")

	if provider == "" {
		http.Error(w, "Provider is required", http.StatusBadRequest)
		return
	}

	if detailURL == "" && infoHash == "" {
		http.Error(w, "Detail URL or info hash is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var details *models.TorrentDetails
	var err error

	if detailURL != "" {
		details, err = h.registry.GetDetails(ctx, provider, detailURL)
	} else {
		// Try to get details by info hash from a provider
		p, ok := h.registry.GetProvider(provider)
		if !ok {
			http.Error(w, "Provider not found", http.StatusNotFound)
			return
		}
		details, err = p.GetDetails(ctx, infoHash)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if details == nil {
		http.Error(w, "Details not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

// Providers returns list of all providers
func (h *Handler) Providers(w http.ResponseWriter, r *http.Request) {
	providers := h.registry.GetProviders()

	var providerInfos []ProviderInfo
	for _, p := range providers {
		categories := make([]string, len(p.GetCategories()))
		for i, c := range p.GetCategories() {
			categories[i] = string(c)
		}

		providerInfos = append(providerInfos, ProviderInfo{
			Name:        p.GetName(),
			DisplayName: p.GetDisplayName(),
			Icon:        p.GetIconURL(),
			Categories:  categories,
			Enabled:     p.GetEnabled(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"providers": providerInfos,
		"count":     len(providerInfos),
	})
}

// Categories returns list of all categories
func (h *Handler) Categories(w http.ResponseWriter, r *http.Request) {
	categories := []models.Category{
		models.CategoryAll,
		models.CategoryMovies,
		models.CategoryTV,
		models.CategoryMusic,
		models.CategoryGames,
		models.CategorySoftware,
		models.CategoryAnime,
		models.CategoryBooks,
		models.CategoryXXX,
		models.CategoryOther,
	}

	categoryInfos := make([]map[string]string, len(categories))
	for i, cat := range categories {
		categoryInfos[i] = map[string]string{
			"value":       string(cat),
			"displayName": cat.DisplayName(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"categories": categoryInfos,
	})
}

// Latest returns latest torrents
func (h *Handler) Latest(w http.ResponseWriter, r *http.Request) {
	categoryStr := r.URL.Query().Get("category")
	category := models.Category(categoryStr)
	if category == "" {
		category = models.CategoryAll
	}

	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	results, err := h.registry.GetLatestTorrents(ctx, category, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
		"category": category,
		"page": page,
	})
}

// Top returns top torrents
func (h *Handler) Top(w http.ResponseWriter, r *http.Request) {
	categoryStr := r.URL.Query().Get("category")
	category := models.Category(categoryStr)
	if category == "" {
		category = models.CategoryAll
	}

	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	results, err := h.registry.GetTopTorrents(ctx, category, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
		"category": category,
		"page": page,
	})
}

// Health returns health status
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// Stats returns statistics about providers
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	providers := h.registry.GetProviders()
	enabledCount := 0
	disabledCount := 0

	for _, p := range providers {
		if p.GetEnabled() {
			enabledCount++
		} else {
			disabledCount++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_providers":  len(providers),
		"enabled_providers": enabledCount,
		"disabled_providers": disabledCount,
	})
}

// getProviderInfos returns provider info for a category
func (h *Handler) getProviderInfos(category models.Category) []ProviderInfo {
	providers := h.registry.GetProviders()
	var providerInfos []ProviderInfo

	for _, p := range providers {
		if !p.GetEnabled() {
			continue
		}

		if category != models.CategoryAll && !p.SupportsCategory(category) {
			continue
		}

		categories := make([]string, len(p.GetCategories()))
		for i, c := range p.GetCategories() {
			categories[i] = string(c)
		}

		providerInfos = append(providerInfos, ProviderInfo{
			Name:        p.GetName(),
			DisplayName: p.GetDisplayName(),
			Icon:        p.GetIconURL(),
			Categories:  categories,
			Enabled:     p.GetEnabled(),
		})
	}

	return providerInfos
}

