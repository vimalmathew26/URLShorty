package sqlite

import (
	"database/sql"
)

// applyMigrations runs schema initialization for the SQLite database.
// We keep it embedded (no external migration tool needed for the 1-day build).
func applyMigrations(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	return err
}

const schemaSQL = `
CREATE TABLE IF NOT EXISTS urls (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  code       TEXT    NOT NULL UNIQUE,
  long_url   TEXT    NOT NULL,
  created_at TIMESTAMP NOT NULL,
  expires_at TIMESTAMP NULL,
  hits       INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_urls_expires_at ON urls(expires_at);
`
