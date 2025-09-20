package core

import "errors"

var (
	// Operational/errors for control flow.
	ErrNotFound    = errors.New("not found")
	ErrConflict    = errors.New("code already exists")
	ErrExpired     = errors.New("link expired")
	ErrInvalidURL  = errors.New("invalid url")
	ErrInvalidCode = errors.New("invalid code")
	ErrRateLimited = errors.New("rate limited")
)

// IsNotFound reports whether err is a not-found condition.
func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

// IsConflict reports whether err indicates a uniqueness conflict.
func IsConflict(err error) bool { return errors.Is(err, ErrConflict) }

// IsExpired reports whether err indicates an expired resource.
func IsExpired(err error) bool { return errors.Is(err, ErrExpired) }

