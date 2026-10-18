package rendering

import (
	"math"

	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"
	"2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawTeleportCharge(x, y, charge float64) {
	if charge <= 0 {
		return
	}
	player := rl.NewVector2(float32(x+simulation.PlayerSize/2), float32(y+simulation.PlayerSize/2))
	center := player
	for _, portal := range worldmap.Default().Portals() {
		position := portal.Position()
		if math.Hypot(position.X-float64(player.X), position.Y-float64(player.Y)) <= 60 {
			center = rl.NewVector2(float32(position.X), float32(position.Y))
			break
		}
	}
	progress := float32(min(1, charge/simulation.PortalSeconds))
	drawTeleportEnergy(player, center, charge, progress*progress, 1)
}

func drawTeleportEnergy(player, center rl.Vector2, charge float64, energy, opacity float32) {
	rotation := float32(charge * 260)
	tint := primitives.RuneColor
	white := rl.RayWhite
	fade := func(tint rl.Color, alpha float32) rl.Color {
		return rl.Fade(tint, alpha*opacity)
	}
	progress := float32(math.Sqrt(float64(energy)))
	rl.DrawCircleGradient(center, 145, fade(tint, .18+energy*.5), rl.Blank)
	rl.DrawRing(center, 81, 84, 0, 360, 64, fade(tint, .6+energy*.4))
	rl.DrawPolyLinesEx(center, 6, 74-progress*12, rotation, 2, fade(tint, 1))
	rl.DrawPolyLinesEx(center, 6, 63+progress*12, -rotation, 2, fade(white, .35+energy*.5))
	for i := 0; i < 12; i++ {
		angle := float64(i)*math.Pi/6 + float64(rotation)*math.Pi/180
		rune := rl.NewVector2(center.X+float32(math.Cos(angle))*94, center.Y+float32(math.Sin(angle))*94)
		rl.DrawPolyLinesEx(rune, 4, 3+energy*2, float32(angle*180/math.Pi), 1, fade(tint, .65))
		phase := float32(math.Mod(charge*1.7+float64(i)/12, 1))
		radius := 70 * (1 - phase)
		height := phase * (45 + energy*95)
		mote := rl.NewVector2(center.X+float32(math.Cos(angle))*radius, center.Y+float32(math.Sin(angle))*radius*.45-height)
		rl.DrawLineEx(rl.NewVector2(mote.X, mote.Y+6+energy*8), mote, 1, fade(tint, .35))
		rl.DrawCircleV(mote, 1.5+energy, fade(white, .4+phase*.5))
	}
	for i := 0; i < 3; i++ {
		phase := float32(math.Mod(charge*2+float64(i)/3, 1))
		rl.DrawEllipseLines(int32(center.X), int32(center.Y-phase*(50+energy*90)), 57*(1-phase*.65), 14*(1-phase*.65), fade(tint, (1-phase)*(.3+energy*.6)))
	}
	beamHeight := 25 + energy*155
	for i := -3; i <= 3; i++ {
		x := player.X + float32(i)*6
		rl.DrawLineEx(rl.NewVector2(x, player.Y+18), rl.NewVector2(x, player.Y-beamHeight), 4, fade(tint, energy*.12))
	}
	rl.DrawCircleGradient(player, 35+energy*20, fade(white, energy*.65), rl.Blank)
}

func drawActivatedPortal(center rl.Vector2) {
	time := rl.GetTime()
	pulse := float32(.85 + .15*math.Sin(time*2))
	tint := primitives.RuneColor
	rl.DrawCircleGradient(center, 150, rl.Fade(tint, .38*pulse), rl.Blank)
	rl.DrawCircleGradient(center, 55, rl.Fade(rl.RayWhite, .2*pulse), rl.Blank)
	rl.DrawRing(center, 87, 92, 0, 360, 96, rl.Fade(tint, .95*pulse))
	rl.DrawRing(center, 59, 61, 0, 360, 64, rl.Fade(rl.RayWhite, .7*pulse))
	rl.DrawPolyLinesEx(center, 6, 74, float32(-time*18), 2, rl.Fade(rl.RayWhite, .65*pulse))
	for i := 0; i < 6; i++ {
		angle := float64(i)*math.Pi/3 + time*.3
		inner := rl.NewVector2(center.X+float32(math.Cos(angle))*94, center.Y+float32(math.Sin(angle))*94)
		outer := rl.NewVector2(center.X+float32(math.Cos(angle))*103, center.Y+float32(math.Sin(angle))*103)
		rl.DrawLineEx(inner, outer, 3, rl.Fade(rl.RayWhite, .85*pulse))
		rl.DrawCircleGradient(outer, 12, rl.Fade(tint, .5*pulse), rl.Blank)
	}
	for i := 0; i < 8; i++ {
		phase := math.Mod(time*.35+float64(i)/8, 1)
		angle := float64(i)*math.Pi/4 + time*.2
		mote := rl.NewVector2(center.X+float32(math.Cos(angle))*48, center.Y+float32(math.Sin(angle))*20-float32(phase)*65)
		rl.DrawCircleV(mote, 2, rl.Fade(rl.RayWhite, float32(math.Sin(phase*math.Pi))*.75))
	}
}
