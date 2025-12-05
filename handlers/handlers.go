package handlers

import (
	"encoding/xml"
	"log"
	"net/http"
	"strings"
	"time"

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

// RSS feed types
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	Language      string `xml:"language"`
	LastBuildDate string `xml:"lastBuildDate"`
	Items         []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

func HandleRSS(w http.ResponseWriter, r *http.Request) {
	posts := models.GetAllPosts()

	baseURL := "https://jordanmurray.xyz"
	if host := r.Host; host != "" && strings.HasPrefix(r.URL.Scheme, "http") {
		baseURL = r.URL.Scheme + "://" + host
	} else if host := r.Host; host != "" {
		// Default to https if scheme not available
		if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
			baseURL = "http://" + host
		} else {
			baseURL = "https://" + host
		}
	}

	var items []Item
	var lastBuildDate time.Time

	for _, post := range posts {
		items = append(items, Item{
			Title:       post.Title,
			Link:        baseURL + "/reflections/" + post.Slug,
			Description: post.Excerpt,
			PubDate:     post.PublishedAt.Format(time.RFC1123Z),
			GUID:        baseURL + "/reflections/" + post.Slug,
		})

		if post.PublishedAt.After(lastBuildDate) {
			lastBuildDate = post.PublishedAt
		}
	}

	if lastBuildDate.IsZero() && len(posts) > 0 {
		lastBuildDate = time.Now()
	}

	feed := RSS{
		Version: "2.0",
		Channel: Channel{
			Title:         "Jordan Murray - Reflections",
			Link:          baseURL,
			Description:   "Thoughts and writings on software development, technology, and more",
			Language:      "en-us",
			LastBuildDate: lastBuildDate.Format(time.RFC1123Z),
			Items:         items,
		},
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")

	if _, err := w.Write([]byte(xml.Header)); err != nil {
		log.Printf("Error writing XML header: %v", err)
		return
	}

	if err := encoder.Encode(feed); err != nil {
		log.Printf("Error encoding RSS feed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
