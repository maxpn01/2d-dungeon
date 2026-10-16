package simulation

import (
	"math"
	"sort"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/worldmap"
)

const (
	Circle              Character = "circle"
	SplitDuration                 = parameters.SplitDuration
	SplitLaunchDuration           = parameters.SplitLaunchDuration
	SplitMinimumMass              = parameters.SplitMinimumMass
	SplitMergeDuration            = parameters.SplitMergeDuration
	SplitTurnResponse             = parameters.SplitTurnResponse
	SplitFollowResponse           = parameters.SplitFollowResponse
)

type SplitCells []SplitCell

type SplitCell struct {
	Active            bool
	X, Y              float64
	Stats             Stats
	Remaining, Launch float64
	Direction         Direction
}

func (cells SplitCells) Clone() SplitCells { return append(SplitCells(nil), cells...) }

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
	if p.Split.Count() == 0 {
		moveCircleBody(p, input, dt)
		return
	}
	speed := p.MovementSpeed()
	turnBlend := 1 - math.Exp(-SplitTurnResponse*dt)
	followBlend := 1 - math.Exp(-SplitFollowResponse*dt)
	oldX, oldY := p.X, p.Y
	p.X, p.Y = Move(p.X, p.Y, input, dt, speed)
	dx, dy := p.X-oldX, p.Y-oldY
	order := make([]int, 0, len(p.Split))
	for i := range p.Split {
		if p.Split[i].Active {
			order = append(order, i)
		}
	}
	sort.Slice(order, func(i, j int) bool {
		a, b := p.Split[order[i]].Stats.HP, p.Split[order[j]].Stats.HP
		return a > b || a == b && order[i] < order[j]
	})
	count := len(order)
	angle := math.Atan2(c.Direction.Y, c.Direction.X)
	if math.Hypot(input.X, input.Y) > 0 {
		desired := math.Atan2(-input.Y, -input.X)
		turn := math.Atan2(math.Sin(desired-angle), math.Cos(desired-angle))
		if math.Abs(math.Abs(turn)-math.Pi) < 1e-6 {
			turn = math.Pi
		}
		angle += turn * turnBlend
		c.Direction = Direction{X: math.Cos(angle), Y: math.Sin(angle)}
	}
	mainBodySize := BodySize(PlayerSize, p.Stats.HP)
	rank, ring, ringStart := 0, 1, 0
	ringDiameter := BodySize(PlayerSize, p.Split[order[0]].Stats.HP)
	ringRadius := (mainBodySize + ringDiameter) / 2
	for _, i := range order {
		cell := &p.Split[i]
		cell.X, cell.Y = worldmap.Default().Move(cell.X, cell.Y, dx, dy, PlayerSize, PlayerSize)
		cell.Remaining = max(0, cell.Remaining-dt)
		cell.Launch = max(0, cell.Launch-dt)
		if rank >= ringStart+6*ring {
			ringStart += 6 * ring
			ring++
			nextDiameter := BodySize(PlayerSize, cell.Stats.HP)
			ringRadius += (ringDiameter + nextDiameter) / 2
			ringDiameter = nextDiameter
		}
		slot, slots := rank-ringStart, min(count-ringStart, 6*ring)
		bodyAngle := angle + 2*math.Pi*float64(slot)/float64(slots)
		oldAngle := math.Atan2(cell.Direction.Y, cell.Direction.X)
		turn := math.Atan2(math.Sin(bodyAngle-oldAngle), math.Cos(bodyAngle-oldAngle))
		if math.Abs(math.Abs(turn)-math.Pi) < 1e-6 {
			turn = math.Pi
		}
		oldAngle += turn * turnBlend
		desiredX, desiredY := math.Cos(oldAngle), math.Sin(oldAngle)
		cell.Direction = Direction{X: desiredX, Y: desiredY}
		spacing := ringRadius
		smooth := func(t float64) float64 { t = max(0, min(1, t)); return t * t * (3 - 2*t) }
		spacing *= smooth(1-cell.Launch/SplitLaunchDuration) * smooth(cell.Remaining/SplitMergeDuration)
		targetX, targetY := worldmap.Default().Move(p.X, p.Y, desiredX*spacing, desiredY*spacing, PlayerSize, PlayerSize)
		if !worldmap.Default().Visible(cell.X+PlayerSize/2, cell.Y+PlayerSize/2, targetX+PlayerSize/2, targetY+PlayerSize/2) {
			if x, y, ok := worldmap.Default().NextStep(cell.X, cell.Y, targetX, targetY, PlayerSize); ok {
				targetX, targetY = x, y
			}
		}
		moveX, moveY := (targetX-cell.X)*followBlend, (targetY-cell.Y)*followBlend
		distance := math.Hypot(moveX, moveY)
		limit := speed * SpeedScale * dt
		if distance > limit {
			moveX *= limit / distance
			moveY *= limit / distance
		}
		cell.X, cell.Y = worldmap.Default().Move(cell.X, cell.Y, moveX, moveY, PlayerSize, PlayerSize)
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
	if p.Stats.MP < parameters.SplitMPCost {
		return
	}
	sources := make([]int, 0, len(p.Split)+1)
	free := make([]int, 0)
	if p.Stats.HP >= SplitMinimumMass {
		sources = append(sources, -1)
	}
	for i, c := range p.Split {
		if !c.Active {
			free = append(free, i)
		} else if c.Stats.HP >= SplitMinimumMass {
			sources = append(sources, i)
		}
	}
	if len(sources) == 0 {
		return
	}
	p.Stats.MP -= parameters.SplitMPCost
	if p.Split.Count() == 0 {
		p.Combat.Direction = direction
	}
	for _, source := range sources {
		stats, x, y := &p.Stats, p.X, p.Y
		if source >= 0 {
			stats = &p.Split[source].Stats
			x, y = p.Split[source].X, p.Split[source].Y
		}
		stats.HP /= 2
		child := SplitCell{Active: true, X: x, Y: y, Stats: *stats, Remaining: SplitDuration, Launch: SplitLaunchDuration, Direction: direction}
		child.Stats.MP = 0
		if len(free) > 0 {
			index := free[len(free)-1]
			free = free[:len(free)-1]
			p.Split[index] = child
		} else {
			p.Split = append(p.Split, child)
		}
	}
	if len(sources) > 0 {
		for i := range p.Split {
			if p.Split[i].Active {
				p.Split[i].Remaining = SplitDuration
			}
		}
	}
}

func (w *World) finishSplitCells() {
	for _, p := range w.players {
		finishActorSplit(&p.X, &p.Y, &p.Stats, &p.Split)
	}
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick == 0 {
			finishActorSplit(&ai.X, &ai.Y, &ai.Stats, &ai.Split)
		}
	}
}

func finishActorSplit(x, y *float64, stats *Stats, split *SplitCells) {
	for i := range *split {
		cell := &(*split)[i]
		if !cell.Active {
			continue
		}
		if cell.Stats.HP <= 0 {
			*cell = SplitCell{}
			continue
		}
		if stats.HP <= 0 {
			*x, *y, stats.HP = cell.X, cell.Y, cell.Stats.HP
			*cell = SplitCell{}
		}
	}
	for len(*split) > 0 && !(*split)[len(*split)-1].Active {
		*split = (*split)[:len(*split)-1]
	}
}
