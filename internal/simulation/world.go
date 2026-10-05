package simulation

import (
	"fmt"
	"math"
	"sort"
	"time"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

const (
	TickRate     = parameters.TickRate
	StepSeconds  = parameters.StepSeconds
	PlayerSize   = parameters.PlayerSize
	SpeedScale   = parameters.SpeedScale
	BaseSpeed    = parameters.BaseSpeed
	MapMin       = parameters.MapMin
	MapMax       = parameters.MapMax
	MinPosition  = parameters.MinPosition
	MaxPosition  = parameters.MaxPosition
	InputTimeout = parameters.InputTimeout
)

type Direction struct {
	Shoot      bool
	Interact   bool
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
	EntranceID        string `json:"entrance_id"`
	HasWings          bool
	Flight            FlightState
	WingDoorOpen      float64
	NorthWingDoorOpen float64
	OpenSecrets       []string
	HasOrb            bool
	Escaped           bool
	ActiveStatue      string
	EntranceGate      EntranceGateState
	ShotCooldown      float64
	EjectCooldown     float64
	Character         Character
	Split             SplitCells
	GatesClosed       bool
	InArena           bool
	TeleportCharge    float64
	portalCooldown    int
	Combat            CombatState
	Dead              bool
	KilledBy          string
	Stats             Stats
	InputSeq          uint64
	ID                string
	X, Y              float64
}

type command struct {
	direction Direction
	sequence  uint64
}

type playerState struct {
	velocityX, velocityY float64
	testing              bool
	god                  bool
	Player
	spawnSlot        uint64
	progressionReset bool
	direction        Direction
	lastInput        time.Time
	commands         []command
	lastSequence     uint64
}

type World struct {
	sentinelTactics   map[string]*sentinelTactic
	vaultArrowStarted uint64
	vaultFire         [1]VaultFire
	vaultFireHits     [1]map[string]bool
	dungeonLayout     string
	wingDoorOpen      float64
	northWingDoorOpen float64
	winnerID          string
	finishedTick      uint64
	secretWallHits    map[string]int
	openSecrets       []string
	orb               OrbState
	statues           []StatueState
	projectiles       []Projectile
	nextProjectile    uint64
	safeZones         []SafeZone
	boss              Boss
	bossHitUntil      map[string]uint64
	ais               []aiState
	nextAI            uint64
	fruits            []Fruit
	patchNext         map[string]uint64
	nextFruit         uint64
	players           map[string]*playerState
	tick              uint64
	aiSpawnTicks      int
	protectedUntil    map[string]uint64
}

func NewWorld() *World {
	w := &World{dungeonLayout: worldmap.RandomizeDungeon(), secretWallHits: make(map[string]int), bossHitUntil: make(map[string]uint64), patchNext: make(map[string]uint64), protectedUntil: make(map[string]uint64), players: make(map[string]*playerState)}
	for _, statue := range worldmap.Default().Statues() {
		at := statue.Position()
		w.statues = append(w.statues, StatueState{ID: statue.ID, X: at.X, Y: at.Y, Cycle: 1})
	}
	at := worldmap.Default().OrbPosition()
	w.orb = OrbState{X: at.X, Y: at.Y}
	w.spawnEntranceGuards()
	w.replenishEntranceFood()
	w.updateOrb()
	return w
}

const MaxPlayers = 2

func (w *World) Join(id string) bool {
	if _, exists := w.players[id]; exists {
		return true
	}
	if len(w.players) >= MaxPlayers {
		return false
	}
	slot := uint64(0)
	for _, p := range w.players {
		if p.spawnSlot == 0 {
			slot = 1
		}
	}
	x, y := w.spawnPosition(slot)
	w.players[id] = &playerState{spawnSlot: slot, Player: Player{EntranceID: worldmap.EntranceID(slot), EntranceGate: EntranceGateState{ID: worldmap.EntranceID(slot)}, Character: Square, Stats: NewStats(), ID: id, X: x, Y: y}}
	return true
}

func (w *World) Leave(id string) {
	w.releaseStatues(id)
	w.releaseOrb(id)
	delete(w.players, id)
}

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
		if sequence != player.lastSequence+1 {
			return false
		}
		player.lastSequence = sequence
		if !player.Dead {
			if len(player.commands) > 0 {
				pending := player.commands[0].direction
				if pending.Shoot || pending.Dash || pending.Interact {
					direction.AimX, direction.AimY = pending.AimX, pending.AimY
				}
				direction.Shoot = direction.Shoot || pending.Shoot
				direction.Dash = direction.Dash || pending.Dash
				direction.Interact = direction.Interact || pending.Interact
				player.commands[0] = command{direction, sequence}
			} else {
				player.commands = append(player.commands, command{direction, sequence})
			}
		}
	}
	if player.Dead {
		player.direction = Direction{}
		player.InputSeq = player.lastSequence
	}
	player.lastInput = now
	return true
}

func (w *World) DiscardInputs() {
	for _, player := range w.players {
		player.direction = Direction{}
		player.commands = nil
		player.InputSeq = player.lastSequence
	}
}

func Move(x, y float64, d Direction, seconds, speed float64) (float64, float64) {
	if seconds <= 0 {
		return x, y
	}
	return MoveBy(x, y, d, speed*SpeedScale*seconds)
}

func MoveBy(x, y float64, d Direction, distance float64) (float64, float64) {
	motion := moveByMotion(x, y, d, distance)
	return motion.Bounds.X, motion.Bounds.Y
}

