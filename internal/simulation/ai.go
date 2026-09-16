package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"
)

const (
	InitialAIs        = 5
	AISize            = PlayerSize
	AISpeed           = 500.0
	AIDetectionRadius = 300.0
)

type AI struct {
	ID    string
	X, Y  float64
	Stats Stats
}

type aiState struct {
	AI
	direction   Direction
	wanderTicks int
}

func (w *World) spawnAIs() {
	for i := 0; i < InitialAIs; i++ {
		stats := NewStats()
		stats.Speed = AISpeed
		w.ais = append(w.ais, aiState{AI: AI{
			ID: fmt.Sprintf("ai-%d", i+1), X: 600 + float64(i%3)*450, Y: 400 + float64(i/3)*650, Stats: stats,
		}})
	}
}

func (w *World) updateAIs() {
	for i := range w.ais {
		ai := &w.ais[i]
		if target, found := w.nearestFruit(ai); found {
			dx := target.X + target.Size/2 - (ai.X + AISize/2)
			dy := target.Y + target.Size/2 - (ai.Y + AISize/2)
			distance := math.Hypot(dx, dy)
			if distance > 0 {
				direction := Direction{X: dx / distance, Y: dy / distance}
				seconds := min(StepSeconds, distance/ai.Stats.Speed)
				ai.X, ai.Y = Move(ai.X, ai.Y, direction, seconds, ai.Stats.Speed)
			}
			ai.wanderTicks = 0
			continue
		}
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
