package rendering

import (
	"math"

	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"
	"2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawPortals(local simulation.Player, orb simulation.OrbState, e Visibility) {
	for _, portal := range worldmap.Default().Portals() {
		position := portal.Position()
		pad := rl.NewVector2(float32(position.X), float32(position.Y))
		if !e.EntityVisible(float64(pad.X), float64(pad.Y), 0) {
			continue
		}
		activated := orb.Consumed || portal.ID != worldmap.DescentPortal
		active := activated || local.HasOrb
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
		if label != "" {
			primitives.DrawTextAt(label, rl.NewVector2(pad.X-float32(primitives.MeasureText(label, 14))/2, pad.Y+105), 14, tint)
		}
	}
}
