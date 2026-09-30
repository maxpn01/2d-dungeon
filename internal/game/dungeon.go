package game

import (
	"image/color"
	"math"
	"raylib-2d-sandbox/internal/terrain"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var stone = color.RGBA{53, 62, 73, 255}
var paleStone = color.RGBA{93, 109, 119, 255}
var runeColor = color.RGBA{102, 196, 192, 255}

func dungeonRect(x, y, w, h int, c color.RGBA) {
	rl.DrawRectangle(int32(x), int32(y), int32(w), int32(h), c)
}

func (m *Map) drawDungeon(ctx RenderContext) {
	zoom := max(ctx.Camera.Zoom, .1)
	left := int(math.Floor(float64(ctx.Camera.Target.X-ctx.Camera.Offset.X/zoom)/40)) - 1
	right := int(math.Ceil(float64(ctx.Camera.Target.X+(float32(ctx.WindowWidth)-ctx.Camera.Offset.X)/zoom)/40)) + 1
	top := int(math.Floor(float64(ctx.Camera.Target.Y-ctx.Camera.Offset.Y/zoom)/40)) - 1
	bottom := min(0, int(math.Ceil(float64(ctx.Camera.Target.Y+(float32(ctx.WindowHeight)-ctx.Camera.Offset.Y)/zoom)/40))+1)
	for y := max(terrain.MinCellY, top); y < bottom; y++ {
		for x := max(terrain.MinCellX, left); x < min(terrain.MaxCellX, right); x++ {
			id := terrain.Cell(x, y)
			px, py := x*40, y*40
			hash := uint32(x*73856093) ^ uint32((y+100)*19349663)
			if id != 0 {
				tone := uint8(hash % 7)
				c := color.RGBA{26 + tone, 32 + tone, 40 + tone, 255}
				dungeonRect(px, py, 40, 40, c)
				dungeonRect(px, py, 40, 1, color.RGBA{43, 49, 57, 255})
				dungeonRect(px, py, 1, 40, color.RGBA{43, 49, 57, 255})
				if (x == 19 || x == 20) && (id == 2 || id == 8 || id == 1) {
					dungeonRect(px+4, py+3, 32, 34, color.RGBA{47, 52, 58, 255})
					dungeonRect(px+7, py+6, 26, 1, color.RGBA{69, 70, 66, 255})
				}
				if terrain.Cover(x, y) {
					// Arcane residue seeps from engraved stone; it confuses creature sight.
					center := rl.NewVector2(float32(px+20), float32(py+20))
					pulse := float32(.7 + .15*math.Sin(rl.GetTime()*1.4+float64(hash%31)))
					dungeonRect(px+2, py+2, 36, 36, color.RGBA{33, 35, 57, 255})
					rl.DrawCircleGradient(center, 26, rl.Fade(color.RGBA{134, 125, 214, 255}, .20*pulse), rl.Blank)
					rl.DrawPolyLines(center, 4, 9, 45, rl.Fade(color.RGBA{149, 139, 218, 255}, .5*pulse))
					for i := 0; i < 3; i++ {
						ox := float32((hash+uint32(i*17))%29) + 5
						oy := float32((hash+uint32(i*11))%27) + 5 + float32(math.Sin(rl.GetTime()+float64(i)))*3
						rl.DrawCircleV(rl.NewVector2(float32(px)+ox, float32(py)+oy), 1.5, rl.Fade(color.RGBA{164, 155, 233, 255}, .65*pulse))
					}
				}
				if terrain.Cell(x, y-1) == 0 {
					dungeonRect(px, py, 40, 10, color.RGBA{8, 13, 20, 150})
				}
			} else if terrain.Pool(x, y) {
				// Pools are drawn as continuous basins below, without tile noise.
			} else {
				near := false
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if terrain.Cell(x+dx, y+dy) != 0 {
							near = true
						}
					}
				}
				if !near {
					continue
				}
				dungeonRect(px, py, 40, 40, color.RGBA{17, 22, 30, 255})
				dungeonRect(px+1, py+1, 38, 30, stone)
				dungeonRect(px+3, py+3, 34, 2, paleStone)
				dungeonRect(px+3, py+6, 2, 22, color.RGBA{67, 78, 87, 255})
				if terrain.Cell(x, y+1) != 0 {
					dungeonRect(px+1, py+31, 38, 9, color.RGBA{32, 39, 49, 255})
				}
			}
		}
	}
	for _, y := range []int{-55, -47} {
		for _, x := range []int{18, 27} {
			if x+5 >= left && x <= right && y+4 >= top && y <= bottom {
				drawDungeonPool(x*40, y*40)
			}
		}
	}
	for _, x := range []int{600, 920} {
		dungeonRect(x, -80, 80, 80, stone)
		dungeonRect(x+6, -74, 68, 5, paleStone)
		dungeonRect(x+12, -60, 56, 48, color.RGBA{31, 39, 49, 255})
		dungeonRect(x+34, -52, 12, 28, runeColor)
	}
	for _, r := range terrain.Rooms {
		if r.Y+r.H < top || r.Y > bottom {
			continue
		}
		cx, cy := float32((r.X*2+r.W)*20), float32((r.Y*2+r.H)*20)
		if r.Name == "THE THRESHOLD HALL" {
			rl.DrawRectangleLines(int32(r.X*40+100), int32(r.Y*40+100), int32(r.W*40-200), int32(r.H*40-200), stone)
			rl.DrawPolyLines(rl.NewVector2(cx, cy), 8, 145, 22.5, stone)
			rl.DrawCircleLines(int32(cx), int32(cy), 110, stone)
		}
	}
	for _, patch := range terrain.Patches {
		px, py := patch.X*terrain.Tile, patch.Y*terrain.Tile
		if px < left*40-160 || px > right*40+160 || py < top*40-160 || py > bottom*40+160 {
			continue
		}
		rl.DrawCircleLines(int32(px+20), int32(py+20), float32(patch.Radius*terrain.Tile), color.RGBA{44, 82, 87, 255})
		rl.DrawPolyLines(rl.NewVector2(float32(px+20), float32(py+20)), 4, 12, 45, runeColor)
		rl.DrawCircleLines(int32(px+20), int32(py+20), 6, runeColor)
	}
	for _, den := range terrain.Dens {
		px, py := den.X*terrain.Tile, den.Y*terrain.Tile
		if px < left*40 || px > right*40 || py < top*40 || py > bottom*40 {
			continue
		}
		for i := range 6 {
			a := float64(i) * math.Pi / 3
			dungeonRect(px+15+int(math.Cos(a)*26), py+15+int(math.Sin(a)*26), 10, 10, color.RGBA{124, 91, 67, 255})
		}
	}

}

