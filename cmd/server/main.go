package main

import (
	"log"
	"net/http"
	"raylib-2d-sandbox/internal/server"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", server.RootHandler)
	mux.HandleFunc("GET /health", server.HealthHandler)

	log.Printf("server has started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
