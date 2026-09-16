package game

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"image/color"
	"math"
	"raylib-2d-sandbox/internal/terrain"
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
	for y := max(-96, top); y < bottom; y++ {
		for x := max(0, left); x < min(50, right); x++ {
			id := terrain.Cell(x, y)
			px, py := x*40, y*40
			hash := uint32(x*73856093) ^ uint32((y+100)*19349663)
			if id != 0 {
				tone := uint8(hash % 7)
				c := color.RGBA{26 + tone, 32 + tone, 40 + tone, 255}
				dungeonRect(px, py, 40, 40, c)
				dungeonRect(px, py, 40, 1, color.RGBA{43, 49, 57, 255})
				dungeonRect(px, py, 1, 40, color.RGBA{43, 49, 57, 255})
				// Worn inset tiles along the axial processional route.
				if (x == 24 || x == 25) && (id == 2 || id == 8 || id == 1) {
					dungeonRect(px+4, py+3, 32, 34, color.RGBA{47, 52, 58, 255})
					dungeonRect(px+7, py+6, 26, 1, color.RGBA{69, 70, 66, 255})
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
	// Clear the field border at the walkable gate opening.
	dungeonRect(880, 0, 240, 10, rl.Black)
	for _, x := range []int{800, 1120} {
		dungeonRect(x, -80, 80, 80, stone)
		dungeonRect(x+6, -74, 68, 5, paleStone)
		dungeonRect(x+12, -60, 56, 48, color.RGBA{31, 39, 49, 255})
		dungeonRect(x+34, -52, 12, 28, runeColor)
	}
	// Floor inlays make each room recognizable without adding props or rewards.
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
	drawDescentCircle()
}

func drawDescentCircle() {
	center := rl.NewVector2(1000, -3620)
	glow := float32(.65 + .15*math.Sin(rl.GetTime()*1.5))
	for _, r := range []float32{82, 90, 116} {
		rl.DrawCircleLines(int32(center.X), int32(center.Y), r, rl.Fade(runeColor, glow))
	}
	rl.DrawPolyLines(center, 6, 78, 30, rl.Fade(runeColor, glow))
	rl.DrawPolyLines(center, 3, 72, -90, rl.Fade(runeColor, glow))
	rl.DrawPolyLines(center, 3, 72, 90, rl.Fade(runeColor, glow))
	for i := 0; i < 12; i++ {
		a := float64(i) * math.Pi / 6
		x, y := center.X+103*float32(math.Cos(a)), center.Y+103*float32(math.Sin(a))
		rl.DrawPolyLines(rl.NewVector2(x, y), 4, 5, float32(i*30), runeColor)
	}
}

// A compact location caption leaves the architecture itself free of labels.
func drawDungeonLocation(x, y float32, width, height int32) {
	name, hint := "OUTER FIELD", "Head north to enter The Hollow"
	if y < 0 {
		name, hint = "THE HOLLOW / FLOOR I", "Entrance lies to the south"
		for _, r := range terrain.Rooms {
			if x >= float32(r.X*40) && x < float32((r.X+r.W)*40) && y >= float32(r.Y*40) && y < float32((r.Y+r.H)*40) {
				name = "I / " + r.Name
				break
			}
		}
		if y < -3480 {
			hint = "The circle is dormant. The deeper floors are yet to open."
		}
	}
	rl.DrawRectangle(0, height-70, width, 70, rl.Fade(rl.Black, .8))
	rl.DrawText(name, 24, height-57, 20, runeColor)
	rl.DrawText(hint, 24, height-29, 16, rl.LightGray)
}

// Keep the coping inside the blocked pool footprint so its edge matches collision.
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
	// Slowly expanding elliptical ripples keep the surface calm and legible.
	for i := 0; i < 3; i++ {
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
