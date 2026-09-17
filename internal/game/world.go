package game

import (
	"math/rand"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GameObject interface {
	update(w *World, dt float32)
	draw()
}

type World struct {
	gameMap      *Map
	player       *Player
	fruitSpawner *FruitSpawner
	ais          []*AI
	entities     []GameObject
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
			hp:     1,
			speed:  500,
			maxLvl: 100,
			maxHP:  100,
		},
		AIGrowthStats{
			HPPerLevel:    0.25,
			SpeedPerLevel: 0.25,
			ExpAsleep:     0.001,
			HPAsleep:      0.001,
		},
		1,
		300,
		120, 60,
	)
}

func NewWorld() *World {
	w := &World{}
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
			hp:               1,
			speed:            600,
			sprintMultiplier: simulation.SprintMultiplier,
			maxLvl:           100,
			maxHP:            100,
		},
		PlayerGrowthStats{
			HPPerLevel:    0.25,
			SpeedPerLevel: 0.25,
		},
		1,
	)
	for range simulation.InitialAIs {
		w.ais = append(w.ais, NewDefaultAI(w.gameMap.size))
	}
	w.fruitSpawner = NewFruitSpawner(
		rl.NewVector2(15, 15),
		rl.Yellow,
		0.25,
		1, 20,
	)

	w.entities = []GameObject{w.player, w.fruitSpawner}
	for _, ai := range w.ais {
		w.entities = append(w.entities, ai)
	}
	return w
}

const maxSimulationStep float32 = 1.0 / 120

func (w *World) update(dt float32) {
	for dt > 0 && !w.player.isDeadFlag {
		step := min(dt, maxSimulationStep)

		for _, entity := range w.entities {
			if w.player.isDeadFlag {
				return
			}

			entity.update(w, step)
		}

		for _, ai := range w.ais {
			if ai.isDeadFlag {
				*ai = *NewDefaultAI(w.gameMap.size)
			}
		}

		dt -= step
	}
}

func (w *World) draw(ctx RenderContext) {
	w.gameMap.draw(ctx)

	for _, entity := range w.entities {
		entity.draw()
	}
}
