package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/worldmap"
)

const WingMPPerSecond = 2.0
const WingSpeedMultiplier = 1.2

type FlightState struct {
	Flying    bool    `json:"flying"`
	Returning bool    `json:"returning"`
	LandingX  float64 `json:"landing_x"`
	LandingY  float64 `json:"landing_y"`
}

func advanceWingPlayer(p *Player, input Direction, dt float64) {
	geometry := worldmap.Default().WithSecrets(p.OpenSecrets).WithVaultGate(p.VaultGateOpen)
	grounded := geometry.Walkable(p.X, p.Y, PlayerSize, PlayerSize)
	f := &p.Flight
	if grounded {
		f.LandingX, f.LandingY = p.X, p.Y
	}
	powered := input.Boost && p.Stats.MP >= WingMPPerSecond*dt
	if f.Flying && !powered && !grounded && !f.Returning {
		if x, y, found := geometry.NearestLanding(p.X, p.Y, PlayerSize); found {
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
	speed := p.MovementSpeed() * WingSpeedMultiplier * SpeedScale
	if f.Returning {
		if x, y, found := geometry.FlightLandingStep(p.X, p.Y, f.LandingX, f.LandingY, PlayerSize); found {
			dx, dy = x-p.X, y-p.Y
		} else {
			dx, dy = 0, 0
		}
		speed *= 2
	} else {
		p.Stats.MP = max(0, p.Stats.MP-WingMPPerSecond*dt)
	}
	length := math.Hypot(dx, dy)
	if length > 0 {
		distance := speed * dt
		if f.Returning {
			distance = min(distance, length)
		}
		p.Combat.Direction = Direction{X: dx / length, Y: dy / length}
		p.X, p.Y = geometry.MoveFlying(p.X, p.Y, dx/length*distance, dy/length*distance, PlayerSize)
		if input.AimX == 0 && input.AimY == 0 {
			p.Combat.Facing = p.Combat.Direction
		}
	}
	if geometry.Walkable(p.X, p.Y, PlayerSize, PlayerSize) {
		f.LandingX, f.LandingY = p.X, p.Y
		if f.Returning {
			*f = FlightState{LandingX: p.X, LandingY: p.Y}
		}
	}
}
