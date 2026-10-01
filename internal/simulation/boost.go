package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
)

const (
	BoostMultiplier  = parameters.BoostMultiplier
	BoostMPPerSecond = parameters.BoostMPPerSecond
	BoostMinimumMass = parameters.BoostMinimumMass
)

func advanceBoostPlayer(p *Player, input Direction, dt float64) {
	facing := p.Combat.Facing
	p.Combat = CombatState{Facing: facing}
	if length := math.Hypot(input.AimX, input.AimY); length > 0 {
		p.Combat.Facing = Direction{X: input.AimX / length, Y: input.AimY / length}
	}
	if p.Dead || !input.Valid() || dt <= 0 {
		return
	}
	length := math.Hypot(input.X, input.Y)
	if length == 0 {
		return
	}
	p.Combat.Direction = Direction{X: input.X / length, Y: input.Y / length}
	if input.AimX == 0 && input.AimY == 0 {
		p.Combat.Facing = p.Combat.Direction
	}
	boosted := 0.0
	if input.Boost && p.Stats.MP > 0 {
		boosted = min(dt, p.Stats.MP/p.Stats.BoostCost)
	}
	oldX, oldY := p.X, p.Y
	p.X, p.Y = Move(p.X, p.Y, input, boosted, p.MovementSpeed()*p.Stats.BoostMultiplier)
	if p.X != oldX || p.Y != oldY {
		p.Stats.MP = max(0, p.Stats.MP-boosted*p.Stats.BoostCost)
		p.Combat.Boosting = boosted > 0
	}
	p.X, p.Y = Move(p.X, p.Y, input, dt-boosted, p.MovementSpeed())
}
