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
	combat      simulation.CombatState
	id          string
	pos, size   rl.Vector2
	color       color.RGBA
	stats       PlayerStats
	actionState AIActionState
	isDeadFlag  bool
}

func (ai *AI) draw() {
	if ai.isDeadFlag {
		return
	}
	drawSafeProtection(float64(ai.pos.X), float64(ai.pos.Y), float64(ai.size.X), float64(ai.stats.hp), ai.safeZones)
	bounds := drawActor(float64(ai.pos.X), float64(ai.pos.Y), float64(ai.size.X), float64(ai.stats.hp), ai.combat, ai.color)
	drawTextAt(fmt.Sprintf("AI Lv.%d %s", ai.stats.lvl, ai.actionState), rl.NewVector2(bounds.X, bounds.Y-22), 18, rl.RayWhite)
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
