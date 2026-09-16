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
	MinPosition  = 10.0
	MaxPosition  = 1990.0 - PlayerSize
	InputTimeout = 500 * time.Millisecond
)

type Direction struct{ X, Y float64 }

func (d Direction) Valid() bool {
	return !math.IsNaN(d.X) && !math.IsNaN(d.Y) &&
		!math.IsInf(d.X, 0) && !math.IsInf(d.Y, 0) &&
		math.Abs(d.X) <= 1 && math.Abs(d.Y) <= 1
}

type Player struct {
	ID   string
	X, Y float64
}

type playerState struct {
	Player
	direction Direction
	lastInput time.Time
}

type World struct {
	players   map[string]*playerState
	nextSpawn uint64
	tick      uint64
}

func NewWorld() *World { return &World{players: make(map[string]*playerState)} }

func (w *World) Join(id string) {
	if _, exists := w.players[id]; exists {
		return
	}
	slot := w.nextSpawn % 400
	w.nextSpawn++
	w.players[id] = &playerState{Player: Player{
		ID: id, X: 300 + float64(slot%20)*60, Y: 300 + float64(slot/20)*60,
	}}
}

func (w *World) Leave(id string) { delete(w.players, id) }

func (w *World) Input(id string, direction Direction, now time.Time) bool {
	player, exists := w.players[id]
	if !exists || !direction.Valid() {
		return false
	}
	player.direction, player.lastInput = direction, now
	return true
}

func (w *World) Step(now time.Time) {
	w.tick++
	for _, player := range w.players {
		if now.Sub(player.lastInput) >= InputTimeout {
			player.direction = Direction{}
		}
		d := player.direction
		length := math.Hypot(d.X, d.Y)
		if length == 0 {
			continue
		}
		distance := PlayerSpeed * StepSeconds
		player.X = max(MinPosition, min(MaxPosition, player.X+d.X/length*distance))
		player.Y = max(MinPosition, min(MaxPosition, player.Y+d.Y/length*distance))
	}
}

func (w *World) Snapshot() (uint64, []Player) {
	players := make([]Player, 0, len(w.players))
	for _, player := range w.players {
		players = append(players, player.Player)
	}
	sort.Slice(players, func(i, j int) bool { return players[i].ID < players[j].ID })
	return w.tick, players
}

func (w *World) Tick() uint64 { return w.tick }
