package simulation

import (
	"fmt"
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
	Eject      bool
	Boost      bool
	Dash       bool
	AimX, AimY float64
	X, Y       float64
}

func (d Direction) Valid() bool {
	return !math.IsNaN(d.X) && !math.IsNaN(d.Y) &&
		!math.IsInf(d.X, 0) && !math.IsInf(d.Y, 0) &&
		!math.IsNaN(d.AimX) && !math.IsNaN(d.AimY) && !math.IsInf(d.AimX, 0) && !math.IsInf(d.AimY, 0) && math.Abs(d.AimX) <= 1 && math.Abs(d.AimY) <= 1 &&
		math.Abs(d.X) <= 1 && math.Abs(d.Y) <= 1
}

type Player struct {
	EjectCooldown  float64
	Character      Character
	Split          SplitCells
	GatesClosed    bool
	InArena        bool
	TeleportCharge float64
	portalCooldown int
	Combat         CombatState
	Dead           bool
	KilledBy       string
	Stats          Stats
	InputSeq       uint64
	ID             string
	X, Y           float64
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
	safeZones      []SafeZone
	boss           Boss
	bossHitUntil   map[string]uint64
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
	w := &World{safeZones: newSafeZones(), bossHitUntil: make(map[string]uint64), patchNext: make(map[string]uint64), protectedUntil: make(map[string]uint64), players: make(map[string]*playerState), aiSpawnTicks: InitialAISpawnDelay * TickRate}
	for range InitialFruits {
		w.spawnFruit()
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
		Character: Square, Stats: NewStats(), ID: id, X: x, Y: y, GatesClosed: w.boss.Phase == "fighting",
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
	if w.aiSpawnTicks > 0 {
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
		for i, cell := range player.Split {
			if cell.Active {
				previous[fmt.Sprintf("%s/split/%d", id, i)] = position{cell.X, cell.Y}
			}
		}
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
		drops := 0
		for _, f := range w.fruits {
			if f.Source != "" {
				drops++
			}
		}
		ejectInput := direction
		if drops >= MaxMassDrops {
			ejectInput.Eject = false
		}
		if food := EjectFood(&player.Player, ejectInput, StepSeconds); food != nil {
			w.nextFruit++
			food.ID = fmt.Sprintf("m%d", w.nextFruit)
			food.Source = id
			food.OwnerPickupTick = w.tick + EjectOwnerDelay
			food.ExpiresTick = w.tick + 30*TickRate
			w.fruits = append(w.fruits, *food)
		}
		AdvancePlayer(&player.Player, direction, StepSeconds, w.safeZones)
	}
	previousAIs := make([]position, len(w.ais))
	for i, ai := range w.ais {
		previousAIs[i] = position{ai.X, ai.Y}
	}
	w.updateAIs()
	w.advanceFood()
	eaters := w.eaters(previous, previousAIs)
	w.resolvePredation(eaters)
	for id, p := range w.players {
		old := previous[id]
		ConstrainArena(&p.Player, old.x, old.y)
	}
	for i := range eaters {
		e := &eaters[i]
		if e.cell != nil {
			e.end = position{e.cell.X, e.cell.Y}
		} else if e.player != nil {
			e.end = position{e.player.X, e.player.Y}
		} else {
			e.end = position{e.ai.X, e.ai.Y}
		}
	}
	w.consumeFruits(eaters)
	w.finishSplitCells()
	for _, p := range w.players {
		RepelPlayer(&p.Player, p.Player, w.safeZones)
	}
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.Stats.HP > SafeZoneMaxMass {
			ai.X, ai.Y = repelBody(ai.X, ai.Y, ai.X, ai.Y, ai.Stats.HP, w.safeZones)
		}
	}
	w.updatePortals()
}

type Snapshot struct {
	SafeZones []SafeZone
	Boss      Boss
	AIs       []AI
	Tick      uint64
	Players   []Player
	Fruits    []Fruit
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
	boss := w.boss
	boss.Tail = append([]Point(nil), boss.Tail...)
	return Snapshot{SafeZones: append([]SafeZone(nil), w.safeZones...), Boss: boss, Tick: w.tick, Players: players, Fruits: fruits, AIs: ais}
}

func (w *World) Tick() uint64 { return w.tick }
