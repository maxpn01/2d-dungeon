package main

import (
	"flag"

	"2d-dungeon/internal/game"
)

func main() {
	testing := flag.Bool("testing", false, "start directly in the testing mode game")
	flag.Parse()
	game.Run(*testing)
}
