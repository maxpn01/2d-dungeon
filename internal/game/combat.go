package game

import (
	"image/color"
	"math"
	"raylib-2d-sandbox/internal/simulation"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func massColor(source, localID string) color.RGBA {
	if source == "" {
		return energyColor
	}
	if source == localID {
		return rl.Red
	}
	if strings.HasPrefix(source, "ai-") {
		return rl.Green
	}
	return rl.SkyBlue
}

func actorRenderSize(size, hp float64) float32 {
	return float32(size * (0.65 + 0.35*math.Log2(1+max(0, hp))/math.Log2(1+simulation.StartingHP)))
}

func actorBounds(x, y, size, hp float64) rl.Rectangle {
	side := actorRenderSize(size, hp)
	return rl.NewRectangle(float32(x+size/2)-side/2, float32(y+size/2)-side/2, side, side)
}

func phaseProgress(remaining, duration float64) float32 {
	return float32(1 - max(0, min(1, remaining/duration)))
}

func drawActor(x, y, size, hp float64, c simulation.CombatState, tint color.RGBA) rl.Rectangle {
	bounds := actorBounds(x, y, size, hp)
	center := rl.NewVector2(float32(x+size/2), float32(y+size/2))
	dir := rl.NewVector2(float32(c.Direction.X), float32(c.Direction.Y))
	stretch, offset := float32(0), float32(0)
	switch c.Phase {
	case simulation.DashWinding:
		p := phaseProgress(c.Remaining, simulation.DashWindup)
		stretch, offset = -0.18*p, -4*p
	case simulation.DashRushing:
		stretch = 0.3 * (1 - 0.4*phaseProgress(c.Remaining, simulation.DashBurst))
		for i := 3; i > 0; i-- {
			trail := bounds
			trail.X -= dir.X * float32(i) * 13
			trail.Y -= dir.Y * float32(i) * 13
			rl.DrawRectangleRec(trail, rl.Fade(tint, 0.19/float32(i)))
		}
	case simulation.DashBumped:
		p := phaseProgress(c.Remaining, simulation.BumpStun)
		stretch = -0.28 * (1 - p)
		offset = float32(math.Sin(float64(p)*math.Pi*4)) * 4 * (1 - p)
	}
	width := bounds.Width * (1 + stretch*float32(math.Abs(float64(dir.X))) - stretch*0.5*float32(math.Abs(float64(dir.Y))))
	height := bounds.Height * (1 + stretch*float32(math.Abs(float64(dir.Y))) - stretch*0.5*float32(math.Abs(float64(dir.X))))
	body := rl.NewRectangle(center.X-width/2+dir.X*offset, center.Y-height/2+dir.Y*offset, width, height)
	shadow := body
	shadow.Y += 3
	rl.DrawRectangleRec(shadow, rl.Fade(rl.Black, 0.35))
	rl.DrawRectangleRec(body, tint)
	rl.DrawRectangleLinesEx(body, 1, rl.Fade(rl.RayWhite, 0.3))
	drawCombat(center, bounds.Width, c)
	return bounds
}

func drawCombat(center rl.Vector2, size float32, c simulation.CombatState) {
	dir := rl.NewVector2(float32(c.Direction.X), float32(c.Direction.Y))
	point := func(distance, sideways float32) rl.Vector2 {
		return rl.NewVector2(center.X+dir.X*distance-dir.Y*sideways, center.Y+dir.Y*distance+dir.X*sideways)
	}
	switch c.Phase {
	case simulation.DashWinding:
		p := phaseProgress(c.Remaining, simulation.DashWindup)
		start, tip := point(size/2+7, 0), point(size/2+48, 0)
		rl.DrawLineEx(start, tip, 2, rl.Fade(rl.Orange, 0.3))
		rl.DrawLineEx(start, point(size/2+7+41*p, 0), 3, rl.Orange)
		rl.DrawLineEx(tip, point(size/2+37, 7), 3, rl.Orange)
		rl.DrawLineEx(tip, point(size/2+37, -7), 3, rl.Orange)
		rl.DrawRing(center, size*0.72+3, size*0.72+5, -90, -90+360*p, 40, rl.Orange)
	case simulation.DashRushing:
		for _, side := range []float32{-1, 1} {
			rl.DrawLineEx(point(-size/2-30, side*size*0.6), point(size/2, side*size*0.6), 2, rl.Fade(rl.RayWhite, 0.65))
		}
	case simulation.DashBumped:
		p := phaseProgress(c.Remaining, simulation.BumpStun)
		impact := point(size/2+simulation.BumpKnockback, 0)
		for i := 0; i < 7; i++ {
			angle := float64(i)*math.Pi/3 + float64(p)*0.4
			dx, dy := float32(math.Cos(angle)), float32(math.Sin(angle))
			a, b := 5+12*p, 10+22*p
			rl.DrawLineEx(rl.NewVector2(impact.X+dx*a, impact.Y+dy*a), rl.NewVector2(impact.X+dx*b, impact.Y+dy*b), 2, rl.Fade(rl.RayWhite, 1-p))
		}
	case simulation.DashRecovering:
		p := phaseProgress(c.Remaining, simulation.DashRecovery)
		rl.DrawRing(center, size*0.72+3, size*0.72+4, -90+360*p, 270, 40, rl.Fade(rl.Gray, 0.65))
	}
}
