package simulation

import (
	"fmt"
	"math"
	"raylib-2d-sandbox/internal/terrain"
	"sort"
)

type encounter struct {
	first, second eater
	time          float64
}

func (w *World) resolvePredation(eaters []eater) {
	contacts := make([]encounter, 0)
	for i, first := range eaters {
		for _, second := range eaters[i+1:] {
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
		if stopped[a.id] || stopped[b.id] || a.dead() || b.dead() || w.tick < w.protectedUntil[a.id] || w.tick < w.protectedUntil[b.id] {
			continue
		}
		attackA, attackB := a.attacking(), b.attacking()
		if attackA || attackB {
			a.end = position{a.start.x + (a.end.x-a.start.x)*hit.time, a.start.y + (a.end.y-a.start.y)*hit.time}
			b.end = position{b.start.x + (b.end.x-b.start.x)*hit.time, b.start.y + (b.end.y-b.start.y)*hit.time}
		}
		if attackA && attackB {
			w.separate(a, b)
			continue
		}
		if attackA || attackB {
			if attackB {
				a, b = b, a
			}
			stopped[a.id] = true
			a.setPosition(a.end.x, a.end.y)
			a.player.Combat.Attack = false
			a.player.Combat.Phase, a.player.Combat.Remaining = DashRecovering, DashRecovery
			if a.stats.HP >= b.stats.HP && a.stats.HP > 0 {
				a.stats.Eat(b.stats.HP*ActorHPFraction, b.stats.HP)
				b.die(a.id, w.tick)
			} else if a.stats.HP < b.stats.HP {
				w.knockMass(b)
				w.separate(a, b)
			}
			continue
		}
		if b.stats.HP > a.stats.HP {
			a, b = b, a
		}
		if a.ai != nil && a.ai.Behavior == "sleeping" {
			continue
		}
		if CanContactEat(a.stats.HP, b.stats.HP) {
			a.stats.Eat(b.stats.HP*ActorHPFraction, b.stats.HP)
			b.die(a.id, w.tick)
		}
	}
}

func (e eater) attacking() bool { return e.player != nil && e.player.Combat.Attack }

func (e eater) setPosition(x, y float64) {
	if e.player != nil {
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
		if e.player != nil {
			e.player.Combat.Attack = false
			e.player.Combat.Phase, e.player.Combat.Remaining = DashRecovering, DashRecovery
		}
	}
}

func (w *World) knockMass(target eater) {
	amount := min(KnockedMass, target.stats.HP)
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
		w.fruits = append(w.fruits, Fruit{ID: fmt.Sprintf("m%d", w.nextFruit), X: x, Y: y, Size: FruitSize, Source: target.id, HP: hp, AvailableTick: w.tick + HitProtectionTicks, ExpiresTick: w.tick + 30*TickRate})
		target.stats.HP -= hp
		amount -= hp
		drops++
	}
}

func (w *World) spawnPosition(slot uint64) (float64, float64) {
	bestX, bestY, bestDistance := 300.0, 300.0, -1.0
	for offset := uint64(0); offset < 400; offset++ {
		cell := (slot + offset) % 400
		x, y := 300+float64(cell%20)*60, 300+float64(cell/20)*60
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
	for _, player := range w.players {
		if player.Dead || threat && player.Stats.HP <= ai.Stats.HP || !threat && player.Stats.HP >= ai.Stats.HP {
			continue
		}
		dx, dy := player.X-ai.X, player.Y-ai.Y
		distance := dx*dx + dy*dy
		if distance <= best && (selected == nil || distance < best || player.ID < selected.ID) {
			selected, best = player, distance
		}
	}
	return selected, best
}

func (ai *aiState) approach(x, y float64) {
	dx, dy := x-ai.X, y-ai.Y
	distance := math.Hypot(dx, dy)
	if distance > 0 {
		ai.X, ai.Y = Move(ai.X, ai.Y, Direction{X: dx / distance, Y: dy / distance}, min(StepSeconds, distance/(ai.Stats.Speed*SpeedScale)), ai.Stats.Speed)
	}
	ai.wanderTicks = 0
}

func (ai *aiState) flee(player *playerState) {
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
	blocked := (ai.X <= MapMin && d.X < 0) || (ai.X >= MaxPosition && d.X > 0) || (ai.Y <= MapMin && d.Y < 0) || (ai.Y >= MaxPosition && d.Y > 0)
	if blocked {
		best := -1.0
		for _, candidate := range [8]Direction{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}, {X: 1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: 1}, {X: -1, Y: -1}} {
			x, y := Move(ai.X, ai.Y, candidate, StepSeconds, ai.Stats.Speed)
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
	ai.X, ai.Y = Move(ai.X, ai.Y, d, StepSeconds, ai.Stats.Speed)
	ai.wanderTicks = 0
}
