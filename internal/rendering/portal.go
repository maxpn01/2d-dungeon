package rendering

import (
	"math"

	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"
	"2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawPortals(local simulation.Player, snapshot simulation.Snapshot, e Visibility) {
	orb := snapshot.Orb
	for _, portal := range worldmap.Default().Portals() {
		position := portal.Position()
		pad := rl.NewVector2(float32(position.X), float32(position.Y))
		if !e.EntityVisible(float64(pad.X), float64(pad.Y), 0) {
			continue
		}
		activated := orb.Consumed || portal.ID != worldmap.DescentPortal && portal.ID != worldmap.ForestExitPortal || portal.ID == worldmap.ForestExitPortal && snapshot.HydraDefeated
		active := activated || local.HasOrb && portal.ID == worldmap.DescentPortal
		tint := rl.Gray
		label := ""
		if active {
			tint = primitives.RuneColor
			label = "[E] ACTIVATE TELEPORT"
			if activated {
				destination, _ := worldmap.Default().Portal(portal.Destination)
				verb := "TRAVEL TO"
				if destination.Floor < portal.Floor {
					verb = "RETURN TO"
				}
				label = "[E] " + verb + " FLOOR " + worldmap.FloorNumeral(destination.Floor)
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
		if activated {
			drawActivatedPortal(pad)
			status := "TELEPORT ACTIVE"
			primitives.DrawTextAt(status, rl.NewVector2(pad.X-float32(primitives.MeasureText(status, 16))/2, pad.Y-125), 16, rl.RayWhite)
		}
		if !active && portal.ID == worldmap.ForestExitPortal {
			label, tint = "SEALED BY THE HYDRA'S BLOOD", rl.NewColor(236, 120, 140, 255)
		}
		if label != "" {
			primitives.DrawTextAt(label, rl.NewVector2(pad.X-float32(primitives.MeasureText(label, 14))/2, pad.Y+105), 14, tint)
		}
	}
}

func DrawSidePortals() {
	t := rl.GetTime()
	for _, portal := range worldmap.Default().Portals() {
		if portal.Floor != worldmap.ThirdFloor {
			continue
		}
		position := portal.Position()
		ground := float32(position.Y) + 20
		x := float32(position.X)
		tint := primitives.RuneColor
		pulse := float32(.7 + .2*math.Sin(t*3))
		rl.DrawCircleGradient(rl.NewVector2(x, ground-60), 130, rl.Fade(tint, .16*pulse), rl.Blank)
		rl.DrawRectangleGradientV(int32(x-44), int32(ground-150), 88, 150, rl.Blank, rl.Fade(tint, .32*pulse))
		rl.DrawRing(rl.NewVector2(x, ground), 66, 72, 180, 360, 48, rl.Fade(tint, pulse))
		rl.DrawRing(rl.NewVector2(x, ground), 78, 80, 180, 360, 48, rl.Fade(tint, .5*pulse))
		for i := range 6 {
			angle := math.Pi + float64(i+1)*math.Pi/7 + math.Sin(t+float64(i))*.03
			at := rl.NewVector2(x+float32(math.Cos(angle))*69, ground+float32(math.Sin(angle))*69)
			rl.DrawPoly(at, 4, 4, float32(t*40), rl.RayWhite)
		}
		for _, radius := range []float32{84, 66, 46} {
			rl.DrawEllipseLines(int32(x), int32(ground), radius, radius*.16, rl.Fade(tint, pulse))
		}
		rl.DrawEllipse(int32(x), int32(ground), 64, 10, rl.Fade(tint, .35*pulse))
		for i := range 12 {
			age := float32(math.Mod(t*.5+float64(i)/12, 1))
			sx := x + float32(math.Sin(float64(i)*2.3+t*.7))*38
			rl.DrawCircleV(rl.NewVector2(sx, ground-age*140), 2, rl.Fade(rl.RayWhite, 1-age))
		}
		destination, _ := worldmap.Default().Portal(portal.Destination)
		label := "[E] RETURN TO FLOOR " + worldmap.FloorNumeral(destination.Floor)
		primitives.DrawTextAt(label, rl.NewVector2(x-float32(primitives.MeasureText(label, 14))/2, ground-110), 14, tint)
	}
}
