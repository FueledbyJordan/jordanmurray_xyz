package handlers

import (
	"log"
	"net/http"

	"jordanmurray.xyz/site/internal/cache"
	"jordanmurray.xyz/site/templates"
)

func HandleHome(w http.ResponseWriter, r *http.Request) {
	withCache(func(w http.ResponseWriter, r *http.Request, c *cache.Cache) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		component := templates.Home(c.AllPosts())
		if err := component.Render(r.Context(), w); err != nil {
			log.Printf("Error rendering home: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})(w, r)
}
