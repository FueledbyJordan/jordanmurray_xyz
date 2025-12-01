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

func HandleBlogList(w http.ResponseWriter, r *http.Request) {
	posts := models.GetAllPosts()
	component := templates.BlogList(posts)

	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering blog list: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func HandleBlogPost(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/blog/")
	if slug == "" {
		http.Redirect(w, r, "/blog", http.StatusSeeOther)
		return
	}

	post := models.GetPostBySlug(slug)
	if post == nil {
		http.NotFound(w, r)
		return
	}

	component := templates.BlogPost(*post)

	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering blog post: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
