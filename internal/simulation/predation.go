package simulation

import (
	"fmt"
	"math"
	"sort"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/terrain"
)

type encounter struct {
	first, second eater
	time          float64
}

func (w *World) resolvePredation(eaters []eater) {
	contacts := make([]encounter, 0)
	for i, first := range eaters {
		for _, second := range eaters[i+1:] {
			if first.player != nil && first.player == second.player {
				continue
			}
			contact, hit := eatingContact(first.start, first.end, first.size, second.start, second.end, second.size)
			if hit {
				contacts = append(contacts, encounter{first: first, second: second, time: contact})
			}
		}
	}
	sort.Slice(contacts, func(i, j int) bool {
		a, b := contacts[i], contacts[j]
		if a.time != b.time {
			return a.time < b.time
		}
		if a.first.id != b.first.id {
			return a.first.id < b.first.id
		}
		return a.second.id < b.second.id
	})
	stopped := make(map[string]bool)
	for _, hit := range contacts {
		a, b := hit.first, hit.second
		if a.protectedAt(hit.time, w.safeZones) || b.protectedAt(hit.time, w.safeZones) {
			continue
		}
		if stopped[a.id] || stopped[b.id] || a.dead() || b.dead() || w.tick < w.protectedUntil[a.id] || w.tick < w.protectedUntil[b.id] {
			continue
		}
		attackA, attackB := a.attacking(), b.attacking()
		if attackA || attackB {
			a.end = position{a.start.x + (a.end.x-a.start.x)*hit.time, a.start.y + (a.end.y-a.start.y)*hit.time}
			b.end = position{b.start.x + (b.end.x-b.start.x)*hit.time, b.start.y + (b.end.y-b.start.y)*hit.time}
		}
		if attackA && attackB {
			stopped[a.id], stopped[b.id] = true, true
			w.separate(a, b)
			continue
		}
		if attackA || attackB {
			if attackB {
				a, b = b, a
			}
			stopped[a.id] = true
			a.setPosition(a.end.x, a.end.y)
			a.combat().Attack = false
			a.combat().Phase, a.combat().Remaining = DashRecovering, DashRecovery
			if a.stats.HP >= b.stats.HP && a.stats.HP > 0 {
				a.gain(b.stats.HP*ActorHPFraction, b.stats.HP)
				killer := a.id
				if a.player != nil {
					killer = a.player.ID
				}
				b.die(killer, w.tick)
			} else if a.stats.HP < b.stats.HP {
				w.knockMass(b)
				stopped[b.id] = true
				w.separate(a, b)
			}
			continue
		}
		if b.stats.HP > a.stats.HP {
			a, b = b, a
		}
		if a.ai != nil && (a.ai.Behavior == "sleeping" || a.ai.Behavior == "emerging") {
			continue
		}
		if a.stats.HP > b.stats.HP {
			a.gain(b.stats.HP*ActorHPFraction, b.stats.HP)
			killer := a.id
			if a.player != nil {
				killer = a.player.ID
			}
			b.die(killer, w.tick)
		}
	}
}

func (e eater) combat() *CombatState {
	if e.player != nil {
		return &e.player.Combat
	}
	return &e.ai.Combat
}

func (e eater) attacking() bool {
	return e.cell == nil && e.player != nil && e.player.Character == Triangle && e.combat().Attack
}

func (e eater) setPosition(x, y float64) {
	if e.cell != nil {
		e.cell.X, e.cell.Y = x, y
	} else if e.player != nil {
		e.player.X, e.player.Y = x, y
	} else {
		e.ai.X, e.ai.Y = x, y
	}
}

func (w *World) separate(a, b eater) {
	dx, dy := b.start.x-a.start.x, b.start.y-a.start.y
	length := math.Hypot(dx, dy)
	if length == 0 {
		dx, dy, length = 1, 0, 1
	}
	for i, e := range []eater{a, b} {
		sign := -1.0
		if i == 1 {
			sign = 1
		}
		x, y := MoveBy(e.end.x, e.end.y, Direction{X: sign * dx / length, Y: sign * dy / length}, PlayerSize)
		e.setPosition(x, y)
		w.protectedUntil[e.id] = w.tick + HitProtectionTicks
		e.combat().Attack = false
		if e.player != nil && e.player.Character == Triangle {
			e.combat().Phase, e.combat().Remaining = DashRecovering, DashRecovery
		}
	}
}

func (w *World) knockMass(target eater) {
	// Raids dislodge mass but leave the victim alive to defend or retreat.
	amount := min(KnockedMass, max(0, target.stats.HP-BoostMinimumMass))
	drops := 0
	for _, f := range w.fruits {
		if f.Source != "" {
			drops++
		}
	}
	for amount > 0 && drops < MaxMassDrops {
		hp := min(1, amount)
		angle := float64(drops) * 2.399963229728653
		x, y := terrain.Move(target.end.x, target.end.y, math.Cos(angle)*45, math.Sin(angle)*45, FruitSize, FruitSize)
		w.nextFruit++
		w.fruits = append(w.fruits, Fruit{ID: fmt.Sprintf("m%d", w.nextFruit), X: x, Y: y, Size: FruitSize, Source: target.id, HP: hp, AvailableTick: w.tick + HitProtectionTicks, ExpiresTick: w.tick + parameters.MassDropLifetimeTicks})
		target.stats.HP -= hp
		amount -= hp
		drops++
	}
}

