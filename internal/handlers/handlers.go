package handlers

import (
	"log"
	"net/http"
	"strings"

	"jordanmurray.xyz/site/internal/cache"
	"jordanmurray.xyz/site/internal/renderer"
	"jordanmurray.xyz/site/templates"
)

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

	renderer.Write(w, r, cachedPost)
}

func HandleRSS(w http.ResponseWriter, r *http.Request) {
	if cache.Posts.Empty() {
		http.Error(w, "RSS feed not available", http.StatusInternalServerError)
		return
	}

	renderer.Write(w, r, &cache.Posts.Generator)
}
