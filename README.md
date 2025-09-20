# urlshorty — URL shortener in Go (Gin + SQLite)

A compact, single-binary URL shortener built with Go, Gin, and SQLite.
Runs locally without Docker or external databases. Provides a minimal HTML page, a clean JSON API, rate limiting, and automated tests.

---

## 1. Prerequisites

Install the following:

* Go 1.22+ — [https://go.dev/dl/](https://go.dev/dl/)
* Git — [https://git-scm.com/downloads](https://git-scm.com/downloads)

No other software is required. SQLite is embedded via a pure-Go driver; CGO is not used.

---

## 2. Getting the code

```bash
git clone https://github.com/<your-username>/urlshorty.git
cd urlshorty
```

---

## 3. Environment configuration

The application runs with sensible defaults and does not require a `.env` file. You may optionally create one by copying `.env.example` to `.env`. Environment variables (or `.env` entries) supported:

| Variable     | Default                                        | Description                                              |
| ------------ | ---------------------------------------------- | -------------------------------------------------------- |
| PORT         | 8080                                           | HTTP listen port                                         |
| BASE\_URL    | [http://localhost:8080](http://localhost:8080) | Used to construct `short_url` values (no trailing slash) |
| DB\_PATH     | ./data/urlshorty.db                            | SQLite file path                                         |
| CODE\_LENGTH | 7                                              | Length of generated Base62 codes                         |
| RATE\_LIMIT  | 10:10                                          | Token bucket for POST /api/shorten, format rps\:burst    |

Examples:

Windows PowerShell:

```powershell
$env:PORT = "8081"
$env:BASE_URL = "http://localhost:8081"
$env:RATE_LIMIT = "20:40"
```

macOS/Linux:

```bash
export PORT=8081
export BASE_URL=http://localhost:8081
export RATE_LIMIT=20:40
```

If `.env` exists, the app loads it without overriding already-set environment variables.

---

## 4. Build and run

```bash
# Fetch Go modules
go mod tidy

# Create data folder for the SQLite database file
# Windows PowerShell:
#   mkdir -Force data
# macOS/Linux:
mkdir -p data

# Run the server
go run ./cmd/urlshorty
```

You should see the router start on the configured port. Open the root page:

* [http://localhost:8080/](http://localhost:8080/)

This page contains a minimal form that calls the API to create short links.

---

## 5. Quick usage examples

Below are self-contained examples using a real URL (a Google Form) to demonstrate functionality.

### Example long URL

```
https://docs.google.com/forms/d/e/1FAIpQLScZvxdUW0VChM-9-5N-yNmlI1n3sZJ9QAIPV37uCWnjJhjbKQ/viewform?usp=header
```

### Windows PowerShell

```powershell
$base = "http://localhost:8080"
$long = "https://docs.google.com/forms/d/e/1FAIpQLScZvxdUW0VChM-9-5N-yNmlI1n3sZJ9QAIPV37uCWnjJhjbKQ/viewform?usp=header"

# Create a short link (random code)
$res = Invoke-RestMethod -Method Post -Uri "$base/api/shorten" `
  -ContentType 'application/json' `
  -Body (@{ url = $long } | ConvertTo-Json)
$res
$code = $res.code
$short = $res.short_url
Start-Process $short  # opens redirect in browser

# Create a short link with a custom alias
$res2 = Invoke-RestMethod -Method Post -Uri "$base/api/shorten" `
  -ContentType 'application/json' `
  -Body (@{ url = $long; custom = "form" } | ConvertTo-Json)
$res2
Start-Process "$base/form"

# Create a short link with expiry in 2 hours
$exp = (Get-Date).AddHours(2).ToUniversalTime().ToString("s") + "Z"
$res3 = Invoke-RestMethod -Method Post -Uri "$base/api/shorten" `
  -ContentType 'application/json' `
  -Body (@{ url = $long; expires_at = $exp } | ConvertTo-Json)
$res3

# Fetch metadata for a code
Invoke-RestMethod -Method Get -Uri "$base/api/$code"
```

### curl (macOS/Linux/WSL)

```bash
BASE=http://localhost:8080
LONG='https://docs.google.com/forms/d/e/1FAIpQLScZvxdUW0VChM-9-5N-yNmlI1n3sZJ9QAIPV37uCWnjJhjbKQ/viewform?usp=header'

# Random code
curl -sS -X POST "$BASE/api/shorten" -H 'Content-Type: application/json' \
  -d "{\"url\":\"$LONG\"}"

# Custom alias
curl -sS -X POST "$BASE/api/shorten" -H 'Content-Type: application/json' \
  -d "{\"url\":\"$LONG\",\"custom\":\"form\"}"

# Expiring in 2 hours (UTC)
EXP=$(date -u -d '+2 hours' +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -v+2H -u +"%Y-%m-%dT%H:%M:%SZ")
curl -sS -X POST "$BASE/api/shorten" -H 'Content-Type: application/json' \
  -d "{\"url\":\"$LONG\",\"expires_at\":\"$EXP\"}"

# Metadata (replace CODE or use "form")
curl -sS "$BASE/api/CODE"
```

---

## 6. API specification

Base URL: `http://localhost:8080`

### POST `/api/shorten`

Create a short link.

Request body:

```json
{
  "url": "https://example.com/very/long/link",
  "custom": "my-alias-123",
  "expires_at": "2025-12-31T23:59:59Z"
}
```

Notes:

* `custom` is optional. Allowed characters: `[A-Za-z0-9_-]`. Length 3 to 64.
* `expires_at` is optional and must be a future RFC3339 timestamp.

Responses:

* `201 Created`

  ```json
  { "code": "Ab3kZpQ", "short_url": "http://localhost:8080/Ab3kZpQ" }
  ```
* `400 Bad Request` for invalid URL, invalid alias, invalid JSON, or past expiry.
* `409 Conflict` if a custom alias already exists.
* `429 Too Many Requests` if rate-limited.

### GET `/:code`

Redirect to the destination URL.

Responses:

* `301 Moved Permanently` and `Location` header with the original URL.
* `410 Gone` if the link has expired.
* `404 Not Found` if the code is unknown.
* `400 Bad Request` if the code format is invalid.

### GET `/api/:code`

Return metadata for a code.

Response:

```json
{
  "code": "Ab3kZpQ",
  "url": "https://example.com/very/long/link",
  "created_at": "2025-09-07T08:15:30Z",
  "expires_at": null,
  "hits": 3,
  "expired": false,
  "short_url": "http://localhost:8080/Ab3kZpQ"
}
```

### GET `/health`

Health check. Returns:

```json
{"ok": true}
```

---

## 7. Architecture and implementation

* Core service layer performs input validation, code generation, expiry checks, and delegates persistence.
* Base62 code generator uses `crypto/rand` for uniform randomness and a configurable length.
* SQLite persistence uses `modernc.org/sqlite` (pure Go). The schema is applied automatically at startup. No external migrations are required.
* HTTP layer uses Gin:

  * `POST /api/shorten` to create short links,
  * `GET /:code` for redirects,
  * `GET /api/:code` for metadata,
  * `GET /health` for readiness checks,
  * a minimal static page at `/`.
* Rate limiting is an in-memory token bucket keyed by client IP for `POST /api/shorten`.
* Server is configured with no trusted proxies for safe local defaults.

---

## 8. Project structure

```
cmd/urlshorty/main.go         # entrypoint

internal/app/                 # wiring of components
internal/config/config.go     # environment and .env loader with defaults
internal/core/                # business logic and interfaces
  types.go
  errors.go
  service.go
internal/http/                # Gin router, handlers, inline static page
  router.go
  handlers.go
  static.go
  middleware/
    logger.go
    recover.go
    ratelimit.go
internal/id/                  # base62 + crypto/rand generator
  base62.go
  rand.go
internal/rate/limiter.go      # token bucket limiter
internal/store/sqlite/        # SQLite persistence
  sqlite.go
  migrations.go

.github/workflows/ci.yml      # CI for test/lint/build
internal/http/handlers_test.go# end-to-end style test
```

---

## 9. Testing and CI

Run tests locally:

```bash
go test ./...
```

The repository includes a GitHub Actions workflow:

* Test job on Go 1.22.x and 1.23.x with race detector and coverage.
* Lint job using golangci-lint (if configured).
* Build job producing a binary artifact if tests and lint pass.

Coverage files are uploaded as workflow artifacts; Codecov upload can be enabled if desired.

---

## 10. Troubleshooting

* Port already in use: change `PORT` or stop the process using 8080.
* Windows firewall prompts: allow local network access on first run.
* Database write issues: ensure the `data/` directory exists and is writable; adjust `DB_PATH` if needed.
* Rate limited: raise `RATE_LIMIT` to a higher rps\:burst value.
* Go module checksum mismatch during `go mod tidy`: clear module cache and regenerate `go.sum`:

  * `go clean -modcache`
  * remove `go.sum`
  * `go mod tidy`

---

## 11. License

MIT. See `LICENSE` for full text.
