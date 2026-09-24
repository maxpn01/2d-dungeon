package simulation

import (
	"math"
	"raylib-2d-sandbox/internal/terrain"
)

type Character string

const (
	Square              Character = "square"
	Triangle            Character = "triangle"
	Circle              Character = "circle"
	SplitDuration                 = 8.0
	SplitLaunchDuration           = .25
	SplitMinimumMass              = 2.0
	MaxCircleCells                = 16
	SplitMergeDuration            = .35
	SplitTurnResponse             = 7.0
	SplitFollowResponse           = 24.0
)

func NormalizeCharacter(c Character) Character {
	switch c {
	case Triangle, Circle:
		return c
	default:
		return Square
	}
}

type SplitCells [MaxCircleCells - 1]SplitCell

func (cells SplitCells) Mass() float64 {
	total := 0.0
	for _, c := range cells {
		if c.Active {
			total += c.Stats.HP
		}
	}
	return total
}
func (cells SplitCells) Count() int {
	n := 0
	for _, c := range cells {
		if c.Active {
			n++
		}
	}
	return n
}

type SplitCell struct {
	Active            bool
	X, Y              float64
	Stats             Stats
	Remaining, Launch float64
	Direction         Direction
}

func (w *World) SetCharacter(id string, character Character) {
	if p := w.players[id]; p != nil {
		p.Stats.HP += p.Split.Mass()
		p.Split = SplitCells{}
		p.Character = NormalizeCharacter(character)
		p.Combat = CombatState{}
	}
}

func AdvancePlayer(p *Player, input Direction, dt float64, zones ...[]SafeZone) {
	if len(zones) > 0 {
		before := *p
		defer func() { RepelPlayer(p, before, zones[0]) }()
	}
	if p.Dead || !input.Valid() || dt <= 0 {
		return
	}
	switch NormalizeCharacter(p.Character) {
	case Triangle:
		AdvanceDashPlayer(p, input, dt)
	case Circle:
		advanceCircle(p, input, dt)
	default:
		advanceBoostPlayer(p, input, dt)
	}
}

func advanceCircle(p *Player, input Direction, dt float64) {
	c := &p.Combat
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
	if pressed {
		splitCircle(p, c.Facing)
	}
	// Before the first split, use the same movement integration as the square.
	// It prevents a separate idle path from fighting client prediction and
	// producing the visible back-and-forth correction.
	if p.Split.Count() == 0 {
		moveCircleBody(p, input, dt)
		return
	}
	oldX, oldY := p.X, p.Y
	p.X, p.Y = Move(p.X, p.Y, input, dt, p.MovementSpeed())
	dx, dy := p.X-oldX, p.Y-oldY
	count := p.Split.Count()
	if count == 0 {
		return
	}
	// Use normalized input, not tiny differences in resolved world coordinates.
	// At a 180-degree reversal always turn the same way, including diagonals.
	angle := math.Atan2(c.Direction.Y, c.Direction.X)
	if math.Hypot(input.X, input.Y) > 0 {
		desired := math.Atan2(-input.Y, -input.X)
		turn := math.Atan2(math.Sin(desired-angle), math.Cos(desired-angle))
		if math.Abs(math.Abs(turn)-math.Pi) < 1e-6 {
			turn = math.Pi
		}
		angle += turn * (1 - math.Exp(-SplitTurnResponse*dt))
		c.Direction = Direction{X: math.Cos(angle), Y: math.Sin(angle)}
	}
	rank := 0
	for i := range p.Split {
		cell := &p.Split[i]
		if !cell.Active {
			continue
		}
		cell.X, cell.Y = terrain.Move(cell.X, cell.Y, dx, dy, PlayerSize, PlayerSize)
		cell.Remaining = max(0, cell.Remaining-dt)
		cell.Launch = max(0, cell.Launch-dt)
		// Pack additional bodies into close rings, instead of stacking them.
		ring, slot, slots := 1.0, rank, min(count, 6)
		if rank >= 6 {
			ring, slot, slots = 2, rank-6, count-6
		}
		bodyAngle := angle + 2*math.Pi*float64(slot)/float64(slots)
		oldAngle := math.Atan2(cell.Direction.Y, cell.Direction.X)
		turn := math.Atan2(math.Sin(bodyAngle-oldAngle), math.Cos(bodyAngle-oldAngle))
		if math.Abs(math.Abs(turn)-math.Pi) < 1e-6 {
			turn = math.Pi
		}
		oldAngle += turn * (1 - math.Exp(-SplitTurnResponse*dt))
		desiredX, desiredY := math.Cos(oldAngle), math.Sin(oldAngle)
		cell.Direction = Direction{X: desiredX, Y: desiredY}
		spacing := ring * (BodySize(PlayerSize, p.Stats.HP) + BodySize(PlayerSize, cell.Stats.HP)) / 2
		smooth := func(t float64) float64 { t = max(0, min(1, t)); return t * t * (3 - 2*t) }
		spacing *= smooth(1-cell.Launch/SplitLaunchDuration) * smooth(cell.Remaining/SplitMergeDuration)
		targetX, targetY := terrain.Move(p.X, p.Y, desiredX*spacing, desiredY*spacing, PlayerSize, PlayerSize)
		if !terrain.Visible(cell.X+PlayerSize/2, cell.Y+PlayerSize/2, targetX+PlayerSize/2, targetY+PlayerSize/2) {
			if x, y, ok := terrain.NextStep(cell.X, cell.Y, targetX, targetY, PlayerSize); ok {
				targetX, targetY = x, y
			}
		}
		blend := 1 - math.Exp(-SplitFollowResponse*dt)
		moveX, moveY := (targetX-cell.X)*blend, (targetY-cell.Y)*blend
		distance := math.Hypot(moveX, moveY)
		limit := p.MovementSpeed() * SpeedScale * dt
		if distance > limit {
			moveX *= limit / distance
			moveY *= limit / distance
		}
		cell.X, cell.Y = terrain.Move(cell.X, cell.Y, moveX, moveY, PlayerSize, PlayerSize)
		if cell.Remaining == 0 && math.Hypot(cell.X-p.X, cell.Y-p.Y) < 1 {
			p.Stats.HP += cell.Stats.HP
			*cell = SplitCell{}
		}
		rank++
	}
}

