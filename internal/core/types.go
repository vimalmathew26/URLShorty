package core

import (
	"context"
	"time"
)

// URL represents a shortened link record.
type URL struct {
	ID        int64      `json:"id"`
	Code      string     `json:"code"`
	LongURL   string     `json:"url"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Hits      int64      `json:"hits"`
}

// CreateRequest is the input to create/shorten a URL.
type CreateRequest struct {
	URL       string     `json:"url"`
	Custom    string     `json:"custom,omitempty"`     // Optional custom alias
	ExpiresAt *time.Time `json:"expires_at,omitempty"` // Optional UTC expiry
}

// Store abstracts persistence for URL records.
type Store interface {
	// Create inserts a new record. Must fail with ErrConflict if code is taken.
	Create(ctx context.Context, u *URL) error
	// FindByCode returns the record for a code (expired ones included).
	FindByCode(ctx context.Context, code string) (*URL, error)
	// IncrementHits increases the hits counter for a code (best-effort).
	IncrementHits(ctx context.Context, code string) error
	// PurgeExpired deletes or disables expired records and returns affected count.
	PurgeExpired(ctx context.Context, now time.Time) (int64, error)
}

// CodeGenerator creates collision-resistant short codes.
type CodeGenerator interface {
	NewCode(ctx context.Context) (string, error)
}
