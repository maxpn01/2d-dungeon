package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"

	"2d-dungeon/internal/appearance"
	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/worldmap"
)

type Behavior string

const (
	BehaviorPatrolling          Behavior = "patrolling"
	BehaviorStoneGuardian       Behavior = "stone guardian"
	BehaviorAwakening           Behavior = "awakening"
	BehaviorReturning           Behavior = "returning"
	BehaviorAttacking           Behavior = "attacking"
	BehaviorRepositioning       Behavior = "repositioning"
	BehaviorWatching            Behavior = "watching"
	BehaviorAiming              Behavior = "aiming"
	BehaviorWaitingForClearShot Behavior = "waiting for clear shot"
	BehaviorInvestigating       Behavior = "investigating"
	BehaviorSearching           Behavior = "searching"
	BehaviorDodging             Behavior = "dodging"
	BehaviorHuntingOrb          Behavior = "hunting orb"
	BehaviorCoveringOrb         Behavior = "covering orb"
	BehaviorStriking            Behavior = "striking"
	BehaviorWindingSword        Behavior = "winding sword"
	BehaviorCharging            Behavior = "charging"
	BehaviorPursuing            Behavior = "pursuing"
	BehaviorStrafing            Behavior = "strafing"
	BehaviorClosingDistance     Behavior = "closing distance"
	BehaviorKeepingDistance     Behavior = "keeping distance"
	BehaviorStomping            Behavior = "stomping"
	BehaviorWindingStomp        Behavior = "winding stomp"
	BehaviorWindingStrike       Behavior = "winding strike"
	BehaviorWindingCharge       Behavior = "winding charge"
	BehaviorRecoveringStrike    Behavior = "recovering strike"
)

type AI struct {
	Type      MobType
	Stone     bool
	Awakening float64
	Aiming    bool
	Frenzied  bool
	Combat    CombatState
	Behavior  Behavior
	Character appearance.Character
	ID        string
	X, Y      float64
	Stats     Stats
}

type aiState struct {
	meleeHits                    map[string]bool
	sentinelDashHits             map[string]bool
	sentinelMoveX, sentinelMoveY float64
	shotTargetID                 string
	aimTargetID                  string
	aimSince                     uint64
	aimLastSeen                  uint64
	aimReady                     bool
	spawnMaxHP                   float64
	frenzyBase                   Stats
	orbTargetID                  string
	aggressorID                  string
	alertUntil                   uint64
	alertX, alertY               float64
	shotCooldown                 float64
	AI
	den                        worldmap.Den
	attackCooldown             int
	returning                  bool
	patrolX                    float64
	patrolY                    float64
	patrolTicks                int
	patrolPause                int
	respawnTick                uint64
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
	x, y := worldmap.CellPosition(den.At.X, den.At.Y, parameters.AISize)
	stats := newMobStats(den.Level, den.Type)
	character := appearance.Square
	if den.Type == Sentinel {
		character = appearance.Triangle
	}
	ai := aiState{spawnMaxHP: stats.MaxHP(), den: den, AI: AI{Type: den.Type, ID: fmt.Sprintf("ai-%d", w.nextAI), X: x, Y: y, Stats: stats, Behavior: BehaviorPatrolling, Character: character}}
	if den.Type == Sentinel {
		ai.Stone, ai.Behavior = true, BehaviorStoneGuardian
	}
	return ai
}

func (w *World) spawnEntranceGuards() {
	for _, den := range worldmap.Default().Dens() {
		w.ais = append(w.ais, w.newDenAI(den))
	}
}

