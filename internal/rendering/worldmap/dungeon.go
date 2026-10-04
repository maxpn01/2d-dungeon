package worldmap

import (
	"image/color"
	"math"

	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"
	mapdata "raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var stone = color.RGBA{53, 62, 73, 255}
var paleStone = color.RGBA{93, 109, 119, 255}

func dungeonRect(x, y, w, h int, c color.RGBA) {
	rl.DrawRectangle(int32(x), int32(y), int32(w), int32(h), c)
}

func (m *Map) drawDungeon(ctx primitives.Context) {
	sanctumID, feastID := 0, 0
	rooms := m.world.Rooms()
	secretDoors := m.world.SecretDoors()
	northBorder := mapdata.CellRect{}
	for i, room := range rooms {
		if room.ID == mapdata.NorthSanctuary {
			area := room.Area
			northBorder = mapdata.Cells(area.X-1, area.Y-1, area.Width+2, area.Height+2)
		}
		if room.HasRole(mapdata.PortalSanctum) {
			sanctumID = i + 2
		}
		if room.ID == mapdata.FeastHall {
			feastID = i + 2
		}
	}
	zoom := max(ctx.Camera.Zoom, .1)
	left := int(math.Floor(float64(ctx.Camera.Target.X-ctx.Camera.Offset.X/zoom)/40)) - 1
	right := int(math.Ceil(float64(ctx.Camera.Target.X+(float32(ctx.WindowWidth)-ctx.Camera.Offset.X)/zoom)/40)) + 1
	top := int(math.Floor(float64(ctx.Camera.Target.Y-ctx.Camera.Offset.Y/zoom)/40)) - 1
	bottom := min(mapdata.MaxCellY, int(math.Ceil(float64(ctx.Camera.Target.Y+(float32(ctx.WindowHeight)-ctx.Camera.Offset.Y)/zoom)/40))+1)
	for y := max(mapdata.MinCellY, top); y < bottom; y++ {
		for x := max(mapdata.MinCellX, left); x < min(mapdata.MaxCellX, right); x++ {
			if m.sanctuaryTile(x, y, northBorder) || m.world.CellFloor(x, y) != m.floor {
				continue
			}
			id := m.world.Cell(x, y)
			px, py := x*40, y*40
			hash := uint32(x*73856093) ^ uint32((y+100)*19349663)
			if id != 0 {
				tone := uint8(hash % 7)
				c := color.RGBA{26 + tone, 32 + tone, 40 + tone, 255}
				if id == sanctumID {
					c = color.RGBA{22 + tone, 40 + tone, 46 + tone, 255}
				}
				if id == feastID {
					c = color.RGBA{33 + tone, 32 + tone, 36 + tone, 255}
				}
				dungeonRect(px, py, 40, 40, c)
				dungeonRect(px, py, 40, 1, color.RGBA{43, 49, 57, 255})
				dungeonRect(px, py, 1, 40, color.RGBA{43, 49, 57, 255})
				if (x == 17 || x == 18) && id == 1 && y >= -24 {
					dungeonRect(px+4, py+3, 32, 34, color.RGBA{47, 52, 58, 255})
					dungeonRect(px+7, py+6, 26, 1, color.RGBA{69, 70, 66, 255})
				}
				if m.world.Cover(x, y) {
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
				if m.world.Cell(x, y-1) == 0 {
					dungeonRect(px, py, 40, 10, color.RGBA{8, 13, 20, 150})
				}
			} else if m.world.Pool(x, y) {
				continue
			} else {
				near := false
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if m.world.Cell(x+dx, y+dy) != 0 {
							near = true
						}
					}
				}
				if !near {
					continue
				}
				sliding := false
				for _, door := range secretDoors {
					if door.Interactable && door.Area.Contains(mapdata.At(x, y)) {
						sliding = true
					}
				}
				if sliding {
					continue
				}
				dungeonRect(px, py, 40, 40, color.RGBA{17, 22, 30, 255})
				wallStone, wallHighlight := stone, paleStone
				cracked := false
				damage := 0
				for _, door := range secretDoors {
					if door.Area.Contains(mapdata.At(x, y)) {
						wallStone = color.RGBA{56, 64, 74, 255}
						wallHighlight = color.RGBA{96, 110, 120, 255}
						damage = m.wallHits[door.ID]
						cracked = !door.Interactable && (damage > 0 || y == door.Area.Y+door.Area.Height/2 && x == door.Area.X+door.Area.Width/2)
					}
				}
				dungeonRect(px+1, py+1, 38, 30, wallStone)
				dungeonRect(px+3, py+3, 34, 2, wallHighlight)
				dungeonRect(px+3, py+6, 2, 22, color.RGBA{67, 78, 87, 255})
				if cracked {
					drawSecretCrack(px, py, damage)
				}
				if m.world.Cell(x, y+1) != 0 {
					dungeonRect(px+1, py+31, 38, 9, color.RGBA{32, 39, 49, 255})
				}
			}
		}
	}
	if m.floor == mapdata.FirstFloor {
		southState := simulation.EntranceGateState{}
		if m.gate.ID == mapdata.EntranceGate {
			southState = m.gate
		}
		drawEntranceHall(southState, m.gatePrompt && m.nearGateID == mapdata.EntranceGate, m.world)
		northGate, _ := m.world.Gate(mapdata.NorthID(mapdata.EntranceGate))
		state := simulation.EntranceGateState{}
		if m.gate.ID == mapdata.NorthID(mapdata.EntranceGate) {
			state = m.gate
		}
		drawStatueGate(northGate, state, m.gatePrompt && m.nearGateID == mapdata.NorthID(mapdata.EntranceGate))

	}
	m.drawFeedingAreas()
	m.drawVault()
	m.drawSecretInscriptions(ctx)
	for _, pool := range m.world.Pools() {
		if m.world.CellFloor(pool.Area.X, pool.Area.Y) != m.floor {
			continue
		}
		if pool.Area.X+pool.Area.Width >= left && pool.Area.X <= right && pool.Area.Y+pool.Area.Height >= top && pool.Area.Y <= bottom {
			drawDungeonPool(pool.Area.X*mapdata.Tile, pool.Area.Y*mapdata.Tile, pool.Area.Width*mapdata.Tile, pool.Area.Height*mapdata.Tile)
		}
	}
	for _, room := range rooms {
		if room.Floor != m.floor {
			continue
		}
		if (room.ID == mapdata.WingChamber || room.ID == mapdata.NorthID(mapdata.WingChamber)) && m.world.Cell(room.Area.X, room.Area.Y) != 0 && room.Area.Right() >= left && room.Area.X <= right && room.Area.Bottom() >= top && room.Area.Y <= bottom {
			drawWingHall(room)
		}
		if room.ID == mapdata.FeastHall && room.Area.Y+room.Area.Height >= top && room.Area.Y <= bottom {
			drawFeastRoom(room)
		}
		if room.HasRole(mapdata.PortalSanctum) && room.Area.Y+room.Area.Height >= top && room.Area.Y <= bottom {
			if len(room.Portals) == 0 {
				continue
			}
			position := room.WorldCell(room.Portals[0].At).World()
			center := rl.NewVector2(float32(position.X), float32(position.Y))
			rl.DrawPolyLines(center, 8, 164, 22.5, rl.Fade(primitives.RuneColor, .35))
			rl.DrawCircleLinesV(center, 150, rl.Fade(primitives.RuneColor, .2))
			rl.DrawRectangleLinesEx(rl.NewRectangle(float32(room.Area.X*mapdata.Tile+24), float32(room.Area.Y*mapdata.Tile+24), float32(room.Area.Width*mapdata.Tile-48), float32(room.Area.Height*mapdata.Tile-48)), 2, rl.Fade(primitives.RuneColor, .25))
			edges := []int{room.Area.Bottom() * mapdata.Tile}
			if _, ok := m.world.Passage(mapdata.NorthID("sanctum-entrance")); ok {
				edges = append(edges, room.Area.Y*mapdata.Tile)
			}
			for _, y := range edges {
				for _, x := range []int{(room.Area.X+4)*mapdata.Tile - 24, (room.Area.Right()-4)*mapdata.Tile + 8} {
					dungeonRect(x, y-24, 16, 48, paleStone)
					dungeonRect(x+5, y-14, 6, 28, primitives.RuneColor)
				}
			}
		}
		if !room.HasRole(mapdata.BossArena) || room.Area.Y+room.Area.Height < top || room.Area.Y > bottom {
			continue
		}
		center := rl.NewVector2(float32((room.Area.X*2+room.Area.Width)*20), float32((room.Area.Y+room.Area.Height*2/3)*mapdata.Tile))
		rl.DrawPolyLines(center, 8, 105, 22.5, stone)
		rl.DrawCircleLinesV(center, 75, paleStone)
	}
	if m.floor == mapdata.FirstFloor {
		for _, id := range []string{"surface-entrance", mapdata.NorthID("surface-entrance")} {
			entrance, ok := m.world.Passage(id)
			if !ok {
				continue
			}
			pillarY := (entrance.Bottom() - 3) * mapdata.Tile
			if id != "surface-entrance" {
				pillarY = (entrance.Y + 1) * mapdata.Tile
			}
			for _, x := range []int{entrance.X*mapdata.Tile - 80, entrance.Right() * mapdata.Tile} {
				dungeonRect(x, pillarY, 80, 80, stone)
				dungeonRect(x+6, pillarY+6, 68, 5, paleStone)
				dungeonRect(x+12, pillarY+20, 56, 48, color.RGBA{31, 39, 49, 255})
				dungeonRect(x+34, pillarY+28, 12, 28, primitives.RuneColor)
			}

		}
	}

}

