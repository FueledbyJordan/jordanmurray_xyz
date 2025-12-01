package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"jordanmurray.xyz/blog/handlers"
	"jordanmurray.xyz/blog/version"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "42069"
	}

	// Static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	http.HandleFunc("/", handlers.HandleHome)
	http.HandleFunc("/blog", handlers.HandleBlogList)
	http.HandleFunc("/blog/", handlers.HandleBlogPost)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on http://localhost%s (version: %s)", addr, version.Version)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
