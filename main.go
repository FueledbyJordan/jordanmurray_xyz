package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"jordanmurray.xyz/site/internal/cache"
	"jordanmurray.xyz/site/internal/handlers"
	"jordanmurray.xyz/site/internal/rss"
)

var (
	//go:embed static
	staticFiles embed.FS

	//go:embed content
	contentFiles embed.FS
)

func main() {
	ctx := context.Background()

	cachedPosts, err := cache.LoadPosts(contentFiles, ctx)
	if err != nil {
		log.Fatal(err)
	}

	rssBaseURL := os.Getenv("RSS_BASE_URL")
	if rssBaseURL == "" {
		rssBaseURL = "https://jordanmurray.xyz"
	}
	rssGen := rss.NewGenerator(
		rssBaseURL,
		"jordanmurray.xyz // reflections",
		"a personal time capsule in a glass box",
	)
	cache.Posts.SetRSSGenerator(rssGen)
	cache.Posts.Load(cachedPosts)
	handlers.SetRSSGenerator(rssGen)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	// Serve embedded static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Routes
	http.HandleFunc("/", handlers.HandleHome)
	http.HandleFunc("/reflections", handlers.HandleReflections)
	http.HandleFunc("/reflections/", handlers.HandleReflection)
	http.HandleFunc("/reflections/feed.rss", handlers.HandleRSS)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
