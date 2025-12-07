package handlers

import (
	"log"
	"net/http"

	"jordanmurray.xyz/site/internal/cache"
)

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
}

func withCache(next func(w http.ResponseWriter, r *http.Request, c *cache.Cache)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := cache.Get()
		if err != nil {
			log.Printf("cache was nil: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		next(w, r, c)
	}
}
