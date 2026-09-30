package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
)

const (
	DashCost            = parameters.DashCost
	DashWindup          = parameters.DashWindup
	DashBurst           = parameters.DashBurst
	DashRecovery        = parameters.DashRecovery
	DashSpeedMultiplier = parameters.DashSpeedMultiplier
	BumpStun            = parameters.BumpStun
	BumpCost            = parameters.BumpCost
	BumpKnockback       = parameters.BumpKnockback
	ContactRatio        = parameters.ContactRatio
	KnockedMass         = parameters.KnockedMass
	HitProtectionTicks  = parameters.HitProtectionTicks
	MaxMassDrops        = parameters.MaxMassDrops
)

type DashPhase string

const (
	DashIdle       DashPhase = ""
	DashWinding    DashPhase = "windup"
	DashRushing    DashPhase = "burst"
	DashRecovering DashPhase = "recovery"
	DashBumped     DashPhase = "bump"
	DashRecoiling  DashPhase = "recoil"
)

const (
	BossHitGrace       = parameters.BossHitGrace
	BossRecoilDuration = parameters.BossRecoilDuration
	BossImpactHold     = parameters.BossImpactHold
	BossRecoilDistance = parameters.BossRecoilDistance
)

type CombatState struct {
	Boosting  bool      `json:"boosting,omitempty"`
	RecoilX   float64   `json:"recoil_x,omitempty"`
	RecoilY   float64   `json:"recoil_y,omitempty"`
	HitGrace  float64   `json:"hit_grace,omitempty"`
	Phase     DashPhase `json:"phase,omitempty"`
	Remaining float64   `json:"remaining,omitempty"`
	Facing    Direction `json:"facing"`
	Direction Direction `json:"direction"`
	Held      bool      `json:"held,omitempty"`
	Attack    bool      `json:"attack,omitempty"`
}

func AdvanceDashPlayer(p *Player, input Direction, dt float64) {
	oldX, oldY := p.X, p.Y
	defer func() { ConstrainArena(p, oldX, oldY) }()
	c := &p.Combat
	c.Attack = false
	if p.Dead || !input.Valid() || dt <= 0 {
		return
	}
	c.HitGrace = max(0, c.HitGrace-dt)
	pressed := input.Dash && !c.Held
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
		if pressed && p.Stats.MP >= DashCost {
			p.Stats.MP -= DashCost
			c.Direction = c.Facing
			c.Phase, c.Remaining = DashWinding, DashWindup
		}
	}
	for dt > 1e-9 {
		if c.Phase == DashIdle {
			p.X, p.Y = Move(p.X, p.Y, input, dt, p.MovementSpeed())
			return
		}
		step := min(dt, c.Remaining)
		switch c.Phase {
		case DashRushing:
			c.Attack = true
			distance := p.MovementSpeed() * SpeedScale * DashSpeedMultiplier * step
			for distance > 1e-9 {
				travel := min(distance, 4)
				motion := moveByMotion(p.X, p.Y, c.Direction, travel)
				p.X, p.Y = motion.Bounds.X, motion.Bounds.Y
				if motion.Blocked() || motion.Rejected {
					p.Stats.HP = max(BoostMinimumMass, p.Stats.HP-BumpCost)
					p.X, p.Y = MoveBy(p.X, p.Y, Direction{X: -c.Direction.X, Y: -c.Direction.Y}, BumpKnockback)
					c.Phase, c.Remaining = DashBumped, BumpStun
					return
				}
				distance -= travel
			}
		case DashRecoiling:
			progress := func(elapsed float64) float64 {
				t := max(0, min(1, (elapsed-BossImpactHold)/(BossRecoilDuration-BossImpactHold)))
				return 1 - (1-t)*(1-t)
			}
			elapsed := BossRecoilDuration - c.Remaining
			fraction := progress(elapsed+step) - progress(elapsed)
			length := math.Hypot(c.RecoilX, c.RecoilY)
			if length > 0 {
				p.X, p.Y = MoveBy(p.X, p.Y, Direction{X: c.RecoilX / length, Y: c.RecoilY / length}, length*fraction)
			}
		case DashRecovering:
			p.X, p.Y = Move(p.X, p.Y, input, step, p.MovementSpeed()*0.35)
		}
		dt -= step
		c.Remaining -= step
		if c.Remaining <= 1e-9 {
			switch c.Phase {
			case DashWinding:
				c.Phase, c.Remaining = DashRushing, DashBurst
			case DashRushing:
				c.Phase, c.Remaining = DashRecovering, DashRecovery
			default:
				c.Phase, c.Remaining = DashIdle, 0
			}
		}
	}
}

func CanContactEat(hp, preyHP float64) bool {
	return hp > 0 && (preyHP == 0 || hp >= preyHP*ContactRatio)
}
