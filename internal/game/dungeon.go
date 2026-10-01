package game

import (
	"fmt"
	"image/color"
	"math"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var stone = color.RGBA{53, 62, 73, 255}
var paleStone = color.RGBA{93, 109, 119, 255}
var runeColor = color.RGBA{102, 196, 192, 255}

func dungeonRect(x, y, w, h int, c color.RGBA) {
	rl.DrawRectangle(int32(x), int32(y), int32(w), int32(h), c)
}

func (m *mapView) drawDungeon(ctx RenderContext) {
	sanctumID := 0
	for i, room := range worldmap.Default().Rooms() {
		if room.HasRole(worldmap.PortalSanctum) {
			sanctumID = i + 2
			break
		}
	}
	zoom := max(ctx.Camera.Zoom, .1)
	left := int(math.Floor(float64(ctx.Camera.Target.X-ctx.Camera.Offset.X/zoom)/40)) - 1
	right := int(math.Ceil(float64(ctx.Camera.Target.X+(float32(ctx.WindowWidth)-ctx.Camera.Offset.X)/zoom)/40)) + 1
	top := int(math.Floor(float64(ctx.Camera.Target.Y-ctx.Camera.Offset.Y/zoom)/40)) - 1
	bottom := min(0, int(math.Ceil(float64(ctx.Camera.Target.Y+(float32(ctx.WindowHeight)-ctx.Camera.Offset.Y)/zoom)/40))+1)
	for y := max(worldmap.MinCellY, top); y < bottom; y++ {
		for x := max(worldmap.MinCellX, left); x < min(worldmap.MaxCellX, right); x++ {
			id := worldmap.Default().Cell(x, y)
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
				if worldmap.Default().Cover(x, y) {
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
				if worldmap.Default().Cell(x, y-1) == 0 {
					dungeonRect(px, py, 40, 10, color.RGBA{8, 13, 20, 150})
				}
			} else if worldmap.Default().Pool(x, y) {
				continue
			} else {
				near := false
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if worldmap.Default().Cell(x+dx, y+dy) != 0 {
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
				if worldmap.Default().Cell(x, y+1) != 0 {
					dungeonRect(px+1, py+31, 38, 9, color.RGBA{32, 39, 49, 255})
				}
			}
		}
	}
	drawEntranceHall(m.gate, m.gatePrompt)
	for _, pool := range worldmap.Default().Pools() {
		if pool.Area.X+pool.Area.Width >= left && pool.Area.X <= right && pool.Area.Y+pool.Area.Height >= top && pool.Area.Y <= bottom {
			drawDungeonPool(pool.Area.X*worldmap.Tile, pool.Area.Y*worldmap.Tile, pool.Area.Width*worldmap.Tile, pool.Area.Height*worldmap.Tile)
		}
	}
	for _, room := range worldmap.Default().Rooms() {
		if room.HasRole(worldmap.PortalSanctum) && room.Area.Y+room.Area.Height >= top && room.Area.Y <= bottom {
			if len(room.Portals) == 0 {
				continue
			}
			position := room.WorldCell(room.Portals[0].At).World()
			center := rl.NewVector2(float32(position.X), float32(position.Y))
			rl.DrawPolyLines(center, 8, 164, 22.5, rl.Fade(runeColor, .35))
			rl.DrawCircleLinesV(center, 150, rl.Fade(runeColor, .2))
			rl.DrawRectangleLinesEx(rl.NewRectangle(float32(room.Area.X*worldmap.Tile+24), float32(room.Area.Y*worldmap.Tile+24), float32(room.Area.Width*worldmap.Tile-48), float32(room.Area.Height*worldmap.Tile-48)), 2, rl.Fade(runeColor, .25))
			for _, x := range []int{(room.Area.X+4)*worldmap.Tile - 24, (room.Area.Right()-4)*worldmap.Tile + 8} {
				y := (room.Area.Y + room.Area.Height) * worldmap.Tile
				dungeonRect(x, y-24, 16, 48, paleStone)
				dungeonRect(x+5, y-14, 6, 28, runeColor)
			}
		}
		if !room.HasRole(worldmap.BossArena) || room.Area.Y+room.Area.Height < top || room.Area.Y > bottom {
			continue
		}
		center := rl.NewVector2(float32((room.Area.X*2+room.Area.Width)*20), float32((room.Area.Y+13)*worldmap.Tile))
		rl.DrawPolyLines(center, 8, 105, 22.5, stone)
		rl.DrawCircleLinesV(center, 75, paleStone)
	}
	entrance, hasEntrance := worldmap.Default().Passage("surface-entrance")
	if !hasEntrance {
		return
	}
	for _, x := range []int{entrance.X*worldmap.Tile - 80, entrance.Right() * worldmap.Tile} {
		dungeonRect(x, (entrance.Bottom()-3)*worldmap.Tile, 80, 80, stone)
		dungeonRect(x+6, (entrance.Bottom()-3)*worldmap.Tile+6, 68, 5, paleStone)
		dungeonRect(x+12, (entrance.Bottom()-3)*worldmap.Tile+20, 56, 48, color.RGBA{31, 39, 49, 255})
		dungeonRect(x+34, (entrance.Bottom()-3)*worldmap.Tile+28, 12, 28, runeColor)
	}

}

func drawDungeonLocation(x, y float32, width, height int32) {
	name, hint := "SPAWN ZONE", "Head north through the dungeon entrance."
	if y < 0 {
		name, hint = "THE HOLLOW / FLOOR I", "Entrance lies to the south"
		passage, ok := worldmap.Default().Passage("surface-entrance")
		if ok && passage.ContainsWorld(float64(x), float64(y)) {
			name, hint = "I / ENTRANCE HALL", "The provision hall opens to the north."
		}
		for _, r := range worldmap.Default().Rooms() {
			if x >= float32(r.Area.X*40) && x < float32((r.Area.X+r.Area.Width)*40) && y >= float32(r.Area.Y*40) && y < float32((r.Area.Y+r.Area.Height)*40) {
				name = floorLabel(r.Floor) + " / " + r.Name
				if r.HasRole(worldmap.Private) {
					hint = "The statues guard the north gate. Press E nearby to open it."
				}
				if r.HasRole(worldmap.Refuge) {
					hint = "Arcane veils confuse creatures. They do not stop attacks."
				}
				if len(r.Portals) > 0 {
					hint = "The circle links the floors. Stand inside to travel."
				}
				break
			}
		}

	}
	if room, ok := worldmap.Default().RoomAt(float64(x), float64(y)); ok && room.Floor > worldmap.FirstFloor {
		name = floorLabel(room.Floor) + " / " + room.Name
		if len(room.Portals) > 0 {
			hint = "Leave and re-enter the circle to return to floor I."
		}
	}

	rl.DrawRectangle(0, height-70, width, 70, rl.Fade(rl.Black, .8))
	drawText(name, 24, height-57, 20, runeColor)
	drawText(hint, 24, height-29, 16, rl.LightGray)
}

func drawEntranceHall(gate simulation.EntranceGateState, prompt bool) {
	const tile = worldmap.Tile
	border := color.RGBA{71, 86, 94, 255}
	world := worldmap.Default()
	gateBounds := world.EntranceGate()
	if hall, ok := world.Room(worldmap.GateHall); ok && gateBounds.Right > gateBounds.Left {
		centerCol := (gateBounds.Left + gateBounds.Right) / (2 * tile)
		for row := hall.Area.Y; row < 0; row++ {
			for _, col := range []int{centerCol - 1, centerCol} {
				if world.Cell(col, row) == 0 {
					continue
				}
				dungeonRect(col*tile+4, row*tile+3, 32, 34, color.RGBA{47, 52, 58, 255})
				dungeonRect(col*tile+7, row*tile+6, 26, 1, color.RGBA{69, 70, 66, 255})
			}
		}
	}
	for _, patch := range worldmap.Default().Patches() {
		cx, cy := patch.At.X*tile+tile/2, patch.At.Y*tile+tile/2
		tint := color.RGBA{58, 115, 112, 255}
		if patch.Style == worldmap.VioletGarden {
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
	room, hasRoom := world.Room(worldmap.ProvisionCourt)
	entrance, hasPassage := world.Passage("surface-entrance")
	if hasRoom {
		area := room.Area
		left, right := area.X*tile+12, area.Right()*tile-12
		top, bottom := area.Y*tile+12, area.Bottom()*tile-12
		if gateBounds.Right > gateBounds.Left {
			line(left, top, gateBounds.Left, top)
			line(gateBounds.Right, top, right, top)
		} else {
			line(left, top, right, top)
		}
		line(left, top, left, bottom)
		line(right, top, right, bottom)
		if hasPassage {
			line(left, bottom, entrance.X*tile, bottom)
			line(entrance.Right()*tile, bottom, right, bottom)
		} else {
			line(left, bottom, right, bottom)
		}
	}
	drawStatueGate(gate, prompt)

}

func drawDungeonPool(x, y, width, height int) {
	sx, sy := float64(width)/200, float64(height)/160
	rect := func(px, py, w, h int, tint color.RGBA) {
		dungeonRect(x+int(float64(px-x)*sx), y+int(float64(py-y)*sy), max(1, int(float64(w)*sx)), max(1, int(float64(h)*sy)), tint)
	}
	rect(x, y, 200, 160, stone)
	rect(x+2, y+2, 196, 2, paleStone)
	rect(x+2, y+2, 2, 156, paleStone)
	rect(x+2, y+156, 196, 3, color.RGBA{25, 32, 41, 255})
	rect(x+196, y+2, 3, 154, color.RGBA{25, 32, 41, 255})
	for offset := 40; offset < 200; offset += 40 {
		rect(x+offset, y, 1, 8, color.RGBA{30, 39, 47, 255})
		rect(x+offset, y+152, 1, 8, color.RGBA{30, 39, 47, 255})
	}
	rect(x+8, y+8, 184, 144, color.RGBA{9, 24, 33, 255})
	rect(x+11, y+17, 178, 132, color.RGBA{19, 49, 62, 255})
	rect(x+16, y+24, 168, 120, color.RGBA{21, 55, 68, 255})
	rect(x+24, y+34, 152, 100, color.RGBA{23, 58, 71, 255})
	for i := range 3 {
		phase := math.Mod(rl.GetTime()*.16+float64(i)/3+float64(x+y)*.001, 1)
		if phase < 0 {
			phase++
		}
		radius := float32(9 + phase*30)
		cx := int32(x) + int32(float64(55+i*42)*sx)
		cy := int32(y) + int32(float64(57+(i%2)*44)*sy)
		tint := rl.Fade(color.RGBA{79, 137, 150, 255}, float32(math.Sin(phase*math.Pi))*.38)
		rl.DrawEllipseLines(cx, cy, radius*float32(sx), radius*.3*float32(sy), tint)
	}
}

func floorLabel(floor worldmap.Floor) string {
	switch floor {
	case worldmap.FirstFloor:
		return "I"
	case worldmap.SecondFloor:
		return "II"
	default:
		return fmt.Sprint(floor)
	}
}
