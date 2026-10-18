package simulation

import (
	"math"
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
	EntranceID     string
	HasWings       bool
	Flight         FlightState
	WingDoorOpen   float64
	VaultGateOpen  float64
	OpenSecrets    []string
	HasOrb         bool
	FloorCleared   bool
	FloorClearedAt uint64
	ActiveStatue   string
	EntranceGate   EntranceGateState
	ShotCooldown   float64
	ShotsFired     uint64
	Character      appearance.Character
	TeleportCharge float64
	portalCooldown int
	Combat         CombatState
	Status         StatusEffects
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
	floorOne
	dungeonLayout  string
	winnerID       string
	finishedTick   uint64
	statues        []StatueState
	projectiles    []Projectile
	nextProjectile uint64
	ais            []aiState
	nextAI         uint64
	fruits         []Fruit
	patchNext      map[string]uint64
	nextFruit      uint64
	player         *playerState
	tick           uint64
}

type floorOne struct {
	orb               OrbState
	openSecrets       []string
	secretWallHits    map[string]int
	wingDoorOpen      float64
	vaultArrowStarted uint64
	vaultFire         [worldmap.VaultFireTrapCount]VaultFire
	vaultFireHits     [worldmap.VaultFireTrapCount]map[string]bool
}

func NewWorld() *World {
	w := &World{floorOne: floorOne{secretWallHits: make(map[string]int)}, dungeonLayout: worldmap.RandomizeDungeon(), patchNext: make(map[string]uint64)}
	for _, statue := range worldmap.Default().Statues() {
		at := statue.Position()
		w.statues = append(w.statues, StatueState{ID: statue.ID, X: at.X, Y: at.Y, Cycle: 1})
	}
	at := worldmap.Default().OrbPosition()
	w.orb = OrbState{X: at.X, Y: at.Y}
	w.spawnEntranceGuards()
	w.replenishEntranceFood()
	x, y := w.spawnPosition()
	w.player = &playerState{Player: Player{
		EntranceID: worldmap.EntranceGate, EntranceGate: EntranceGateState{ID: worldmap.EntranceGate},
		Character: appearance.Square, Stats: NewStats(), ID: PlayerID, X: x, Y: y,
	}}
	w.updateOrb()
	return w
}

const PlayerID = "player"

func (w *World) playerWithID(id string) *playerState {
	if w.player.ID == id {
		return w.player
	}
	return nil
}

func (w *World) Input(direction Direction) bool {
	if !direction.Valid() {
		return false
	}
	if w.player.Dead {
		direction = Direction{}
	}
	w.player.direction = direction
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
	previous := make(map[string]position, 1+len(w.ais))
	previous[w.player.ID] = position{w.player.X, w.player.Y}
	w.advancePlayer(previous[w.player.ID])
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
	w.advanceStatusEffects()
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
	clampPlayerBodies(&w.player.Player)
	w.updateOrb()
	w.updatePortals()
	w.updateStatues()
	if w.player.god {
		w.refillTestingPlayer(w.player)
	}
}

func (w *World) advancePlayer(previous position) {
	player := w.player
	if player.Dead {
		return
	}
	if player.god {
		w.refillTestingPlayer(player)
	}
	direction := player.direction
	player.VaultGateOpen = w.orb.GateOpen
	player.OpenSecrets = w.openSecrets
	player.WingDoorOpen = w.wingDoorOpen
	if direction.Interact && w.interact(player, direction) {
		direction.Interact = false
	}
	AdvancePlayer(&player.Player, direction, parameters.StepSeconds)
	player.velocityX = (player.X - previous.x) / parameters.StepSeconds
	player.velocityY = (player.Y - previous.y) / parameters.StepSeconds
	w.revealFlownSecrets(player)
	w.fire(&player.Player, direction)
}

func (w *World) interact(p *playerState, direction Direction) bool {
	return w.interactLibrary(p, direction) || w.interactVaultGate(p) || w.interactSecret(p) || w.interactObjective(p) || w.interactStatue(p)
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
	Player            Player
	Fruits            []Fruit
}

func (w *World) Snapshot() Snapshot {
	player := w.player.Player
	player.VaultGateOpen = w.orb.GateOpen
	player.WingDoorOpen = w.wingDoorOpen
	player.OpenSecrets = append([]string(nil), w.openSecrets...)
	fruits := make([]Fruit, len(w.fruits))
	copy(fruits, w.fruits)
	ais := make([]AI, 0, len(w.ais))
	for _, ai := range w.ais {
		if ai.respawnTick != 0 {
			continue
		}
		state := ai.AI
		state.Segments = append([]physics.Vector(nil), ai.Segments...)
		ais = append(ais, state)
	}
	statues := w.statueSnapshots()
	hits := make(map[string]int, len(w.secretWallHits))
	for id, count := range w.secretWallHits {
		hits[id] = count
	}
	return Snapshot{
		VaultArrowStarted: w.vaultArrowStarted, VaultFire: append([]VaultFire(nil), w.vaultFire[:]...),
		SecretWallHits: hits, DungeonLayout: w.dungeonLayout, WingDoorOpen: w.wingDoorOpen,
		WinnerID: w.winnerID, FinishedTick: w.finishedTick, OpenSecrets: append([]string(nil), w.openSecrets...),
		Orb: w.orb, Statues: statues, Projectiles: append([]Projectile(nil), w.projectiles...),
		Tick: w.tick, Player: player, Fruits: fruits, AIs: ais,
	}
}

func (w *World) Tick() uint64 { return w.tick }

func (w *World) geometry() *worldmap.Map {
	return worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen)
}

func (p *Player) geometry() *worldmap.Map {
	return worldmap.Default().WithSecrets(p.OpenSecrets).WithVaultGate(p.VaultGateOpen)
}
