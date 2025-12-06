package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"jordanmurray.xyz/site/internal/cache"
	"jordanmurray.xyz/site/internal/handlers"
	"jordanmurray.xyz/site/internal/middleware"
	"jordanmurray.xyz/site/internal/rss"
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

	ctx := context.Background()

	cache.Posts.Initialize(
		contentFiles,
		rss.Config{
			BaseURL:     rssBaseURL,
			Title:       "jordanmurray.xyz // reflections",
			Description: "a personal time capsule in a glass box",
		},
		ctx,
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}

	getOnly := []string{http.MethodGet}
	http.HandleFunc("/", middleware.MethodFilter(getOnly, handlers.HandleHome))
	http.HandleFunc("/reflections", middleware.MethodFilter(getOnly, handlers.HandleReflections))
	http.HandleFunc("/reflections/", middleware.MethodFilter(getOnly, handlers.HandleReflection))
	http.HandleFunc("/reflections/feed.rss", middleware.MethodFilter([]string{http.MethodGet, http.MethodHead}, handlers.HandleRSS))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	addr := fmt.Sprintf(":%s", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
