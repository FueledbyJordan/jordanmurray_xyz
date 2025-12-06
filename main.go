package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"jordanmurray.xyz/site/internal/cache"
	"jordanmurray.xyz/site/internal/handlers"
	"jordanmurray.xyz/site/internal/models"
)

var (
	//go:embed static
	staticFiles embed.FS

	//go:embed content
	contentFiles embed.FS
)

func main() {
	rssBaseURL := os.Getenv("RSS_BASE_URL")
	if rssBaseURL == "" {
		rssBaseURL = "https://jordanmurray.xyz"
	}

	rssConfig := models.RSSConfig{
		BaseURL:     rssBaseURL,
		Title:       "jordanmurray.xyz // reflections",
		Description: "a personal time capsule in a glass box",
	}

	ctx := context.Background()

	hydrateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cache.Hydrate(contentFiles, rssConfig, hydrateCtx)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("GET /{$}", handlers.HandleHome)
	http.HandleFunc("GET /reflections", handlers.HandleReflections)
	http.HandleFunc("GET /reflections/{slug}", handlers.HandleReflection)
	http.HandleFunc("GET /reflections/feed.rss", handlers.HandleRSS)
	http.HandleFunc("HEAD /reflections/feed.rss", handlers.HandleRSS)
	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	addr := fmt.Sprintf(":%s", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
