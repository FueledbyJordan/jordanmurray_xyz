package handlers

import (
	"log"
	"net/http"
	"strings"

	"jordanmurray.xyz/site/internal/cache"
	"jordanmurray.xyz/site/templates"
)

type encodedResponse interface {
	Data() []byte
	CompressedData() []byte
	ContentType() string
}

func writeResponse(w http.ResponseWriter, r *http.Request, resp encodedResponse) {
	w.Header().Set("Content-Type", resp.ContentType())

	if strings.Contains(r.Header.Get("Accept-Encoding"), "br") {
		w.Header().Set("Content-Encoding", "br")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Write(resp.CompressedData())
	} else {
		w.Write(resp.Data())
	}
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

	writeResponse(w, r, cachedPost)
}

func HandleRSS(w http.ResponseWriter, r *http.Request) {
	if cache.Posts.Empty() {
		http.Error(w, "RSS feed not available", http.StatusInternalServerError)
		return
	}

	writeResponse(w, r, &cache.Posts.Generator)
}