func moveByMotion(x, y float64, d Direction, distance float64) physics.Motion {
	body := physics.Square(x, y, PlayerSize)
	if !d.Valid() || math.IsNaN(distance) || math.IsInf(distance, 0) {
		return physics.Motion{Bounds: body, Rejected: true}
	}
	if distance <= 0 {
		return physics.Motion{Bounds: body}
	}
	delta := (physics.Vector{X: d.X, Y: d.Y}).Normalized().Scale(distance)
	return worldmap.Default().MoveBody(body, delta)
}

func (w *World) Step(now time.Time) {
	w.tick++
	w.advanceWingDoor()
	for id, until := range w.protectedUntil {
		if w.tick >= until {
			delete(w.protectedUntil, id)
		}
	}
	previousCapacity := len(w.players) + len(w.ais)
	for _, player := range w.players {
		previousCapacity += len(player.Split)
	}
	for i := range w.ais {
		previousCapacity += len(w.ais[i].Split)
	}
	previous := make(map[string]position, previousCapacity)
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
		if player.god {
			w.refillTestingPlayer(player)
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
		player.OpenSecrets = w.openSecrets
		player.WingDoorOpen = w.wingDoorOpen
		player.NorthWingDoorOpen = w.northWingDoorOpen
		if direction.Interact && (w.interactSecret(player) || w.interactObjective(player) || w.interactStatue(player)) {
			direction.Interact = false
		}
		AdvancePlayer(&player.Player, direction, StepSeconds, w.safeZones)
		player.velocityX = (player.X - previous[id].x) / StepSeconds
		player.velocityY = (player.Y - previous[id].y) / StepSeconds
		w.revealFlownSecrets(player)
		w.fire(&player.Player, direction)
	}
	for _, ai := range w.ais {
		previous[ai.ID] = position{ai.X, ai.Y}
		for i, cell := range ai.Split {
			if cell.Active {
				previous[fmt.Sprintf("%s/split/%d", ai.ID, i)] = position{cell.X, cell.Y}
			}
		}
	}
	w.updateAIs()
	w.respawnEntranceGuards()
	w.replenishEntranceFood()
	w.advanceFood()
	eaters := w.eaters(previous)
	w.advanceVaultArrows(eaters)
	w.advanceProjectiles(eaters)
	w.advanceMobMelee(eaters)
	w.advanceVaultTraps(eaters)
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
		clampPlayerBodies(&p.Player)
	}
	w.updateOrb()
	w.updatePortals()
	w.updateStatues()
	for _, p := range w.players {
		if p.god {
			w.refillTestingPlayer(p)
		}
	}
	// The surface is only a respawn area while floor one is being redesigned.
	remaining := w.fruits[:0]
	for _, fruit := range w.fruits {
		if fruit.Y < 0 || fruit.Source != "" {
			remaining = append(remaining, fruit)
		}
	}
	w.fruits = remaining
}

type Snapshot struct {
	VaultArrowStarted uint64
	VaultFire         []VaultFire
	SecretWallHits    map[string]int
	DungeonLayout     string
	WingDoorOpen      float64
	NorthWingDoorOpen float64
	WinnerID          string
	FinishedTick      uint64
	OpenSecrets       []string
	Orb               OrbState
	Statues           []StatueState
	Projectiles       []Projectile
	SafeZones         []SafeZone
	Boss              Boss
	AIs               []AI
	Tick              uint64
	Players           []Player
	Fruits            []Fruit
}

func (w *World) Snapshot() Snapshot {
	players := make([]Player, 0, len(w.players))
	for _, player := range w.players {
		state := player.Player
		state.WingDoorOpen = w.wingDoorOpen
		state.NorthWingDoorOpen = w.northWingDoorOpen
		state.OpenSecrets = append([]string(nil), w.openSecrets...)
		state.Split = state.Split.Clone()
		players = append(players, state)
	}
	sort.Slice(players, func(i, j int) bool { return players[i].ID < players[j].ID })
	fruits := make([]Fruit, len(w.fruits))
	copy(fruits, w.fruits)
	ais := make([]AI, 0, len(w.ais))
	for _, ai := range w.ais {
		if ai.respawnTick != 0 {
			continue
		}
		state := ai.AI
		state.Split = state.Split.Clone()
		ais = append(ais, state)
	}
	boss := w.boss
	boss.Tail = append([]Point(nil), boss.Tail...)
	statues := w.StatuesForPlayer("")
	hits := make(map[string]int, len(w.secretWallHits))
	for id, count := range w.secretWallHits {
		hits[id] = count
	}
	return Snapshot{VaultArrowStarted: w.vaultArrowStarted, VaultFire: append([]VaultFire(nil), w.vaultFire[:]...), SecretWallHits: hits, DungeonLayout: w.dungeonLayout, NorthWingDoorOpen: w.northWingDoorOpen, WingDoorOpen: w.wingDoorOpen, WinnerID: w.winnerID, FinishedTick: w.finishedTick, OpenSecrets: append([]string(nil), w.openSecrets...), Orb: w.orb, Statues: statues, Projectiles: append([]Projectile(nil), w.projectiles...), SafeZones: append([]SafeZone(nil), w.safeZones...), Boss: boss, Tick: w.tick, Players: players, Fruits: fruits, AIs: ais}
}

func (w *World) Tick() uint64 { return w.tick }
