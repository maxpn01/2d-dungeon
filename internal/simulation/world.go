package simulation

import (
	"math"
	"sort"
	"time"
)

const (
	TickRate     = 30
	StepSeconds  = 1.0 / TickRate
	PlayerSize   = 30.0
	PlayerSpeed  = 600.0
	MapMin       = 10.0
	MapMax       = 1990.0
	MinPosition  = MapMin
	MaxPosition  = MapMax - PlayerSize
	InputTimeout = 500 * time.Millisecond
)

type Direction struct {
	X, Y float64
}

func (d Direction) Valid() bool {
	return !math.IsNaN(d.X) && !math.IsNaN(d.Y) &&
		!math.IsInf(d.X, 0) && !math.IsInf(d.Y, 0) &&
		math.Abs(d.X) <= 1 && math.Abs(d.Y) <= 1
}

type Player struct {
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
	ais       []aiState
	fruits    []Fruit
	nextFruit uint64
	players   map[string]*playerState
	nextSpawn uint64
	tick      uint64
}

func NewWorld() *World {
	w := &World{players: make(map[string]*playerState)}
	w.spawnAIs()
	for range InitialFruits {
		w.spawnFruit()
	}
	return w
}

func (w *World) Join(id string) {
	if _, exists := w.players[id]; exists {
		return
	}
	slot := w.nextSpawn % 400
	w.nextSpawn++
	w.players[id] = &playerState{Player: Player{
		Stats: NewStats(), ID: id, X: 300 + float64(slot%20)*60, Y: 300 + float64(slot/20)*60,
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
		player.commands = append(player.commands, command{direction, sequence})
	}
	player.lastInput = now
	return true
}

func Move(x, y float64, d Direction, seconds, speed float64) (float64, float64) {
	length := math.Hypot(d.X, d.Y)
	if length == 0 || !d.Valid() || seconds <= 0 {
		return x, y
	}
	distance := speed * seconds
	return max(MinPosition, min(MaxPosition, x+d.X/length*distance)), max(MinPosition, min(MaxPosition, y+d.Y/length*distance))
}

func (w *World) Step(now time.Time) {
	w.tick++
	if w.tick%fruitSpawnTicks == 0 {
		w.spawnFruit()
	}
	previous := make(map[string]position, len(w.players))
	for id, player := range w.players {
		previous[id] = position{player.X, player.Y}
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
		player.X, player.Y = Move(player.X, player.Y, direction, StepSeconds, player.Stats.Speed)
	}
	previousAIs := make([]position, len(w.ais))
	for i, ai := range w.ais {
		previousAIs[i] = position{ai.X, ai.Y}
	}
	w.updateAIs()
	w.consumeFruits(previous, previousAIs)
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
	ais := make([]AI, len(w.ais))
	for i, ai := range w.ais {
		ais[i] = ai.AI
	}
	return Snapshot{Tick: w.tick, Players: players, Fruits: fruits, AIs: ais}
}

func (w *World) Tick() uint64 { return w.tick }
