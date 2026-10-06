package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

const (
	InitialAIs            = parameters.InitialAIs
	InitialAISpawnDelay   = parameters.InitialAISpawnDelay
	AIRespawnSeconds      = parameters.AIRespawnSeconds
	AIAwakeSeconds        = parameters.AIAwakeSeconds
	AIAsleepSeconds       = parameters.AIAsleepSeconds
	AISize                = parameters.AISize
	AIDetectionRadius     = parameters.AIDetectionRadius
	AISettleTicks         = parameters.AISettleTicks
	AIReactionTicks       = parameters.AIReactionTicks
	AIAttackIntervalTicks = parameters.AIAttackIntervalTicks
)

type AI struct {
	Type      MobType
	Stone     bool
	Awakening float64
	Aiming    bool
	Frenzied  bool
	Combat    CombatState
	Behavior  string
	Character Character
	Split     SplitCells
	ID        string
	X, Y      float64
	Stats     Stats
}

type aiState struct {
	sentinelMoveX, sentinelMoveY float64
	shotTargetID                 string
	aimTargetID                  string
	aimSince                     uint64
	aimLastSeen                  uint64
	aimReady                     bool
	spawnMaxHP                   float64
	frenzyBase                   Stats
	orbTargetID                  string
	orbPathFound                 bool
	orbPathUntil                 uint64
	orbPathX, orbPathY           float64
	aggressorID                  string
	alertUntil                   uint64
	alertX, alertY               float64
	shotCooldown                 float64
	AI
	den                        worldmap.Den
	settleTicks                int
	targetID                   string
	targetTicks                int
	attackCooldown             int
	returning                  bool
	direction                  Direction
	wanderTicks                int
	patrolX                    float64
	patrolY                    float64
	patrolTicks                int
	patrolPause                int
	respawnTick                uint64
	awakeTicks                 int
	asleepTicks                int
	escapeTicks                int
	escapeDirection            Direction
	evadeTicks                 int
	evadeDirection             Direction
	shotDodgeReactionTicks     int
	shotDodgeCooldown          int
	strafeDirection            float64
	strafeUntil                uint64
	strafeTurnUntil            uint64
	shooterLastInput           Direction
	shooterSightUntil          uint64
	shooterMoveX, shooterMoveY float64
	shooterRangeMode           int
	navigationActive           bool
	navigationX, navigationY   float64
	navigationGoal             [2]int
	navigationMap              *worldmap.Map
	mobChargeCooldown          int
	deaths                     int
	perkMilestones             int
}

func (w *World) newDenAI(den worldmap.Den) aiState {
	w.nextAI++
	x, y := worldmap.CellPosition(den.At.X, den.At.Y, AISize)
	stats := newMobStats(den.Level, den.Type)
	character := Square
	if den.Type == Sentinel {
		character = Triangle
	}
	ai := aiState{spawnMaxHP: stats.MaxHP(), den: den, AI: AI{Type: den.Type, ID: fmt.Sprintf("ai-%d", w.nextAI), X: x, Y: y, Stats: stats, Behavior: "patrolling", Character: character}}
	if den.Type == Sentinel {
		ai.Stone, ai.Behavior = true, "stone guardian"
	}
	return ai
}

func (w *World) spawnEntranceGuards() {
	for _, den := range worldmap.Default().Dens() {
		w.ais = append(w.ais, w.newDenAI(den))
	}
}

func (w *World) spawnAIs() {
	for i := 0; i < InitialAIs; i++ {
		ai := w.newSurfaceAI(i)
		if !w.denClear(ai.X, ai.Y, -1) {
			ai.respawnTick = w.tick + TickRate
		}
		w.ais = append(w.ais, ai)
	}
}

func (w *World) newSurfaceAI(slot int) aiState {
	w.nextAI++
	stats := newMobStats(1, SpawnGuard)
	return aiState{spawnMaxHP: stats.MaxHP(), settleTicks: AISettleTicks, AI: AI{
		Type: SpawnGuard, ID: fmt.Sprintf("ai-%d", w.nextAI), X: 600 + float64(slot%3)*450,
		Y: 400 + float64(slot/3)*650, Stats: stats, Behavior: "emerging", Character: Square,
	}}
}

