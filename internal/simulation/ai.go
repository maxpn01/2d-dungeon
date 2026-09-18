package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"
	"raylib-2d-sandbox/internal/terrain"
)

const (
	InitialAISpawnDelay   = 5
	AIRespawnSeconds      = 45
	AIAwakeSeconds        = 120
	AIAsleepSeconds       = 60
	AISize                = PlayerSize
	AISpeed               = PlayerSpeed * 0.75
	AIDetectionRadius     = 300.0
	AISettleTicks         = 2 * TickRate
	AIReactionTicks       = TickRate
	AIAttackIntervalTicks = 2 * TickRate
)

type AI struct {
	Combat   CombatState
	Behavior string
	ID       string
	X, Y     float64
	Stats    Stats
}

type aiState struct {
	AI
	den             terrain.Den
	settleTicks     int
	targetID        string
	targetTicks     int
	attackCooldown  int
	returning       bool
	direction       Direction
	wanderTicks     int
	respawnTick     uint64
	awakeTicks      int
	asleepTicks     int
	escapeTicks     int
	escapeDirection Direction
	evadeTicks      int
	evadeDirection  Direction
}

func (w *World) newDenAI(den terrain.Den) aiState {
	w.nextAI++
	x, y := terrain.CellPosition(den.X, den.Y, AISize)
	stats := NewStats()
	stats.Speed = AISpeed
	stats.Level = den.Level
	stats.HP = den.HP
	return aiState{den: den, settleTicks: AISettleTicks, awakeTicks: AIAwakeSeconds * TickRate, AI: AI{ID: fmt.Sprintf("ai-%d", w.nextAI), X: x, Y: y, Stats: stats, Behavior: "emerging"}}
}
func (w *World) spawnAIs() {
	for _, den := range terrain.Dens {
		ai := w.newDenAI(den)
		if !w.denClear(ai.X, ai.Y, -1) {
			ai.respawnTick = w.tick + TickRate
		}
		w.ais = append(w.ais, ai)
	}
}

func (w *World) updateAIs() {
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick != 0 {
			continue
		}
		if ai.attackCooldown > 0 {
			ai.attackCooldown--
		}
		input := w.aiInput(ai)
		if input.Dash {
			ai.attackCooldown = AIAttackIntervalTicks
		}
		actor := Player{X: ai.X, Y: ai.Y, Stats: ai.Stats, Combat: ai.Combat}
		AdvancePlayer(&actor, input, StepSeconds)
		ai.X, ai.Y, ai.Stats, ai.Combat = actor.X, actor.Y, actor.Stats, actor.Combat
	}
}

func (w *World) aiInput(ai *aiState) Direction {
	if ai.settleTicks > 0 {
		ai.settleTicks--
		ai.Behavior = "emerging"
		return Direction{}
	}
	hx, hy := terrain.CellPosition(ai.den.X, ai.den.Y, AISize)
	distance := math.Hypot(ai.X-hx, ai.Y-hy)
	if distance > float64(ai.den.Radius*terrain.Tile) {
		ai.returning = true
	}
	if ai.returning {
		if distance < terrain.Tile {
			ai.returning = false
		} else {
			ai.Behavior = "returning"
			return ai.navigate(hx, hy)
		}
	}
	incoming := w.incomingDash(ai)
	threat, _ := w.nearestPlayer(ai, true)
	prey, preyDistance := w.nearestPlayer(ai, false)
	target := prey
	if threat != nil {
		target = threat
	}
	ai.observeTarget(target)
	if ai.updateSleep(threat != nil || incoming != nil || ai.evadeTicks > 0 || ai.Combat.Phase != DashIdle) {
		return Direction{}
	}
	if ai.evadeTicks > 0 {
		ai.evadeTicks--
		ai.Behavior = "evading"
		return ai.evadeDirection
	}
	if incoming != nil {
		ai.Behavior = "evading"
		ai.wanderTicks, ai.escapeTicks = 0, 0
		d := ai.evade(incoming)
		duration := incoming.Combat.Remaining
		if incoming.Combat.Phase == DashWinding {
			duration += DashBurst
		}
		ai.evadeTicks = int(math.Ceil(duration/StepSeconds)) + 1
		ai.evadeDirection = Direction{X: d.X, Y: d.Y}
		return d
	}
	if threat != nil {
		if ai.canAttack(threat) {
			ai.Behavior = "attacking"
			d := ai.approach(threat.X, threat.Y)
			d.Dash = true
			return d
		}
		ai.Behavior = "fleeing"
		return ai.flee(threat)
	}
	ai.escapeTicks = 0
	fruit, found := w.nearestFruit(ai)
	fruitX, fruitY := fruit.X+fruit.Size/2-AISize/2, fruit.Y+fruit.Size/2-AISize/2
	fruitDistance := (fruitX-ai.X)*(fruitX-ai.X) + (fruitY-ai.Y)*(fruitY-ai.Y)
	if prey != nil && (!found || preyDistance <= fruitDistance) {
		if ai.targetTicks < AIReactionTicks {
			ai.Behavior = "alerted"
			return Direction{}
		}
		ai.Behavior = "chasing"
		d := ai.navigate(prey.X, prey.Y)
		d.Dash = ai.canAttack(prey)
		if d.Dash {
			ai.Behavior = "attacking"
		}
		return d
	}
	if found {
		ai.Behavior = "seeking"
		return ai.navigate(fruitX, fruitY)
	}
	ai.Behavior = "wandering"
	if ai.wanderTicks == 0 {
		directions := [8]Direction{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}, {X: 1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: 1}, {X: -1, Y: -1}}
		ai.direction = directions[rand.IntN(len(directions))]
		ai.wanderTicks = 2*TickRate + rand.IntN(2*TickRate+1)
	}
	d := ai.direction
	nx, ny := Move(ai.X, ai.Y, d, .15, ai.Stats.Speed)
	if math.Hypot(nx-hx, ny-hy) > float64(ai.den.Radius*terrain.Tile)*.75 || math.Hypot(nx-ai.X, ny-ai.Y) < 1 {
		ai.wanderTicks = 1
		d = ai.navigate(hx, hy)
	}
	ai.direction = d
	ai.wanderTicks--
	return d
}

