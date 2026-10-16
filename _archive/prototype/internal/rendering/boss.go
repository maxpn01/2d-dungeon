package rendering

import (
	"math"

	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawBoss(b simulation.Boss, e Visibility) {
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
