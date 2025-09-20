package http

import (
	"log"

	"github.com/gin-gonic/gin"

	"urlshorty/internal/core"
	"urlshorty/internal/http/middleware"
	"urlshorty/internal/rate"
)

type Options struct {
	BaseURL     string
	RateLimiter *rate.Limiter // used for POST /api/shorten only
}

// NewRouter sets up all routes and middleware.
func NewRouter(svc *core.Service, opts Options) *gin.Engine {
	r := gin.New()
	// Treat all upstreams as untrusted (removes the warning).
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Printf("SetTrustedProxies: %v", err)
	}

	r.Use(middleware.Logger())
	r.Use(middleware.Recover())

	h := NewHandlers(svc, opts.BaseURL)

	// Health
	r.GET("/health", h.Health)

	// Optional tiny UI (inline HTML)
	RegisterStatic(r)

	// API
	api := r.Group("/api")
	// POST /api/shorten (rate-limited if limiter provided)
	if opts.RateLimiter != nil {
		api.POST("/shorten", middleware.RateLimit(opts.RateLimiter), h.Shorten)
	} else {
		api.POST("/shorten", h.Shorten)
	}
	api.GET("/:code", h.Metadata)

	// Redirect
	r.GET("/:code", h.Redirect)

	return r
}
