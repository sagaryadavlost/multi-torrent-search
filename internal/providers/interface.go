package providers

import (
	"context"

	"torrentsearch-web/internal/models"
)

// Provider defines the interface that all torrent providers must implement
type Provider interface {
	GetName() string
	GetBaseURL() string
	GetEnabled() bool
	SetEnabled(bool)
	GetSupportedCategories() []models.Category
	GetCategories() []models.Category
	GetDisplayName() string
	GetIconURL() string
	SupportsCategory(category models.Category) bool
	Search(ctx context.Context, query string, category models.Category, page int) ([]models.Torrent, error)
	GetDetails(ctx context.Context, infoHash string) (*models.TorrentDetails, error)
	GetDetailsFromURL(ctx context.Context, detailURL string) (*models.TorrentDetails, error)
}
