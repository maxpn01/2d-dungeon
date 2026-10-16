package main

import (
	"flag"

	"raylib-2d-sandbox/internal/game"
)

func main() {
	testing := flag.Bool("testing", false, "start directly in the testing mode game")
	flag.Parse()
	game.Run(*testing)
}
