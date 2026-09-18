package game

import (
	"fmt"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type AIActionState int

const (
	ActionSearchingFood AIActionState = iota
	ActionEating
	ActionFleeing
	ActionSleeping
)

type AI struct {
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
	rl.DrawRectangleV(ai.pos, ai.size, ai.color)
	drawTextAt(fmt.Sprintf("AI Lv.%d %s", ai.stats.lvl, ai.actionState), rl.NewVector2(ai.pos.X, ai.pos.Y-22), 18, rl.RayWhite)
}

func (state AIActionState) String() string {
	switch state {
	case ActionSearchingFood:
		return "searching"
	case ActionEating:
		return "eating"
	case ActionFleeing:
		return "fleeing"
	case ActionSleeping:
		return "sleeping"
	default:
		return "unknown"
	}
}
