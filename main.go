package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"jordanmurray.xyz/site/internal/cache"
	"jordanmurray.xyz/site/internal/handlers"
	"jordanmurray.xyz/site/internal/models"
	"jordanmurray.xyz/site/internal/rss"
	"jordanmurray.xyz/site/templates"
)

var (
	//go:embed static
	staticFiles embed.FS

	//go:embed content
	contentFiles embed.FS
)

type templRenderer struct{}

func (templRenderer) Render(post *models.Post) ([]byte, error) {
	var buf bytes.Buffer
	component := templates.Reflection(*post)
	if err := component.Render(context.Background(), &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func main() {
	cache.Posts.SetContentFS(contentFiles)
	cache.Posts.SetRenderer(templRenderer{})

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
