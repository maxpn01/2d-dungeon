package simulation

import (
	"math"
	"math/rand/v2"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
	"2d-dungeon/internal/worldmap"
)

type HydraPhase string

const (
	HydraSleeping  HydraPhase = "sleeping"
	HydraForaging  HydraPhase = "foraging"
	HydraHunting   HydraPhase = "hunting"
	HydraReturning HydraPhase = "returning"
)

type Hydra struct {
	Phase          HydraPhase
	PhaseRemaining float64
	Engaged        bool
	FoeID          string
	Facing         Direction
	Scale          float64
	Heads          [parameters.HydraHeadCount]HydraHead
	Severed        int
	Regrown        int
}

type hydraState struct {
	maxHP          float64
	lastHP         float64
	foeSeen        uint64
	anchorX        float64
	anchorY        float64
	lastX, lastY   float64
	calmUntil      uint64
	lungeReady     uint64
	volleyReady    uint64
	prey           string
	preyRethink    uint64
	roaming        bool
	roamX, roamY   float64
	roamTicks      int
	strafe         float64
	strafeUntil    uint64
	grazing        bool
	grazeX, grazeY float64
	lungeHit       [parameters.HydraHeadCount]bool
	sated          bool
}

type hydraTarget struct {
	id     string
	x, y   float64
	size   float64
	vx, vy float64
	player *playerState
	ai     *aiState
}

func HydraCenter(ai AI) physics.Vector {
	return physics.Vector{X: ai.X + parameters.AISize/2, Y: ai.Y + parameters.AISize/2}
}

func HydraBodyRadius(h Hydra) float64 { return parameters.HydraBodySize * h.Scale / 2 }

func hydraScale(level int) float64 {
	return min(parameters.HydraMaxGrowth, 1+parameters.HydraGrowthPerLevel*float64(level-1))
}

func (w *World) initHydra(ai *aiState) {
	ai.Hydra = Hydra{Phase: HydraSleeping, PhaseRemaining: parameters.HydraSleepSeconds, Facing: Direction{Y: 1}, Scale: hydraScale(ai.Stats.Level)}
	ai.hydra = hydraState{maxHP: ai.Stats.MaxHP(), lastHP: ai.Stats.HP}
	for i := range ai.Hydra.Heads {
		head := &ai.Hydra.Heads[i]
		head.MaxHP = ai.Stats.MaxHP() * parameters.HydraHeadHealth
		head.HP = head.MaxHP
	}
	center := HydraCenter(ai.AI)
	ai.Segments = make([]physics.Vector, parameters.HydraTailSegments)
	for i := range ai.Segments {
		ai.Segments[i] = physics.Vector{X: ai.X, Y: ai.Y - float64(i+1)*parameters.HydraTailSpacing}
	}
	for i := range ai.Hydra.Heads {
		x, y := ai.Hydra.restPoint(i, center, 0)
		ai.Hydra.Heads[i].X, ai.Hydra.Heads[i].Y = x, y
	}
}

func (w *World) hydraDefeated() bool {
	for _, ai := range w.ais {
		if ai.Type == CaveHydra {
			return ai.respawnTick != 0
		}
	}
	return false
}

func (w *World) hydraLair(ai *aiState) (float64, float64) {
	return worldmap.CellPosition(ai.den.At.X, ai.den.At.Y, parameters.AISize)
}

