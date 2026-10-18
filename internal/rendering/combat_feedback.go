package rendering

import (
	"math"
	"time"

	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (f *Feedback) DrawHit(id string, bounds rl.Rectangle, character simulation.Character) {
	if f == nil {
		return
	}
	at, ok := f.hits[id]
	if !ok {
		return
	}
	elapsed := time.Since(at).Seconds()
	if elapsed < 0 || elapsed > .22 {
		return
	}
	p := float32(elapsed / .22)
	primitives.DrawCharacterShape(character, bounds, rl.Fade(rl.RayWhite, .9*(1-p)))
	outline := bounds
	outline.X -= 3 + 4*p
	outline.Y -= 3 + 4*p
	outline.Width += 6 + 8*p
	outline.Height += 6 + 8*p
	rl.DrawRectangleLinesEx(outline, 2, rl.Fade(rl.Orange, 1-p))
}

func (f *Feedback) CheckCombatInput(now time.Time, input simulation.Direction, player simulation.Player) {
	if player.Dead {
		return
	}
	failed := input.Shoot && player.Stats.MP < player.Stats.ShotMPCost()
	if input.Boost && (input.X != 0 || input.Y != 0 || player.HasWings) {
		failed = failed || player.Stats.MP <= 1e-6 || player.HasWings && player.Stats.MP < simulation.WingMPPerSecond*simulation.StepSeconds
	}
	if failed && now.Sub(f.manaAt) >= 900*time.Millisecond {
		f.manaAt = now
		playFeedbackSound("mana")
	}
}

func (f *Feedback) DrawCombatOverlay(now time.Time, player simulation.Player, width, height int32) {
	if player.Dead {
		return
	}
	hp := player.Stats.HP
	low := hp <= player.Stats.MaxHP()*.25
	pulse := float32(.5 + .5*math.Sin(float64(now.UnixMilli())*.009))
	if low {
		rl.DrawRectangleLinesEx(rl.NewRectangle(2, 2, float32(width-4), float32(height-4)), 4, rl.Fade(rl.Red, .18+.2*pulse))
	}
	hurt := now.Sub(f.hurtAt).Seconds()
	if hurt >= 0 && hurt < .3 {
		fade := float32(1 - hurt/.3)
		rl.DrawRectangleLinesEx(rl.NewRectangle(0, 0, float32(width), float32(height)), 12, rl.Fade(rl.Red, .6*fade))
	}
	warning := func(label string, y int32, tint rl.Color) {
		textWidth := primitives.MeasureText(label, 20)
		x := (width - textWidth) / 2
		rl.DrawRectangle(x-12, y-6, textWidth+24, 32, rl.Fade(rl.Black, .85))
		primitives.DrawText(label, x, y, 20, tint)
	}
	if low {
		warning("LOW HEALTH", height-145, rl.Fade(rl.Red, .65+.35*pulse))
	}
	if now.Sub(f.manaAt) >= 0 && now.Sub(f.manaAt) < 800*time.Millisecond {
		warning("NOT ENOUGH MANA", height-105, rl.SkyBlue)
	} else if player.Stats.MP < player.Stats.ShotMPCost() {
		warning("LOW MANA / Cannot shoot", height-105, rl.SkyBlue)
	}
}
