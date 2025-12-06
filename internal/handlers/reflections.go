package handlers

import (
	"log"
	"net/http"
	"strings"

	"jordanmurray.xyz/site/internal/cache"
	"jordanmurray.xyz/site/internal/renderer"
	"jordanmurray.xyz/site/templates"
)

func HandleReflections(w http.ResponseWriter, r *http.Request) {
	withCache(func(w http.ResponseWriter, r *http.Request, c *cache.Cache) {
		component := templates.Reflections(c.AllPosts())
		if err := component.Render(r.Context(), w); err != nil {
			log.Printf("Error rendering reflections list: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})(w, r)
}

func HandleReflection(w http.ResponseWriter, r *http.Request) {
	withCache(func(w http.ResponseWriter, r *http.Request, c *cache.Cache) {
		slug := strings.TrimPrefix(r.URL.Path, "/reflections/")
		if slug == "" {
			http.Error(w, "reflection id must be set", http.StatusBadRequest)
			return
		}

		cachedPost, err := c.PostBySlug(slug)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		renderer.Write(w, r, cachedPost)
	})(w, r)
}

func HandleRSS(w http.ResponseWriter, r *http.Request) {
	withCache(func(w http.ResponseWriter, r *http.Request, c *cache.Cache) {
		if c.RSS().Empty() {
			http.Error(w, "RSS feed not available", http.StatusInternalServerError)
			return
		}

		renderer.Write(w, r, c.RSS())
	})(w, r)
}