func (w *World) hydraInput(ai *aiState) Direction {
	h := &ai.Hydra
	w.growHydra(ai)
	w.advanceHydraRegrowth(ai)
	hurt := ai.Stats.HP < ai.hydra.lastHP-1e-9
	ai.hydra.lastHP = ai.Stats.HP
	ai.hydra.grazing = false
	if h.Phase == HydraSleeping {
		if !w.hydraWakes(ai, hurt) {
			ai.Behavior = BehaviorSleeping
			ai.Stats.HP = min(ai.Stats.MaxHP(), ai.Stats.HP+ai.Stats.MaxHP()*parameters.HydraSleepHealing*parameters.StepSeconds)
			ai.Stats.GainMP(parameters.HydraSleepManaRegen * parameters.StepSeconds)
			ai.hydra.lastHP = ai.Stats.HP
			if h.PhaseRemaining -= parameters.StepSeconds; h.PhaseRemaining <= 0 {
				h.Phase, h.PhaseRemaining = HydraForaging, parameters.HydraForageSeconds
			}
			return Direction{}
		}
		h.Phase, h.PhaseRemaining = HydraForaging, parameters.HydraForageSeconds
		ai.Behavior = BehaviorAwakening
	}
	ai.Stats.GainMP(parameters.HydraManaRegen * parameters.StepSeconds)
	if foe, ok := w.hydraFoe(ai, hurt); ok {
		h.Engaged, h.FoeID = true, foe.id
		if foe.player == nil {
			h.PhaseRemaining = max(0, h.PhaseRemaining-parameters.StepSeconds)
		}
		return w.hydraFight(ai, foe)
	}
	h.Engaged, h.FoeID = false, ""
	switch h.Phase {
	case HydraForaging:
		if h.PhaseRemaining -= parameters.StepSeconds; h.PhaseRemaining <= 0 {
			h.Phase, h.PhaseRemaining = HydraHunting, parameters.HydraHuntSeconds
			ai.hydra.prey, ai.hydra.preyRethink, ai.hydra.roaming, ai.hydra.sated = "", 0, false, false
			return w.hydraHunt(ai)
		}
		return w.hydraForage(ai, true)
	case HydraHunting:
		if h.PhaseRemaining -= parameters.StepSeconds; h.PhaseRemaining <= 0 {
			h.Phase, h.PhaseRemaining = HydraReturning, 0
			return w.hydraReturn(ai)
		}
		return w.hydraHunt(ai)
	}
	return w.hydraReturn(ai)
}

func (w *World) hydraWakes(ai *aiState, hurt bool) bool {
	if hurt {
		return true
	}
	p := w.player
	center := HydraCenter(ai.AI)
	return !p.Dead && math.Hypot(p.X+parameters.PlayerSize/2-center.X, p.Y+parameters.PlayerSize/2-center.Y) <= parameters.HydraWakeRadius &&
		w.geometry().Visible(center.X, center.Y, p.X+parameters.PlayerSize/2, p.Y+parameters.PlayerSize/2)
}

func playerTarget(p *playerState) hydraTarget {
	return hydraTarget{id: p.ID, x: p.X + parameters.PlayerSize/2, y: p.Y + parameters.PlayerSize/2, size: parameters.PlayerSize, vx: p.velocityX, vy: p.velocityY, player: p}
}

func snakeTarget(ai *aiState) hydraTarget {
	return hydraTarget{id: ai.ID, x: ai.X + parameters.AISize/2, y: ai.Y + parameters.AISize/2, size: parameters.AISize, ai: ai}
}

func (w *World) aiByID(id string) *aiState {
	for i := range w.ais {
		if w.ais[i].ID == id && w.ais[i].respawnTick == 0 {
			return &w.ais[i]
		}
	}
	return nil
}

func (w *World) hydraTargetByID(id string) (hydraTarget, bool) {
	if p := w.player; p.ID == id {
		return playerTarget(p), !p.Dead
	}
	if snake := w.aiByID(id); snake != nil && snake.Type == ForestSnake {
		return snakeTarget(snake), true
	}
	return hydraTarget{}, false
}

func (w *World) hydraSees(ai *aiState, x, y, radius float64) bool {
	center := HydraCenter(ai.AI)
	return math.Hypot(x-center.X, y-center.Y) <= radius && w.geometry().Visible(center.X, center.Y, x, y)
}

