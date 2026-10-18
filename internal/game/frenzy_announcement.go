package game

import (
	"math"
	"time"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/rendering"
	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"
	"2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const frenzyAnnouncementDuration = float32(simulation.OrbFrenzyDelayTicks)/parameters.TickRate + 2

func frenzyAnnouncementAge(snapshot simulation.Snapshot, local simulation.Player) (float32, bool) {
	if snapshot.Orb.FrenzyAt < simulation.OrbFrenzyDelayTicks {
		return 0, false
	}
	at := snapshot.Orb.FrenzyAt - simulation.OrbFrenzyDelayTicks
	age := float32(snapshot.Tick-at) / parameters.TickRate
	onFirstFloor := worldmap.Default().FloorAt(local.X+parameters.PlayerSize/2, local.Y+parameters.PlayerSize/2) == worldmap.FirstFloor
	if snapshot.Orb.Carrier == "" || snapshot.Tick < at || age >= frenzyAnnouncementDuration || snapshot.WinnerID != "" || !onFirstFloor {
		return 0, false
	}
	return float32(snapshot.Tick-at) / parameters.TickRate, true
}

func (p *presentation) dungeonAwakeningCamera(camera rl.Camera2D, snapshot simulation.Snapshot, local simulation.Player, now time.Time) rl.Camera2D {
	age, visible := frenzyAnnouncementAge(snapshot, local)
	if !visible {
		if snapshot.Orb.FrenzyAt == 0 {
			p.dungeonShakeAt, p.dungeonShakeStarted = 0, time.Time{}
		}
		return camera
	}
	if p.dungeonShakeAt != snapshot.Orb.FrenzyAt {
		p.dungeonShakeAt = snapshot.Orb.FrenzyAt
		p.dungeonShakeStarted = now.Add(-time.Duration(float64(age) * float64(time.Second)))
		if age < .25 {
			rendering.PlayDungeonAwakeningSound()
		}
	}
	elapsed := max(0, now.Sub(p.dungeonShakeStarted).Seconds())
	const duration = rendering.DungeonAwakeningShakeDuration
	if elapsed >= duration {
		return camera
	}
	strength := 8 * min(1, elapsed/.06) * math.Pow(1-elapsed/duration, 2)
	camera.Offset.X += float32(strength * (.65*math.Sin(elapsed*57) + .35*math.Sin(elapsed*91)))
	camera.Offset.Y += float32(strength * (.65*math.Sin(elapsed*71) + .35*math.Sin(elapsed*43)))
	return camera
}

func drawFrenzyAnnouncement(snapshot simulation.Snapshot, local simulation.Player, width int32) {
	age, visible := frenzyAnnouncementAge(snapshot, local)
	if !visible {
		return
	}
	alpha := min(float32(1), (frenzyAnnouncementDuration-age)/.75)
	title, subtitle := "THE FIRST FLOOR AWAKENS", "Its guardians enter a frenzy!"
	if snapshot.Tick < snapshot.Orb.FrenzyAt {
		subtitle = "Its guardians are stirring..."
	}
	fontSize := int32(24)
	available := max(int32(240), width-480)
	y := int32(108)
	if width < 800 {
		available, y = width-32, 218
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