func (w *World) updateAIs() {
	w.awakenVaultSentinels()
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick != 0 {
			continue
		}
		if worldmap.Default().WithSecrets(w.openSecrets).Cell(ai.den.At.X, ai.den.At.Y) == 0 && ai.den.ID != "" {
			continue
		}
		if ai.Stone {
			continue
		}
		if ai.Awakening > 0 {
			ai.Awakening = max(0, ai.Awakening-StepSeconds)
			ai.Behavior = "awakening"
			continue
		}
		ai.mobChargeCooldown = max(0, ai.mobChargeCooldown-1)
		if ai.Type == DungeonPursuer || ai.Type == DungeonHeavy {
			ai.attackCooldown = max(0, ai.attackCooldown-1)
		}
		ai.setFrenzy(w.orbTarget(ai) != nil)
		input := w.aiInput(ai)
		input = w.avoidPlayerShots(ai, input)
		if ai.Type == Sentinel || ai.Frenzied {
			input = w.steerSentinel(ai, input)
		}
		if ai.Type == DungeonShooter {
			input = w.steerShooter(ai, input)
		}
		oldX, oldY := ai.X, ai.Y
		actor := Player{OpenSecrets: w.openSecrets, ID: ai.ID, ShotCooldown: ai.shotCooldown, X: ai.X, Y: ai.Y, Stats: ai.Stats, Combat: ai.Combat, Character: ai.Character, Split: ai.Split}
		actor.GatesClosed = w.boss.Phase == "fighting"
		AdvancePlayer(&actor, input, StepSeconds, w.safeZones)
		w.fire(&actor, input)
		ai.shotCooldown = actor.ShotCooldown
		ai.X, ai.Y, ai.Stats, ai.Combat, ai.Split = actor.X, actor.Y, actor.Stats, actor.Combat, actor.Split
		if worldmap.Default().PrivateGateCrossed(oldX, oldY, ai.X, ai.Y, AISize) || ai.den.ID != "" && !w.aiWithinTerritory(ai) {
			ai.X, ai.Y = oldX, oldY
			if ai.aggressorID != "" {
				ai.aggressorID, ai.alertUntil = "", 0
				ai.returning = true
				ai.Behavior = "returning"
			}
		}
	}
}