func (w *World) hydraFoe(ai *aiState, hurt bool) (hydraTarget, bool) {
	h := &ai.Hydra
	center := HydraCenter(ai.AI)
	if p := w.player; !p.Dead && worldmap.Default().BodyFloor(p.X, p.Y, parameters.PlayerSize) == worldmap.SecondFloor {
		target := playerTarget(p)
		engaged := h.FoeID == p.ID
		if engaged && math.Hypot(target.x-ai.hydra.anchorX, target.y-ai.hydra.anchorY) > parameters.HydraChaseLeash {
			ai.hydra.calmUntil = w.tick + parameters.HydraLoseSightTicks
			engaged = false
		}
		territory := h.Phase == HydraHunting || worldmap.Default().InCave(target.x, target.y) || worldmap.Default().InCave(center.X, center.Y)
		sees := w.hydraSees(ai, target.x, target.y, parameters.HydraDetectionRadius) && !w.geometry().OverlapsCover(physics.Square(p.X, p.Y, parameters.PlayerSize))
		provoked := hurt && math.Hypot(target.x-center.X, target.y-center.Y) <= parameters.HydraChaseLeash
		if sees && (territory || engaged) && w.tick >= ai.hydra.calmUntil || provoked {
			if !engaged {
				ai.hydra.anchorX, ai.hydra.anchorY = center.X, center.Y
			}
			ai.hydra.foeSeen = w.tick
			ai.hydra.lastX, ai.hydra.lastY = target.x, target.y
			return target, true
		}
		if engaged && w.tick < ai.hydra.foeSeen+parameters.HydraLoseSightTicks {
			target.x, target.y, target.vx, target.vy = ai.hydra.lastX, ai.hydra.lastY, 0, 0
			return target, true
		}
	}
	if h.FoeID != "" && h.FoeID != w.player.ID {
		snake := w.aiByID(h.FoeID)
		if snake == nil && h.Phase == HydraHunting {
			ai.hydra.prey, ai.hydra.sated = "", true
			h.PhaseRemaining = min(h.PhaseRemaining, parameters.HydraSatedHuntSeconds)
		}
		if snake != nil && (h.Phase == HydraHunting || snake.snake.foe == ai.ID) && hydraCanBeat(ai, snake) &&
			w.hydraSees(ai, snake.X+parameters.AISize/2, snake.Y+parameters.AISize/2, parameters.HydraDetectionRadius) {
			return snakeTarget(snake), true
		}
	}
	if prey := w.aiByID(ai.hydra.prey); prey != nil && h.Phase == HydraHunting && !ai.hydra.sated && hydraCanBeat(ai, prey) &&
		w.hydraSees(ai, prey.X+parameters.AISize/2, prey.Y+parameters.AISize/2, parameters.HydraDetectionRadius) {
		return snakeTarget(prey), true
	}
	return hydraTarget{}, false
}

func (w *World) provokeHydra(ai *aiState, foe *aiState) {
	if ai.Hydra.FoeID == "" && ai.Hydra.Phase != HydraSleeping && hydraCanBeat(ai, foe) {
		ai.Hydra.FoeID = foe.ID
	}
}

func hydraCanBeat(hydra, snake *aiState) bool {
	return hydra.Stats.HP*hydra.Stats.Damage >= parameters.HydraSafeMargin*snake.Stats.HP*snake.Stats.Damage
}

func (w *World) turnHydra(ai *aiState, dx, dy float64) {
	if math.Hypot(dx, dy) < 1e-6 {
		return
	}
	facing := ai.Hydra.Facing
	current := math.Atan2(facing.Y, facing.X)
	turn := math.Remainder(math.Atan2(dy, dx)-current, 2*math.Pi)
	limit := parameters.HydraTurnRate * parameters.StepSeconds
	angle := current + max(-limit, min(limit, turn))
	ai.Hydra.Facing = Direction{X: math.Cos(angle), Y: math.Sin(angle)}
}

func (w *World) hydraMove(ai *aiState, x, y, speed float64) Direction {
	d := ai.navigateMap(w.geometry(), x-parameters.AISize/2, y-parameters.AISize/2)
	if d.X != 0 || d.Y != 0 {
		w.turnHydra(ai, d.X, d.Y)
	}
	d.X, d.Y = d.X*speed, d.Y*speed
	return d
}

func (w *World) hydraFight(ai *aiState, foe hydraTarget) Direction {
	h := &ai.Hydra
	center := HydraCenter(ai.AI)
	dx, dy := foe.x-center.X, foe.y-center.Y
	distance := math.Hypot(dx, dy)
	w.turnHydra(ai, dx, dy)
	aim := Direction{AimX: dx / max(1, distance), AimY: dy / max(1, distance)}
	if h.aliveHeads() == 0 {
		ai.Behavior = BehaviorCowering
		return w.hydraBackOff(ai, foe, aim, 1)
	}
	w.startHydraAttacks(ai, foe, distance)
	if h.lunging() {
		ai.Behavior = BehaviorWindingLunge
		return aim
	}
	ai.Behavior = BehaviorPursuing
	if h.spitting() {
		ai.Behavior = BehaviorWindingSpit
	}
	switch {
	case distance > parameters.HydraPreferredRange+40:
		d := w.hydraMove(ai, foe.x, foe.y, 1)
		d.AimX, d.AimY = aim.AimX, aim.AimY
		return d
	case distance < parameters.HydraCloseRange:
		return w.hydraBackOff(ai, foe, aim, .7)
	}
	if ai.hydra.strafe == 0 || w.tick >= ai.hydra.strafeUntil {
		ai.hydra.strafe = 1
		if rand.IntN(2) == 0 {
			ai.hydra.strafe = -1
		}
		ai.hydra.strafeUntil = w.tick + uint64(parameters.TickRate*(2+rand.IntN(3)))
	}
	move := Direction{X: -aim.AimY * ai.hydra.strafe * .35, Y: aim.AimX * ai.hydra.strafe * .35}
	body := MobBody(ai.Type, ai.X+move.X*40, ai.Y+move.Y*40)
	if !w.geometry().CanOccupy(body) {
		ai.hydra.strafe = -ai.hydra.strafe
		move.X, move.Y = -move.X, -move.Y
	}
	move.AimX, move.AimY = aim.AimX, aim.AimY
	return move
}

