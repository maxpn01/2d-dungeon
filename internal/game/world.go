package game

import (
	"time"

	"2d-dungeon/internal/appearance"
	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/rendering"
	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"
	"2d-dungeon/internal/ui"
	"2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type World struct {
	presentation
	simulation          *simulation.World
	accumulator         float64
	previousSnapshot    simulation.Snapshot
	snapshot            simulation.Snapshot
	player              simulation.Player
	ais                 []simulation.AI
	teleporterDiscovery ui.TeleporterDiscovery
	libraryDiscovered   bool
	wingRoomDiscovered  bool
	controls            controls
	direction           simulation.Direction
	name                string
	started             bool
}

func NewWorld() *World {
	w := &World{simulation: simulation.NewWorld(), presentation: newPresentation()}
	w.syncSimulation()
	w.gameMap.Exploration.SetSecrets(w.snapshot.OpenSecrets, w.snapshot.Orb.GateOpen)
	w.gameMap.Exploration.Reveal(w.Center(), w.player.HasWings, w.player.Flight.Flying)
	return w
}

func (w *World) Pause() {
	w.direction = simulation.Direction{}
	w.accumulator = 0
	w.previousSnapshot = w.snapshot
	w.player = w.snapshot.Player
	w.ais = append(w.ais[:0], w.snapshot.AIs...)
	w.controls.suppressShootUntilRelease = true
}

func (w *World) Update(dt float32) {
	w.accumulator += float64(min(dt, 0.25))
	for w.accumulator >= parameters.StepSeconds && !w.Dead() && !w.ResultsOpen() && !w.StatsOpen() && !w.PerksOpen() {
		now := time.Now()
		w.simulation.Input(w.direction)
		w.direction.Shoot, w.direction.Dash, w.direction.Interact = false, false, false
		w.simulation.Step(now)
		w.accumulator -= parameters.StepSeconds
		w.syncSimulation()
	}
	if w.StatsOpen() || w.PerksOpen() {
		w.Pause()
		return
	}
	w.interpolateActors()
	w.gameMap.Exploration.SetSecrets(w.snapshot.OpenSecrets, w.snapshot.Orb.GateOpen)
	w.gameMap.Exploration.Reveal(w.Center(), w.player.HasWings, w.player.Flight.Flying)
	w.updateDiscoveries(dt)
}

func (w *World) ReadInput(camera rl.Camera2D, width, height int32) {
	p := w.snapshot.Player
	statue := w.perks.Sync(p.ActiveStatue, w.snapshot.Statues)
	if statue != nil {
		w.CloseStats()
		w.direction = simulation.Direction{}
		w.controls.suppressShootUntilRelease = true
		if card := w.perks.Read(*statue, p.Stats, width, height); card >= 0 {
			if worldmap.IsWingStatue(statue.ID) && card == 1 {
				w.ClosePerks()
			} else {
				w.simulation.ChoosePerk(statue.ID, statue.Cycle, card)
			}
			w.syncSimulation()
		}
		return
	}
	if w.StatsOpen() {
		w.direction = simulation.Direction{}
		w.controls.suppressShootUntilRelease = true
		return
	}
	if !rl.IsWindowFocused() {
		w.direction = simulation.Direction{}
		return
	}
	next := w.controls.read(camera, w.Center(), p.Combat.Facing)
	w.feedback.CheckCombatInput(time.Now(), next, p)
	next.Shoot = !w.controls.suppressShootUntilRelease && (next.Shoot || w.direction.Shoot)
	next.Dash = next.Dash || w.direction.Dash
	next.Interact = next.Interact || w.direction.Interact
	w.direction = next
}

func (w *World) ClosePerks() {
	w.simulation.CloseOffer()
	w.controls.suppressShootUntilRelease = true
	w.syncSimulation()
}

func (w *World) syncSimulation() {
	snapshot := w.simulation.Snapshot()
	w.feedback.Observe(time.Now(), snapshot)
	w.previousSnapshot, w.snapshot = w.snapshot, snapshot
	w.player = snapshot.Player
	w.ais = append(w.ais[:0], snapshot.AIs...)
}

func (w *World) Configure(name string, character appearance.Character) {
	w.name = name
	w.simulation.SetCharacter(character)
	w.syncSimulation()
}
func (w *World) SetName(name string) {
	w.name = name
}
func (w *World) Character() appearance.Character { return w.player.Character }
func (w *World) Dead() bool                      { return w.player.Dead }
func (w *World) PerksOpen() bool                 { return w.snapshot.Player.ActiveStatue != "" }
func (w *World) Center() rl.Vector2              { return playerCenter(w.player) }

func (w *World) Start(name string) {
	if err := worldmap.SetDungeonLayout(w.snapshot.DungeonLayout); err != nil {
		panic(err)
	}
	w.CloseStats()
	if !w.started {
		w.feedback.Spawn(simulation.PlayerID, time.Now())
		w.started = true
	}
	w.SetName(name)
	w.direction = simulation.Direction{}
	w.controls.suppressShootUntilRelease = true
}

func (w *World) scene() rendering.Scene {
	return rendering.Scene{
		Snapshot: w.snapshot, PlayerName: w.name, Local: w.snapshot.Player, RenderedPlayer: w.player, RenderedAIs: w.ais,
		PreviousStatues: w.previousSnapshot.Statues, PreviousTick: w.previousSnapshot.Tick, PreviousProjectiles: w.previousSnapshot.Projectiles, ProjectileAlpha: w.accumulator / parameters.StepSeconds,
	}
}

func (w *World) Draw(camera rl.Camera2D, width, height int32) {
	w.presentation.draw(primitives.Context{Camera: camera, WindowWidth: width, WindowHeight: height}, w.scene())
	if !w.StatsOpen() && !w.PerksOpen() {
		w.teleporterDiscovery.Draw(width, height)
	}
	p := w.snapshot.Player
	w.perks.Draw(p.ActiveStatue, w.snapshot.Statues, p.Stats, p.HasWings, width, height)
	if w.StatsOpen() {
		w.stats.Draw(p, w.name, width, height)
	}
	w.drawTesting(width, height)
}
func (w *World) DrawMap(width, height int32) { w.presentation.drawRadar(width, height, w.scene()) }
