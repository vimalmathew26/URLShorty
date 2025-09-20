package app

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"urlshorty/internal/config"
	"urlshorty/internal/core"
	httpapi "urlshorty/internal/http"
	"urlshorty/internal/id"
	"urlshorty/internal/rate"
	"urlshorty/internal/store/sqlite"
)

// App wires config, storage, core service, rate limiters, and the HTTP router.
type App struct {
	Cfg     config.Config
	Store   *sqlite.Store
	Service *core.Service
	Limiter *rate.Limiter
	Router  *gin.Engine
}

// New builds a fully-wired application instance.
func New(ctx context.Context, cfg config.Config) (*App, error) {
	// Open SQLite store (creates DB file and applies schema if missing).
	store, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// ID generator and core service.
	gen := id.NewGenerator(cfg.CodeLength)
	svc := core.NewService(store, gen)

	// In-memory rate limiter for POST /api/shorten
	var limiter *rate.Limiter
	if cfg.RateLimitRPS > 0 {
		limiter = rate.NewLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	}

	// HTTP router
	router := httpapi.NewRouter(svc, httpapi.Options{
		BaseURL:     cfg.BaseURL,
		RateLimiter: limiter,
	})

	return &App{
		Cfg:     cfg,
		Store:   store,
		Service: svc,
		Limiter: limiter,
		Router:  router,
	}, nil
}

// Addr returns the HTTP listen address, e.g. ":8080".
func (a *App) Addr() string {
	return fmt.Sprintf(":%d", a.Cfg.Port)
}

// Start runs the HTTP server (blocking).
func (a *App) Start() error {
	return a.Router.Run(a.Addr())
}

// Close releases resources (call on shutdown if you wire graceful stop).
func (a *App) Close() error {
	return a.Store.Close()
}
