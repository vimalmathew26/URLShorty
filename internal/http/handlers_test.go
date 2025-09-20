package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"urlshorty/internal/app"
	"urlshorty/internal/config"
)

func newTestServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		Port:           0,
		BaseURL:        "http://example", // response builds short_url from this
		DBPath:         ":memory:",
		CodeLength:     7,
		RateLimitRPS:   0, // disable limiter in tests
		RateLimitBurst: 0,
	}

	a, err := app.New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	srv := httptest.NewServer(a.Router)
	cleanup := func() {
		srv.Close()
		_ = a.Close()
	}
	return srv, cleanup
}

func postJSON(t *testing.T, client *http.Client, url string, body any) (*http.Response, []byte) {
	t.Helper()
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		buf = bytes.NewBuffer(b)
	}
	req, _ := http.NewRequest(http.MethodPost, url, buf)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	data, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	return res, data
}

func get(t *testing.T, client *http.Client, url string) (*http.Response, []byte) {
	t.Helper()
	res, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	data, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	return res, data
}

func TestURLShorty_EndToEnd(t *testing.T) {
	ts, done := newTestServer(t)
	defer done()

	base := ts.URL

	// 1) Health
	{
		res, body := get(t, ts.Client(), base+"/health")
		if res.StatusCode != http.StatusOK {
			t.Fatalf("health: status=%d body=%s", res.StatusCode, string(body))
		}
	}

	// 2) Shorten a URL
	var code string
	{
		res, body := postJSON(t, ts.Client(), base+"/api/shorten", map[string]any{
			"url": "https://go.dev/",
		})
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("shorten: status=%d body=%s", res.StatusCode, string(body))
		}
		var out struct {
			Code     string `json:"code"`
			ShortURL string `json:"short_url"`
		}
		_ = json.Unmarshal(body, &out)
		if out.Code == "" || out.ShortURL == "" {
			t.Fatalf("shorten: bad payload: %s", string(body))
		}
		code = out.Code
	}

	// 3) Metadata before redirect (hits should be 0)
	{
		res, body := get(t, ts.Client(), base+"/api/"+code)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("metadata: status=%d body=%s", res.StatusCode, string(body))
		}
		var meta struct {
			Hits int64 `json:"hits"`
		}
		_ = json.Unmarshal(body, &meta)
		if meta.Hits != 0 {
			t.Fatalf("expected hits=0, got %d", meta.Hits)
		}
	}

	// 4) Redirect (do NOT follow; inspect Location)
	nfClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	{
		res, _ := nfClient.Get(base + "/" + code)
		if res.StatusCode != http.StatusMovedPermanently {
			t.Fatalf("redirect: expected 301, got %d", res.StatusCode)
		}
		loc := res.Header.Get("Location")
		if loc != "https://go.dev/" {
			t.Fatalf("redirect: bad Location %q", loc)
		}
	}

	// 5) Metadata after redirect (hits should be >= 1; allow async)
	{
		var hits int64
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			res, body := get(t, ts.Client(), base+"/api/"+code)
			if res.StatusCode != http.StatusOK {
				t.Fatalf("metadata 2: status=%d body=%s", res.StatusCode, string(body))
			}
			var meta struct {
				Hits int64 `json:"hits"`
			}
			_ = json.Unmarshal(body, &meta)
			hits = meta.Hits
			if hits >= 1 {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		if hits < 1 {
			t.Fatalf("expected hits>=1 after redirect, got %d", hits)
		}
	}

	// 6) Custom alias OK and duplicate -> 409
	{
		// create custom
		res, body := postJSON(t, ts.Client(), base+"/api/shorten", map[string]any{
			"url":    "https://example.com",
			"custom": "vimal",
		})
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("custom: status=%d body=%s", res.StatusCode, string(body))
		}
		// duplicate custom
		res2, _ := postJSON(t, ts.Client(), base+"/api/shorten", map[string]any{
			"url":    "https://example.com/2",
			"custom": "vimal",
		})
		if res2.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409 for duplicate custom, got %d", res2.StatusCode)
		}
	}

	// 7) Expiry flow: create link that expires in ~1s
	var expCode string
	{
		exp := time.Now().Add(1 * time.Second).UTC().Format(time.RFC3339)
		res, body := postJSON(t, ts.Client(), base+"/api/shorten", map[string]any{
			"url":        "https://httpbin.org/get",
			"expires_at": exp,
		})
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("shorten exp: status=%d body=%s", res.StatusCode, string(body))
		}
		var out struct{ Code string }
		_ = json.Unmarshal(body, &out)
		expCode = out.Code
	}

	// should redirect before expiry
	{
		res, _ := nfClient.Get(base + "/" + expCode)
		if res.StatusCode != http.StatusMovedPermanently {
			t.Fatalf("redirect pre-exp: expected 301, got %d", res.StatusCode)
		}
	}
	// wait a bit and ensure it expires
	time.Sleep(1200 * time.Millisecond)
	{
		res, _ := nfClient.Get(base + "/" + expCode)
		if res.StatusCode != http.StatusGone {
			t.Fatalf("redirect post-exp: expected 410 Gone, got %d", res.StatusCode)
		}
	}

	// 8) Invalid URL rejected
	{
		res, _ := postJSON(t, ts.Client(), base+"/api/shorten", map[string]any{
			"url": "notaurl",
		})
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("invalid url: expected 400, got %d", res.StatusCode)
		}
	}

	// 9) Invalid code rejected on metadata (too short)
	{
		res, _ := get(t, ts.Client(), base+"/api/ab")
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("invalid code: expected 400, got %d", res.StatusCode)
		}
	}
}
