package game

import (
	"image/color"
	"raylib-2d-sandbox/internal/client"
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/terrain"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GameObject interface {
	draw()
}

type World struct {
	feedback         combatFeedback
	gameMap          *Map
	player           *Player
	fruitSpawner     *FruitSpawner
	ais              []*AI
	entities         []GameObject
	simulation       *simulation.World
	accumulator      float64
	previousSnapshot simulation.Snapshot
	snapshot         simulation.Snapshot
	eating           *eatingScene
	eatingAI         *AI
}

func NewWorld() *World {
	w := &World{simulation: simulation.NewWorld()}
	w.simulation.Join("player")
	w.gameMap = NewMap(
		rl.NewVector2(terrain.SurfaceCells*terrain.Tile/10, terrain.SurfaceCells*terrain.Tile/10),
		rl.NewVector2(10, 10),
		rl.Black, rl.Gray,
	)
	w.player = NewPlayer(
		rl.NewVector2(300, 300),
		rl.NewVector2(30, 30),
		rl.Red,
		PlayerStats{
			lvl:    1,
			hp:     simulation.StartingHP,
			speed:  simulation.PlayerSpeed,
			maxLvl: simulation.MaxLevel,
			maxHP:  simulation.StartingMaxHP,
		},
	)
	w.fruitSpawner = &FruitSpawner{}

	w.syncSimulation()
	w.gameMap.exploration.reveal(entityCenter(w.player.pos, w.player.size))
	return w
}

func (w *World) update(dt float32) {
	w.accumulator += float64(min(dt, 0.25))
	for w.accumulator >= simulation.StepSeconds && !w.player.isDeadFlag {
		now := time.Now()
		input := w.player.input
		w.simulation.Input("player", simulation.Direction{X: float64(input.Direction.X), Y: float64(input.Direction.Y), Eject: input.Eject, Boost: input.Boost, Dash: input.Dash, AimX: float64(input.Aim.X), AimY: float64(input.Aim.Y)}, now)
		w.player.input.Dash = false
		w.player.input.Eject = false
		w.simulation.Step(now)
		w.accumulator -= simulation.StepSeconds
		w.syncSimulation()
	}
	w.interpolateActors()
	w.gameMap.exploration.reveal(entityCenter(w.player.pos, w.player.size))
}

func gameStats(s simulation.Stats) PlayerStats {
	return PlayerStats{lvl: s.Level, exp: s.XP, hp: float32(s.HP), speed: float32(s.Speed), maxLvl: simulation.MaxLevel, maxHP: float32(simulation.MaxHPForLevel(s.Level))}
}

func (w *World) syncSimulation() {
	snapshot := w.simulation.Snapshot()
	w.gameMap.safeZones = snapshot.SafeZones
	w.player.safeZones = snapshot.SafeZones
	w.feedback.snapshot(time.Now(), snapshot)
	w.previousSnapshot, w.snapshot = w.snapshot, snapshot
	p := snapshot.Players[0]
	previous := make(map[string]rl.Vector2)
	for _, ai := range w.ais {
		previous[ai.id] = ai.pos
	}
	w.player.pos = rl.NewVector2(float32(p.X), float32(p.Y))
	w.player.stats, w.player.combat = gameStats(p.Stats), p.Combat
	w.player.character, w.player.split, w.player.color = p.Character, p.Split.Clone(), characterColor(p.Character)
	w.player.isDeadFlag = p.Dead
	w.ais = nil
	w.entities = []GameObject{w.player, w.fruitSpawner}
	for _, state := range snapshot.AIs {
		ai := &AI{safeZones: snapshot.SafeZones, character: state.Character, combat: state.Combat, split: state.Split.Clone(), id: state.ID, pos: rl.NewVector2(float32(state.X), float32(state.Y)), size: rl.NewVector2(simulation.AISize, simulation.AISize), color: rl.Green, stats: gameStats(state.Stats)}
		switch state.Behavior {
		case "chasing":
			ai.actionState = ActionChasing
		case "alerted":
			ai.actionState = ActionAlerted
		case "emerging":
			ai.actionState = ActionEmerging
		case "attacking":
			ai.actionState = ActionAttacking
		case "evading":
			ai.actionState = ActionEvading
		case "sleeping":
			ai.actionState = ActionSleeping
		case "fleeing":
			ai.actionState = ActionFleeing
		case "patrolling":
			ai.actionState = ActionPatrolling
		case "watching":
			ai.actionState = ActionWatching
		}
		w.ais = append(w.ais, ai)
		w.entities = append(w.entities, ai)
		if p.Dead && p.KilledBy == state.ID && w.eating == nil {
			start, ok := previous[state.ID]
			if !ok {
				start = ai.pos
			}
			w.startEating(ai, start)
		}
	}
	w.fruitSpawner.fruits = nil
	playerColors := make(map[string]color.RGBA, len(snapshot.Players))
	for _, player := range snapshot.Players {
		playerColors[player.ID] = characterColor(player.Character)
	}
	for _, fruit := range snapshot.Fruits {
		w.fruitSpawner.fruits = append(w.fruitSpawner.fruits, Fruit{mass: fruit.Source != "", pos: rl.NewVector2(float32(fruit.X), float32(fruit.Y)), size: rl.NewVector2(float32(fruit.Size), float32(fruit.Size)), color: massColor(fruit.Source, playerColors)})
	}
}

