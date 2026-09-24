package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/terrain"
)

const EjectInterval = parameters.EjectInterval
const EjectOwnerDelay = parameters.EjectOwnerDelay

func EjectFood(p *Player, input Direction, dt float64) *Fruit {
	p.EjectCooldown = max(0, p.EjectCooldown-dt)
	if p.Dead || !input.Valid() || !input.Eject || p.EjectCooldown > 1e-9 {
		return nil
	}
	stats, x, y := &p.Stats, p.X, p.Y
	for i := range p.Split {
		cell := &p.Split[i]
		if cell.Active && cell.Stats.HP > stats.HP {
			stats, x, y = &cell.Stats, cell.X, cell.Y
		}
	}
	if stats.HP < 2 {
		return nil
	}
	dx, dy := input.AimX, input.AimY
	if math.Hypot(dx, dy) == 0 {
		dx, dy = input.X, input.Y
	}
	if math.Hypot(dx, dy) == 0 {
		dx, dy = p.Combat.Facing.X, p.Combat.Facing.Y
	}
	length := math.Hypot(dx, dy)
	if length == 0 {
		dx, dy, length = 1, 0, 1
	}
	dx, dy = dx/length, dy/length
	offset := max(PlayerSize, BodySize(PlayerSize, stats.HP))/2 + FruitSize + 5
	startX, startY := x+(PlayerSize-FruitSize)/2, y+(PlayerSize-FruitSize)/2
	fx, fy := terrain.Move(startX, startY, dx*offset, dy*offset, FruitSize, FruitSize)
	stats.HP--
	p.EjectCooldown = EjectInterval
	return &Fruit{X: fx, Y: fy, Size: FruitSize, HP: 1, Ejected: true, VelocityX: dx * 450, VelocityY: dy * 450}
}

func (w *World) advanceFood() {
	for i := range w.fruits {
		f := &w.fruits[i]
		if !f.Ejected {
			continue
		}
		f.X, f.Y = terrain.Move(f.X, f.Y, f.VelocityX*StepSeconds, f.VelocityY*StepSeconds, f.Size, f.Size)
		decay := math.Exp(-7 * StepSeconds)
		f.VelocityX *= decay
		f.VelocityY *= decay
	}
}
