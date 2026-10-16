package game

import (
	"time"

	"raylib-2d-sandbox/internal/rendering"
	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/ui"
	"raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const localPlayerID = "player"

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
	w.simulation.Join(localPlayerID)
	w.syncSimulation()
	w.gameMap.Exploration.SetSecrets(w.snapshot.OpenSecrets, w.snapshot.Orb.GateOpen)
	w.gameMap.Exploration.Reveal(w.Center(), w.player.HasWings, w.player.Flight.Flying)
	return w
}

func (w *World) Update(dt float32) {
	w.accumulator += float64(min(dt, 0.25))
	for w.accumulator >= simulation.StepSeconds && !w.Dead() && !w.ResultsOpen() {
		now := time.Now()
		w.simulation.Input(localPlayerID, w.direction)
		w.direction.Shoot, w.direction.Dash, w.direction.Interact = false, false, false
		w.simulation.Step(now)
		w.accumulator -= simulation.StepSeconds
		w.syncSimulation()
	}
	w.interpolateActors()
	w.gameMap.Exploration.SetSecrets(w.snapshot.OpenSecrets, w.snapshot.Orb.GateOpen)
	w.gameMap.Exploration.Reveal(w.Center(), w.player.HasWings, w.player.Flight.Flying)
	w.updateDiscoveries(dt)
}

func (w *World) ReadInput(camera rl.Camera2D, width, height int32) {
	p := w.snapshot.Players[0]
	statue := w.perks.Sync(p.ActiveStatue, w.snapshot.Statues)
	if statue != nil {
		w.CloseStats()
		w.direction = simulation.Direction{}
		w.controls.suppressShootUntilRelease = true
		if card := w.perks.Read(*statue, p.Stats, width, height); card >= 0 {
			if worldmap.IsWingStatue(statue.ID) && card == 1 {
				w.ClosePerks()
			} else {
				w.simulation.ChoosePerk(localPlayerID, statue.ID, statue.Cycle, card)
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
	w.simulation.CloseOffer(localPlayerID)
	w.controls.suppressShootUntilRelease = true
	w.syncSimulation()
}

func (w *World) syncSimulation() {
	snapshot := w.simulation.Snapshot()
	snapshot.Statues = w.simulation.StatuesForPlayer(localPlayerID)
	w.feedback.Observe(time.Now(), snapshot, localPlayerID)
	w.previousSnapshot, w.snapshot = w.snapshot, snapshot
	w.player = snapshot.Players[0]
	w.player.Split = w.player.Split.Clone()
	w.ais = append(w.ais[:0], snapshot.AIs...)
	for i := range w.ais {
		w.ais[i].Split = w.ais[i].Split.Clone()
	}
}

func (w *World) Configure(name string, character simulation.Character) {
	w.name = name
	w.simulation.SetCharacter(localPlayerID, character)
	w.syncSimulation()
}
func (w *World) SetName(name string) {
	w.name = name
}
func (w *World) Character() simulation.Character { return w.player.Character }
func (w *World) Dead() bool                      { return w.player.Dead }
func (w *World) PerksOpen() bool                 { return w.snapshot.Players[0].ActiveStatue != "" }
func (w *World) Center() rl.Vector2              { return playerCenter(w.player) }

func (w *World) Start(name string) {
	if err := worldmap.SetDungeonLayout(w.snapshot.DungeonLayout); err != nil {
		panic(err)
	}
	w.CloseStats()
	if !w.started {
		w.feedback.Spawn(localPlayerID, time.Now())
		w.started = true
	}
	w.SetName(name)
	w.direction = simulation.Direction{}
	w.controls.suppressShootUntilRelease = true
}

func (w *World) scene() rendering.Scene {
	return rendering.Scene{
		Snapshot: w.snapshot, PlayerName: w.name, Local: w.snapshot.Players[0], RenderedPlayers: []simulation.Player{w.player}, RenderedAIs: w.ais,
		PreviousStatues: w.previousSnapshot.Statues, PreviousTick: w.previousSnapshot.Tick, PreviousProjectiles: w.previousSnapshot.Projectiles, ProjectileAlpha: w.accumulator / simulation.StepSeconds,
	}
}

func (w *World) Draw(camera rl.Camera2D, width, height int32) {
	w.presentation.draw(primitives.Context{Camera: camera, WindowWidth: width, WindowHeight: height}, w.scene())
	if !w.StatsOpen() && !w.PerksOpen() {
		w.teleporterDiscovery.Draw(width, height)
	}
	p := w.snapshot.Players[0]
	w.perks.Draw(p.ActiveStatue, w.snapshot.Statues, p.Stats, width, height)
	if w.StatsOpen() {
		ui.DrawStats(p, w.name, width, height)
	}
	w.drawTesting(width, height)
}
func (w *World) DrawMap(width, height int32) { w.presentation.drawRadar(width, height, w.scene()) }