func (w *World) updateAIs() {
	w.awakenVaultSentinels()
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick != 0 {
			continue
		}
		if w.geometry().Cell(ai.den.At.X, ai.den.At.Y) == 0 && ai.den.ID != "" {
			continue
		}
		if ai.Stone {
			continue
		}
		if ai.Awakening > 0 {
			ai.Awakening = max(0, ai.Awakening-parameters.StepSeconds)
			ai.Behavior = BehaviorAwakening
			continue
		}
		ai.mobChargeCooldown = max(0, ai.mobChargeCooldown-1)
		if ai.Type == DungeonPursuer || ai.Type.IsHeavy() {
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
		actor := Player{VaultGateOpen: w.orb.GateOpen, OpenSecrets: w.openSecrets, ID: ai.ID, ShotCooldown: ai.shotCooldown, X: ai.X, Y: ai.Y, Stats: ai.Stats, Combat: ai.Combat, Character: ai.Character}
		AdvancePlayer(&actor, input, parameters.StepSeconds)
		w.fire(&actor, input)
		ai.shotCooldown = actor.ShotCooldown
		ai.X, ai.Y, ai.Stats, ai.Combat = actor.X, actor.Y, actor.Stats, actor.Combat
		if worldmap.Default().PrivateGateCrossed(oldX, oldY, ai.X, ai.Y, parameters.AISize) || ai.den.ID != "" && !w.aiWithinTerritory(ai) {
			ai.X, ai.Y = oldX, oldY
			if ai.aggressorID != "" {
				ai.aggressorID, ai.alertUntil = "", 0
				ai.returning = true
				ai.Behavior = BehaviorReturning
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
		if ai.Type == DungeonPursuer || ai.Type.IsHeavy() {
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
	if ai.returning || ai.aggressorID != "" && w.tick >= ai.alertUntil {
		if d, returning := w.investigateAI(ai); returning {
			return d
		}
	}
	if target := w.rangedTarget(ai); target != nil {
		if ai.Type == DungeonPursuer || ai.Type.IsHeavy() {
			return w.meleeGuardInput(ai, target)
		}
		if ai.Type == DungeonShooter {
			return w.shooterInput(ai, target)
		}
		ai.shotTargetID = target.ID
		dx, dy := target.X-ai.X, target.Y-ai.Y
		length := math.Hypot(dx, dy)
		d := Direction{AimX: dx / max(1, length), AimY: dy / max(1, length), Shoot: true}
		ai.Behavior = BehaviorAttacking
		if ai.aggressorID != "" && w.tick < ai.alertUntil {
			movement, _ := w.investigateAI(ai)
			d.X, d.Y = movement.X, movement.Y
		}
		return d
	}
	if ai.Type == DungeonShooter && w.tick < ai.shooterSightUntil {
		input := ai.shooterLastInput
		input.Shoot = false
		if !w.shotDodgeFits(ai, w.geometry(), input, parameters.StepSeconds) {
			input.X, input.Y = 0, 0
		}
		ai.Behavior = BehaviorRepositioning
		return input
	}
	if d, investigating := w.investigateAI(ai); investigating {
		return d
	}
	return ai.patrolInput(w.geometry())
}

func (ai *aiState) patrolInput(geometry *worldmap.Map) Direction {
	if ai.patrolPause > 0 {
		ai.patrolPause--
		ai.Behavior = BehaviorWatching
		return Direction{}
	}
	if ai.patrolTicks == 0 {
		ai.choosePatrolTarget(geometry)
	}
	if math.Hypot(ai.patrolX-ai.X, ai.patrolY-ai.Y) <= 18 {
		if ai.Character != appearance.Triangle {
			ai.X, ai.Y = ai.patrolX, ai.patrolY
		}
		ai.patrolTicks = 0
		ai.patrolPause = parameters.TickRate/3 + rand.IntN(parameters.TickRate/2)
		ai.Behavior = BehaviorWatching
		return Direction{}
	}
	ai.patrolTicks--
	ai.Behavior = BehaviorPatrolling
	return ai.navigateMap(geometry, ai.patrolX, ai.patrolY)
}

func (ai *aiState) choosePatrolTarget(geometry *worldmap.Map) {
	left := float64(ai.den.Patrol.X*parameters.Tile + 8)
	right := float64(ai.den.Patrol.Right()*parameters.Tile - parameters.AISize - 8)
	top := float64(ai.den.Patrol.Y*parameters.Tile + 8)
	bottom := float64(ai.den.Patrol.Bottom()*parameters.Tile - parameters.AISize - 8)
	minimumDistance := min(100.0, max(18.0, min(right-left, bottom-top)/2))
	for range 24 {
		x := left + rand.Float64()*(right-left)
		y := top + rand.Float64()*(bottom-top)
		body := MobBody(ai.Type, x, y)
		if math.Hypot(x-ai.X, y-ai.Y) < minimumDistance || !geometry.RoomContains(ai.den.RoomID, body) || !geometry.CanOccupy(body) || ai.Type == DungeonShooter && geometry.OverlapsCover(body) {
			continue
		}
		ai.patrolX, ai.patrolY = x, y
		ai.patrolTicks = 5 * parameters.TickRate
		return
	}
	ai.patrolX, ai.patrolY = worldmap.CellPosition(ai.den.At.X, ai.den.At.Y, parameters.AISize)
	ai.patrolTicks = 5 * parameters.TickRate
}

func (w *World) respawnEntranceGuards() {
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.den.RoomID == worldmap.OrbVault || ai.Type == GateHeavy || ai.den.ID == "" || ai.respawnTick == 0 || w.tick < ai.respawnTick {
			continue
		}
		x, y := worldmap.CellPosition(ai.den.At.X, ai.den.At.Y, parameters.AISize)
		if !w.denClear(x, y, i) {
			ai.respawnTick = w.tick + parameters.TickRate
			continue
		}
		*ai = w.respawnDenAI(*ai)
	}
}

func (w *World) denClear(x, y float64, skip int) bool {
	if p := w.player; !p.Dead && math.Hypot(p.X-x, p.Y-y) < parameters.AIDetectionRadius+parameters.PlayerSize {
		return false
	}
	for j := range w.ais {
		a := &w.ais[j]
		if j != skip && a.respawnTick == 0 {
			if math.Hypot(a.X-x, a.Y-y) < parameters.AISize*2 {
				return false
			}
		}
	}
	return true
}
