package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"urlshorty/internal/core"

	_ "modernc.org/sqlite" // pure-Go SQLite driver (no CGO)
)

// Store implements core.Store backed by SQLite.
type Store struct {
	db *sql.DB
}

// Open opens (or creates) the SQLite DB at path and applies migrations.
func Open(path string) (*Store, error) {
	// For modernc.org/sqlite, the DSN can be a simple file path.
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// Conservative pool settings for SQLite.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	// Pragmas to improve concurrency & reliability.
	_, _ = db.Exec("PRAGMA busy_timeout = 5000;")
	_, _ = db.Exec("PRAGMA journal_mode = WAL;")
	_, _ = db.Exec("PRAGMA foreign_keys = ON;")

	if err := applyMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close releases the underlying DB.
func (s *Store) Close() error { return s.db.Close() }

// Create inserts a new URL record. Returns core.ErrConflict if code already exists.
func (s *Store) Create(ctx context.Context, u *core.URL) error {
	const q = `
INSERT INTO urls(code, long_url, created_at, expires_at, hits)
VALUES (?, ?, ?, ?, 0);`
	var exp interface{}
	if u.ExpiresAt != nil {
		exp = u.ExpiresAt.UTC()
	} else {
		exp = nil
	}
	_, err := s.db.ExecContext(ctx, q, u.Code, u.LongURL, u.CreatedAt.UTC(), exp)
	if err != nil {
		// Map unique violations to ErrConflict (driver-specific error codes vary,
		// so we conservatively detect by message to keep deps minimal).
		if strings.Contains(strings.ToLower(err.Error()), "unique constraint failed") ||
			strings.Contains(strings.ToLower(err.Error()), "constraint failed") ||
			strings.Contains(strings.ToLower(err.Error()), "unique") {
			return core.ErrConflict
		}
		return err
	}
	return nil
}

// FindByCode returns a URL record for the given code (expired included).
func (s *Store) FindByCode(ctx context.Context, code string) (*core.URL, error) {
	const q = `
SELECT id, code, long_url, created_at, expires_at, hits
FROM urls
WHERE code = ?
LIMIT 1;`
	row := s.db.QueryRowContext(ctx, q, code)

	var rec core.URL
	var created time.Time
	var expires sql.NullTime

	if err := row.Scan(&rec.ID, &rec.Code, &rec.LongURL, &created, &expires, &rec.Hits); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	rec.CreatedAt = created.UTC()
	if expires.Valid {
		t := expires.Time.UTC()
		rec.ExpiresAt = &t
	} else {
		rec.ExpiresAt = nil
	}
	return &rec, nil
}

// IncrementHits increases the hits counter for code.
// If the code doesn't exist, return ErrNotFound so the caller can log it.
func (s *Store) IncrementHits(ctx context.Context, code string) error {
	const q = `UPDATE urls SET hits = hits + 1 WHERE code = ?;`
	res, err := s.db.ExecContext(ctx, q, code)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound
	}
	return nil
}

// PurgeExpired deletes expired links and returns deleted row count.
func (s *Store) PurgeExpired(ctx context.Context, now time.Time) (int64, error) {
	const q = `
DELETE FROM urls
WHERE expires_at IS NOT NULL AND expires_at <= ?;`
	res, err := s.db.ExecContext(ctx, q, now.UTC())
	if err != nil {
		return 0, err
	}
	affected, _ := res.RowsAffected()
	return affected, nil
}

// Compile-time check: *Store implements core.Store.
var _ core.Store = (*Store)(nil)
