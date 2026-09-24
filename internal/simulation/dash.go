package simulation

import "math"

const (
	DashCost            = 1.0
	DashWindup          = 0.14
	DashBurst           = 0.15
	DashRecovery        = 0.45
	DashSpeedMultiplier = 3.0
	BumpStun            = 0.2
	BumpCost            = 1.0
	BumpKnockback       = 12.0
	ContactRatio        = 2.0
	KnockedMass         = 2.0
	HitProtectionTicks  = 8
	MaxMassDrops        = 128
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
	BossHitGrace       = 1.2
	BossRecoilDuration = .26
	BossImpactHold     = .07
	BossRecoilDistance = 85.0
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
		if pressed && p.Stats.HP >= DashCost {
			p.Stats.HP -= DashCost
			c.Direction = c.Facing
			c.Phase, c.Remaining = DashWinding, DashWindup
		}
	}
	for dt > 1e-9 {
		if c.Phase == DashIdle {
			p.X, p.Y = Move(p.X, p.Y, input, dt, p.Stats.Speed)
			return
		}
		step := min(dt, c.Remaining)
		switch c.Phase {
		case DashRushing:
			c.Attack = true
			distance := p.Stats.Speed * SpeedScale * DashSpeedMultiplier * step
			for distance > 1e-9 {
				travel := min(distance, 4)
				x, y := MoveBy(p.X, p.Y, c.Direction, travel)
				blocked := math.Abs(x-p.X-c.Direction.X*travel) > 1e-6 || math.Abs(y-p.Y-c.Direction.Y*travel) > 1e-6
				p.X, p.Y = x, y
				if blocked {
					p.Stats.HP = max(0, p.Stats.HP-BumpCost)
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
			p.X, p.Y = Move(p.X, p.Y, input, step, p.Stats.Speed*0.35)
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
