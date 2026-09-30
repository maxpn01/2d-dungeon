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
	mp    float32
	speed float32

	maxLvl int
	maxHP  float32
}

type PlayerInput struct {
	Shoot     bool
	Interact  bool
	Direction rl.Vector2
	Aim       rl.Vector2
	Boost     bool
	Dash      bool
}

type Player struct {
	feedback                  *combatFeedback
	suppressShootUntilRelease bool
	safeZones                 []simulation.SafeZone
	character                 simulation.Character
	split                     simulation.SplitCells
	combat                    simulation.CombatState
	name                      string
	input                     PlayerInput
	pos                       rl.Vector2
	size                      rl.Vector2
	color                     color.RGBA

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
	if p.isDeadFlag {
		return
	}
	mass := float64(p.stats.hp) + p.split.Mass()
	count := p.split.Count()
	drawSafeProtection(float64(p.pos.X), float64(p.pos.Y), float64(p.size.X), mass, p.safeZones)
	bounds := p.feedback.drawPlayer("player", float64(p.pos.X), float64(p.pos.Y), float64(p.size.X), float64(p.stats.hp), p.combat, p.color, p.character)
	for _, cell := range p.split {
		if cell.Active {
			drawSafeProtection(cell.X, cell.Y, float64(p.size.X), mass, p.safeZones)
			drawSplitBody(cell.X, cell.Y, float64(p.size.X), cell.Stats.HP, p.color, p.character, count)
		}
	}
	drawTextAt(fmt.Sprintf("%s Lv.%d (you)", p.name, p.stats.lvl), rl.NewVector2(bounds.X, bounds.Y-22), 18, rl.RayWhite)
}

func (p *Player) readInput(camera rl.Camera2D) {
	if !rl.IsMouseButtonDown(rl.MouseLeftButton) {
		p.suppressShootUntilRelease = false
	}
	input := PlayerInput{Shoot: !p.suppressShootUntilRelease && (p.input.Shoot || rl.IsMouseButtonDown(rl.MouseLeftButton)), Boost: rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift), Dash: p.input.Dash || rl.IsKeyPressed(rl.KeyLeftShift) || rl.IsKeyPressed(rl.KeyRightShift)}
	input.Interact = p.input.Interact || rl.IsKeyPressed(rl.KeyE)
	input.Aim = mouseAim(camera, entityCenter(p.pos, p.size), p.combat.Facing)
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

// Convert through the actual camera, so aim remains correct at any zoom.
func mouseAim(camera rl.Camera2D, center rl.Vector2, fallback simulation.Direction) rl.Vector2 {
	target := rl.GetScreenToWorld2D(rl.GetMousePosition(), camera)
	delta := rl.Vector2Subtract(target, center)
	if rl.Vector2Length(delta) > 1 {
		return rl.Vector2Normalize(delta)
	}
	if fallback.X != 0 || fallback.Y != 0 {
		return rl.NewVector2(float32(fallback.X), float32(fallback.Y))
	}
	return rl.NewVector2(1, 0)
}