func (w *World) hydraBackOff(ai *aiState, foe hydraTarget, aim Direction, speed float64) Direction {
	center := HydraCenter(ai.AI)
	away := math.Atan2(center.Y-foe.y, center.X-foe.x)
	geometry := w.geometry()
	for _, turn := range []float64{0, .5, -.5, 1, -1, 1.6, -1.6} {
		mx, my := math.Cos(away+turn), math.Sin(away+turn)
		if geometry.CanOccupy(MobBody(ai.Type, ai.X+mx*50, ai.Y+my*50)) {
			aim.X, aim.Y = mx*speed, my*speed
			return aim
		}
	}
	return aim
}

func (w *World) hydraCanEat(fruit Fruit) bool {
	return fruit.Source != "" || len(fruit.Patch) > len(worldmap.CaveFeedingPrefix) && fruit.Patch[:len(worldmap.CaveFeedingPrefix)] == worldmap.CaveFeedingPrefix
}

func (w *World) hydraFood(ai *aiState, radius float64, inCave bool) (float64, float64, bool) {
	center := HydraCenter(ai.AI)
	best, bestX, bestY := radius, 0.0, 0.0
	for _, fruit := range w.fruits {
		if w.tick < fruit.AvailableTick || fruit.ExpiresTick != 0 && w.tick >= fruit.ExpiresTick || !w.hydraCanEat(fruit) {
			continue
		}
		x, y := fruit.X+fruit.Size/2, fruit.Y+fruit.Size/2
		if inCave && !worldmap.Default().InCave(x, y) {
			continue
		}
		if distance := math.Hypot(x-center.X, y-center.Y); distance < best {
			best, bestX, bestY = distance, x, y
		}
	}
	return bestX, bestY, best < radius
}

func (w *World) hydraGraze(ai *aiState, x, y float64) Direction {
	center := HydraCenter(ai.AI)
	ai.Behavior = BehaviorFeeding
	reach := parameters.HydraNeckLength * ai.Hydra.Scale * parameters.HydraGrazeReach
	if math.Hypot(x-center.X, y-center.Y) <= reach {
		ai.hydra.grazing, ai.hydra.grazeX, ai.hydra.grazeY = true, x, y
		w.turnHydra(ai, x-center.X, y-center.Y)
		if math.Hypot(x-center.X, y-center.Y) > reach*.6 {
			return w.hydraMove(ai, x, y, .35)
		}
		return Direction{}
	}
	return w.hydraMove(ai, x, y, .6)
}

func (w *World) hydraForage(ai *aiState, inCave bool) Direction {
	if x, y, ok := w.hydraFood(ai, parameters.HydraForageRadius, inCave); ok {
		ai.hydra.roaming = false
		return w.hydraGraze(ai, x, y)
	}
	ai.Behavior = BehaviorPatrolling
	center := HydraCenter(ai.AI)
	if !ai.hydra.roaming || ai.hydra.roamTicks <= 0 || math.Hypot(ai.hydra.roamX-center.X, ai.hydra.roamY-center.Y) < 30 {
		var spots []worldmap.FeedingPatch
		for _, patch := range worldmap.Default().Patches() {
			if len(patch.ID) > len(worldmap.CaveFeedingPrefix) && patch.ID[:len(worldmap.CaveFeedingPrefix)] == worldmap.CaveFeedingPrefix {
				spots = append(spots, patch)
			}
		}
		x, y := w.hydraLair(ai)
		ai.hydra.roamX, ai.hydra.roamY = x+parameters.AISize/2, y+parameters.AISize/2
		if len(spots) > 0 && rand.IntN(len(spots)+1) > 0 {
			at := spots[rand.IntN(len(spots))].At.Center()
			ai.hydra.roamX, ai.hydra.roamY = at.X, at.Y
		}
		ai.hydra.roaming, ai.hydra.roamTicks = true, 12*parameters.TickRate
	}
	ai.hydra.roamTicks--
	return w.hydraMove(ai, ai.hydra.roamX, ai.hydra.roamY, .45)
}