func moveCircleBody(p *Player, input Direction, dt float64) {
	length := math.Hypot(input.X, input.Y)
	if length == 0 {
		return
	}
	p.Combat.Direction = Direction{X: input.X / length, Y: input.Y / length}
	p.Combat.Facing = p.Combat.Direction
	p.X, p.Y = Move(p.X, p.Y, input, dt, p.MovementSpeed())
}

func splitCircle(p *Player, direction Direction) {
	sources := []int{}
	if p.Stats.HP >= SplitMinimumMass {
		sources = append(sources, -1)
	}
	for i, c := range p.Split {
		if c.Active && c.Stats.HP >= SplitMinimumMass {
			sources = append(sources, i)
		}
	}
	if p.Split.Count() == 0 {
		p.Combat.Direction = direction
	}
	created := false
	for _, source := range sources {
		free := -1
		for i, c := range p.Split {
			if !c.Active {
				free = i
				break
			}
		}
		if free < 0 {
			break
		}
		stats, x, y := &p.Stats, p.X, p.Y
		if source >= 0 {
			stats = &p.Split[source].Stats
			x, y = p.Split[source].X, p.Split[source].Y
		}
		created = true
		stats.HP /= 2
		p.Split[free] = SplitCell{Active: true, X: x, Y: y, Stats: *stats, Remaining: SplitDuration, Launch: SplitLaunchDuration, Direction: direction}
	}
	if created {
		for i := range p.Split {
			if p.Split[i].Active {
				p.Split[i].Remaining = SplitDuration
			}
		}
	}
}

func (w *World) finishSplitCells() {
	for _, p := range w.players {
		for i := range p.Split {
			cell := &p.Split[i]
			if !cell.Active {
				continue
			}
			if cell.Stats.HP <= 0 {
				*cell = SplitCell{}
				continue
			}
			if p.Stats.HP <= 0 {
				p.X, p.Y, p.Stats.HP = cell.X, cell.Y, cell.Stats.HP
				*cell = SplitCell{}
			}
		}
	}
}

// BodySize is shared with rendering so clustered circles touch at any mass.
func BodySize(size, hp float64) float64 {
	return size * math.Pow(max(1, hp)/StartingHP, .3)
}
