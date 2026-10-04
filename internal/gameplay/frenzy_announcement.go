package gameplay

import (
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawFrenzyAnnouncement(snapshot protocol.Snapshot, local protocol.Player, width int32) {
	if snapshot.Orb.FrenzyAt < simulation.OrbFrenzyDelayTicks {
		return
	}
	at := snapshot.Orb.FrenzyAt - simulation.OrbFrenzyDelayTicks
	if snapshot.Orb.Carrier == "" || snapshot.Tick < at || snapshot.Tick-at >= 5*simulation.TickRate || snapshot.WinnerID != "" || worldmap.Default().FloorAt(local.X+local.Size/2, local.Y+local.Size/2) != worldmap.FirstFloor {
		return
	}
	age := float32(snapshot.Tick-at) / simulation.TickRate
	alpha := min(float32(1), (5-age)/.75)
	title, subtitle := "THE FIRST FLOOR AWAKENS", "Its guardians enter a frenzy!"
	if snapshot.Tick < snapshot.Orb.FrenzyAt {
		subtitle = "Its guardians are stirring..."
	}
	fontSize := int32(24)
	available := max(int32(240), width-480)
	y := int32(108)
	if width < 800 {
		available, y = width-32, 218
		if snapshot.Boss.Phase != "" {
			y = 270
		}
	}
	for fontSize > 14 && primitives.MeasureText(title, fontSize)+40 > available {
		fontSize--
	}
	panelWidth := max(primitives.MeasureText(title, fontSize), primitives.MeasureText(subtitle, 18)) + 40
	x := (width - panelWidth) / 2
	rl.DrawRectangle(x, y, panelWidth, 65, rl.Fade(rl.NewColor(19, 15, 31, 255), .92*alpha))
	rl.DrawRectangle(x, y, panelWidth, 2, rl.Fade(primitives.RuneColor, alpha))
	primitives.DrawText(title, (width-primitives.MeasureText(title, fontSize))/2, y+10, fontSize, rl.Fade(rl.Gold, alpha))
	primitives.DrawText(subtitle, (width-primitives.MeasureText(subtitle, 18))/2, y+39, 18, rl.Fade(rl.NewColor(208, 190, 244, 255), alpha))
}