func (w *World) aiIntent(ai *aiState) Direction {
	if ai.Type == Sentinel {
		return w.sentinelInput(ai)
	}
	if ai.Combat.SwordRemaining > 0 || ai.Combat.StompRemaining > 0 {
		return w.meleeGuardInput(ai, nil)
	}
	if carrier := w.orbTarget(ai); carrier != nil {
		if ai.Type == DungeonPursuer || ai.Type == DungeonHeavy {
			return w.meleeGuardInput(ai, carrier)
		}
		d := w.huntOrb(ai, carrier)
		offset := uint64(math.Abs(float64(ai.den.At.X*17 + ai.den.At.Y*31)))
		d.Boost = ai.Frenzied && (d.X != 0 || d.Y != 0) && (w.tick+offset)%parameters.AIFrenzyBoostCycleTicks < parameters.AIFrenzyBoostTicks
		return d
	}
	if ai.orbTargetID != "" {
		ai.returning = true
		ai.aggressorID, ai.alertUntil = "", 0
	}
	ai.orbTargetID = ""
	ai.orbPathUntil = 0
	ai.orbPathFound = false
	if ai.returning || ai.aggressorID != "" && w.tick >= ai.alertUntil {
		if d, returning := w.investigateAI(ai); returning {
			return d
		}
	}
	if target := w.rangedTarget(ai); target != nil {
		if ai.Type == DungeonPursuer || ai.Type == DungeonHeavy {
			return w.meleeGuardInput(ai, target)
		}
		if ai.Type == DungeonShooter {
			return w.shooterInput(ai, target)
		}
		ai.shotTargetID = target.ID
		dx, dy := target.X-ai.X, target.Y-ai.Y
		length := math.Hypot(dx, dy)
		d := Direction{AimX: dx / max(1, length), AimY: dy / max(1, length), Shoot: true}
		ai.Behavior = "attacking"
		if ai.den.ID == "" && length > parameters.AIFiringDistance {
			ai.Behavior = "pursuing"
			movement := ai.navigateMap(worldmap.Default().WithSecrets(w.openSecrets), target.X, target.Y)
			d.X, d.Y = movement.X, movement.Y
		} else if ai.aggressorID != "" && w.tick < ai.alertUntil {
			movement, _ := w.investigateAI(ai)
			d.X, d.Y = movement.X, movement.Y
		}
		return d
	}
	if ai.Type == DungeonShooter && w.tick < ai.shooterSightUntil {
		input := ai.shooterLastInput
		input.Shoot = false
		if !w.shotDodgeFits(ai, worldmap.Default().WithSecrets(w.openSecrets), input, StepSeconds) {
			input.X, input.Y = 0, 0
		}
		ai.Behavior = "repositioning"
		return input
	}
	if d, investigating := w.investigateAI(ai); investigating {
		return d
	}
	if ai.den.ID != "" {
		return ai.patrolInput()
	}
	if ai.settleTicks > 0 {
		ai.settleTicks--
		ai.Behavior = "emerging"
		return Direction{}
	}
	if fruit, found := w.nearestFruit(ai); found {
		ai.Behavior = "seeking"
		return ai.navigate(fruit.X+(fruit.Size-AISize)/2, fruit.Y+(fruit.Size-AISize)/2)
	}
	ai.Behavior = "wandering"
	if ai.wanderTicks == 0 {
		directions := [8]Direction{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: -1}}
		ai.direction = directions[rand.IntN(len(directions))]
		ai.wanderTicks = 2*TickRate + rand.IntN(2*TickRate+1)
	}
	d := ai.direction
	if ai.X <= MapMin+8 {
		d.X = math.Abs(d.X)
	}
	if ai.X >= MapMax-AISize-8 {
		d.X = -math.Abs(d.X)
	}
	if ai.Y <= MapMin+8 {
		d.Y = math.Abs(d.Y)
	}
	if ai.Y >= MapMax-AISize-8 {
		d.Y = -math.Abs(d.Y)
	}
	ai.direction = d
	ai.wanderTicks--
	return d
}

func (ai *aiState) patrolInput() Direction {
	if ai.patrolPause > 0 {
		ai.patrolPause--
		ai.Behavior = "watching"
		return Direction{}
	}
	if ai.patrolTicks == 0 {
		ai.choosePatrolTarget()
	}
	if math.Hypot(ai.patrolX-ai.X, ai.patrolY-ai.Y) <= 18 {
		if ai.Character != Triangle {
			ai.X, ai.Y = ai.patrolX, ai.patrolY
		}
		ai.patrolTicks = 0
		ai.patrolPause = TickRate/3 + rand.IntN(TickRate/2)
		ai.Behavior = "watching"
		return Direction{}
	}
	ai.patrolTicks--
	ai.Behavior = "patrolling"
	dx, dy := ai.patrolX-ai.X, ai.patrolY-ai.Y
	distance := math.Hypot(dx, dy)
	return Direction{X: dx / distance, Y: dy / distance}
}

func (ai *aiState) choosePatrolTarget() {
	left := float64(ai.den.Patrol.X*worldmap.Tile + 8)
	right := float64(ai.den.Patrol.Right()*worldmap.Tile - AISize - 8)
	top := float64(ai.den.Patrol.Y*worldmap.Tile + 8)
	bottom := float64(ai.den.Patrol.Bottom()*worldmap.Tile - AISize - 8)
	for range 24 {
		x := left + rand.Float64()*(right-left)
		y := top + rand.Float64()*(bottom-top)
		if math.Hypot(x-ai.X, y-ai.Y) < 100 || !worldmap.Default().RoomContains(ai.den.RoomID, physics.Square(x, y, AISize)) {
			continue
		}
		ai.patrolX, ai.patrolY = x, y
		ai.patrolTicks = 5 * TickRate
		return
	}
	// A temporarily crowded region still gives the guard somewhere to go.
	ai.patrolX, ai.patrolY = left, top
	ai.patrolTicks = 5 * TickRate
}

