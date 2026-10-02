package worldmap

import (
	"math"

	"raylib-2d-sandbox/internal/rendering/primitives"
	mapdata "raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var sanctuaryFloor = rl.NewColor(14, 27, 31, 255)
var sanctuarySeam = rl.NewColor(22, 38, 42, 255)
var sanctuaryGold = rl.NewColor(133, 117, 70, 255)

func (m *Map) drawSanctuary(ctx primitives.Context) {
	m.drawSanctuaryAt(ctx, 0, false)
	if room, ok := m.world.Room(mapdata.NorthSanctuary); ok {
		m.drawSanctuaryAt(ctx, room.Area.Y*mapdata.Tile, true)
	}
}

func (m *Map) drawSanctuaryAt(ctx primitives.Context, originY int, north bool) {
	camera := ctx.Camera
	camera.Target.Y -= float32(originY)
	cs, rs, ce, re := m.visibleCellRange(camera, ctx.WindowWidth, ctx.WindowHeight)
	if cs == ce || rs == re {
		return
	}
	left, top := cs/4, rs/4
	right, bottom := (ce+3)/4, (re+3)/4
	w, h := int(m.size.X), int(m.size.Y)
	rl.DrawRectangle(0, int32(originY), int32(w), int32(h), sanctuaryFloor)
	for row := top; row < bottom; row++ {
		for col := left; col < right; col++ {
			x, y := col*mapdata.Tile, originY+row*mapdata.Tile
			patternRow := row
			if north {
				patternRow = mapdata.SurfaceCells - 1 - row
			}
			hash := uint32(col*73856093) ^ uint32(patternRow*19349663)
			tone := uint8(hash % 5)
			dungeonRect(x+1, y+1, 39, 39, rl.NewColor(14+tone, 27+tone, 31+tone, 255))
			dungeonRect(x, y, 40, 1, sanctuarySeam)
			dungeonRect(x, y, 1, 40, sanctuarySeam)
		}
	}
	var wingDoor mapdata.CellRect
	doorID := mapdata.WingDoor
	if north {
		doorID = mapdata.NorthID(doorID)
	}
	for _, door := range m.world.SecretDoors() {
		if door.ID == doorID {
			wingDoor = door.Area
		}
	}
	for col := left; col < right; col++ {
		entranceRow, entranceY, closedY := -1, originY, originY+h-10
		if north {
			entranceRow, entranceY, closedY = originY/mapdata.Tile+mapdata.SurfaceCells, originY+h-10, originY
		}
		connected := false
		for dx := -2; dx <= 2; dx++ {
			if mapdata.Default().Cell(col+dx, entranceRow) != 0 {
				connected = true
			}
		}
		if !connected {
			drawSanctuaryBorder(col*40, entranceY, 40, 10, col)
		}
		drawSanctuaryBorder(col*40, closedY, 40, 10, col)
	}
	for row := top; row < bottom; row++ {
		y := originY + row*40
		index := row
		if north {
			index = mapdata.SurfaceCells - 1 - row
		}
		if y/mapdata.Tile < wingDoor.Y || y/mapdata.Tile >= wingDoor.Bottom() {
			drawSanctuaryBorder(0, y, 10, 40, index)
		}
		drawSanctuaryBorder(w-10, y, 10, 40, index)
	}
	for _, p := range []rl.Vector2{{X: 22, Y: float32(originY + 22)}, {X: float32(w - 22), Y: float32(originY + 22)}, {X: 22, Y: float32(originY + h - 22)}, {X: float32(w - 22), Y: float32(originY + h - 22)}} {
		rl.DrawRectangle(int32(p.X-10), int32(p.Y-8), 20, 24, rl.NewColor(8, 17, 22, 255))
		rl.DrawRectangle(int32(p.X-8), int32(p.Y-10), 16, 20, stone)
		rl.DrawRectangle(int32(p.X-6), int32(p.Y-8), 12, 2, paleStone)
		drawSanctuaryRune(p, 5, rl.Fade(primitives.RuneColor, .75))
	}
}

func (m *Map) drawSanctuaryAltar(ctx primitives.Context) {
	drawSpawnAltar(rl.NewVector2(mapdata.SpawnAltarCenter, mapdata.SpawnAltarCenter))
	if room, ok := m.world.Room(mapdata.NorthSanctuary); ok {
		bounds := room.Bounds()
		drawSpawnAltar(rl.NewVector2(float32(bounds.X+bounds.W/2), float32(bounds.Y+bounds.H/2)))
	}
}

func drawSanctuaryBorder(x, y, w, h, index int) {
	dungeonRect(x, y, w, h, rl.NewColor(9, 18, 23, 255))
	dungeonRect(x+1, y+1, w-2, h-2, stone)
	dungeonRect(x+2, y+2, w-4, 1, rl.Fade(paleStone, .5))
	horizontal := w > h
	if horizontal {
		dungeonRect(x+2, y+h-2, w-4, 1, rl.Fade(primitives.RuneColor, .35))
	} else {
		dungeonRect(x+w-2, y+2, 1, h-4, rl.Fade(primitives.RuneColor, .35))
	}
	if index%4 == 1 {
		drawSanctuaryRune(rl.NewVector2(float32(x+w/2), float32(y+h/2)), 3, rl.Fade(sanctuaryGold, .75))
	}
}

func drawSanctuaryRune(p rl.Vector2, size float32, c rl.Color) {
	rl.DrawPolyLines(p, 4, size, 0, c)
	rl.DrawLineV(rl.NewVector2(p.X, p.Y-size-2), rl.NewVector2(p.X, p.Y+size+2), c)
}

func sanctuaryPoint(center rl.Vector2, angle, radius float64) rl.Vector2 {
	return rl.NewVector2(center.X+float32(math.Cos(angle)*radius), center.Y+float32(math.Sin(angle)*radius))
}

func drawSpawnAltar(center rl.Vector2) {
	t := rl.GetTime()
	pulse := float32(.65 + .15*math.Sin(t*1.4))
	rl.DrawCircleGradient(center, 140, rl.Fade(primitives.RuneColor, .06*pulse), rl.Blank)
	for i := 0; i < 4; i++ {
		angle := float64(i) * math.Pi / 2
		start, end := sanctuaryPoint(center, angle, 58), sanctuaryPoint(center, angle, 132)
		rl.DrawLineEx(start, end, 4, rl.NewColor(9, 21, 25, 255))
		rl.DrawLineEx(start, end, 1, rl.Fade(primitives.RuneColor, .22*pulse))
		drawSanctuaryRune(end, 5, rl.Fade(sanctuaryGold, .55))
	}
	rl.DrawRing(center, 75, 77, 0, 360, 96, rl.Fade(sanctuaryGold, .32))
	for i := 0; i < 8; i++ {
		angle := float32(i) * 45
		rl.DrawRing(center, 65, 66, angle+5, angle+32, 12, rl.Fade(primitives.RuneColor, .5*pulse))
		drawSanctuaryRune(sanctuaryPoint(center, float64(i)*math.Pi/4, 86), 5, rl.Fade(primitives.RuneColor, .55*pulse))
	}
	shadow := rl.NewVector2(center.X, center.Y+8)
	rl.DrawPoly(shadow, 8, 52, 22.5, rl.NewColor(7, 16, 21, 255))
	rl.DrawPoly(center, 8, mapdata.SpawnAltarRadius, 22.5, stone)
	rl.DrawPolyLines(center, 8, 48, 22.5, rl.Fade(paleStone, .6))
	rl.DrawPoly(center, 8, 36, 22.5, rl.NewColor(30, 43, 49, 255))
	rl.DrawPolyLines(center, 8, 35, 22.5, sanctuaryGold)
	rl.DrawPoly(center, 8, 25, 22.5, rl.NewColor(43, 58, 64, 255))
	rl.DrawPolyLines(center, 8, 24, 22.5, rl.Fade(paleStone, .65))
	rl.DrawRing(center, 17, 18, 0, 360, 48, rl.Fade(primitives.RuneColor, pulse))
	crystal := rl.NewVector2(center.X, center.Y-10+float32(math.Sin(t*1.8))*3)
	rl.DrawCircleGradient(crystal, 32, rl.Fade(primitives.RuneColor, .22*pulse), rl.Blank)
	upper := rl.NewVector2(crystal.X, crystal.Y-17)
	lower := rl.NewVector2(crystal.X, crystal.Y+14)
	left := rl.NewVector2(crystal.X-9, crystal.Y)
	right := rl.NewVector2(crystal.X+9, crystal.Y)
	rl.DrawTriangle(upper, left, lower, rl.NewColor(72, 158, 161, 255))
	rl.DrawTriangle(upper, lower, right, rl.NewColor(125, 218, 207, 255))
	rl.DrawLineV(upper, lower, rl.NewColor(182, 240, 218, 255))
	for i := 0; i < 7; i++ {
		phase := math.Mod(t*.18+float64(i)/7, 1)
		p := sanctuaryPoint(center, float64(i)*2.4+t*.15, 22+phase*17)
		p.Y -= float32(phase * 30)
		rl.DrawCircleV(p, 1.2, rl.Fade(primitives.RuneColor, float32(math.Sin(phase*math.Pi))*.7))
	}
}
