package game

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GameOver struct {
	menu *Menu
}

func NewGameOver() *GameOver {
	return &GameOver{menu: NewMenu([]*MenuButton{
		NewMenuButton("play again", ActionPlay),
		NewMenuButton("multiplayer", ActionMultiplayer),
		NewMenuButton("exit", ActionExit),
	}, 30, rl.Black)}
}

func (g *GameOver) draw() Action {
	rl.ClearBackground(g.menu.BackgroundColor)

	cx := int32(rl.GetScreenWidth()) / 2
	cy := int32(rl.GetScreenHeight()) / 2

	// Integer-sized pixels keep the title sharp without a font or texture asset.
	pixelSize := max(int32(1), min(int32(rl.GetScreenWidth())/64, int32(rl.GetScreenHeight())/64))
	titleY := cy - 10*pixelSize

	// A shattered version of the red square the player controls.
	rl.DrawRectangle(cx-2*pixelSize, titleY-8*pixelSize, 2*pixelSize, 2*pixelSize, rl.Red)
	rl.DrawRectangle(cx+pixelSize/2, titleY-7*pixelSize, pixelSize, 2*pixelSize, rl.Red)
	rl.DrawRectangle(cx-pixelSize, titleY-5*pixelSize, pixelSize, pixelSize, rl.Red)
	rl.DrawRectangle(cx+2*pixelSize, titleY-4*pixelSize, pixelSize/2, pixelSize/2, rl.Red)

	drawGameOverTitle(cx+pixelSize/3, titleY+pixelSize/3, pixelSize, color.RGBA{R: 95, G: 20, B: 28, A: 255})
	drawGameOverTitle(cx, titleY, pixelSize, rl.RayWhite)

	return g.menu.drawOptions(float32(cy + 5*pixelSize))
}

func drawGameOverTitle(centerX, y, pixelSize int32, tint color.RGBA) {
	const title = "GAME OVER"

	var gameOverGlyphs = map[rune][7]uint8{
		'G': {14, 17, 16, 23, 17, 17, 14},
		'A': {14, 17, 17, 31, 17, 17, 17},
		'M': {17, 27, 21, 21, 17, 17, 17},
		'E': {31, 16, 16, 30, 16, 16, 31},
		'O': {14, 17, 17, 17, 17, 17, 14},
		'V': {17, 17, 17, 17, 17, 10, 4},
		'R': {30, 17, 17, 30, 20, 18, 17},
	}

	const width = int32(len(title)*6 - 1)

	x := centerX - width*pixelSize/2

	for _, letter := range title {
		for row, bits := range gameOverGlyphs[letter] {
			for col := range 5 {
				if bits&(1<<uint(4-col)) != 0 {
					rl.DrawRectangle(x+int32(col)*pixelSize, y+int32(row)*pixelSize, pixelSize, pixelSize, tint)
				}
			}
		}
		x += 6 * pixelSize
	}
}
