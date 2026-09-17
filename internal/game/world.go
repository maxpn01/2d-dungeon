package game

import (
	"math/rand"
	"raylib-2d-sandbox/internal/client"
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GameObject interface {
	update(w *World, dt float32)
	draw()
}

type World struct {
	gameMap        *Map
	player         *Player
	fruitSpawner   *FruitSpawner
	ais            []*AI
	entities       []GameObject
	aiSpawnSeconds float32
	eating         *eatingScene
	eatingAI       *AI
}

func NewDefaultAI(mapSize rl.Vector2) *AI {
	size := rl.NewVector2(30, 30)
	x := 10 + float32(rand.Intn(int(mapSize.X-size.X-20)))
	y := 10 + float32(rand.Intn(int(mapSize.Y-size.Y-20)))

	return NewAI(
		rl.NewVector2(x, y),
		size,
		rl.Green,
		AIStats{
			lvl:    1,
			hp:     simulation.StartingHP,
			speed:  simulation.AISpeed,
			maxLvl: simulation.MaxLevel,
			maxHP:  simulation.StartingMaxHP,
		},
		300,
		simulation.AIAwakeSeconds, simulation.AIAsleepSeconds,
	)
}

func NewWorld() *World {
	w := &World{aiSpawnSeconds: simulation.InitialAISpawnDelay}
	w.gameMap = NewMap(
		rl.NewVector2(200, 200),
		rl.NewVector2(10, 10),
		rl.Black, rl.Gray,
	)
	w.player = NewPlayer(
		rl.NewVector2(300, 300),
		rl.NewVector2(30, 30),
		rl.Red,
		PlayerStats{
			lvl:              1,
			hp:               simulation.StartingHP,
			speed:            simulation.PlayerSpeed,
			sprintMultiplier: simulation.SprintMultiplier,
			maxLvl:           simulation.MaxLevel,
			maxHP:            simulation.StartingMaxHP,
		},
	)
	w.fruitSpawner = NewFruitSpawner(
		rl.NewVector2(15, 15),
		rl.Yellow,
		1, 20,
	)

	w.entities = []GameObject{w.player, w.fruitSpawner}
	w.gameMap.exploration.reveal(entityCenter(w.player.pos, w.player.size))
	return w
}

const maxSimulationStep float32 = 1.0 / 120

func (w *World) update(dt float32) {
	for dt > 0 && !w.player.isDeadFlag {
		step := min(dt, maxSimulationStep)

		if w.aiSpawnSeconds > 0 {
			w.aiSpawnSeconds = max(0, w.aiSpawnSeconds-step)
			if w.aiSpawnSeconds == 0 {
				for range simulation.InitialAIs {
					ai := NewDefaultAI(w.gameMap.size)
					w.ais = append(w.ais, ai)
					w.entities = append(w.entities, ai)
				}
			}
		}

		for _, entity := range w.entities {
			if w.player.isDeadFlag {
				return
			}

			var aiStart rl.Vector2
			if ai, ok := entity.(*AI); ok {
				aiStart = ai.pos
			}
			entity.update(w, step)
			if w.player.isDeadFlag {
				if ai, ok := entity.(*AI); ok {
					w.startEating(ai, aiStart)
				}
				return
			}
		}

		for _, ai := range w.ais {
			if ai.isDeadFlag {
				ai.respawnSeconds = max(0, ai.respawnSeconds-step)
				if ai.respawnSeconds == 0 {
					*ai = *NewDefaultAI(w.gameMap.size)
				}
			}
		}

		w.gameMap.exploration.reveal(entityCenter(w.player.pos, w.player.size))
		dt -= step
	}
}

func (w *World) draw(ctx RenderContext) {
	w.gameMap.draw(ctx)

	for _, entity := range w.entities {
		if w.eating != nil && (entity == w.player || entity == w.eatingAI) {
			continue
		}
		entity.draw()
	}
	if w.eating != nil {
		prey, killer := w.eating.actors(time.Now())
		ai := *w.eatingAI
		ai.pos = rl.NewVector2(float32(killer.X), float32(killer.Y))
		ai.draw()
		if prey.Size > 0 {
			rl.DrawRectangleV(rl.NewVector2(float32(prey.X), float32(prey.Y)), rl.NewVector2(float32(prey.Size), float32(prey.Size)), w.player.color)
		}
	}
}

func (w *World) startEating(ai *AI, start rl.Vector2) {
	p := w.player
	prey := protocol.Player{ID: "player", KilledBy: "ai", Dead: true, X: float64(p.pos.X), Y: float64(p.pos.Y), Size: float64(p.size.X)}
	killer := protocol.AI{ID: "ai", X: float64(ai.pos.X), Y: float64(ai.pos.Y), Size: float64(ai.size.X)}
	previous := killer
	previous.X, previous.Y = float64(start.X), float64(start.Y)
	w.eating = newEatingScene(time.Now(), client.Update{Local: prey, PlayerID: prey.ID, AIs: []protocol.AI{killer}}, []protocol.Player{prey}, []protocol.AI{previous})
	w.eatingAI = ai
}