func (w *World) hydraHunt(ai *aiState) Direction {
	if x, y, ok := w.hydraFood(ai, parameters.HydraLootRadius, false); ok {
		return w.hydraGraze(ai, x, y)
	}
	center := HydraCenter(ai.AI)
	if ai.hydra.sated {
		return w.hydraReturn(ai)
	}
	if w.tick >= ai.hydra.preyRethink {
		ai.hydra.preyRethink = w.tick + parameters.HydraHuntRethinkTicks
		ai.hydra.prey = ""
		best := parameters.HydraHuntRadius
		for i := range w.ais {
			snake := &w.ais[i]
			if snake.Type != ForestSnake || snake.respawnTick != 0 || !hydraCanBeat(ai, snake) {
				continue
			}
			if distance := math.Hypot(snake.X-ai.X, snake.Y-ai.Y); distance < best {
				best, ai.hydra.prey = distance, snake.ID
			}
		}
	}
	ai.Behavior = BehaviorHunting
	if prey := w.aiByID(ai.hydra.prey); prey != nil {
		ai.hydra.roaming = false
		return w.hydraMove(ai, prey.X+parameters.AISize/2, prey.Y+parameters.AISize/2, 1)
	}
	if !ai.hydra.roaming || ai.hydra.roamTicks <= 0 || math.Hypot(ai.hydra.roamX-center.X, ai.hydra.roamY-center.Y) < 40 {
		var spots []worldmap.FeedingPatch
		for _, patch := range worldmap.Default().Patches() {
			at := patch.At.Center()
			if patch.Style == worldmap.VioletGarden && !worldmap.Default().InCave(at.X, at.Y) && math.Hypot(at.X-center.X, at.Y-center.Y) < parameters.HydraHuntRadius {
				spots = append(spots, patch)
			}
		}
		if len(spots) == 0 {
			return w.hydraReturn(ai)
		}
		at := spots[rand.IntN(len(spots))].At.Center()
		ai.hydra.roamX, ai.hydra.roamY = at.X, at.Y
		ai.hydra.roaming, ai.hydra.roamTicks = true, 20*parameters.TickRate
	}
	ai.hydra.roamTicks--
	return w.hydraMove(ai, ai.hydra.roamX, ai.hydra.roamY, .8)
}

func (w *World) hydraReturn(ai *aiState) Direction {
	h := &ai.Hydra
	x, y := w.hydraLair(ai)
	if math.Hypot(x-ai.X, y-ai.Y) <= 24 {
		h.Phase, h.PhaseRemaining = HydraSleeping, parameters.HydraSleepSeconds
		ai.Behavior = BehaviorSleeping
		return Direction{}
	}
	ai.Behavior = BehaviorReturning
	return w.hydraMove(ai, x+parameters.AISize/2, y+parameters.AISize/2, .8)
}

func (w *World) growHydra(ai *aiState) {
	maxHP := ai.Stats.MaxHP()
	ai.Hydra.Scale = hydraScale(ai.Stats.Level)
	if math.Abs(maxHP-ai.hydra.maxHP) < 1e-9 || ai.hydra.maxHP <= 0 {
		return
	}
	ratio := maxHP / ai.hydra.maxHP
	ai.Stats.HP *= ratio
	ai.hydra.lastHP *= ratio
	for i := range ai.Hydra.Heads {
		head := &ai.Hydra.Heads[i]
		head.MaxHP = maxHP * parameters.HydraHeadHealth
		head.HP *= ratio
	}
	ai.hydra.maxHP = maxHP
	ai.spawnMaxHP = max(ai.spawnMaxHP, maxHP)
}

func (w *World) hydraEat(ai *aiState, food float64) {
	ai.Stats.GainXP(food * parameters.HydraFoodXP)
}
