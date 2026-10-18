package simulation

import (
	"math"

	"dungeon-raylib-go/internal/parameters"
)

const WingMPPerSecond = 2.0
const WingSpeedMultiplier = 1.2

type FlightState struct {
	Flying    bool
	Returning bool
	LandingX  float64
	LandingY  float64
}

func FlightSpeedMultiplier(stats Stats) float64 {
	return WingSpeedMultiplier * (1 + PerkBonus(BoostPerk, stats.Perks[BoostPerk]))
}

func FlightMPCost(stats Stats) float64 {
	return WingMPPerSecond * (1 + PerkPenalty(BoostPerk, stats.Perks[BoostPerk]))
}

func advanceWingPlayer(p *Player, input Direction, dt float64) {
	geometry := p.geometry()
	grounded := geometry.Walkable(p.X, p.Y, parameters.PlayerSize, parameters.PlayerSize)
	f := &p.Flight
	if grounded {
		f.LandingX, f.LandingY = p.X, p.Y
	}
	cost := FlightMPCost(p.Stats)
	powered := input.Boost && p.Stats.MP >= cost*dt
	if f.Flying && !powered && !grounded && !f.Returning {
		if x, y, found := geometry.NearestLanding(p.X, p.Y, parameters.PlayerSize); found {
			f.LandingX, f.LandingY = x, y
		}
		f.Returning = true
	}
	if grounded && (!powered || f.Returning) {
		*f = FlightState{LandingX: p.X, LandingY: p.Y}
	}
	if !f.Flying && powered {
		f.Flying = true
	}
	if !f.Flying {
		input.Boost = false
		advanceBoostPlayer(p, input, dt)
		return
	}
	p.Combat = CombatState{Facing: p.Combat.Facing}
	if length := math.Hypot(input.AimX, input.AimY); length > 0 {
		p.Combat.Facing = Direction{X: input.AimX / length, Y: input.AimY / length}
	}
	dx, dy := input.X, input.Y
	speed := p.MovementSpeed() * FlightSpeedMultiplier(p.Stats) * parameters.SpeedScale
	if f.Returning {
		if x, y, found := geometry.FlightLandingStep(p.X, p.Y, f.LandingX, f.LandingY, parameters.PlayerSize); found {
			dx, dy = x-p.X, y-p.Y
		} else {
			dx, dy = 0, 0
		}
		speed *= 2
	} else {
		p.Stats.MP = max(0, p.Stats.MP-cost*dt)
	}
	length := math.Hypot(dx, dy)
	if length > 0 {
		distance := speed * dt
		if f.Returning {
			distance = min(distance, length)
		}
		p.Combat.Direction = Direction{X: dx / length, Y: dy / length}
		p.X, p.Y = geometry.MoveFlying(p.X, p.Y, dx/length*distance, dy/length*distance, parameters.PlayerSize)
		if input.AimX == 0 && input.AimY == 0 {
			p.Combat.Facing = p.Combat.Direction
		}
	}
	if geometry.Walkable(p.X, p.Y, parameters.PlayerSize, parameters.PlayerSize) {
		f.LandingX, f.LandingY = p.X, p.Y
		if f.Returning {
			*f = FlightState{LandingX: p.X, LandingY: p.Y}
		}
	}
}
