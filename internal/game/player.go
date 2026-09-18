package game

import (
	"fmt"
	"image/color"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type PlayerStats struct {
	lvl int
	exp float64

	hp    float32
	speed float32

	maxLvl int
	maxHP  float32
}

type PlayerInput struct {
	Direction rl.Vector2
	Dash      bool
}

type Player struct {
	combat simulation.CombatState
	name   string
	input  PlayerInput
	pos    rl.Vector2
	size   rl.Vector2
	color  color.RGBA

	stats PlayerStats

	isDeadFlag bool
}

func NewPlayer(
	pos, size rl.Vector2,
	color color.RGBA,
	stats PlayerStats,
) *Player {
	return &Player{
		pos:        pos,
		size:       size,
		color:      color,
		stats:      stats,
		isDeadFlag: false,
	}
}

func (p *Player) draw() {
	rl.DrawRectangleV(p.pos, p.size, p.color)
	drawCombat(float64(p.pos.X), float64(p.pos.Y), float64(p.size.X), p.combat)
	drawTextAt(fmt.Sprintf("%s Lv.%d (you)", p.name, p.stats.lvl), rl.NewVector2(p.pos.X, p.pos.Y-22), 18, rl.RayWhite)
}

func (p *Player) readInput() {
	input := PlayerInput{Dash: p.input.Dash || rl.IsKeyPressed(rl.KeyLeftShift)}
	if rl.IsKeyDown(rl.KeyW) {
		input.Direction.Y -= 1
	}
	if rl.IsKeyDown(rl.KeyS) {
		input.Direction.Y += 1
	}
	if rl.IsKeyDown(rl.KeyA) {
		input.Direction.X -= 1
	}
	if rl.IsKeyDown(rl.KeyD) {
		input.Direction.X += 1
	}
	p.input = input
}
