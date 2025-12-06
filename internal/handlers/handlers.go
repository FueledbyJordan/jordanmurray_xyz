package handlers

import (
	"log"
	"net/http"
	"strings"

	"jordanmurray.xyz/site/internal/cache"
	"jordanmurray.xyz/site/templates"
)

func HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	posts := cache.Posts.GetAllPosts()
	component := templates.Reflections(posts)

	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering reflections list: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func HandleReflection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	var resp []byte
	acceptEncoding := r.Header.Get("Accept-Encoding")
	if strings.Contains(acceptEncoding, "br") {
		w.Header().Set("Content-Encoding", "br")
		w.Header().Set("Vary", "Accept-Encoding")
		resp = cachedPost.CompressedHTML
	} else {
		resp = cachedPost.HTML
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(resp)
}

func HandleRSS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(cache.Posts.RssFeed()) == 0 {
		http.Error(w, "RSS feed not available", http.StatusInternalServerError)
		return
	}

	acceptEncoding := r.Header.Get("Accept-Encoding")

	var resp []byte
	if strings.Contains(acceptEncoding, "br") {
		resp = cache.Posts.CompressedRssFeed()
		w.Header().Set("Content-Encoding", "br")
		w.Header().Set("Vary", "Accept-Encoding")
	} else {
		resp = cache.Posts.RssFeed()
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Write(resp)
}
