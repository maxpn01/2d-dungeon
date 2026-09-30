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
	sanctumID := 0
	for i, room := range terrain.Rooms {
		if room.ID == "the-descent" {
			sanctumID = i + 2
			break
		}
	}
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
				if id == sanctumID {
					c = color.RGBA{22 + tone, 40 + tone, 46 + tone, 255}
				}
				dungeonRect(px, py, 40, 40, c)
				dungeonRect(px, py, 40, 1, color.RGBA{43, 49, 57, 255})
				dungeonRect(px, py, 1, 40, color.RGBA{43, 49, 57, 255})
				if (x == 17 || x == 18) && id == 1 && y >= -24 {
					dungeonRect(px+4, py+3, 32, 34, color.RGBA{47, 52, 58, 255})
					dungeonRect(px+7, py+6, 26, 1, color.RGBA{69, 70, 66, 255})
				}
				if terrain.Cover(x, y) {
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
				continue
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
	if top <= 0 && bottom >= -76 && left <= 34 && right >= 2 {
		drawEntranceHall()
	}
	for _, pool := range terrain.Pools {
		if pool.X+pool.W >= left && pool.X <= right && pool.Y+pool.H >= top && pool.Y <= bottom {
			drawDungeonPool(pool.X*terrain.Tile, pool.Y*terrain.Tile)
		}
	}
	for _, room := range terrain.Rooms {
		if room.ID == "the-descent" && room.Y+room.H >= top && room.Y <= bottom {
			center := rl.NewVector2(terrain.DescentCircleX, terrain.DescentCircleY)
			rl.DrawPolyLines(center, 8, 164, 22.5, rl.Fade(runeColor, .35))
			rl.DrawCircleLinesV(center, 150, rl.Fade(runeColor, .2))
			rl.DrawRectangleLinesEx(rl.NewRectangle(float32(room.X*terrain.Tile+24), float32(room.Y*terrain.Tile+24), float32(room.W*terrain.Tile-48), float32(room.H*terrain.Tile-48)), 2, rl.Fade(runeColor, .25))
			for _, x := range []int{16*terrain.Tile - 24, 20*terrain.Tile + 8} {
				y := (room.Y + room.H) * terrain.Tile
				dungeonRect(x, y-24, 16, 48, paleStone)
				dungeonRect(x+5, y-14, 6, 28, runeColor)
			}
		}
		if room.ID != "the-crown-boss-arena" || room.Y+room.H < top || room.Y > bottom {
			continue
		}
		center := rl.NewVector2(float32((room.X*2+room.W)*20), -99*terrain.Tile)
		rl.DrawPolyLines(center, 8, 105, 22.5, stone)
		rl.DrawCircleLinesV(center, 75, paleStone)
	}
	for _, x := range []int{520, 840} {
		dungeonRect(x, -80, 80, 80, stone)
		dungeonRect(x+6, -74, 68, 5, paleStone)
		dungeonRect(x+12, -60, 56, 48, color.RGBA{31, 39, 49, 255})
		dungeonRect(x+34, -52, 12, 28, runeColor)
	}

}

func drawDungeonLocation(x, y float32, width, height int32) {
	name, hint := "SPAWN ZONE", "Head north through the dungeon entrance."
	if y < 0 {
		name, hint = "THE HOLLOW / FLOOR I", "Entrance lies to the south"
		if x >= 15*terrain.Tile && x < 21*terrain.Tile && y >= -24*terrain.Tile {
			name, hint = "I / ENTRANCE HALL", "The provision hall opens to the north."
		}
		for _, r := range terrain.Rooms {
			if x >= float32(r.X*40) && x < float32((r.X+r.W)*40) && y >= float32(r.Y*40) && y < float32((r.Y+r.H)*40) {
				name = "I / " + r.Name
				if r.ID == "the-provision-hall" || r.ID == "western-garden" || r.ID == "eastern-garden" {
					hint = "Gather food in the gardens. The north gate leads to the gate hall."
				}
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

func drawEntranceHall() {
	const tile = terrain.Tile
	border := color.RGBA{71, 86, 94, 255}
	for row := -64; row < 0; row++ {
		for _, col := range []int{17, 18} {
			dungeonRect(col*tile+4, row*tile+3, 32, 34, color.RGBA{47, 52, 58, 255})
			dungeonRect(col*tile+7, row*tile+6, 26, 1, color.RGBA{69, 70, 66, 255})
		}
	}
	for _, patch := range terrain.Patches {
		cx, cy := patch.X*tile+tile/2, patch.Y*tile+tile/2
		tint := color.RGBA{58, 115, 112, 255}
		if patch.ID == "northeast" || patch.ID == "southwest" {
			tint = color.RGBA{104, 93, 139, 255}
		}
		half := patch.Radius * tile
		dungeonRect(cx-half, cy-half, half*2, half*2, rl.Fade(tint, .18))
		rl.DrawRectangleLines(int32(cx-half), int32(cy-half), int32(half*2), int32(half*2), tint)
		rl.DrawRectangleLines(int32(cx-half+12), int32(cy-half+12), int32(half*2-24), int32(half*2-24), rl.Fade(tint, .5))
		for _, dx := range []int{-half, half - 8} {
			for _, dy := range []int{-half, half - 8} {
				dungeonRect(cx+dx, cy+dy, 8, 8, runeColor)
			}
		}
		rl.DrawPolyLines(rl.NewVector2(float32(cx), float32(cy)), 4, 12, 45, runeColor)
	}
	line := func(x1, y1, x2, y2 int) {
		rl.DrawLineEx(rl.NewVector2(float32(x1), float32(y1)), rl.NewVector2(float32(x2), float32(y2)), 3, border)
	}
	line(2*tile+12, -56*tile+12, 14*tile, -56*tile+12)
	line(22*tile, -56*tile+12, 34*tile-12, -56*tile+12)
	line(2*tile+12, -56*tile+12, 2*tile+12, -24*tile-12)
	line(34*tile-12, -56*tile+12, 34*tile-12, -24*tile-12)
	line(2*tile+12, -24*tile-12, 15*tile, -24*tile-12)
	line(21*tile, -24*tile-12, 34*tile-12, -24*tile-12)
	gateX, gateY := terrain.EntranceGateLeft, terrain.EntranceGateY
	gateW := terrain.EntranceGateRight - gateX
	for _, x := range []int{gateX - tile, gateX + gateW} {
		dungeonRect(x-4, gateY-48, tile+8, 96, color.RGBA{22, 29, 38, 255})
		dungeonRect(x, gateY-44, tile, 88, stone)
		dungeonRect(x+4, gateY-40, tile-8, 5, paleStone)
		dungeonRect(x+14, gateY-22, 12, 44, runeColor)
	}
	dungeonRect(gateX, gateY-34, gateW, 68, color.RGBA{24, 31, 44, 255})
	dungeonRect(gateX, gateY-34, gateW, 5, paleStone)
	dungeonRect(gateX, gateY+29, gateW, 5, paleStone)
	phase := rl.GetTime()
	pulse := float32(.48 + .10*math.Sin(phase*1.6))
	for row := -28; row < 28; row++ {
		distance := float64(row) / 28
		glow := float32(1 - distance*distance)
		shimmer := float32(.025 * math.Sin(phase*2+float64(row)*.06))
		tint := color.RGBA{143, 123, 220, 255}
		dungeonRect(gateX, gateY+row, gateW, 1, rl.Fade(tint, pulse*(.65+.35*glow)+shimmer))
	}

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
