package simulation

import (
	"math"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
)

type DashPhase string

const (
	DashIdle       DashPhase = ""
	DashWinding    DashPhase = "windup"
	DashRushing    DashPhase = "burst"
	DashRecovering DashPhase = "recovery"
	DashBumped     DashPhase = "bump"
)

type CombatState struct {
	StompRemaining float64   `json:"stomp_remaining,omitempty"`
	StompDirection Direction `json:"stomp_direction"`
	StunRemaining  float64   `json:"stun_remaining,omitempty"`
	dashDuration   float64
	SwordRemaining float64   `json:"sword_remaining,omitempty"`
	SwordDirection Direction `json:"sword_direction"`
	Boosting       bool      `json:"boosting,omitempty"`
	Phase          DashPhase `json:"phase,omitempty"`
	Remaining      float64   `json:"remaining,omitempty"`
	Facing         Direction `json:"facing"`
	Direction      Direction `json:"direction"`
	Held           bool      `json:"held,omitempty"`
	Attack         bool      `json:"attack,omitempty"`
}

func DashBurstDuration(stats Stats) float64 {
	if stats.Mob && stats.MobType == Sentinel {
		return parameters.SentinelDashBurst
	}
	return parameters.DashBurst
}

func AdvanceDashPlayer(p *Player, input Direction, dt float64) {
	c := &p.Combat
	c.Attack = false
	if p.Dead || !input.Valid() || dt <= 0 {
		return
	}
	geometry := p.geometry()
	move := func(d Direction, distance float64) physics.Motion {
		delta := (physics.Vector{X: d.X, Y: d.Y}).Normalized().Scale(distance)
		body := physics.Square(p.X, p.Y, parameters.PlayerSize)
		if p.Stats.Mob {
			body = MobBody(p.Stats.MobType, p.X, p.Y)
		}
		motion := geometry.MoveBody(body, delta)
		motion.Bounds.X += p.X - body.X
		motion.Bounds.Y += p.Y - body.Y
		return motion
	}
	pressed := input.Dash && !c.Held
	walkSpeed := p.MovementSpeed()
	if p.Stats.Mob {
		walkSpeed *= min(1, math.Hypot(input.X, input.Y))
		attack := MobMelee(p.Stats.MobType)
		if c.SwordRemaining > 0 && attack.Duration-c.SwordRemaining < attack.Windup+attack.Active {
			walkSpeed *= attack.AdvanceMultiplier
		}
	}
	c.Held = input.Dash
	if length := math.Hypot(input.AimX, input.AimY); length > 0 {
		c.Facing = Direction{X: input.AimX / length, Y: input.AimY / length}
	} else if length := math.Hypot(input.X, input.Y); length > 0 {
		c.Facing = Direction{X: input.X / length, Y: input.Y / length}
	}
	if c.Facing.X == 0 && c.Facing.Y == 0 {
		c.Facing.X = 1
	}
	if c.Phase == DashIdle {
		if pressed && p.Stats.MP >= parameters.DashCost && !(p.Stats.Mob && p.Stats.MobType.IsHeavy()) {
			p.Stats.MP -= parameters.DashCost
			c.Direction = c.Facing
			c.Phase, c.Remaining = DashWinding, parameters.DashWindup
		}
	}
	for dt > 1e-9 {
		if c.Phase == DashIdle {
			motion := move(input, walkSpeed*parameters.SpeedScale*dt)
			p.X, p.Y = motion.Bounds.X, motion.Bounds.Y
			return
		}
		step := min(dt, c.Remaining)
		switch c.Phase {
		case DashRushing:
			c.Attack = true
			distance := p.MovementSpeed() * parameters.SpeedScale * parameters.DashSpeedMultiplier * step
			for distance > 1e-9 {
				travel := min(distance, 4)
				motion := move(c.Direction, travel)
				p.X, p.Y = motion.Bounds.X, motion.Bounds.Y
				if motion.Blocked() || motion.Rejected {
					p.Stats.HP = max(parameters.BoostMinimumMass, p.Stats.HP-parameters.BumpCost)
					bump := move(Direction{X: -c.Direction.X, Y: -c.Direction.Y}, parameters.BumpKnockback)
					p.X, p.Y = bump.Bounds.X, bump.Bounds.Y
					c.Phase, c.Remaining = DashBumped, parameters.BumpStun
					return
				}
				distance -= travel
			}
		case DashRecovering:
			motion := move(input, walkSpeed*parameters.SpeedScale*step*.35)
			p.X, p.Y = motion.Bounds.X, motion.Bounds.Y
		}
		dt -= step
		c.Remaining -= step
		if c.Remaining <= 1e-9 {
			switch c.Phase {
			case DashWinding:
				c.Phase, c.Remaining = DashRushing, DashBurstDuration(p.Stats)
				if p.Stats.Mob && p.Stats.MobType == Sentinel && c.dashDuration > 0 {
					c.Remaining = min(c.Remaining, c.dashDuration)
				}
			case DashRushing:
				c.Phase, c.Remaining = DashRecovering, parameters.DashRecovery
			default:
				c.Phase, c.Remaining = DashIdle, 0
			}
		}
	}
}
