package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"raylib-2d-sandbox/internal/server"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP server listen address")
	flag.Parse()

	hub := server.NewHub()
	go hub.Run(context.Background())

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", server.RootHandler)
	mux.HandleFunc("GET /health", server.HealthHandler)
	mux.HandleFunc("GET /ws", hub.WebSocketHandler)

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("server listening on %s", listener.Addr())
	log.Fatal(http.Serve(listener, mux))
}
