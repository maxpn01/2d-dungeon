package game

import (
	"raylib-2d-sandbox/internal/terrain"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func newFullMapView(windowWidth, windowHeight int32, m *Map) radarView {
	const margin float32 = 72
	width := float32((terrain.MaxCellX - terrain.MinCellX) * terrain.Tile)
	height := m.size.Y - terrain.North
	scale := min(max(1, float32(windowWidth)-2*margin)/width, max(1, float32(windowHeight)-2*margin)/height)
	return radarView{
		bounds: rl.NewRectangle((float32(windowWidth)-width*scale)/2, (float32(windowHeight)-height*scale)/2, width*scale, height*scale),
		focus:  rl.NewVector2(terrain.MinCellX*terrain.Tile+width/2, terrain.North+height/2),
		scale:  scale,
	}
}

func (g *Game) drawFullMap(multiplayer bool) {
	width, height := int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight())
	rl.ClearBackground(g.menu.BackgroundColor)
	drawCenteredText("WORLD MAP", 24, 28, rl.RayWhite)
	if multiplayer {
		g.multiplayer.drawRadar(width, height)
	} else {
		g.world.drawRadar(newFullMapView(width, height, g.world.gameMap))
	}
	drawCenteredText("M / Esc to close", height-44, 20, rl.LightGray)
}
