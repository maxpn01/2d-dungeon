package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"
)

const (
	InitialAIs          = 5
	InitialAISpawnDelay = 10
	AISize              = PlayerSize
	AISpeed             = PlayerSpeed
	AIDetectionRadius   = 300.0
)

type AI struct {
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
	escapeTicks     int
	escapeDirection Direction
}

func (w *World) spawnAIs() {
	for i := 0; i < InitialAIs; i++ {
		w.nextAI++
		stats := NewStats()
		stats.Speed = AISpeed
		w.ais = append(w.ais, aiState{AI: AI{
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
		if threat, _ := w.nearestPlayer(ai, true); threat != nil {
			ai.Behavior = "fleeing"
			ai.flee(threat)
			continue
		}
		ai.escapeTicks = 0
		prey, preyDistance := w.nearestPlayer(ai, false)
		fruit, found := w.nearestFruit(ai)
		fruitX, fruitY := fruit.X+fruit.Size/2-AISize/2, fruit.Y+fruit.Size/2-AISize/2
		fruitDistance := (fruitX-ai.X)*(fruitX-ai.X) + (fruitY-ai.Y)*(fruitY-ai.Y)
		if prey != nil && (!found || preyDistance <= fruitDistance) {
			ai.Behavior = "chasing"
			ai.approach(prey.X, prey.Y)
			continue
		}
		if found {
			ai.Behavior = "seeking"
			ai.approach(fruitX, fruitY)
			continue
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
		ai.X, ai.Y = Move(ai.X, ai.Y, d, StepSeconds, ai.Stats.Speed)
		ai.wanderTicks--
	}
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
		*ai = aiState{AI: AI{ID: fmt.Sprintf("ai-%d", w.nextAI), X: x, Y: y, Stats: stats, Behavior: "wandering"}}
	}
}
