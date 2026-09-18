package game

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"image/color"
	"math"
	"raylib-2d-sandbox/internal/simulation"
	"strings"
)

func massColor(source, localID string) color.RGBA {
	if source == "" {
		return rl.Yellow
	}
	if source == localID {
		return rl.Red
	}
	if strings.HasPrefix(source, "ai-") {
		return rl.Green
	}
	return rl.SkyBlue
}

func drawCombat(x, y, size float64, c simulation.CombatState) {
	center := rl.NewVector2(float32(x+size/2), float32(y+size/2))
	switch c.Phase {
	case simulation.DashWinding, simulation.DashRushing:
		length := float32(55)
		tint := rl.Orange
		if c.Phase == simulation.DashRushing {
			length = 80
			tint = rl.RayWhite
		}
		dir := rl.NewVector2(float32(c.Direction.X), float32(c.Direction.Y))
		tip := rl.NewVector2(center.X+dir.X*length, center.Y+dir.Y*length)
		rl.DrawLineEx(center, tip, 3, tint)
		rl.DrawLineEx(tip, rl.NewVector2(tip.X-dir.X*12-dir.Y*7, tip.Y-dir.Y*12+dir.X*7), 3, tint)
		rl.DrawLineEx(tip, rl.NewVector2(tip.X-dir.X*12+dir.Y*7, tip.Y-dir.Y*12-dir.X*7), 3, tint)
		rl.DrawRectangleLinesEx(rl.NewRectangle(float32(x)-3, float32(y)-3, float32(size)+6, float32(size)+6), 2, tint)
	case simulation.DashBumped:
		shake := float32(math.Sin(c.Remaining*100) * 4)
		rl.DrawRectangleLinesEx(rl.NewRectangle(float32(x)-4+shake, float32(y)-4, float32(size)+8, float32(size)+8), 3, rl.RayWhite)
	case simulation.DashRecovering:
		rl.DrawRectangleLinesEx(rl.NewRectangle(float32(x)-2, float32(y)-2, float32(size)+4, float32(size)+4), 1, rl.Gray)
	}
}