func (w *World) nearestFruit(ai *aiState) (Fruit, bool) {
	var nearest Fruit
	best := AIDetectionRadius * AIDetectionRadius
	found := false
	for _, fruit := range w.fruits {
		if !ai.sees(fruit.X, fruit.Y) || !ai.inTerritory(fruit.X, fruit.Y) || ai.Stats.HP >= MaxHPForLevel(ai.Stats.Level) {
			continue
		}
		dx := fruit.X + fruit.Size/2 - (ai.X + AISize/2)
		dy := fruit.Y + fruit.Size/2 - (ai.Y + AISize/2)
		distance := dx*dx + dy*dy
		if distance <= best && (!found || distance < best || fruit.ID < nearest.ID) {
			nearest, best, found = fruit, distance, true
		}
	}
	return nearest, found
}

func (w *World) respawnAIs() {
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick == 0 || w.tick < ai.respawnTick {
			continue
		}
		den := ai.den
		x, y := terrain.CellPosition(den.X, den.Y, AISize)
		if !w.denClear(x, y, i) {
			ai.respawnTick = w.tick + TickRate
			continue
		}
		*ai = w.newDenAI(den)
	}
}

func (ai *aiState) updateSleep(threat bool) bool {
	if threat {
		ai.awakeTicks = max(0, ai.awakeTicks-1)
		return false
	}
	if ai.Behavior == "sleeping" {
		ai.asleepTicks--
		if ai.asleepTicks > 0 {
			return true
		}
		ai.awakeTicks = AIAwakeSeconds * TickRate
		return false
	}
	ai.awakeTicks = max(0, ai.awakeTicks-1)
	if ai.awakeTicks == 0 {
		ai.Behavior = "sleeping"
		ai.asleepTicks = AIAsleepSeconds * TickRate
		ai.wanderTicks, ai.escapeTicks = 0, 0
		return true
	}
	return false
}

// Delay arrivals anywhere a player could immediately see or collide with them.
func (w *World) denClear(x, y float64, skip int) bool {
	for _, p := range w.players {
		if !p.Dead && math.Hypot(p.X-x, p.Y-y) < AIDetectionRadius+PlayerSize {
			return false
		}
	}
	for j := range w.ais {
		a := &w.ais[j]
		if j != skip && a.respawnTick == 0 && math.Hypot(a.X-x, a.Y-y) < AISize*2 {
			return false
		}
	}
	return true
}
func (ai *aiState) observeTarget(target *playerState) {
	if target == nil {
		ai.targetID = ""
		ai.targetTicks = 0
		return
	}
	if target.ID != ai.targetID {
		ai.targetID = target.ID
		ai.targetTicks = 0
	}
	ai.targetTicks++
}
