# urlshorty — a tiny URL shortener in Go 

A **single-binary** URL shortener built with **Go + Gin + SQLite**.
Runs locally on Windows, macOS, or Linux with just Go installed.
Includes a tiny web page, clean JSON API, optional `.env` config, and basic tests.

---

## ✨ Features

* **Shorten links** with randomly generated Base62 codes
* **Custom aliases** (e.g., `http://localhost:8080/vimal`)
* **Optional expiry** (`expires_at` in ISO-8601 / RFC3339)
* **301 redirects** and **metadata API** (hits, timestamps, expiry)
* **In-memory rate limiting** for the shorten endpoint
* **Embedded schema** (SQLite) — no external DB or migration tools
* **Tiny UI** at `/` (one HTML page calling the API)
* **Tests** using `httptest` (no extra deps)
* **No CGO** (pure Go SQLite driver `modernc.org/sqlite`)

---

## 🧰 Prerequisites

* **Go 1.22+** → [https://go.dev/dl/](https://go.dev/dl/)
* **Git** → [https://git-scm.com/downloads](https://git-scm.com/downloads)

> **No Docker required.** No separate SQLite install required.

---

## 🚀 Quick Start (Windows/macOS/Linux)

Open a terminal (PowerShell on Windows) and run:

```bash
# 1) Get the code
git clone https://github.com/<your-username>/urlshorty.git
cd urlshorty

# 2) (Optional) use a local .env file instead of env vars
# Windows PowerShell:
#   Copy-Item .env.example .env
# macOS/Linux:
#   cp .env.example .env

# 3) Fetch dependencies
go mod tidy

# 4) Create data folder (SQLite file lives here)
mkdir -p data  # PowerShell: mkdir -Force data

# 5) Run the server
go run ./cmd/urlshorty
```

Open **[http://localhost:8080](http://localhost:8080)** in your browser.
You’ll see a tiny page where you can paste a long URL and click **Shorten**.

---

## 🧪 Try the API (copy/paste)

### PowerShell (Windows)

```powershell
$base = "http://localhost:8080"

# Create a short link
$res = Invoke-RestMethod -Method Post -Uri "$base/api/shorten" `
  -ContentType 'application/json' `
  -Body '{"url":"https://go.dev"}'
$res
$code = $res.code
$short = $res.short_url

# Open the short URL (should redirect to https://go.dev)
Start-Process $short

# Metadata for the code
Invoke-RestMethod -Method Get -Uri "$base/api/$code"

# Health
Invoke-RestMethod -Method Get -Uri "$base/health"
```

### curl (macOS/Linux/WSL)

```bash
BASE=http://localhost:8080

# Create a short link
curl -sS -X POST "$BASE/api/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://go.dev"}'

# Metadata (replace CODE)
curl -sS "$BASE/api/CODE"

# Health
curl -sS "$BASE/health"
```

---

## ⚙️ Configuration

The app runs with sensible defaults — **no config required**.
You can customize via **environment variables** or a local **`.env`** file.

| Variable      | Default                 | What it does                                      |
| ------------- | ----------------------- | ------------------------------------------------- |
| `PORT`        | `8080`                  | HTTP port                                         |
| `BASE_URL`    | `http://localhost:8080` | Used to build `short_url` in responses (no slash) |
| `DB_PATH`     | `./data/urlshorty.db`   | SQLite file path                                  |
| `CODE_LENGTH` | `7`                     | Base62 code length                                |
| `RATE_LIMIT`  | `10:10`                 | rps\:burst (also accepts `10` or `10rps`)         |

**Windows PowerShell example:**

```powershell
$env:PORT="8081"
$env:BASE_URL="http://localhost:8081"
$env:RATE_LIMIT="20:40"
go run .\cmd\urlshorty
```

**macOS/Linux example:**

```bash
export PORT=8081
export BASE_URL=http://localhost:8081
export RATE_LIMIT=20:40
go run ./cmd/urlshorty
```

> If `.env` exists, the app loads it automatically (lines like `PORT=8081`).
> It won’t override variables already set in your environment.

---

## 📡 API Reference

### POST `/api/shorten` — Create a short link

**Request (JSON)**

```json
{
  "url": "https://example.com/very/long/link",
  "custom": "my-alias-123",         // optional; [A-Za-z0-9_-], 3..64 chars
  "expires_at": "2025-12-31T23:59:59Z"  // optional; RFC3339/ISO-8601
}
```

**Responses**

* `201 Created`

  ```json
  { "code": "Ab3kZpQ", "short_url": "http://localhost:8080/Ab3kZpQ" }
  ```
* `400 Bad Request` → invalid URL or alias, bad JSON, past expiry
* `409 Conflict` → custom alias already exists
* `429 Too Many Requests` → rate limited

---

### GET `/:code` — Redirect

* `301 Moved Permanently` → `Location: https://destination`
* `410 Gone` → link expired
* `404 Not Found` → unknown code
* `400 Bad Request` → invalid code format

---

### GET `/api/:code` — Metadata

**Response (200)**

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

---

### GET `/health` — Health check

```json
{"ok": true}
```

---

## 🗂️ Project Structure

```
cmd/urlshorty/main.go         # entrypoint (reads config, starts HTTP server)

internal/app/                 # wires config + store + service + rate limiter + router
internal/config/config.go     # env & .env loader (with defaults)
internal/core/                # business logic and interfaces
  types.go
  errors.go
  service.go
internal/http/                # Gin router, handlers, tiny UI
  router.go
  handlers.go
  static.go
  middleware/
    logger.go
    recover.go
    ratelimit.go
internal/id/                  # base62 + crypto/rand code generator
  base62.go
  rand.go
internal/rate/limiter.go      # simple per-IP token bucket
internal/store/sqlite/        # SQLite store with embedded schema
  sqlite.go
  migrations.go

.github/workflows/ci.yml      # runs `go test ./...` on push/PR
internal/http/handlers_test.go# end-to-end style test with httptest

go.mod / .gitignore / .env.example / LICENSE / README.md
```

---

## 🧱 Build a single binary (optional)

```bash
# Windows
go build -ldflags="-s -w" -o .\bin\urlshorty.exe .\cmd\urlshorty
.\bin\urlshorty.exe

# macOS/Linux
go build -ldflags="-s -w" -o ./bin/urlshorty ./cmd/urlshorty
./bin/urlshorty
```

---

## 🧰 Run tests

```bash
go test ./...
```

---

## 🧩 How it works (high level)

* **Core service** validates input, generates **Base62** codes, checks expiry, and calls the store
* **SQLite store** uses `modernc.org/sqlite` (pure Go) and creates the schema on startup
* **Gin handlers** expose `/api/shorten`, redirect `/:code`, metadata `/api/:code`, and `/health`
* **Rate limiter** is a simple in-memory token bucket keyed by client IP for `/api/shorten`
* A tiny **static HTML** page at `/` calls the API to make trying it out easy

---

## 🛡️ Notes & Troubleshooting

* **Port in use**: change `PORT` or stop the other service on 8080
* **Firewall prompt (Windows)**: allow local network access on first run
* **Database file path**: default `./data/urlshorty.db` (create `data/` if needed)
* **Too many requests**: you’re hitting the in-memory limiter; raise `RATE_LIMIT`
* **Trusted proxies**: router is configured to **trust none** (good for local use)
* **Checksum mismatch on `go mod tidy`**:
  Try: `go clean -modcache` then `Remove-Item go.sum` (Windows) or `rm go.sum` (macOS/Linux), then `go mod tidy`

---

## 🔒 Security basics

* No auth — intended for local demo / learning.
* If you expose it publicly, consider:

  * auth for `/api/shorten`
  * stricter rate limiting / captcha
  * domain allowlists / URL validation rules
  * HTTPS + reverse proxy (Caddy/Nginx)
  * persistence backups and retention policies

---

## 📄 License

[MIT](./LICENSE)

---

## 🙌 Credits

* Web framework: **Gin**
* SQLite driver: **modernc.org/sqlite** (pure Go)

---


