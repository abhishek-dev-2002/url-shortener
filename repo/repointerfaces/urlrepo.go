package repointerfaces

import (
	"context"

	"github.com/abhishekmaurya/url-shortner/models"
)

// URLStore defines the contract for URL persistence operations.
type URLStore interface {
	// CreateURL stores a new URL mapping.
	CreateURL(ctx context.Context, url *models.URL) (*models.URL, error)

	// GetByShortCode retrieves a URL by its short code. Returns nil, nil if not found.
	GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error)

	// GetByOriginalURL retrieves the first non-alias URL by original URL. Returns nil, nil if not found.
	GetByOriginalURL(ctx context.Context, originalURL string) (*models.URL, error)

	// ShortCodeExists checks if a short code already exists in the database.
	ShortCodeExists(ctx context.Context, shortCode string) (bool, error)

	// IncrementClickCount increments the click count for a short code.
	IncrementClickCount(ctx context.Context, shortCode string) error

	// AllocateIDBlock atomically allocates a block of sequential IDs.
	// Returns the start of the block. The caller owns IDs [start, start+blockSize).
	AllocateIDBlock(ctx context.Context, blockSize int64) (int64, error)
}
