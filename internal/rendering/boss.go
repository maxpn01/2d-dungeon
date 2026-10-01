package rendering

import (
	"math"

	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawEncounter(b simulation.Boss, e Visibility) {
	for _, portal := range worldmap.Default().Portals() {
		position := portal.Position()
		pad := rl.NewVector2(float32(position.X), float32(position.Y))
		if !e.EntityVisible(float64(pad.X), float64(pad.Y), 0) {
			continue
		}
		active := b.Phase == "" || b.Defeats > 0
		tint := rl.Gray
		label := "DEFEAT THE CROWN DEVOURER"
		if active {
			tint = primitives.RuneColor
			label = "STAND IN THE CIRCLE TO DESCEND"
			destination, _ := worldmap.Default().Portal(portal.Destination)
			if destination.Floor < portal.Floor {
				label = "STAND IN THE CIRCLE TO RETURN"
			}
		}
		pulse := float32(.65 + .2*math.Sin(rl.GetTime()*3))
		if active {
			rl.DrawCircleGradient(pad, 110, rl.Fade(tint, .22*pulse), rl.Blank)
		}
		for _, radius := range []float32{60, 82, 90} {
			rl.DrawCircleLines(int32(pad.X), int32(pad.Y), radius, rl.Fade(tint, pulse))
		}
		rotation := float32(30)
		if active {
			rotation += float32(rl.GetTime() * 25)
		}
		rl.DrawPolyLines(pad, 6, 78, rotation, tint)
		primitives.DrawTextAt(label, rl.NewVector2(pad.X-float32(primitives.MeasureText(label, 14))/2, pad.Y+105), 14, tint)
	}
	if b.Phase == "" || b.Phase == "waiting" || b.Phase == "defeated" || !e.EntityVisible(b.X, b.Y, simulation.BossSize) {
		return
	}
	if b.Phase == "fighting" && b.Behavior == "winding" {
		start := rl.NewVector2(float32(b.X+simulation.BossSize/2), float32(b.Y+simulation.BossSize/2))
		target := rl.NewVector2(float32(b.AttackTarget.X+simulation.BossSize/2), float32(b.AttackTarget.Y+simulation.BossSize/2))
		rl.DrawLineEx(start, target, simulation.BossSize, rl.Fade(rl.Orange, .12))
		rl.DrawLineEx(start, target, 4, rl.Fade(rl.Orange, .85))
		rl.DrawCircleLines(int32(target.X), int32(target.Y), 24, rl.Orange)
	}
	tint := rl.NewColor(161, 97, 191, 255)
	if b.Enraged {
		tint = rl.NewColor(225, 60, 85, 255)
	}
	if b.Accelerating {
		tint = rl.Orange
	}
	for i := len(b.Tail) - 1; i >= 0; i-- {
		s := b.Tail[i]
		r := rl.NewRectangle(float32(s.X), float32(s.Y), simulation.BossSize, simulation.BossSize)
		rl.DrawRectangleRec(r, rl.Fade(tint, .85))
		rl.DrawRectangleLinesEx(r, 2, rl.Fade(rl.Black, .5))
		rl.DrawCircleV(rl.NewVector2(r.X+18, r.Y+18), 4, rl.Fade(rl.Gold, .5))
	}
	center := rl.NewVector2(float32(b.X+simulation.BossSize/2), float32(b.Y+simulation.BossSize/2))
	if b.Enraged {
		rl.DrawRing(center, 43, 46, 0, 360, 40, rl.Fade(rl.Red, float32(.5+.2*math.Sin(rl.GetTime()*12))))
	}
	if b.Accelerating {
		direction := rl.Vector2Normalize(rl.Vector2Subtract(rl.NewVector2(float32(b.AttackTarget.X+simulation.BossSize/2), float32(b.AttackTarget.Y+simulation.BossSize/2)), center))
		for i := 1; i <= 4; i++ {
			trail := rl.Vector2Subtract(center, rl.Vector2Scale(direction, float32(i)*18))
			rl.DrawCircleV(trail, float32(18-i*2), rl.Fade(tint, float32(5-i)*.08))
		}
	}
	rl.DrawCircleGradient(center, 48, rl.Fade(tint, .35), rl.Blank)
	rl.DrawRectangleRec(rl.NewRectangle(float32(b.X), float32(b.Y), simulation.BossSize, simulation.BossSize), tint)
	rl.DrawRectangleLinesEx(rl.NewRectangle(float32(b.X), float32(b.Y), simulation.BossSize, simulation.BossSize), 2, rl.Gold)
	rl.DrawCircleV(rl.NewVector2(center.X-8, center.Y-5), 4, rl.RayWhite)
	rl.DrawCircleV(rl.NewVector2(center.X+8, center.Y-5), 4, rl.RayWhite)
}

func DrawTeleportCharge(x, y, charge float64) {
	if charge <= 0 {
		return
	}
	center := rl.NewVector2(float32(x+simulation.PlayerSize/2), float32(y+simulation.PlayerSize/2))
	progress := float32(min(1, charge/simulation.PortalSeconds))
	rl.DrawCircleGradient(center, 80, rl.Fade(primitives.RuneColor, progress*.65), rl.Blank)
	rl.DrawRing(center, 35, 40, -90, -90+360*progress, 48, rl.RayWhite)
}
