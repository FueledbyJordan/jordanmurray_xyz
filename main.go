package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"jordanmurray.xyz/site/handlers"
	"jordanmurray.xyz/site/models"
	"jordanmurray.xyz/site/templates"
)

//go:embed static
var staticFiles embed.FS

//go:embed content
var contentFiles embed.FS

func main() {
	// Set up embedded content filesystem
	models.SetContentFS(contentFiles)

	// Set up pre-rendering for posts
	models.SetRenderFunc(func(post *models.Post) ([]byte, error) {
		var buf bytes.Buffer
		component := templates.Reflection(*post)
		if err := component.Render(context.Background(), &buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	// Serve embedded static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Routes
	http.HandleFunc("/", handlers.HandleHome)
	http.HandleFunc("/reflections", handlers.HandleReflections)
	http.HandleFunc("/reflections/", handlers.HandleReflection)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
