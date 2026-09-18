package simulation

import (
	"math"
	"raylib-2d-sandbox/internal/terrain"
	"sort"
	"time"
)

const (
	TickRate     = 30
	StepSeconds  = 1.0 / TickRate
	PlayerSize   = terrain.ActorSize
	SpeedScale   = 100.0
	PlayerSpeed  = 6.0
	MapMin       = float64(terrain.SurfaceMin)
	MapMax       = float64(terrain.SurfaceMax)
	MinPosition  = MapMin
	MaxPosition  = MapMax - PlayerSize
	InputTimeout = 500 * time.Millisecond
)

type Direction struct {
	Dash bool
	X, Y float64
}

func (d Direction) Valid() bool {
	return !math.IsNaN(d.X) && !math.IsNaN(d.Y) &&
		!math.IsInf(d.X, 0) && !math.IsInf(d.Y, 0) &&
		math.Abs(d.X) <= 1 && math.Abs(d.Y) <= 1
}

type Player struct {
	Combat   CombatState
	Dead     bool
	KilledBy string
	Stats    Stats
	InputSeq uint64
	ID       string
	X, Y     float64
}

type command struct {
	direction Direction
	sequence  uint64
}

type playerState struct {
	Player
	direction    Direction
	lastInput    time.Time
	commands     []command
	lastSequence uint64
}

type World struct {
	ais            []aiState
	nextAI         uint64
	fruits         []Fruit
	patchNext      map[string]uint64
	nextFruit      uint64
	players        map[string]*playerState
	nextSpawn      uint64
	tick           uint64
	aiSpawnTicks   int
	protectedUntil map[string]uint64
}

func NewWorld() *World {
	w := &World{patchNext: make(map[string]uint64), protectedUntil: make(map[string]uint64), players: make(map[string]*playerState), aiSpawnTicks: InitialAISpawnDelay * TickRate}
	for _, p := range terrain.Patches {
		for range p.Capacity {
			w.spawnPatchFruit(p)
		}
		w.patchNext[p.ID] = uint64(p.Seconds * TickRate)
	}
	return w
}

func (w *World) Join(id string) {
	if _, exists := w.players[id]; exists {
		return
	}
	x, y := w.spawnPosition(w.nextSpawn)
	w.nextSpawn++
	w.players[id] = &playerState{Player: Player{
		Stats: NewStats(), ID: id, X: x, Y: y,
	}}
}

func (w *World) Leave(id string) { delete(w.players, id) }

func (w *World) Input(id string, direction Direction, now time.Time) bool {
	return w.Command(id, direction, 0, now)
}

func (w *World) Command(id string, direction Direction, sequence uint64, now time.Time) bool {
	player, exists := w.players[id]
	if !exists || !direction.Valid() {
		return false
	}
	if sequence == 0 {
		if player.lastSequence != 0 {
			return false
		}
		player.direction = direction
	} else {
		if sequence != player.lastSequence+1 || len(player.commands) >= 32 {
			return false
		}
		player.lastSequence = sequence
		if !player.Dead {
			player.commands = append(player.commands, command{direction, sequence})
		}
	}
	if player.Dead {
		player.direction = Direction{}
		player.InputSeq = player.lastSequence
	}
	player.lastInput = now
	return true
}

func Move(x, y float64, d Direction, seconds, speed float64) (float64, float64) {
	if seconds <= 0 {
		return x, y
	}
	return MoveBy(x, y, d, speed*SpeedScale*seconds)
}

func MoveBy(x, y float64, d Direction, distance float64) (float64, float64) {
	length := math.Hypot(d.X, d.Y)
	if length == 0 || !d.Valid() || distance <= 0 {
		return x, y
	}
	return terrain.Move(x, y, d.X/length*distance, d.Y/length*distance, PlayerSize, PlayerSize)
}

func (w *World) Step(now time.Time) {
	w.tick++
	for id, until := range w.protectedUntil {
		if w.tick >= until {
			delete(w.protectedUntil, id)
		}
	}
	if w.aiSpawnTicks > 0 && len(w.players) > 0 {
		w.aiSpawnTicks--
		if w.aiSpawnTicks == 0 {
			w.spawnAIs()
		}
	}
	w.respawnAIs()
	if w.tick%fruitSpawnTicks == 0 {
		w.spawnFruit()
	}
	previous := make(map[string]position, len(w.players))
	for id, player := range w.players {
		previous[id] = position{player.X, player.Y}
		if player.Dead {
			continue
		}
		if now.Sub(player.lastInput) >= InputTimeout {
			player.direction = Direction{}
			player.commands = nil
			player.InputSeq = player.lastSequence
		}
		direction := player.direction
		if player.lastSequence != 0 {
			direction = Direction{}
			if len(player.commands) > 0 {
				next := player.commands[0]
				player.commands = player.commands[1:]
				direction = next.direction
				player.InputSeq = next.sequence
			}
		}
		AdvancePlayer(&player.Player, direction, StepSeconds)
	}
	previousAIs := make([]position, len(w.ais))
	for i, ai := range w.ais {
		previousAIs[i] = position{ai.X, ai.Y}
	}
	w.updateAIs()
	eaters := w.eaters(previous, previousAIs)
	w.resolvePredation(eaters)
	for i := range eaters {
		e := &eaters[i]
		if e.player != nil {
			e.end = position{e.player.X, e.player.Y}
		} else {
			e.end = position{e.ai.X, e.ai.Y}
		}
	}
	w.consumeFruits(eaters)
}

type Snapshot struct {
	AIs     []AI
	Tick    uint64
	Players []Player
	Fruits  []Fruit
}

func (w *World) Snapshot() Snapshot {
	players := make([]Player, 0, len(w.players))
	for _, player := range w.players {
		players = append(players, player.Player)
	}
	sort.Slice(players, func(i, j int) bool { return players[i].ID < players[j].ID })
	fruits := make([]Fruit, len(w.fruits))
	copy(fruits, w.fruits)
	ais := make([]AI, 0, len(w.ais))
	for _, ai := range w.ais {
		if ai.respawnTick != 0 {
			continue
		}
		ais = append(ais, ai.AI)
	}
	return Snapshot{Tick: w.tick, Players: players, Fruits: fruits, AIs: ais}
}

func (w *World) Tick() uint64 { return w.tick }
