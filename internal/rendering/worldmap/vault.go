package worldmap

import (
	"math"

	"2d-dungeon/internal/simulation"
	mapdata "2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (m *Map) drawVault() {
	room, ok := m.world.Room(mapdata.OrbVault)
	if !ok || room.Floor != m.floor || m.world.Cell(room.Area.X, room.Area.Y) == 0 {
		return
	}
	center := m.world.OrbPosition()
	at := rl.NewVector2(float32(center.X), float32(center.Y))
	rune := rl.NewColor(156, 128, 227, 255)
	pulse := float32(.75 + .25*math.Sin(rl.GetTime()*1.2))
	m.drawVaultApproach(room, rune)
	rl.DrawCircleGradient(at, 210, rl.Fade(rune, .12*pulse), rl.Blank)
	rl.DrawPoly(at, 8, 105, 22.5, rl.NewColor(37, 35, 63, 255))
	for _, radius := range []float32{87, 105, 145} {
		rl.DrawCircleLinesV(at, radius, rl.Fade(rune, .3*pulse))
	}
	for i := 0; i < 12; i++ {
		angle := float64(i) * math.Pi / 6
		pos := rl.NewVector2(at.X+float32(math.Cos(angle))*125, at.Y+float32(math.Sin(angle))*125)
		rl.DrawPolyLines(pos, 4, 5, float32(i*30)+45, rl.Fade(rune, .7*pulse))
	}
	m.drawVaultBorder(room, rune)
	for _, cell := range []mapdata.Cell{mapdata.At(2, 2), mapdata.At(room.Area.Width-3, 2), mapdata.At(2, room.Area.Height-3), mapdata.At(room.Area.Width-3, room.Area.Height-3)} {
		p := room.WorldCell(cell).Center()
		pos := rl.NewVector2(float32(p.X), float32(p.Y))
		rl.DrawPoly(pos, 6, 24, 30, stone)
		rl.DrawPolyLines(pos, 6, 24, 30, paleStone)
		rl.DrawCircleGradient(pos, 90, rl.Fade(rune, .22*pulse), rl.Blank)
		rl.DrawPoly(rl.NewVector2(pos.X, pos.Y-20), 4, 10, 45, rune)
		rl.DrawLineEx(rl.NewVector2(pos.X, pos.Y-32), rl.NewVector2(pos.X, pos.Y-8), 2, rl.RayWhite)
	}
	for i := 0; i < 24; i++ {
		angle := float64(i)*2.399 + rl.GetTime()*.07
		radius := float32(170 + i%5*48)
		pos := rl.NewVector2(at.X+float32(math.Cos(angle))*radius, at.Y+float32(math.Sin(angle))*radius+float32(math.Sin(rl.GetTime()+float64(i)))*10)
		rl.DrawCircleV(pos, 1.5, rl.Fade(rune, .3+.35*float32(math.Sin(rl.GetTime()+float64(i))*.5+.5)))
	}
}

func (m *Map) drawVaultTraps(snapshot simulation.Snapshot) {
	if !m.showsRoom(mapdata.OrbVault) {
		return
	}
	rune := rl.NewColor(156, 128, 227, 255)
	for i, trap := range m.world.VaultTraps() {
		p := trap.Tile.World()
		dungeonRect(int(p.X)+3, int(p.Y)+3, 34, 34, rl.NewColor(42, 44, 52, 255))
		rl.DrawRectangleLinesEx(rl.NewRectangle(float32(p.X+5), float32(p.Y+5), 30, 30), 1, rl.NewColor(48, 50, 57, 255))
		rl.DrawLineEx(rl.NewVector2(float32(p.X+7), float32(p.Y+34)), rl.NewVector2(float32(p.X+33), float32(p.Y+34)), 2, rl.NewColor(29, 31, 38, 255))
		if i >= len(snapshot.VaultFire) {
			continue
		}
		state := snapshot.VaultFire[i]
		if state.TriggeredTick == 0 || snapshot.Tick < state.TriggeredTick || snapshot.Tick-state.TriggeredTick >= simulation.VaultTrapWarningTicks+simulation.VaultTrapFireTicks {
			continue
		}
		for _, vent := range vaultVents(trap) {
			rl.DrawPoly(vent, 6, 22, 30, stone)
			rl.DrawPolyLines(vent, 6, 22, 30, paleStone)
			rl.DrawCircleV(vent, 10, rl.NewColor(17, 13, 26, 255))
			rl.DrawLineEx(rl.NewVector2(vent.X-8, vent.Y-14), rl.NewVector2(vent.X+8, vent.Y-14), 3, rune)
		}
	}

	if snapshot.VaultArrowStarted == 0 || snapshot.Tick < snapshot.VaultArrowStarted || snapshot.Tick-snapshot.VaultArrowStarted >= simulation.VaultArrowDurationTicks {
		return
	}
	for _, port := range m.world.VaultArrowPorts() {
		at := rl.NewVector2(float32(port.X), float32(port.Y))
		rl.DrawPoly(at, 4, 12, 45, stone)
		rl.DrawPolyLines(at, 4, 12, 45, paleStone)
		rl.DrawLineEx(rl.NewVector2(at.X-3, at.Y-6), rl.NewVector2(at.X+3, at.Y+6), 4, rl.NewColor(12, 16, 24, 255))
	}
}