func drawEntranceHall(gate simulation.EntranceGateState, prompt bool, world *mapdata.Map) {
	const tile = mapdata.Tile
	border := color.RGBA{71, 86, 94, 255}
	gateBounds := world.EntranceGate()
	if hall, ok := world.Room(mapdata.GateHall); ok && gateBounds.Right > gateBounds.Left {
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

	line := func(x1, y1, x2, y2 int) {
		rl.DrawLineEx(rl.NewVector2(float32(x1), float32(y1)), rl.NewVector2(float32(x2), float32(y2)), 3, border)
	}
	room, hasRoom := world.Room(mapdata.ProvisionCourt)
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
	drawStatueGate(gateBounds, gate, prompt)

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

func drawSecretCrack(x, y, damage int) {
	tint := color.RGBA{26, 34, 43, 255}
	points := [][2]int{{20, 9}, {23, 13}, {21, 16}, {24, 21}}
	for i := 1; i < len(points); i++ {
		a, b := points[i-1], points[i]
		rl.DrawLine(int32(x+a[0]), int32(y+a[1]), int32(x+b[0]), int32(y+b[1]), tint)
	}
	rl.DrawLine(int32(x+22), int32(y+15), int32(x+27), int32(y+13), rl.Fade(tint, .75))
	rl.DrawPixel(int32(x+24), int32(y+14), color.RGBA{77, 87, 96, 255})
	light := rl.NewColor(152, 181, 207, 255)
	pulse := float32(.65 + .35*math.Sin(rl.GetTime()*1.7+float64(x+y)))
	center := rl.NewVector2(float32(x+22), float32(y+16))
	rl.DrawCircleGradient(center, float32(14+damage*4), rl.Fade(light, (.05+float32(damage)*.025)*pulse), rl.Blank)
	rl.DrawLineEx(rl.NewVector2(float32(x+21), float32(y+16)), rl.NewVector2(float32(x+24), float32(y+21)), 1, rl.Fade(light, (.22+float32(damage)*.12)*pulse))
	branches := [][4]int{{21, 16, 11, 23}, {23, 13, 31, 7}, {24, 21, 29, 29}, {11, 23, 5, 14}}
	for i := 0; i < min(damage, len(branches)); i++ {
		p := branches[i]
		a, b := rl.NewVector2(float32(x+p[0]), float32(y+p[1])), rl.NewVector2(float32(x+p[2]), float32(y+p[3]))
		rl.DrawLineEx(a, b, float32(2+damage/2), rl.NewColor(12, 19, 27, 255))
		rl.DrawLineEx(a, b, 1, rl.Fade(light, .12+float32(damage)*.1))
	}
	for i := 0; i < damage; i++ {
		dungeonRect(x+6+i*7, y+30+i%2*3, 4+i%2, 3, color.RGBA{100, 109, 117, 255})
	}
}

func drawFeastRoom(room mapdata.Room) {
	bounds := room.Bounds()
	tint := color.RGBA{181, 155, 97, 255}
	rl.DrawRectangleLinesEx(rl.NewRectangle(float32(bounds.X+18), float32(bounds.Y+18), float32(bounds.W-36), float32(bounds.H-36)), 2, rl.Fade(tint, .3))
	center := room.WorldCell(mapdata.At(room.Area.Width/2, room.Area.Height/2)).Center()
	at := rl.NewVector2(float32(center.X), float32(center.Y))
	pulse := float32(.7 + .12*math.Sin(rl.GetTime()*1.2))
	rl.DrawCircleGradient(at, 180, rl.Fade(tint, .10*pulse), rl.Blank)
	rl.DrawPolyLines(at, 8, 120, 22.5, rl.Fade(tint, .25*pulse))
	rl.DrawCircleLinesV(at, 96, rl.Fade(tint, .2*pulse))
	for _, offset := range []mapdata.Cell{mapdata.At(2, 2), mapdata.At(room.Area.Width-3, 2), mapdata.At(2, room.Area.Height-3), mapdata.At(room.Area.Width-3, room.Area.Height-3)} {
		p := room.WorldCell(offset).Center()
		pos := rl.NewVector2(float32(p.X), float32(p.Y))
		rl.DrawCircleGradient(pos, 52, rl.Fade(tint, .18*pulse), rl.Blank)
		rl.DrawPolyLines(pos, 4, 10, 45, rl.Fade(tint, .6*pulse))
	}
}

func (m *Map) drawFeedingAreas() {
	const tile = mapdata.Tile
	world := m.world
	for _, patch := range world.Patches() {
		if world.Cell(patch.At.X, patch.At.Y) == 0 || world.CellFloor(patch.At.X, patch.At.Y) != m.floor {
			continue
		}
		cx, cy := patch.At.X*tile+tile/2, patch.At.Y*tile+tile/2
		tint := color.RGBA{58, 115, 112, 255}
		if patch.Style == mapdata.VioletGarden {
			tint = color.RGBA{104, 93, 139, 255}
		}
		half := patch.Radius * tile
		dungeonRect(cx-half, cy-half, half*2, half*2, rl.Fade(tint, .18))
		rl.DrawRectangleLines(int32(cx-half), int32(cy-half), int32(half*2), int32(half*2), tint)
		rl.DrawRectangleLines(int32(cx-half+12), int32(cy-half+12), int32(half*2-24), int32(half*2-24), rl.Fade(tint, .5))
		for _, dx := range []int{-half, half - 8} {
			for _, dy := range []int{-half, half - 8} {
				dungeonRect(cx+dx, cy+dy, 8, 8, primitives.RuneColor)
			}
		}
		rl.DrawPolyLines(rl.NewVector2(float32(cx), float32(cy)), 4, 12, 45, primitives.RuneColor)
	}
}