func (w *World) draw(ctx RenderContext) {
	w.gameMap.draw(ctx)

	for _, entity := range w.entities {
		if w.eating != nil && (entity == w.player || entity == w.eatingAI) {
			continue
		}
		if ai, ok := entity.(*AI); ok {
			ai.drawVisible(&w.gameMap.exploration)
			continue
		}
		if entity == w.fruitSpawner {
			for _, f := range w.fruitSpawner.fruits {
				if w.gameMap.exploration.entityVisible(float64(f.pos.X), float64(f.pos.Y), float64(f.size.X)) {
					drawFruit(f.pos.X, f.pos.Y, f.size.X, f.color, f.mass)
				}
			}
			continue
		}
		entity.draw()
	}
	drawEncounter(w.snapshot.Boss, &w.gameMap.exploration)
	for _, p := range w.snapshot.Players {
		drawTeleportCharge(p.X, p.Y, p.TeleportCharge)
	}
	w.feedback.draw(time.Now(), &w.gameMap.exploration)
	if w.eating != nil {
		prey, killer := w.eating.actors(time.Now())
		ai := *w.eatingAI
		ai.pos = rl.NewVector2(float32(killer.X), float32(killer.Y))
		ai.draw()
		if prey.Size > 0 {
			drawActor(prey.X, prey.Y, prey.Size, prey.Stats.HP, simulation.CombatState{}, w.player.color, w.player.character)
		}
	}
}

func (w *World) startEating(ai *AI, start rl.Vector2) {
	p := w.player
	prey := protocol.Player{ID: "player", KilledBy: "ai", Dead: true, X: float64(p.pos.X), Y: float64(p.pos.Y), Size: float64(p.size.X), Stats: protocol.PlayerStats{HP: float64(p.stats.hp)}}
	killer := protocol.AI{ID: "ai", X: float64(ai.pos.X), Y: float64(ai.pos.Y), Size: float64(ai.size.X), Stats: protocol.PlayerStats{HP: float64(ai.stats.hp)}}
	previous := killer
	previous.X, previous.Y = float64(start.X), float64(start.Y)
	w.eating = newEatingScene(time.Now(), client.Update{Local: prey, PlayerID: prey.ID, AIs: []protocol.AI{killer}}, []protocol.Player{prey}, []protocol.AI{previous})
	w.eatingAI = ai
}

func (w *World) interpolateActors() {
	if w.player.isDeadFlag || len(w.previousSnapshot.Players) == 0 {
		return
	}
	alpha := w.accumulator / simulation.StepSeconds
	interpolate := func(ax, ay, bx, by float64) rl.Vector2 {
		if terrain.FloorTwo(ax, ay) != terrain.FloorTwo(bx, by) {
			return rl.NewVector2(float32(bx), float32(by))
		}
		x, y := terrain.Move(ax, ay, (bx-ax)*alpha, (by-ay)*alpha, simulation.PlayerSize, simulation.PlayerSize)
		return rl.NewVector2(float32(x), float32(y))
	}
	before, after := w.previousSnapshot.Players[0], w.snapshot.Players[0]
	w.player.pos = interpolate(before.X, before.Y, after.X, after.Y)
	w.player.split = append(w.player.split[:0], after.Split...)
	for i, cell := range after.Split {
		if !cell.Active {
			continue
		}
		startX, startY := cell.X-after.X, cell.Y-after.Y
		if i < len(before.Split) && before.Split[i].Active {
			startX, startY = before.Split[i].X-before.X, before.Split[i].Y-before.Y
		}
		ox := startX + (cell.X-after.X-startX)*alpha
		oy := startY + (cell.Y-after.Y-startY)*alpha
		w.player.split[i].X = float64(w.player.pos.X) + ox
		w.player.split[i].Y = float64(w.player.pos.Y) + oy
	}
	for i, ai := range w.ais {
		current := w.snapshot.AIs[i]
		for _, previous := range w.previousSnapshot.AIs {
			if ai.id == previous.ID {
				ai.pos = interpolate(previous.X, previous.Y, current.X, current.Y)
				for j, cell := range current.Split {
					if cell.Active && j < len(previous.Split) && previous.Split[j].Active {
						point := interpolate(previous.Split[j].X, previous.Split[j].Y, cell.X, cell.Y)
						ai.split[j].X, ai.split[j].Y = float64(point.X), float64(point.Y)
					}
				}
				break
			}
		}
	}
}
