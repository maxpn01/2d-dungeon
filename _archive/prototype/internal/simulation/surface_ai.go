package simulation

import (
	"fmt"
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
)

const (
	InitialAIs            = parameters.InitialAIs
	InitialAISpawnDelay   = parameters.InitialAISpawnDelay
	AIAwakeSeconds        = parameters.AIAwakeSeconds
	AIAsleepSeconds       = parameters.AIAsleepSeconds
	AISettleTicks         = parameters.AISettleTicks
	AIReactionTicks       = parameters.AIReactionTicks
	AIAttackIntervalTicks = parameters.AIAttackIntervalTicks
)

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
	stats := newMobStats(1, SpawnGuard)
	return aiState{spawnMaxHP: stats.MaxHP(), settleTicks: AISettleTicks, AI: AI{
		Type: SpawnGuard, ID: fmt.Sprintf("ai-%d", w.nextAI), X: 600 + float64(slot%3)*450,
		Y: 400 + float64(slot/3)*650, Stats: stats, Behavior: "emerging", Character: Square,
	}}
}

func (w *World) nearestFruit(ai *aiState) (Fruit, bool) {
	var nearest Fruit
	best := AIDetectionRadius * AIDetectionRadius
	found := false
	for _, fruit := range w.fruits {
		if !ai.sees(fruit.X, fruit.Y) || !ai.inTerritory(fruit.X, fruit.Y) || (fruit.MP > 0 && ai.Stats.MP >= ai.Stats.MaxMP() || fruit.MP == 0 && ai.Stats.HP >= ai.Stats.MaxHP() && ai.Stats.Level >= MaxLevel) {
			continue
		}
		fruitCenter := (physics.Rect{X: fruit.X, Y: fruit.Y, W: fruit.Size, H: fruit.Size}).Center()
		aiCenter := (physics.Rect{X: ai.X, Y: ai.Y, W: AISize, H: AISize}).Center()
		distance := fruitCenter.DistanceSquared(aiCenter)
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

func (w *World) surfaceAIIntent(ai *aiState) Direction {
	if ai.settleTicks > 0 {
		ai.settleTicks--
		ai.Behavior = "emerging"
		return Direction{}
	}
	if fruit, found := w.nearestFruit(ai); found {
		ai.Behavior = "seeking"
		return ai.navigate(fruit.X+(fruit.Size-AISize)/2, fruit.Y+(fruit.Size-AISize)/2)
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
