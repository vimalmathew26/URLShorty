package main

import (
	"context"
	"log"

	"urlshorty/internal/app"
	"urlshorty/internal/config"
)

func main() {
	cfg := config.FromEnv()

	a, err := app.New(context.Background(), cfg)
	if err != nil {
		log.Fatalf("boot: %v", err)
	}
	log.Printf("urlshorty listening on %s (BASE_URL=%s, DB=%s)", a.Addr(), cfg.BaseURL, cfg.DBPath)

	// Blocking; press Ctrl+C to stop the process.
	if err := a.Start(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
