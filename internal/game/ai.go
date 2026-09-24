package game

import (
	"fmt"
	"image/color"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type AIActionState int

const (
	ActionSearchingFood AIActionState = iota
	ActionEating
	ActionChasing
	ActionFleeing
	ActionSleeping
	ActionAttacking
	ActionEvading
	ActionAlerted
	ActionEmerging
)

type AI struct {
	safeZones   []simulation.SafeZone
	character   simulation.Character
	combat      simulation.CombatState
	split       simulation.SplitCells
	id          string
	pos, size   rl.Vector2
	color       color.RGBA
	stats       PlayerStats
	actionState AIActionState
	isDeadFlag  bool
}

func (ai *AI) draw() {
	ai.drawVisible(nil)
}

func (ai *AI) drawVisible(exploration *mapExploration) {
	if ai.isDeadFlag {
		return
	}
	mass := float64(ai.stats.hp) + ai.split.Mass()
	if exploration == nil || exploration.entityVisible(float64(ai.pos.X), float64(ai.pos.Y), float64(ai.size.X)) {
		drawSafeProtection(float64(ai.pos.X), float64(ai.pos.Y), float64(ai.size.X), mass, ai.safeZones)
		bounds := drawActor(float64(ai.pos.X), float64(ai.pos.Y), float64(ai.size.X), float64(ai.stats.hp), ai.combat, ai.color, ai.character)
		drawTextAt(fmt.Sprintf("AI Lv.%d %s", ai.stats.lvl, ai.actionState), rl.NewVector2(bounds.X, bounds.Y-22), 18, rl.RayWhite)
	}
	for _, cell := range ai.split {
		if cell.Active && (exploration == nil || exploration.entityVisible(cell.X, cell.Y, float64(ai.size.X))) {
			drawSafeProtection(cell.X, cell.Y, float64(ai.size.X), mass, ai.safeZones)
			drawActor(cell.X, cell.Y, float64(ai.size.X), cell.Stats.HP, simulation.CombatState{}, ai.color, ai.character)
		}
	}
}

func (state AIActionState) String() string {
	switch state {
	case ActionSearchingFood:
		return "searching"
	case ActionChasing:
		return "chasing"
	case ActionEating:
		return "eating"
	case ActionFleeing:
		return "fleeing"
	case ActionAttacking:
		return "attacking"
	case ActionEvading:
		return "evading"
	case ActionAlerted:
		return "alerted"
	case ActionEmerging:
		return "emerging"
	case ActionSleeping:
		return "sleeping"
	default:
		return "unknown"
	}
}