func drawDungeonLocation(x, y float32, width, height int32) {
	name, hint := "SPAWN ZONE", "Head north through the dungeon entrance."
	if y < 0 {
		name, hint = "THE HOLLOW / FLOOR I", "Entrance lies to the south"
		for _, r := range terrain.Rooms {
			if x >= float32(r.X*40) && x < float32((r.X+r.W)*40) && y >= float32(r.Y*40) && y < float32((r.Y+r.H)*40) {
				name = "I / " + r.Name
				for _, tag := range r.Tags {
					if tag == "refuge" {
						hint = "Arcane veils confuse creatures. They do not stop attacks."
					}
				}
				break
			}
		}
		if y < terrain.DescentCircleY+160 {
			hint = "The circle links the floors. Stand inside to travel."
		}
	}
	if terrain.FloorTwo(float64(x), float64(y)) {
		name = "II / THE FEAST"
		hint = "Leave and re-enter the circle to return to floor I."
	}
	rl.DrawRectangle(0, height-70, width, 70, rl.Fade(rl.Black, .8))
	drawText(name, 24, height-57, 20, runeColor)
	drawText(hint, 24, height-29, 16, rl.LightGray)
}

func drawDungeonPool(x, y int) {
	dungeonRect(x, y, 200, 160, stone)
	dungeonRect(x+2, y+2, 196, 2, paleStone)
	dungeonRect(x+2, y+2, 2, 156, paleStone)
	dungeonRect(x+2, y+156, 196, 3, color.RGBA{25, 32, 41, 255})
	dungeonRect(x+196, y+2, 3, 154, color.RGBA{25, 32, 41, 255})
	for offset := 40; offset < 200; offset += 40 {
		dungeonRect(x+offset, y, 1, 8, color.RGBA{30, 39, 47, 255})
		dungeonRect(x+offset, y+152, 1, 8, color.RGBA{30, 39, 47, 255})
	}
	dungeonRect(x+8, y+8, 184, 144, color.RGBA{9, 24, 33, 255})
	dungeonRect(x+11, y+17, 178, 132, color.RGBA{19, 49, 62, 255})
	dungeonRect(x+16, y+24, 168, 120, color.RGBA{21, 55, 68, 255})
	dungeonRect(x+24, y+34, 152, 100, color.RGBA{23, 58, 71, 255})
	for i := range 3 {
		phase := math.Mod(rl.GetTime()*.16+float64(i)/3+float64(x+y)*.001, 1)
		if phase < 0 {
			phase++
		}
		radius := float32(9 + phase*30)
		cx := int32(x + 55 + i*42)
		cy := int32(y + 57 + (i%2)*44)
		tint := rl.Fade(color.RGBA{79, 137, 150, 255}, float32(math.Sin(phase*math.Pi))*.38)
		rl.DrawEllipseLines(cx, cy, radius, radius*.3, tint)
	}
}
