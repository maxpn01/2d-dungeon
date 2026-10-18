package rendering

import (
	"math"
	"time"

	"2d-dungeon/internal/parameters"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (f *Feedback) shake(now time.Time, strength float64) {
	if now.Sub(f.shakeAt).Seconds() < parameters.ScreenShakeSeconds && f.shakeStrength >= strength {
		return
	}
	f.shakeAt, f.shakeStrength = now, strength
}

func (f *Feedback) ShakeCamera(camera rl.Camera2D, now time.Time) rl.Camera2D {
	elapsed := now.Sub(f.shakeAt).Seconds()
	if f.shakeStrength <= 0 || elapsed >= parameters.ScreenShakeSeconds {
		return camera
	}
	decay := 1 - elapsed/parameters.ScreenShakeSeconds
	strength := f.shakeStrength * decay * decay
	camera.Offset.X += float32(strength * (.6*math.Sin(elapsed*83) + .4*math.Sin(elapsed*131)))
	camera.Offset.Y += float32(strength * (.6*math.Sin(elapsed*97) + .4*math.Sin(elapsed*59)))
	return camera
}
