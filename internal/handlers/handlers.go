package handlers

import (
	"log"
	"net/http"
	"strings"

	"jordanmurray.xyz/site/internal/cache"
	"jordanmurray.xyz/site/templates"
)

func writeWithEncoding(w http.ResponseWriter, r *http.Request, data, compressedData []byte, contentType string) {
	acceptEncoding := r.Header.Get("Accept-Encoding")

	var resp []byte
	if strings.Contains(acceptEncoding, "br") && len(compressedData) > 0 {
		w.Header().Set("Content-Encoding", "br")
		w.Header().Set("Vary", "Accept-Encoding")
		resp = compressedData
	} else {
		resp = data
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(resp)
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

	writeWithEncoding(w, r, cachedPost.HTML, cachedPost.CompressedHTML, "text/html; charset=utf-8")
}

func HandleRSS(w http.ResponseWriter, r *http.Request) {
	if len(cache.Posts.RssFeed()) == 0 {
		http.Error(w, "RSS feed not available", http.StatusInternalServerError)
		return
	}

	writeWithEncoding(w, r, cache.Posts.RssFeed(), cache.Posts.CompressedRssFeed(), "application/rss+xml; charset=utf-8")
}