func (w *World) respawnEntranceGuards() {
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.den.RoomID == worldmap.OrbVault || ai.den.ID == "" || ai.respawnTick == 0 || w.tick < ai.respawnTick {
			continue
		}
		x, y := worldmap.CellPosition(ai.den.At.X, ai.den.At.Y, AISize)
		if !w.denClear(x, y, i) {
			ai.respawnTick = w.tick + TickRate
			continue
		}
		*ai = w.respawnDenAI(*ai)
	}
}

func (w *World) nearestFruit(ai *aiState) (Fruit, bool) {
	var nearest Fruit
	best := AIDetectionRadius * AIDetectionRadius
	found := false
	for _, fruit := range w.fruits {
		if !ai.sees(fruit.X, fruit.Y) || !ai.inTerritory(fruit.X, fruit.Y) || (fruit.MP > 0 && ai.Stats.MP >= ai.Stats.MaxMP() || fruit.MP == 0 && ai.Stats.HP >= ai.Stats.MaxHP() && ai.Stats.Level >= MaxLevel) {
			continue
		}
		fruitCenter := (physics.Rect{X: fruit.X, Y: fruit.Y, W: fruit.Size, H: fruit.Size}).Center()
		aiCenter := (physics.Rect{X: ai.X, Y: ai.Y, W: AISize, H: AISize}).Center()
		distance := fruitCenter.DistanceSquared(aiCenter)
		if distance <= best && (!found || distance < best || fruit.ID < nearest.ID) {
			nearest, best, found = fruit, distance, true
		}
	}
	return nearest, found
}

func (w *World) respawnAIs() {
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick == 0 || w.tick < ai.respawnTick {
			continue
		}
		next := w.newSurfaceAI(i)
		if !w.denClear(next.X, next.Y, i) {
			ai.respawnTick = w.tick + TickRate
			continue
		}
		*ai = next
	}
}

func (ai *aiState) updateSleep(threat bool) bool {
	if threat {
		ai.awakeTicks = max(0, ai.awakeTicks-1)
		return false
	}
	if ai.Behavior == "sleeping" {
		ai.asleepTicks--
		if ai.asleepTicks > 0 {
			return true
		}
		ai.awakeTicks = AIAwakeSeconds * TickRate
		return false
	}
	ai.awakeTicks = max(0, ai.awakeTicks-1)
	if ai.awakeTicks == 0 {
		ai.Behavior = "sleeping"
		ai.asleepTicks = AIAsleepSeconds * TickRate
		ai.wanderTicks, ai.escapeTicks = 0, 0
		return true
	}
	return false
}

func (w *World) denClear(x, y float64, skip int) bool {
	for _, p := range w.players {
		if !p.Dead && math.Hypot(p.X-x, p.Y-y) < AIDetectionRadius+PlayerSize {
			return false
		}
	}
	for j := range w.ais {
		a := &w.ais[j]
		if j != skip && a.respawnTick == 0 {
			if math.Hypot(a.X-x, a.Y-y) < AISize*2 {
				return false
			}
			for _, cell := range a.Split {
				if cell.Active && math.Hypot(cell.X-x, cell.Y-y) < AISize*2 {
					return false
				}
			}
		}
	}
	return true
}
func (ai *aiState) observeTarget(target *playerState) {
	if target == nil {
		ai.targetID = ""
		ai.targetTicks = 0
		return
	}
	if target.ID != ai.targetID {
		ai.targetID = target.ID
		ai.targetTicks = 0
	}
	ai.targetTicks++
}
