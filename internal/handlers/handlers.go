package handlers

import (
	"log"
	"net/http"
	"strings"

	"jordanmurray.xyz/site/internal/cache"
	"jordanmurray.xyz/site/internal/rss"
	"jordanmurray.xyz/site/templates"
)

var rssGenerator *rss.Generator

func SetRSSGenerator(gen *rss.Generator) {
	rssGenerator = gen
}

func HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	posts := cache.Posts.GetAllPosts()
	component := templates.Home(posts)

	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering home: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func HandleReflections(w http.ResponseWriter, r *http.Request) {
	posts := cache.Posts.GetAllPosts()
	component := templates.Reflections(posts)

	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering reflections list: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func HandleReflection(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/reflections/")
	if slug == "" {
		http.Error(w, "reflection id must be set", http.StatusBadRequest)
		return
	}

	cachedPost, err := cache.Posts.GetPostBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	acceptEncoding := r.Header.Get("Accept-Encoding")
	if strings.Contains(acceptEncoding, "br") {
		w.Header().Set("Content-Encoding", "br")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Write(cachedPost.CompressedHTML)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(cachedPost.HTML)
}

func HandleRSS(w http.ResponseWriter, r *http.Request) {
	if rssGenerator == nil {
		log.Printf("RSS generator not configured")
		http.Error(w, "RSS feed not available", http.StatusInternalServerError)
		return
	}

	acceptEncoding := r.Header.Get("Accept-Encoding")
	if strings.Contains(acceptEncoding, "br") {
		rssFeed := rssGenerator.GetFeedBrotli()
		if len(rssFeed) > 0 {
			w.Header().Set("Content-Encoding", "br")
			w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
			w.Header().Set("Vary", "Accept-Encoding")
			w.Write(rssFeed)
			return
		}
	}

	rssFeed := rssGenerator.GetFeed()
	if len(rssFeed) > 0 {
		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		w.Write(rssFeed)
		return
	}

	// If feed wasn't generated, return error
	log.Printf("RSS feed not available")
	http.Error(w, "RSS feed not available", http.StatusInternalServerError)
}