func (w *World) spawnPosition(slot uint64) (float64, float64) {
	bestX, bestY, bestDistance := 900.0, 100.0, -1.0
	for offset := uint64(0); offset < 400; offset++ {
		cell := (slot + offset) % 400
		x, y := 700+float64(cell%10)*60, 60+float64(cell/10)*15
		distance := math.Inf(1)
		for _, ai := range w.ais {
			if ai.respawnTick == 0 {
				distance = min(distance, math.Hypot(x-ai.X, y-ai.Y))
			}
		}
		for _, player := range w.players {
			if !player.Dead {
				distance = min(distance, math.Hypot(x-player.X, y-player.Y))
			}
		}
		if distance > bestDistance {
			bestX, bestY, bestDistance = x, y, distance
		}
		if distance > AIDetectionRadius+PlayerSize {
			return x, y
		}
	}
	return bestX, bestY
}

func (w *World) nearestPlayer(ai *aiState, threat bool) (*playerState, float64) {
	radius := AIDetectionRadius
	if threat && ai.Behavior == "fleeing" {
		radius *= 1.2
	}
	best := radius * radius
	var selected *playerState
	for _, owner := range w.players {
		candidates := []playerState{*owner}
		for _, part := range owner.Split {
			if !part.Active {
				continue
			}
			cell := *owner
			cell.X, cell.Y, cell.Stats = part.X, part.Y, part.Stats
			candidates = append(candidates, cell)
		}
		for _, player := range candidates {
			if InSafeZone(player.X, player.Y, PlayerSize, owner.Stats.HP+owner.Split.Mass(), w.safeZones) {
				continue
			}
			if !ai.sees(player.X, player.Y) || !ai.inTerritory(player.X, player.Y) || player.Dead || player.Stats.HP <= 0 || threat && player.Stats.HP <= ai.Stats.HP || !threat && player.Stats.HP >= ai.Stats.HP {
				continue
			}
			dx, dy := player.X-ai.X, player.Y-ai.Y
			distance := dx*dx + dy*dy
			if distance <= best && (selected == nil || distance < best || player.ID < selected.ID) {
				copy := player
				selected, best = &copy, distance
			}
		}
	}

	return selected, best
}

func (ai *aiState) approach(x, y float64) Direction {
	ai.wanderTicks = 0
	dx, dy := x-ai.X, y-ai.Y
	if distance := math.Hypot(dx, dy); distance > 0 {
		return Direction{X: dx / distance, Y: dy / distance}
	}
	return Direction{}
}

func (ai *aiState) flee(player *playerState) Direction {
	dx, dy := ai.X-player.X, ai.Y-player.Y
	length := math.Hypot(dx, dy)
	d := Direction{X: 1}
	if length > 0 {
		d = Direction{X: dx / length, Y: dy / length}
	}
	if ai.escapeTicks > 0 {
		d = ai.escapeDirection
		ai.escapeTicks--
	}
	nx, ny := Move(ai.X, ai.Y, d, .15, ai.Stats.MovementSpeed())
	blocked := math.Hypot(nx-ai.X, ny-ai.Y) < ai.Stats.MovementSpeed()*SpeedScale*.1
	if blocked {
		best := -1.0
		for _, candidate := range [8]Direction{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}, {X: 1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: 1}, {X: -1, Y: -1}} {
			x, y := Move(ai.X, ai.Y, candidate, StepSeconds, ai.Stats.MovementSpeed())
			if x == ai.X && y == ai.Y {
				continue
			}
			score := (x-player.X)*(x-player.X) + (y-player.Y)*(y-player.Y)
			if score > best {
				best = score
				d = candidate
			}
		}
		ai.escapeDirection, ai.escapeTicks = d, 9
	}
	ai.wanderTicks = 0
	return d
}

func (ai *aiState) sees(x, y float64) bool {
	return terrain.Visible(ai.X+AISize/2, ai.Y+AISize/2, x+AISize/2, y+AISize/2)
}
func (ai *aiState) inTerritory(x, y float64) bool {
	return x >= MapMin && y >= MapMin && x <= MaxPosition && y <= MaxPosition
}
func (ai *aiState) navigate(x, y float64) Direction {
	dx, dy := x-ai.X, y-ai.Y
	mx, my := terrain.Move(ai.X, ai.Y, dx, dy, AISize, AISize)
	if math.Hypot(mx-x, my-y) < .1 {
		return ai.approach(x, y)
	}
	nx, ny, ok := terrain.NextStep(ai.X, ai.Y, x, y, AISize)
	if !ok {
		return Direction{}
	}
	return ai.approach(nx, ny)
}
