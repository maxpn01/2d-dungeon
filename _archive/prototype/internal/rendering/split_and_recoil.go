package rendering

import (
	"image/color"
	"math"

	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawSplitBody(x, y, size, hp float64, tint color.RGBA, character simulation.Character, count int) {
	if character == simulation.Circle && count >= 64 {
		bounds := actorBounds(x, y, size, hp)
		radius := bounds.Width / 2
		segments := int32(12)
		if radius > 12 {
			segments = 18
		}
		rl.DrawCircleSector(rl.NewVector2(bounds.X+radius, bounds.Y+radius), radius, 0, 360, segments, tint)
		return
	}
	DrawActor(x, y, size, hp, simulation.CombatState{}, tint, character)
}

func drawBossHitGrace(center rl.Vector2, bounds rl.Rectangle, c simulation.CombatState) {
	if c.HitGrace > 0 {
		pulse := float32(.6 + .25*math.Sin(rl.GetTime()*20))
		rl.DrawRing(center, bounds.Width*.75+4, bounds.Width*.75+7, 0, 360, 40, rl.Fade(rl.SkyBlue, pulse))
	}
}

func bossRecoilDeformation(c simulation.CombatState) (float32, float32) {
	p := phaseProgress(c.Remaining, simulation.BossRecoilDuration)
	stretch := float32(-0.28) * (1 - p)
	offset := float32(math.Sin(float64(p)*math.Pi*4)) * 4 * (1 - p)
	return stretch, offset
}

func drawBossRecoilImpact(center rl.Vector2, size float32, c simulation.CombatState) {
	dir := rl.NewVector2(float32(c.Direction.X), float32(c.Direction.Y))
	point := func(distance, sideways float32) rl.Vector2 {
		return rl.NewVector2(center.X+dir.X*distance-dir.Y*sideways, center.Y+dir.Y*distance+dir.X*sideways)
	}
	p := phaseProgress(c.Remaining, simulation.BossRecoilDuration)
	impact := point(size/2+simulation.BumpKnockback, 0)
	for i := 0; i < 7; i++ {
		angle := float64(i)*math.Pi/3 + float64(p)*0.4
		dx, dy := float32(math.Cos(angle)), float32(math.Sin(angle))
		a, b := 5+12*p, 10+22*p
		rl.DrawLineEx(rl.NewVector2(impact.X+dx*a, impact.Y+dy*a), rl.NewVector2(impact.X+dx*b, impact.Y+dy*b), 2, rl.Fade(rl.RayWhite, 1-p))
	}

}
