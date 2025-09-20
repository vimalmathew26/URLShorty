package core

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	maxURLLength    = 2048
	minAliasLength  = 3
	maxAliasLength  = 64
	generateRetries = 6
)

var aliasRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// Service implements the business logic for creating and resolving short URLs.
type Service struct {
	store   Store
	gen     CodeGenerator
	nowFunc func() time.Time
}

func NewService(store Store, gen CodeGenerator) *Service {
	return &Service{
		store:   store,
		gen:     gen,
		nowFunc: time.Now,
	}
}

// Shorten validates input, optionally accepts a custom alias, or generates one.
// It returns the created record (without guaranteeing Hits is updated concurrently).
func (s *Service) Shorten(ctx context.Context, in CreateRequest) (*URL, error) {
	longURL, err := normalizeAndValidateURL(in.URL)
	if err != nil {
		return nil, ErrInvalidURL
	}
	if in.ExpiresAt != nil && in.ExpiresAt.Before(s.nowFunc()) {
		// Past expiry is not allowed.
		return nil, ErrInvalidURL
	}

	var code string
	if strings.TrimSpace(in.Custom) != "" {
		if !validAlias(in.Custom) {
			return nil, ErrInvalidCode
		}
		code = in.Custom
		// Single attempt for custom alias; surface conflict back to caller.
		rec := &URL{
			Code:      code,
			LongURL:   longURL,
			CreatedAt: s.nowFunc(),
			ExpiresAt: in.ExpiresAt,
			Hits:      0,
		}
		if err := s.store.Create(ctx, rec); err != nil {
			if IsConflict(err) {
				return nil, ErrConflict
			}
			return nil, err
		}
		return rec, nil
	}

	// Auto-generate codes; retry on rare collisions.
	for i := 0; i < generateRetries; i++ {
		code, err = s.gen.NewCode(ctx)
		if err != nil {
			return nil, err
		}
		if !validAlias(code) {
			// Defensive: if generator returns something invalid, retry.
			continue
		}
		rec := &URL{
			Code:      code,
			LongURL:   longURL,
			CreatedAt: s.nowFunc(),
			ExpiresAt: in.ExpiresAt,
			Hits:      0,
		}
		err = s.store.Create(ctx, rec)
		if err == nil {
			return rec, nil
		}
		if !IsConflict(err) {
			return nil, err
		}
	}
	// Extremely unlikely after multiple retries.
	return nil, ErrConflict
}

// Resolve returns the destination URL for a code if it exists and is not expired.
func (s *Service) Resolve(ctx context.Context, code string) (*URL, error) {
	if !validAlias(code) {
		return nil, ErrInvalidCode
	}
	rec, err := s.store.FindByCode(ctx, code)
	if err != nil {
		if IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	// Pass the function, not its result.
	if isExpired(rec, s.nowFunc) {
		return nil, ErrExpired
	}
	return rec, nil
}

// Metadata returns the record whether or not it is expired.
// Callers can decide how to present expiry status.
func (s *Service) Metadata(ctx context.Context, code string) (*URL, error) {
	if !validAlias(code) {
		return nil, ErrInvalidCode
	}
	rec, err := s.store.FindByCode(ctx, code)
	if err != nil {
		if IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return rec, nil
}

// RecordHit increments the hits counter; failures are returned for caller to log.
// Handlers may call this in a goroutine for best-effort accounting.
func (s *Service) RecordHit(ctx context.Context, code string) error {
	if !validAlias(code) {
		return ErrInvalidCode
	}
	return s.store.IncrementHits(ctx, code)
}

// CleanupExpired purges expired links and returns the number of rows affected.
func (s *Service) CleanupExpired(ctx context.Context) (int64, error) {
	return s.store.PurgeExpired(ctx, s.nowFunc())
}

// ---- helpers ----

func validAlias(a string) bool {
	if len(a) < minAliasLength || len(a) > maxAliasLength {
		return false
	}
	return aliasRe.MatchString(a)
}

func isExpired(u *URL, now func() time.Time) bool {
	if u == nil || u.ExpiresAt == nil {
		return false
	}
	return now().After(*u.ExpiresAt)
}

func normalizeAndValidateURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > maxURLLength {
		return "", ErrInvalidURL
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", ErrInvalidURL
	}
	// Require explicit http/https scheme and non-empty host.
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", ErrInvalidURL
	}
	if parsed.Host == "" {
		return "", ErrInvalidURL
	}
	// Normalize: strip default ports and unnecessary whitespace (already trimmed).
	// Keep as-is otherwise to avoid surprising the user.
	return parsed.String(), nil
}
