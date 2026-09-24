package simulation

import "math"

const (
	BoostMultiplier    = 1.5
	BoostMassPerSecond = 1.0
	BoostMinimumMass   = 1.0
)

func AdvancePlayer(p *Player, input Direction, dt float64) {
	p.Combat = CombatState{}
	if p.Dead || !input.Valid() || dt <= 0 {
		return
	}
	length := math.Hypot(input.X, input.Y)
	if length == 0 {
		return
	}
	p.Combat.Direction = Direction{X: input.X / length, Y: input.Y / length}
	p.Combat.Facing = p.Combat.Direction
	boosted := 0.0
	if input.Boost && p.Stats.HP > BoostMinimumMass {
		boosted = min(dt, (p.Stats.HP-BoostMinimumMass)/BoostMassPerSecond)
	}
	oldX, oldY := p.X, p.Y
	p.X, p.Y = Move(p.X, p.Y, input, boosted, p.Stats.Speed*BoostMultiplier)
	if p.X != oldX || p.Y != oldY {
		p.Stats.HP = max(BoostMinimumMass, p.Stats.HP-boosted*BoostMassPerSecond)
		p.Combat.Boosting = boosted > 0
	}
	p.X, p.Y = Move(p.X, p.Y, input, dt-boosted, p.Stats.Speed)
}
