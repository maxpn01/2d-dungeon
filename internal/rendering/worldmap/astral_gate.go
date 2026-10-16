package worldmap

import (
	"fmt"
	"math"

	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"
	mapdata "raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var astralGold = rl.NewColor(206, 171, 104, 255)
var astralViolet = rl.NewColor(162, 139, 238, 255)

func drawAstralSigil(at rl.Vector2, radius float32, tint rl.Color) {
	rl.DrawCircleLinesV(at, radius, rl.Fade(tint, .7))
	rl.DrawPolyLines(at, 4, radius*.7, 0, tint)
	rl.DrawCircleV(at, radius*.2, tint)
	for _, side := range []float32{-1, 1} {
		rl.DrawLineEx(rl.NewVector2(at.X+side*radius, at.Y-4), rl.NewVector2(at.X+side*(radius+7), at.Y-4), 2, tint)
		rl.DrawLineEx(rl.NewVector2(at.X+side*radius, at.Y+4), rl.NewVector2(at.X+side*(radius+7), at.Y+4), 2, tint)
	}
}

func (m *Map) drawAstralRoute() {
	if m.floor != mapdata.FirstFloor {
		return
	}
	room, ok := m.world.Room(mapdata.VaultCourt)
	if !ok {
		return
	}
	paving := map[[2]int]bool{}
	add := func(area mapdata.CellRect) {
		for y := area.Y; y < area.Bottom(); y++ {
			for x := area.X; x < area.Right(); x++ {
				if m.world.Cell(x, y) != 0 {
					paving[[2]int{x, y}] = true
				}
			}
		}
	}
	for _, id := range []string{"vault-gallery-approach", "vault-processional", "vault-court-entrance"} {
		if area, found := m.world.Passage(id); found {
			add(area)
		}
	}
	gate, _ := m.world.Gate(mapdata.VaultGate)
	entry, _ := m.world.Passage("vault-court-entrance")
	cx := (gate.Left + gate.Right) / (2 * mapdata.Tile)
	cy := entry.Y + entry.Height/2
	add(mapdata.Cells(room.Area.X, cy-2, cx-room.Area.X+3, 5))
	add(mapdata.Cells(cx-2, room.Area.Y, 5, cy-room.Area.Y+3))
	for cell := range paving {
		x, y := float32(cell[0]*mapdata.Tile), float32(cell[1]*mapdata.Tile)
		tone := uint8((cell[0]*13 + cell[1]*7) & 3)
		rl.DrawRectangleRec(rl.NewRectangle(x+1, y+1, mapdata.Tile-2, mapdata.Tile-2), rl.NewColor(37+tone, 39+tone, 48+tone, 255))
		for _, edge := range []struct {
			dx, dy         int
			ax, ay, bx, by float32
		}{
			{-1, 0, 0, 0, 0, mapdata.Tile}, {1, 0, mapdata.Tile, 0, mapdata.Tile, mapdata.Tile},
			{0, -1, 0, 0, mapdata.Tile, 0}, {0, 1, 0, mapdata.Tile, mapdata.Tile, mapdata.Tile},
		} {
			if !paving[[2]int{cell[0] + edge.dx, cell[1] + edge.dy}] && m.world.Cell(cell[0]+edge.dx, cell[1]+edge.dy) != 0 {
				rl.DrawLineEx(rl.NewVector2(x+edge.ax, y+edge.ay), rl.NewVector2(x+edge.bx, y+edge.by), 1, rl.Fade(astralGold, .3))
			}
		}
	}
	for y := room.Area.Y + 5; y < cy; y += 6 {
		at := mapdata.At(cx, y).Center()
		drawAstralSigil(rl.NewVector2(float32(at.X), float32(at.Y)), 12, rl.Fade(astralViolet, .55))
	}
	for _, cell := range []mapdata.Cell{mapdata.At(3, 3), mapdata.At(23, 3), mapdata.At(3, 27), mapdata.At(23, 27)} {
		p := room.WorldCell(cell).Center()
		at := rl.NewVector2(float32(p.X), float32(p.Y))
		rl.DrawEllipse(int32(at.X), int32(at.Y+26), 30, 12, rl.Fade(rl.Black, .45))
		rl.DrawRectangleRec(rl.NewRectangle(at.X-19, at.Y-26, 38, 52), stone)
		rl.DrawRectangleRec(rl.NewRectangle(at.X-25, at.Y-30, 50, 9), paleStone)
		rl.DrawRectangleRec(rl.NewRectangle(at.X-13, at.Y-21, 26, 38), rl.NewColor(48, 38, 74, 255))
		drawAstralSigil(rl.NewVector2(at.X, at.Y-4), 9, astralGold)
		rl.DrawCircleGradient(rl.NewVector2(at.X, at.Y-32), 75, rl.Fade(astralViolet, .17), rl.Blank)
	}
}

func (m *Map) drawVaultGate(orb simulation.OrbState, local simulation.Player) {
	if m.floor != mapdata.FirstFloor {
		return
	}
	gate, ok := m.world.Gate(mapdata.VaultGate)
	if !ok {
		return
	}
	left, right, y := float32(gate.Left), float32(gate.Right), float32(gate.Y)
	center := (left + right) / 2
	for _, x := range []float32{left - 48, right + 12} {
		rl.DrawRectangleRec(rl.NewRectangle(x-7, y+24, 50, 18), rl.NewColor(31, 32, 44, 255))
		rl.DrawRectangleRec(rl.NewRectangle(x, y-95, 36, 126), stone)
		rl.DrawRectangleRec(rl.NewRectangle(x+5, y-88, 26, 111), rl.NewColor(39, 39, 58, 255))
		rl.DrawRectangleRec(rl.NewRectangle(x+15, y-75, 6, 86), astralGold)
		rl.DrawRectangleRec(rl.NewRectangle(x-8, y-101, 52, 15), paleStone)
		rl.DrawPoly(rl.NewVector2(x+18, y-119), 4, 20, 0, astralGold)
		rl.DrawPoly(rl.NewVector2(x+18, y-119), 4, 13, 0, astralViolet)
	}
	rl.DrawRectangleRec(rl.NewRectangle(left-20, y-96, right-left+40, 19), stone)
	rl.DrawRectangleRec(rl.NewRectangle(left-20, y-96, right-left+40, 3), astralGold)
	rl.DrawLineEx(rl.NewVector2(left-24, y-96), rl.NewVector2(center, y-151), 9, stone)
	rl.DrawLineEx(rl.NewVector2(center, y-151), rl.NewVector2(right+24, y-96), 9, stone)
	rl.DrawLineEx(rl.NewVector2(left-24, y-101), rl.NewVector2(center, y-156), 2, astralGold)
	rl.DrawLineEx(rl.NewVector2(center, y-156), rl.NewVector2(right+24, y-101), 2, astralGold)
	drawAstralSigil(rl.NewVector2(center, y-119), 19, astralViolet)
	width := (right - left) / 2 * (1 - float32(orb.GateOpen))
	for side := 0; side < 2; side++ {
		x := left
		if side == 1 {
			x = right - width
		}
		if width < 1 {
			continue
		}
		rl.DrawRectangleRec(rl.NewRectangle(x, y-76, width, 94), rl.NewColor(24, 25, 39, 245))
		rl.DrawRectangleLinesEx(rl.NewRectangle(x, y-76, width, 94), 3, astralGold)
		for bar := float32(16); bar < width-5; bar += 20 {
			rl.DrawRectangleRec(rl.NewRectangle(x+bar, y-72, 4, 85), paleStone)
			rl.DrawRectangleRec(rl.NewRectangle(x+bar, y-28, 4, 10), astralViolet)
		}
		rl.DrawRectangleRec(rl.NewRectangle(x, y-57, width, 5), astralGold)
		rl.DrawRectangleRec(rl.NewRectangle(x, y-8, width, 5), astralGold)
		lockX := x + width - 13
		if side == 1 {
			lockX = x + 13
		}
		rl.DrawPoly(rl.NewVector2(lockX, y-31), 4, 14, 0, astralGold)
		ink := rl.NewColor(75, 66, 100, 255)
		if orb.GateUnlocked {
			ink = astralViolet
		}
		rl.DrawPoly(rl.NewVector2(lockX, y-31), 4, 9, 0, ink)
	}
	if !local.Dead && simulation.NearVaultGate(local.X, local.Y, simulation.PlayerSize) {
		label := fmt.Sprintf("%d GATE WARDENS REMAIN", orb.GateGuardsRemaining)
		if orb.GateUnlocked {
			label = "E / OPEN THE ASTRAL GATE"
		}
		if orb.GateActivated {
			label = "THE ASTRAL GATE OPENS"
		}
		if orb.GateOpen >= 1 {
			label = "THE ORB AWAITS WITHIN"
		}
		w := float32(primitives.MeasureText(label, 20))
		rl.DrawRectangleRec(rl.NewRectangle(center-w/2-12, y+52, w+24, 32), rl.Fade(rl.Black, .8))
		primitives.DrawTextAt(label, rl.NewVector2(center-w/2, y+59), 20, astralGold)
	}
}

func (m *Map) drawVaultSeals(orb simulation.OrbState) {
	if m.floor != mapdata.FirstFloor {
		return
	}
	center := rl.NewVector2(float32(orb.X), float32(orb.Y))
	for i, offset := range []rl.Vector2{rl.NewVector2(-180, -180), rl.NewVector2(180, -180), rl.NewVector2(-180, 180), rl.NewVector2(180, 180)} {
		at := rl.NewVector2(center.X+offset.X, center.Y+offset.Y)
		tint := rl.NewColor(71, 66, 94, 255)
		if orb.Seals[i] {
			tint = astralViolet
			pulse := float32(.8 + .2*math.Sin(rl.GetTime()*2+float64(i)))
			rl.DrawCircleGradient(at, 95, rl.Fade(tint, .3*pulse), rl.Blank)
			rl.DrawLineEx(at, center, 3, rl.Fade(tint, .45*pulse))
			for j := 0; j < 5; j++ {
				angle := rl.GetTime()*.9 + float64(j)*math.Pi*2/5
				rl.DrawCircleV(rl.NewVector2(at.X+float32(math.Cos(angle))*38, at.Y+float32(math.Sin(angle))*38), 2, rl.Fade(rl.RayWhite, .75))
			}
		}
		rl.DrawPoly(at, 8, 38, 22.5, rl.NewColor(29, 28, 43, 255))
		rl.DrawPolyLines(at, 8, 38, 22.5, astralGold)
		rl.DrawCircleLinesV(at, 30, tint)
		drawAstralSigil(at, 19, tint)
		label := []string{"I", "II", "III", "IV"}[i]
		w := primitives.MeasureText(label, 16)
		primitives.DrawTextAt(label, rl.NewVector2(at.X-float32(w)/2, at.Y+43), 16, tint)
	}
}
