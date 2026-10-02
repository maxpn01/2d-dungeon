package rendering

import (
	"image/color"

	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Actor struct {
	Frenzied       bool
	ID, Label      string
	X, Y, Size, HP float64
	Dead           bool
	Character      simulation.Character
	Combat         simulation.CombatState
	Split          simulation.SplitCells
	Tint           color.RGBA
}

func DrawPlayerActor(p Actor, feedback *Feedback, exploration Visibility) {
	if p.Dead {
		return
	}
	count := p.Split.Count()
	bounds := feedback.DrawPlayer(p.ID, p.X, p.Y, p.Size, p.HP, p.Combat, p.Tint, p.Character)
	for _, cell := range p.Split {
		if cell.Active {
			if exploration == nil || exploration.EntityVisible(cell.X, cell.Y, p.Size) {
				DrawSplitBody(cell.X, cell.Y, p.Size, cell.Stats.HP, p.Tint, p.Character, count)
			}
		}
	}
	primitives.DrawTextAt(p.Label, rl.NewVector2(bounds.X, bounds.Y-22), 18, rl.RayWhite)
}

func DrawAIActor(ai Actor, exploration Visibility) {
	if ai.Dead {
		return
	}
	count := ai.Split.Count()
	if exploration == nil || exploration.EntityVisible(ai.X, ai.Y, ai.Size) {
		if ai.Frenzied {
			drawPossession(ai)
		}
		bounds := DrawActor(ai.X, ai.Y, ai.Size, ai.HP, ai.Combat, ai.Tint, ai.Character)
		if ai.Frenzied {
			drawPossessedEyes(bounds)
		}
		primitives.DrawTextAt(ai.Label, rl.NewVector2(bounds.X, bounds.Y-22), 18, rl.RayWhite)
	}
	for _, cell := range ai.Split {
		if cell.Active && (exploration == nil || exploration.EntityVisible(cell.X, cell.Y, ai.Size)) {
			DrawSplitBody(cell.X, cell.Y, ai.Size, cell.Stats.HP, ai.Tint, ai.Character, count)
		}
	}
}
