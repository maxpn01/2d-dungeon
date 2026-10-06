package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

const (
	DashCost            = parameters.DashCost
	DashWindup          = parameters.DashWindup
	DashBurst           = parameters.DashBurst
	SentinelDashBurst   = parameters.SentinelDashBurst
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
	StompRemaining float64   `json:"stomp_remaining,omitempty"`
	StompDirection Direction `json:"stomp_direction"`
	StunRemaining  float64   `json:"stun_remaining,omitempty"`
	dashDuration   float64
	SwordRemaining float64   `json:"sword_remaining,omitempty"`
	SwordDirection Direction `json:"sword_direction"`
	Boosting       bool      `json:"boosting,omitempty"`
	RecoilX        float64   `json:"recoil_x,omitempty"`
	RecoilY        float64   `json:"recoil_y,omitempty"`
	HitGrace       float64   `json:"hit_grace,omitempty"`
	Phase          DashPhase `json:"phase,omitempty"`
	Remaining      float64   `json:"remaining,omitempty"`
	Facing         Direction `json:"facing"`
	Direction      Direction `json:"direction"`
	Held           bool      `json:"held,omitempty"`
	Attack         bool      `json:"attack,omitempty"`
}

func DashBurstDuration(stats Stats) float64 {
	if stats.Mob && stats.MobType == Sentinel {
		return SentinelDashBurst
	}
	return DashBurst
}

func AdvanceDashPlayer(p *Player, input Direction, dt float64) {
	oldX, oldY := p.X, p.Y
	defer func() { ConstrainArena(p, oldX, oldY) }()
	c := &p.Combat
	c.Attack = false
	if p.Dead || !input.Valid() || dt <= 0 {
		return
	}
	geometry := worldmap.Default().WithSecrets(p.OpenSecrets)
	move := func(d Direction, distance float64) physics.Motion {
		delta := (physics.Vector{X: d.X, Y: d.Y}).Normalized().Scale(distance)
		return geometry.MoveBody(physics.Square(p.X, p.Y, PlayerSize), delta)
	}
	c.HitGrace = max(0, c.HitGrace-dt)
	pressed := input.Dash && !c.Held
	walkSpeed := p.MovementSpeed()
	if p.Stats.Mob {
		walkSpeed *= min(1, math.Hypot(input.X, input.Y))
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
		if pressed && p.Stats.MP >= DashCost && !(p.Stats.Mob && p.Stats.MobType == DungeonHeavy) {
			p.Stats.MP -= DashCost
			c.Direction = c.Facing
			c.Phase, c.Remaining = DashWinding, DashWindup
		}
	}
	for dt > 1e-9 {
		if c.Phase == DashIdle {
			motion := move(input, walkSpeed*SpeedScale*dt)
			p.X, p.Y = motion.Bounds.X, motion.Bounds.Y
			return
		}
		step := min(dt, c.Remaining)
		switch c.Phase {
		case DashRushing:
			c.Attack = true
			distance := p.MovementSpeed() * SpeedScale * DashSpeedMultiplier * step
			for distance > 1e-9 {
				travel := min(distance, 4)
				motion := move(c.Direction, travel)
				p.X, p.Y = motion.Bounds.X, motion.Bounds.Y
				if motion.Blocked() || motion.Rejected {
					p.Stats.HP = max(BoostMinimumMass, p.Stats.HP-BumpCost)
					bump := move(Direction{X: -c.Direction.X, Y: -c.Direction.Y}, BumpKnockback)
					p.X, p.Y = bump.Bounds.X, bump.Bounds.Y
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
				motion := move(Direction{X: c.RecoilX / length, Y: c.RecoilY / length}, length*fraction)
				p.X, p.Y = motion.Bounds.X, motion.Bounds.Y
			}
		case DashRecovering:
			motion := move(input, walkSpeed*SpeedScale*step*.35)
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
