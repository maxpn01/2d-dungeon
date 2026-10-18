package rendering

import (
	"image/color"

	"2d-dungeon/internal/appearance"
	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Actor struct {
	Type           simulation.MobType
	Level          int
	Stone          bool
	Awakening      float64
	Aiming         bool
	Frenzied       bool
	ID, Label      string
	X, Y, Size, HP float64
	Dead           bool
	Character      appearance.Character
	Combat         simulation.CombatState
	Tint           color.RGBA
}

func DrawPlayerActor(p Actor, feedback *Feedback, exploration Visibility) {
	if p.Dead {
		return
	}
	bounds := feedback.DrawPlayer(p.ID, p.X, p.Y, p.Size, p.HP, p.Combat, p.Tint, p.Character)
	feedback.DrawHit(p.ID, bounds, p.Character)
	primitives.DrawTextAt(p.Label, rl.NewVector2(bounds.X, bounds.Y-22), 18, rl.RayWhite)
}

func DrawAIActor(ai Actor, exploration Visibility, feedback *Feedback) {
	if ai.Dead {
		return
	}
	if ai.Type.IsHeavy() {
		multiplier := parameters.HeavySizeMultiplier
		if ai.Type == simulation.GateHeavy {
			multiplier = parameters.GateHeavySizeMultiplier
		}
		size := ai.Size * multiplier
		ai.X -= (size - ai.Size) / 2
		ai.Y -= (size - ai.Size) / 2
		ai.Size = size
	}
	if exploration == nil || exploration.EntityVisible(ai.X, ai.Y, ai.Size) {
		if ai.Stone || ai.Awakening > 0 {
			drawStoneSentinel(ai)
			return
		}
		if ai.Frenzied {
			drawPossession(ai)
		}
		bounds := DrawActor(ai.X, ai.Y, ai.Size, ai.HP, ai.Combat, ai.Tint, ai.Character)
		drawStomp(rl.NewVector2(bounds.X+bounds.Width/2, bounds.Y+bounds.Height/2), bounds.Width, ai.Combat, ai.Type)
		drawMobMark(ai.Type, bounds)
		if ai.Type == simulation.DungeonPursuer || ai.Type.IsHeavy() {
			drawMobWeapon(rl.NewVector2(bounds.X+bounds.Width/2, bounds.Y+bounds.Height/2), bounds.Width, ai.Combat, ai.Type)
		}
		feedback.DrawHit(ai.ID, bounds, ai.Character)
		if ai.Aiming {
			center := rl.NewVector2(bounds.X+bounds.Width/2, bounds.Y+bounds.Height/2)
			radius := max(bounds.Width, bounds.Height)*.75 + 4
			rl.DrawRing(center, radius, radius+2, 0, 360, 32, rl.Orange)
		}
		primitives.DrawTextAt(ai.Label, rl.NewVector2(bounds.X, bounds.Y-22), 18, rl.RayWhite)
	}
}
