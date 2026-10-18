package rendering

import (
	"image/color"
	"math"
	"time"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (f *Feedback) observePickups(now time.Time, local feedbackActor, fruits []simulation.Fruit) {
	current := make(map[string]simulation.Fruit, len(fruits))
	for _, fruit := range fruits {
		current[fruit.ID] = fruit
	}
	if f.fruits != nil && !local.dead {
		reach := local.size + 20
		for id, fruit := range f.fruits {
			if _, exists := current[id]; exists {
				continue
			}
			x, y := fruit.X+fruit.Size/2, fruit.Y+fruit.Size/2
			if math.Hypot(x-local.x-local.size/2, y-local.y-local.size/2) > reach+fruit.Size {
				continue
			}
			tint := HealthFoodColor
			if fruit.MP > 0 {
				tint = ManaFoodColor
			}
			f.effects = append(f.effects, feedbackEffect{started: now, x: float32(x), y: float32(y), size: float32(fruit.Size), tint: tint, kind: "pop"})
		}
	}
	f.fruits = current
}

func drawPickupPop(center rl.Vector2, size, p float32, tint color.RGBA) {
	eased := 1 - (1-p)*(1-p)
	radius := size*.4 + 16*eased
	rl.DrawRing(center, radius, radius+2.5*(1-p), 0, 360, 32, rl.Fade(tint, 1-p))
	rl.DrawCircleV(center, size*.45*(1-eased), rl.Fade(rl.RayWhite, .8*(1-p)))
	for i := range 6 {
		angle := float64(i)*math.Pi/3 + math.Pi/6
		distance := size*.3 + 22*eased
		spark := 3 * (1 - p)
		at := rl.NewVector2(center.X+float32(math.Cos(angle))*distance, center.Y+float32(math.Sin(angle))*distance)
		rl.DrawRectangleRec(rl.NewRectangle(at.X-spark/2, at.Y-spark/2, spark, spark), rl.Fade(tint, 1-p))
	}
}

func drawDeathBurst(center rl.Vector2, size, p float32, tint color.RGBA, seed uint32) {
	eased := 1 - (1-p)*(1-p)*(1-p)
	glow := (1 - p) * (1 - p)
	rl.DrawCircleGradient(center, size*(.8+.6*eased), rl.Fade(tint, .35*glow), rl.Blank)
	for i := range parameters.DeathBurstParticles {
		seed = seed*1664525 + 1013904223
		angle := float64(seed>>8) / 16777216 * 2 * math.Pi
		seed = seed*1664525 + 1013904223
		speed := float32(.45 + float64(seed>>8)/16777216*.55)
		distance := size*.3 + parameters.DeathBurstDistance*speed*eased
		side := (3 + 4*speed) * (1 - p*.8)
		at := rl.NewVector2(center.X+float32(math.Cos(angle))*distance, center.Y+float32(math.Sin(angle))*distance)
		particle := tint
		if i%4 == 0 {
			particle = rl.RayWhite
		}
		rl.DrawRectanglePro(rl.NewRectangle(at.X, at.Y, side, side), rl.NewVector2(side/2, side/2), float32(angle*57.3)+360*p*speed, rl.Fade(particle, 1-p))
	}
}
