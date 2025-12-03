package handlers

import (
	"log"
	"net/http"
	"strings"

	"jordanmurray.xyz/site/models"
	"jordanmurray.xyz/site/templates"
)

func HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	posts := models.GetAllPosts()
	component := templates.Home(posts)

	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering home: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func HandleReflections(w http.ResponseWriter, r *http.Request) {
	posts := models.GetAllPosts()
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

	post := models.GetPostBySlug(slug)
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
