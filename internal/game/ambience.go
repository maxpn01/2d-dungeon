package game

import (
	"dungeon-raylib-go/internal/rendering"
	"dungeon-raylib-go/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (g *Game) updateAmbience(current screen, dt float32) {
	active := current == screenSinglePlayer && !g.world.Dead() && !g.world.ResultsOpen() && rl.IsWindowFocused()
	if active {
		center := g.world.Center()
		geometry := worldmap.Default()
		active = geometry.FloorAt(float64(center.X), float64(center.Y)) == worldmap.FirstFloor
	}
	rendering.UpdateFloorAmbience(active, dt)
}
