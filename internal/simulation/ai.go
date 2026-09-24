package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"
	"raylib-2d-sandbox/internal/terrain"
)

const (
	InitialAIs            = 5
	InitialAISpawnDelay   = 10
	AIRespawnSeconds      = 45
	AIAwakeSeconds        = 120
	AIAsleepSeconds       = 60
	AISize                = PlayerSize
	AISpeed               = PlayerSpeed
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
	for i := 0; i < InitialAIs; i++ {
		ai := w.newSurfaceAI(i)
		if !w.denClear(ai.X, ai.Y, -1) {
			ai.respawnTick = w.tick + TickRate
		}
		w.ais = append(w.ais, ai)
	}
}

func (w *World) newSurfaceAI(slot int) aiState {
	w.nextAI++
	stats := NewStats()
	stats.Speed = AISpeed
	return aiState{settleTicks: AISettleTicks, AI: AI{
		ID: fmt.Sprintf("ai-%d", w.nextAI), X: 600 + float64(slot%3)*450,
		Y: 400 + float64(slot/3)*650, Stats: stats, Behavior: "emerging",
	}}
}

func (w *World) updateAIs() {
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick != 0 {
			continue
		}
		input := w.aiInput(ai)
		actor := Player{X: ai.X, Y: ai.Y, Stats: ai.Stats, Combat: ai.Combat}
		actor.GatesClosed = w.boss.Phase == "fighting"
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
	if threat, distance := w.nearestPlayer(ai, true); threat != nil {
		ai.Behavior = "fleeing"
		d := ai.flee(threat)
		d.Boost = distance < 180*180 && ai.Stats.HP > BoostMinimumMass
		return d
	}
	ai.escapeTicks = 0
	prey, preyDistance := w.nearestPlayer(ai, false)
	fruit, found := w.nearestFruit(ai)
	fx, fy := fruit.X+(fruit.Size-AISize)/2, fruit.Y+(fruit.Size-AISize)/2
	foodDistance := (fx-ai.X)*(fx-ai.X) + (fy-ai.Y)*(fy-ai.Y)
	if prey != nil && (!found || preyDistance <= foodDistance) {
		ai.Behavior = "chasing"
		d := ai.approach(prey.X, prey.Y)
		catchSeconds := math.Sqrt(preyDistance) / (ai.Stats.Speed * SpeedScale * (BoostMultiplier - 1))
		d.Boost = ai.Stats.HP-BoostMassPerSecond*catchSeconds > max(BoostMinimumMass, prey.Stats.HP+1)
		return d
	}
	if found {
		ai.Behavior = "seeking"
		return ai.approach(fx, fy)
	}
	ai.Behavior = "wandering"
	if ai.wanderTicks == 0 {
		directions := [8]Direction{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: -1}}
		ai.direction = directions[rand.IntN(len(directions))]
		ai.wanderTicks = 2*TickRate + rand.IntN(2*TickRate+1)
	}
	d := ai.direction
	if ai.X <= MapMin+8 {
		d.X = math.Abs(d.X)
	}
	if ai.X >= MapMax-AISize-8 {
		d.X = -math.Abs(d.X)
	}
	if ai.Y <= MapMin+8 {
		d.Y = math.Abs(d.Y)
	}
	if ai.Y >= MapMax-AISize-8 {
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
		if !ai.sees(fruit.X, fruit.Y) || !ai.inTerritory(fruit.X, fruit.Y) || (ai.Stats.HP >= MaxHPForLevel(ai.Stats.Level) && ai.Stats.Level >= MaxLevel) {
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
		next := w.newSurfaceAI(i)
		if !w.denClear(next.X, next.Y, i) {
			ai.respawnTick = w.tick + TickRate
			continue
		}
		*ai = next
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