func (m *Map) DrawVaultArrowWarning(snapshot simulation.Snapshot) {
	if snapshot.VaultArrowStarted == 0 || snapshot.Tick < snapshot.VaultArrowStarted {
		return
	}
	age := snapshot.Tick - snapshot.VaultArrowStarted
	if age >= simulation.VaultArrowDurationTicks {
		return
	}
	for _, port := range m.world.VaultArrowPorts() {
		if !m.Exploration.EntityVisible(port.X, port.Y, 0) {
			continue
		}
		at := rl.NewVector2(float32(port.X), float32(port.Y))
		pulse := float32(.4 + .3*math.Sin(rl.GetTime()*16))
		rl.DrawCircleGradient(at, 24, rl.Fade(rl.Red, pulse), rl.Blank)
		rl.DrawPolyLines(at, 4, 13, 45, rl.Fade(rl.Orange, pulse+.2))
	}
}

func vaultVents(trap mapdata.VaultTrap) []rl.Vector2 {
	b := trap.Lane
	if b.H > b.W {
		return []rl.Vector2{rl.NewVector2(float32(b.X+b.W/2), float32(b.Y)), rl.NewVector2(float32(b.X+b.W/2), float32(b.Y+b.H))}
	}
	return []rl.Vector2{rl.NewVector2(float32(b.X), float32(b.Y+b.H/2)), rl.NewVector2(float32(b.X+b.W), float32(b.Y+b.H/2))}
}

func (m *Map) DrawVaultFire(snapshot simulation.Snapshot) {
	traps := m.world.VaultTraps()
	for i, state := range snapshot.VaultFire {
		if i >= len(traps) || state.TriggeredTick == 0 || snapshot.Tick < state.TriggeredTick {
			continue
		}
		trap := traps[i]
		center := trap.Tile.Center()
		if !m.Exploration.EntityVisible(center.X, center.Y, 0) {
			continue
		}
		age := int(snapshot.Tick - state.TriggeredTick)
		if age >= simulation.VaultTrapWarningTicks+simulation.VaultTrapFireTicks {
			continue
		}
		vents := vaultVents(trap)
		if age < simulation.VaultTrapWarningTicks {
			pulse := float32(.35 + .25*math.Sin(rl.GetTime()*25))
			for _, vent := range vents {
				rl.DrawCircleGradient(vent, 35, rl.Fade(rl.Orange, pulse), rl.Blank)
				rl.DrawCircleV(vent, 7, rl.Orange)
			}
			b := trap.Lane
			rl.DrawRectangleRec(rl.NewRectangle(float32(b.X), float32(b.Y), float32(b.W), float32(b.H)), rl.Fade(rl.Orange, .06))
			continue
		}
		life := float32(age-simulation.VaultTrapWarningTicks) / float32(simulation.VaultTrapFireTicks)
		fade := min(float32(1), (1-life)*5)
		for side, vent := range vents {
			length := float32(trap.Lane.W / 2)
			dir := rl.NewVector2(1, 0)
			if trap.Lane.H > trap.Lane.W {
				length, dir = float32(trap.Lane.H/2), rl.NewVector2(0, 1)
			}
			if side == 1 {
				dir.X = -dir.X
				dir.Y = -dir.Y
			}
			end := rl.NewVector2(vent.X+dir.X*length, vent.Y+dir.Y*length)
			rl.DrawLineEx(vent, end, 75, rl.Fade(rl.Red, .12*fade))
			rl.DrawLineEx(vent, end, 48, rl.Fade(rl.Orange, .24*fade))
			for j := 0; j < 55; j++ {
				phase := float32(math.Mod(float64(j)*.618+rl.GetTime()*2.8, 1))
				distance := phase * length
				wave := float32(math.Sin(float64(j)*12.9898+math.Floor(rl.GetTime()*18)*3.14)) * (5 + phase*13)
				pos := rl.NewVector2(vent.X+dir.X*distance-dir.Y*wave, vent.Y+dir.Y*distance+dir.X*wave)
				radius := (9 + phase*10) * fade
				rl.DrawCircleGradient(pos, radius*2, rl.Fade(rl.Red, .3*fade), rl.Blank)
				rl.DrawCircleV(pos, radius, rl.Fade(rl.NewColor(255, 91, 22, 255), .5*fade))
				rl.DrawCircleV(pos, radius*.6, rl.Fade(rl.Gold, .4*fade))
				flame := func(scale float32, tint rl.Color) {
					a := rl.NewVector2(pos.X+dir.X*radius*2.4*scale, pos.Y+dir.Y*radius*2.4*scale)
					b := rl.NewVector2(pos.X-dir.X*radius*.7*scale-dir.Y*radius*scale, pos.Y-dir.Y*radius*.7*scale+dir.X*radius*scale)
					c := rl.NewVector2(pos.X-dir.X*radius*.7*scale+dir.Y*radius*scale, pos.Y-dir.Y*radius*.7*scale-dir.X*radius*scale)
					rl.DrawTriangle(a, c, b, tint)
				}
				if j%3 == 0 {
					flame(1, rl.Fade(rl.NewColor(255, 91, 22, 255), .7*fade))
					flame(.65, rl.Fade(rl.Gold, .8*fade))
					flame(.3, rl.Fade(rl.NewColor(255, 242, 177, 255), fade))
				}
				if j%3 == 0 {
					ember := rl.NewVector2(pos.X-dir.Y*32*phase, pos.Y+dir.X*32*phase-phase*30)
					rl.DrawLineEx(ember, rl.NewVector2(ember.X-dir.X*9, ember.Y-dir.Y*9), 2, rl.Fade(rl.Gold, (1-phase)*fade))
				}
			}
		}
	}
}
