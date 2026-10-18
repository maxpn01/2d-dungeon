package simulation

import (
	"math"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/worldmap"
)

func SideView(x, y float64) bool {
	return worldmap.Default().BodyFloor(x, y, parameters.PlayerSize) == worldmap.ThirdFloor
}

func advanceSideViewPlayer(p *Player, input Direction, dt float64) {
	facing := p.Combat.Facing
	p.Combat = CombatState{Facing: facing}
	if length := math.Hypot(input.AimX, input.AimY); length > 0 {
		p.Combat.Facing = Direction{X: input.AimX / length, Y: input.AimY / length}
	}
	size := float64(parameters.PlayerSize)
	geometry := p.geometry()
	jump := input.Jump || input.Y < 0
	grounded := !geometry.Walkable(p.X, p.Y+1, size, size)
	if grounded && p.FallSpeed >= 0 {
		p.FallSpeed = 0
		if jump {
			p.FallSpeed = -parameters.SideJumpSpeed
			grounded = false
		}
	}
	if !jump && p.FallSpeed < -parameters.SideJumpCutSpeed {
		p.FallSpeed = -parameters.SideJumpCutSpeed
	}
	p.FallSpeed = min(parameters.SideMaxFallSpeed, p.FallSpeed+parameters.SideGravity*dt)
	if grounded {
		p.FallSpeed = 0
	}
	if input.X != 0 {
		side := math.Copysign(1, input.X)
		p.Combat.Direction = Direction{X: side}
		if input.AimX == 0 && input.AimY == 0 {
			p.Combat.Facing = p.Combat.Direction
		}
		speed := p.MovementSpeed()
		if input.Boost && p.Stats.MP > 0 {
			boosted := min(dt, p.Stats.MP/p.Stats.BoostCost)
			p.Stats.MP = max(0, p.Stats.MP-boosted*p.Stats.BoostCost)
			speed *= 1 + (p.Stats.BoostMultiplier-1)*boosted/dt
			p.Combat.Boosting = boosted > 0
		}
		p.X, _ = geometry.Move(p.X, p.Y, side*speed*parameters.SpeedScale*dt, 0, size, size)
	}
	if p.FallSpeed != 0 {
		dy := p.FallSpeed * dt
		y := p.Y
		_, p.Y = geometry.Move(p.X, p.Y, 0, dy, size, size)
		if math.Abs(p.Y-(y+dy)) > 1e-6 {
			p.FallSpeed = 0
		}
	}
	p.Grounded = !geometry.Walkable(p.X, p.Y+1, size, size)
}
