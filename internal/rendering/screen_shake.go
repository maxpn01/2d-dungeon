package rendering

import (
	"math"
	"time"

	"dungeon-raylib-go/internal/parameters"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (f *Feedback) shake(now time.Time, strength, seconds float64) {
	if now.Sub(f.shakeAt).Seconds() < f.shakeSeconds && f.shakeStrength >= strength {
		return
	}
	f.shakeAt, f.shakeStrength, f.shakeSeconds = now, strength, seconds
}

func (f *Feedback) ShakeCamera(camera rl.Camera2D, now time.Time) rl.Camera2D {
	elapsed := now.Sub(f.shakeAt).Seconds()
	if f.shakeStrength <= 0 || elapsed >= f.shakeSeconds {
		return camera
	}
	decay := 1 - elapsed/f.shakeSeconds
	strength := f.shakeStrength * decay * decay
	camera.Offset.X += float32(strength * (.6*math.Sin(elapsed*83) + .4*math.Sin(elapsed*131)))
	camera.Offset.Y += float32(strength * (.6*math.Sin(elapsed*97) + .4*math.Sin(elapsed*59)))
	return camera
}

func drawPlayerDeath(center rl.Vector2, size, p float32, tint rl.Color, seed uint32) {
	flash := max(0, 1-p/.18)
	rl.DrawCircleV(center, size*(.6+.5*(1-flash)), rl.Fade(rl.RayWhite, .85*flash))
	ring := size*.6 + parameters.PlayerDeathRingRadius*(1-(1-p)*(1-p))
	rl.DrawRing(center, ring, ring+4*(1-p), 0, 360, 48, rl.Fade(tint, 1-p))
	drawDeathBurst(center, size*1.8, p, tint, seed)
}
