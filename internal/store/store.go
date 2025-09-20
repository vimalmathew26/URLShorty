package store

import (
	"context"

	"urlshorty/internal/core"
)

// Store defines the interface for URL storage operations
type Store interface {
	// Create creates a new URL record
	Create(ctx context.Context, url *core.URL) error

	// GetByCode retrieves a URL by its short code
	GetByCode(ctx context.Context, code string) (*core.URL, error)

	// GetByOriginalURL retrieves a URL by its original URL
	GetByOriginalURL(ctx context.Context, originalURL string) (*core.URL, error)

	// IncrementClickCount increments the click count for a URL
	IncrementClickCount(ctx context.Context, code string) error

	// CodeExists checks if a code already exists
	CodeExists(ctx context.Context, code string) (bool, error)

	// Close closes the store connection
	Close() error
}

