package ui

import (
	"image/color"

	"raylib-2d-sandbox/internal/display"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GameOver struct {
	menu *Menu
}

func NewGameOver() *GameOver {
	return &GameOver{menu: NewMenu([]*MenuButton{
		NewMenuButton("restart", ActionRestart),
		NewMenuButton("back to menu", ActionMenu),
		NewMenuButton("exit", ActionExit),
	}, 30, rl.Black)}
}

func (g *GameOver) Draw() Action {
	rl.ClearBackground(g.menu.BackgroundColor)

	cx := int32(display.Width()) / 2
	cy := int32(display.Height()) / 2

	// Integer-sized pixels keep the title sharp without a font or texture asset.
	pixelSize := max(int32(1), min(int32(display.Width())/64, int32(display.Height())/64))
	titleY := cy - 10*pixelSize

	// A shattered version of the red square the player controls.
	rl.DrawRectangle(cx-2*pixelSize, titleY-8*pixelSize, 2*pixelSize, 2*pixelSize, rl.Red)
	rl.DrawRectangle(cx+pixelSize/2, titleY-7*pixelSize, pixelSize, 2*pixelSize, rl.Red)
	rl.DrawRectangle(cx-pixelSize, titleY-5*pixelSize, pixelSize, pixelSize, rl.Red)
	rl.DrawRectangle(cx+2*pixelSize, titleY-4*pixelSize, pixelSize/2, pixelSize/2, rl.Red)

	drawPixelTitle("GAME OVER", cx+pixelSize/3, titleY+pixelSize/3, pixelSize, color.RGBA{R: 95, G: 20, B: 28, A: 255})
	drawPixelTitle("GAME OVER", cx, titleY, pixelSize, rl.RayWhite)

	return g.menu.drawOptions(float32(cy + 5*pixelSize))
}
