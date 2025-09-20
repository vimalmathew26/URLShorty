package config

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Config holds runtime configuration with sensible defaults for local dev.
type Config struct {
	Port           int    // HTTP port (default 8080)
	BaseURL        string // e.g., http://localhost:8080 (no trailing slash)
	DBPath         string // e.g., ./data/urlshorty.db
	CodeLength     int    // base62 code length (default 7)
	RateLimitRPS   int    // requests per second for POST /api/shorten (default 10)
	RateLimitBurst int    // burst tokens (default = RateLimitRPS)
}

// FromEnv loads configuration from environment variables, falling back to defaults.
// Recognized: PORT, BASE_URL, DB_PATH, CODE_LENGTH, RATE_LIMIT.
// Also (best-effort) loads a local ".env" file first if present.
func FromEnv() Config {
	loadDotEnv() // best-effort: sets env vars if not already set

	cfg := Config{
		Port:           getEnvInt("PORT", 8080),
		BaseURL:        sanitizeBaseURL(getEnv("BASE_URL", "http://localhost:8080")),
		DBPath:         getDBPath(getEnv("DB_PATH", "./data/urlshorty.db")),
		CodeLength:     getEnvInt("CODE_LENGTH", 7),
		RateLimitRPS:   10,
		RateLimitBurst: 10,
	}

	// Parse RATE_LIMIT if provided.
	if rl := strings.TrimSpace(os.Getenv("RATE_LIMIT")); rl != "" {
		rps, burst := parseRateLimit(rl)
		if rps > 0 {
			cfg.RateLimitRPS = rps
		}
		if burst > 0 {
			cfg.RateLimitBurst = burst
		} else {
			cfg.RateLimitBurst = cfg.RateLimitRPS
		}
	}

	// Ensure burst >= rps
	if cfg.RateLimitBurst < cfg.RateLimitRPS {
		cfg.RateLimitBurst = cfg.RateLimitRPS
	}
	if cfg.CodeLength <= 0 {
		cfg.CodeLength = 7
	}
	return cfg
}

func getEnv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func sanitizeBaseURL(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, "/")
	if s == "" {
		return "http://localhost:8080"
	}
	return s
}

func getDBPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		p = "./data/urlshorty.db"
	}
	// Normalize to OS-specific path; create parent dir if possible (best-effort).
	p = filepath.Clean(p)
	if dir := filepath.Dir(p); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	return p
}

var rateRe = regexp.MustCompile(`^\s*(\d+)\s*(?:rps)?\s*(?::\s*(\d+)\s*)?$`)

// parseRateLimit accepts "10", "10rps", or "10:20" (rps:burst).
func parseRateLimit(s string) (rps, burst int) {
	s = strings.ToLower(strings.TrimSpace(s))
	m := rateRe.FindStringSubmatch(s)
	if len(m) == 0 {
		return 0, 0
	}
	rps, _ = strconv.Atoi(m[1])
	if len(m) >= 3 && m[2] != "" {
		burst, _ = strconv.Atoi(m[2])
	} else {
		burst = rps
	}
	return rps, burst
}

// loadDotEnv loads KEY=VALUE pairs from a local ".env" file if present.
// - Ignores blank lines and lines starting with "#" or ";".
// - Strips surrounding single/double quotes from values.
// - Does NOT override variables already set in the environment.
func loadDotEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return // no .env; silently skip
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		i := strings.Index(line, "=")
		if i <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:i])
		val := strings.TrimSpace(line[i+1:])
		// Trim matching quotes if present
		val = strings.Trim(val, `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue // don't override real env
		}
		_ = os.Setenv(key, val)
	}
}
