package main

import (
	"flag"
	"raylib-2d-sandbox/internal/game"
)

func main() {
	serverURL := flag.String("server", "ws://localhost:8080/ws", "multiplayer WebSocket server URL")
	testing := flag.Bool("testing", false, "start directly in the testing mode game")
	flag.Parse()
	game.RunWithOptions(*serverURL, *testing)
}
