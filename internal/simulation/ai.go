package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/terrain"
)

const (
	InitialAIs            = parameters.InitialAIs
	InitialAISpawnDelay   = parameters.InitialAISpawnDelay
	AIRespawnSeconds      = parameters.AIRespawnSeconds
	AIAwakeSeconds        = parameters.AIAwakeSeconds
	AIAsleepSeconds       = parameters.AIAsleepSeconds
	AISize                = parameters.AISize
	AISpeed               = parameters.AISpeed
	AIDetectionRadius     = parameters.AIDetectionRadius
	AISettleTicks         = parameters.AISettleTicks
	AIReactionTicks       = parameters.AIReactionTicks
	AIAttackIntervalTicks = parameters.AIAttackIntervalTicks
)

type AI struct {
	Combat    CombatState
	Behavior  string
	Character Character
	Split     SplitCells
	ID        string
	X, Y      float64
	Stats     Stats
}

type aiState struct {
	shotCooldown float64
	AI
	den             terrain.Den
	settleTicks     int
	targetID        string
	targetTicks     int
	attackCooldown  int
	returning       bool
	direction       Direction
	wanderTicks     int
	patrolX         float64
	patrolY         float64
	patrolTicks     int
	patrolPause     int
	respawnTick     uint64
	awakeTicks      int
	asleepTicks     int
	escapeTicks     int
	escapeDirection Direction
	evadeTicks      int
	evadeDirection  Direction
}

func (w *World) newDenAI(den terrain.Den) aiState {
	w.nextAI++
	x, y := terrain.CellPosition(den.X, den.Y, AISize)
	stats := NewStats()
	stats.Level, stats.HP, stats.Speed = den.Level, den.HP, AISpeed*.65
	return aiState{den: den, AI: AI{ID: fmt.Sprintf("ai-%d", w.nextAI), X: x, Y: y, Stats: stats, Behavior: "patrolling", Character: Square}}
}

func (w *World) spawnEntranceGuards() {
	for _, den := range terrain.Dens {
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
	stats := NewStats()
	stats.Speed = AISpeed
	return aiState{settleTicks: AISettleTicks, AI: AI{
		ID: fmt.Sprintf("ai-%d", w.nextAI), X: 600 + float64(slot%3)*450,
		Y: 400 + float64(slot/3)*650, Stats: stats, Behavior: "emerging", Character: Square,
	}}
}

func (w *World) updateAIs() {
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick != 0 {
			continue
		}
		input := w.aiInput(ai)
		oldX, oldY := ai.X, ai.Y
		actor := Player{ID: ai.ID, ShotCooldown: ai.shotCooldown, X: ai.X, Y: ai.Y, Stats: ai.Stats, Combat: ai.Combat, Character: ai.Character, Split: ai.Split}
		actor.GatesClosed = w.boss.Phase == "fighting"
		AdvancePlayer(&actor, input, StepSeconds, w.safeZones)
		w.fire(&actor, input)
		ai.shotCooldown = actor.ShotCooldown
		ai.X, ai.Y, ai.Stats, ai.Combat, ai.Split = actor.X, actor.Y, actor.Stats, actor.Combat, actor.Split
		if terrain.PrivateGateCrossed(oldX, oldY, ai.X, ai.Y, AISize) || ai.den.ID != "" && !terrain.InEntranceRoom(ai.X, ai.Y, AISize) {
			ai.X, ai.Y = oldX, oldY
		}
	}
}

func (w *World) aiInput(ai *aiState) Direction {
	if target := w.rangedTarget(ai); target != nil {
		dx, dy := target.X-ai.X, target.Y-ai.Y
		length := math.Hypot(dx, dy)
		d := Direction{AimX: dx / max(1, length), AimY: dy / max(1, length), Shoot: true}
		ai.Behavior = "attacking"
		if ai.den.ID == "" && length > 200 {
			movement := ai.navigate(target.X, target.Y)
			d.X, d.Y = movement.X, movement.Y
		}
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
		ai.X, ai.Y = ai.patrolX, ai.patrolY
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
	left := float64(ai.den.Left*terrain.Tile + 8)
	right := float64((ai.den.Right+1)*terrain.Tile - AISize - 8)
	top := float64(ai.den.Top*terrain.Tile + 8)
	bottom := float64((ai.den.Bottom+1)*terrain.Tile - AISize - 8)
	for range 24 {
		x := left + rand.Float64()*(right-left)
		y := top + rand.Float64()*(bottom-top)
		if math.Hypot(x-ai.X, y-ai.Y) < 100 || !terrain.InEntranceRoom(x, y, AISize) {
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
		if ai.den.ID == "" || ai.respawnTick == 0 || w.tick < ai.respawnTick {
			continue
		}
		x, y := terrain.CellPosition(ai.den.X, ai.den.Y, AISize)
		if !w.denClear(x, y, i) {
			ai.respawnTick = w.tick + TickRate
			continue
		}
		*ai = w.newDenAI(ai.den)
	}
}

func (w *World) nearestFruit(ai *aiState) (Fruit, bool) {
	var nearest Fruit
	best := AIDetectionRadius * AIDetectionRadius
	found := false
	for _, fruit := range w.fruits {
		if !ai.sees(fruit.X, fruit.Y) || !ai.inTerritory(fruit.X, fruit.Y) || (fruit.MP > 0 && ai.Stats.MP >= MaxMP || fruit.MP == 0 && ai.Stats.HP >= MaxHPForLevel(ai.Stats.Level) && ai.Stats.Level >= MaxLevel) {
			continue
		}
		dx := fruit.X + fruit.Size/2 - (ai.X + AISize/2)
		dy := fruit.Y + fruit.Size/2 - (ai.Y + AISize/2)
		distance := dx*dx + dy*dy
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
