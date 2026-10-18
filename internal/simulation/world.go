package simulation

import (
	"math"
	"sort"
	"time"

	"2d-dungeon/internal/appearance"
	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
	"2d-dungeon/internal/worldmap"
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
	EntranceID     string `json:"entrance_id"`
	HasWings       bool
	Flight         FlightState
	WingDoorOpen   float64
	VaultGateOpen  float64
	OpenSecrets    []string
	HasOrb         bool
	Escaped        bool
	ActiveStatue   string
	EntranceGate   EntranceGateState
	ShotCooldown   float64
	Character      appearance.Character
	TeleportCharge float64
	portalCooldown int
	Combat         CombatState
	Dead           bool
	KilledBy       string
	Stats          Stats
	ID             string
	X, Y           float64
}

type playerState struct {
	velocityX, velocityY float64
	testing              bool
	god                  bool
	Player
	progressionReset bool
	direction        Direction
}

type World struct {
	vaultArrowStarted uint64
	vaultFire         [worldmap.VaultFireTrapCount]VaultFire
	vaultFireHits     [worldmap.VaultFireTrapCount]map[string]bool
	dungeonLayout     string
	wingDoorOpen      float64
	winnerID          string
	finishedTick      uint64
	secretWallHits    map[string]int
	openSecrets       []string
	orb               OrbState
	statues           []StatueState
	projectiles       []Projectile
	nextProjectile    uint64
	ais               []aiState
	nextAI            uint64
	fruits            []Fruit
	patchNext         map[string]uint64
	nextFruit         uint64
	players           map[string]*playerState
	tick              uint64
}

func NewWorld() *World {
	w := &World{dungeonLayout: worldmap.RandomizeDungeon(), secretWallHits: make(map[string]int), patchNext: make(map[string]uint64), players: make(map[string]*playerState)}
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

const MaxPlayers = 1

func (w *World) Join(id string) bool {
	if _, exists := w.players[id]; exists {
		return true
	}
	if len(w.players) >= MaxPlayers {
		return false
	}
	x, y := w.spawnPosition()
	w.players[id] = &playerState{Player: Player{
		EntranceID: worldmap.EntranceGate, EntranceGate: EntranceGateState{ID: worldmap.EntranceGate},
		Character: appearance.Square, Stats: NewStats(), ID: id, X: x, Y: y,
	}}
	return true
}

func (w *World) Input(id string, direction Direction) bool {
	player, exists := w.players[id]
	if !exists || !direction.Valid() {
		return false
	}
	if player.Dead {
		direction = Direction{}
	}
	player.direction = direction
	return true
}

func Move(x, y float64, d Direction, seconds, speed float64) (float64, float64) {
	if seconds <= 0 {
		return x, y
	}
	return MoveBy(x, y, d, speed*parameters.SpeedScale*seconds)
}

func MoveBy(x, y float64, d Direction, distance float64) (float64, float64) {
	motion := moveByMotion(x, y, d, distance)
	return motion.Bounds.X, motion.Bounds.Y
}

func moveByMotion(x, y float64, d Direction, distance float64) physics.Motion {
	body := physics.Square(x, y, parameters.PlayerSize)
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
	if w.orb.GateActivated {
		w.orb.GateOpen = min(1, w.orb.GateOpen+parameters.StepSeconds*.65)
	}
	previousCapacity := len(w.players) + len(w.ais)
	previous := make(map[string]position, previousCapacity)
	for id, player := range w.players {
		previous[id] = position{player.X, player.Y}
		if player.Dead {
			continue
		}
		if player.god {
			w.refillTestingPlayer(player)
		}

		direction := player.direction

		player.VaultGateOpen = w.orb.GateOpen
		player.OpenSecrets = w.openSecrets
		player.WingDoorOpen = w.wingDoorOpen
		if direction.Interact && (w.interactLibrary(player, direction) || w.interactVaultGate(player) || w.interactSecret(player) || w.interactObjective(player) || w.interactStatue(player)) {
			direction.Interact = false
		}
		AdvancePlayer(&player.Player, direction, parameters.StepSeconds)
		player.velocityX = (player.X - previous[id].x) / parameters.StepSeconds
		player.velocityY = (player.Y - previous[id].y) / parameters.StepSeconds
		w.revealFlownSecrets(player)
		w.fire(&player.Player, direction)
	}
	for _, ai := range w.ais {
		previous[ai.ID] = position{ai.X, ai.Y}
	}
	w.updateAIs()
	w.respawnEntranceGuards()
	w.replenishEntranceFood()
	w.advanceFood()
	eaters := w.eaters(previous)
	w.advanceVaultArrows(eaters)
	w.advanceProjectiles(eaters)
	w.advanceMobMelee(eaters)
	w.advanceHeavyStomps()
	w.advanceVaultTraps(eaters)
	w.advanceLibraries()
	for i := range eaters {
		e := &eaters[i]
		if e.player != nil {
			e.end = position{e.player.X, e.player.Y}
		} else {
			body := MobBody(e.ai.Type, e.ai.X, e.ai.Y)
			e.end = position{body.X, body.Y}
		}
	}
	w.consumeFruits(eaters)
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
}

type Snapshot struct {
	VaultArrowStarted uint64
	VaultFire         []VaultFire
	SecretWallHits    map[string]int
	DungeonLayout     string
	WingDoorOpen      float64
	WinnerID          string
	FinishedTick      uint64
	OpenSecrets       []string
	Orb               OrbState
	Statues           []StatueState
	Projectiles       []Projectile
	AIs               []AI
	Tick              uint64
	Players           []Player
	Fruits            []Fruit
}

func (w *World) Snapshot() Snapshot {
	players := make([]Player, 0, len(w.players))
	for _, player := range w.players {
		state := player.Player
		state.VaultGateOpen = w.orb.GateOpen
		state.WingDoorOpen = w.wingDoorOpen
		state.OpenSecrets = append([]string(nil), w.openSecrets...)
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
		ais = append(ais, state)
	}
	statues := w.StatuesForPlayer("")
	hits := make(map[string]int, len(w.secretWallHits))
	for id, count := range w.secretWallHits {
		hits[id] = count
	}
	return Snapshot{
		VaultArrowStarted: w.vaultArrowStarted, VaultFire: append([]VaultFire(nil), w.vaultFire[:]...),
		SecretWallHits: hits, DungeonLayout: w.dungeonLayout, WingDoorOpen: w.wingDoorOpen,
		WinnerID: w.winnerID, FinishedTick: w.finishedTick, OpenSecrets: append([]string(nil), w.openSecrets...),
		Orb: w.orb, Statues: statues, Projectiles: append([]Projectile(nil), w.projectiles...),
		Tick: w.tick, Players: players, Fruits: fruits, AIs: ais,
	}
}

func (w *World) Tick() uint64 { return w.tick }

func (w *World) geometry() *worldmap.Map {
	return worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen)
}

func (p *Player) geometry() *worldmap.Map {
	return worldmap.Default().WithSecrets(p.OpenSecrets).WithVaultGate(p.VaultGateOpen)
}
