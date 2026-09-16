package main

import (
	"flag"
	"raylib-2d-sandbox/internal/game"
)

func main() {
	serverURL := flag.String("server", "ws://localhost:8080/ws", "multiplayer WebSocket server URL")
	flag.Parse()
	game.RunWithServer(*serverURL)
}
