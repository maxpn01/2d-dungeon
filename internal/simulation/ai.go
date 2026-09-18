package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"
)

const (
	InitialAIs          = 5
	InitialAISpawnDelay = 10
	AIRespawnSeconds    = 3
	AIAwakeSeconds      = 120
	AIAsleepSeconds     = 60
	AISize              = PlayerSize
	AISpeed             = PlayerSpeed
	AIDetectionRadius   = 300.0
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

func (w *World) spawnAIs() {
	for i := 0; i < InitialAIs; i++ {
		w.nextAI++
		stats := NewStats()
		stats.Speed = AISpeed
		w.ais = append(w.ais, aiState{awakeTicks: AIAwakeSeconds * TickRate, AI: AI{
			ID: fmt.Sprintf("ai-%d", w.nextAI), X: 600 + float64(i%3)*450, Y: 400 + float64(i/3)*650, Stats: stats, Behavior: "wandering",
		}})
	}
}

func (w *World) updateAIs() {
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick != 0 {
			continue
		}
		input := w.aiInput(ai)
		actor := Player{X: ai.X, Y: ai.Y, Stats: ai.Stats, Combat: ai.Combat}
		AdvancePlayer(&actor, input, StepSeconds)
		ai.X, ai.Y, ai.Stats, ai.Combat = actor.X, actor.Y, actor.Stats, actor.Combat
	}
}

func (w *World) aiInput(ai *aiState) Direction {
	incoming := w.incomingDash(ai)
	threat, _ := w.nearestPlayer(ai, true)
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
	prey, preyDistance := w.nearestPlayer(ai, false)
	fruit, found := w.nearestFruit(ai)
	fruitX, fruitY := fruit.X+fruit.Size/2-AISize/2, fruit.Y+fruit.Size/2-AISize/2
	fruitDistance := (fruitX-ai.X)*(fruitX-ai.X) + (fruitY-ai.Y)*(fruitY-ai.Y)
	if prey != nil && (!found || preyDistance <= fruitDistance) {
		ai.Behavior = "chasing"
		d := ai.approach(prey.X, prey.Y)
		d.Dash = ai.canAttack(prey)
		if d.Dash {
			ai.Behavior = "attacking"
		}
		return d
	}
	if found {
		ai.Behavior = "seeking"
		return ai.approach(fruitX, fruitY)
	}
	ai.Behavior = "wandering"
	if ai.wanderTicks == 0 {
		directions := [8]Direction{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}, {X: 1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: 1}, {X: -1, Y: -1}}
		ai.direction = directions[rand.IntN(len(directions))]
		ai.wanderTicks = 2*TickRate + rand.IntN(2*TickRate+1)
	}
	d := ai.direction
	if ai.X <= MapMin {
		d.X = math.Abs(d.X)
	}
	if ai.X >= MapMax-AISize {
		d.X = -math.Abs(d.X)
	}
	if ai.Y <= MapMin {
		d.Y = math.Abs(d.Y)
	}
	if ai.Y >= MapMax-AISize {
		d.Y = -math.Abs(d.Y)
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
		x, y := w.spawnPosition(uint64(i)*57 + w.tick)
		w.nextAI++
		stats := NewStats()
		stats.Speed = AISpeed
		*ai = aiState{awakeTicks: AIAwakeSeconds * TickRate, AI: AI{ID: fmt.Sprintf("ai-%d", w.nextAI), X: x, Y: y, Stats: stats, Behavior: "wandering"}}
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
