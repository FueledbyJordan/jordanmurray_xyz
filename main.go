package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"jordanmurray.xyz/site/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	// Static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

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
