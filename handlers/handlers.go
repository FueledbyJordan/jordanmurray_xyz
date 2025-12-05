package handlers

import (
	"log"
	"net/http"
	"strings"

	"jordanmurray.xyz/site/cache"
	"jordanmurray.xyz/site/rss"
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
		http.Redirect(w, r, "/reflections", http.StatusSeeOther)
		return
	}

	post := cache.Posts.GetPostBySlug(slug)
	if post == nil {
		http.NotFound(w, r)
		return
	}

	// Check if client accepts brotli encoding
	acceptEncoding := r.Header.Get("Accept-Encoding")
	if strings.Contains(acceptEncoding, "br") && len(post.RenderedHTMLBrotli) > 0 {
		// Serve pre-compressed brotli version
		w.Header().Set("Content-Encoding", "br")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Write(post.RenderedHTMLBrotli)
		return
	}

	// Serve uncompressed pre-rendered version
	if len(post.RenderedHTML) > 0 {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(post.RenderedHTML)
		return
	}

	// Fallback to dynamic rendering (if pre-rendering failed)
	component := templates.Reflection(*post)
	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering reflection: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func HandleRSS(w http.ResponseWriter, r *http.Request) {
	if rssGenerator == nil {
		log.Printf("RSS generator not configured")
		http.Error(w, "RSS feed not available", http.StatusInternalServerError)
		return
	}

	// Check if client accepts brotli encoding
	acceptEncoding := r.Header.Get("Accept-Encoding")
	if strings.Contains(acceptEncoding, "br") {
		rssFeed := rssGenerator.GetFeedBrotli()
		if len(rssFeed) > 0 {
			// Serve pre-compressed brotli version
			w.Header().Set("Content-Encoding", "br")
			w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
			w.Header().Set("Vary", "Accept-Encoding")
			w.Write(rssFeed)
			return
		}
	}

	// Serve uncompressed pre-generated version
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
