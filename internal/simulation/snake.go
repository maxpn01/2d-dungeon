package simulation

import (
	"math"
	"math/rand/v2"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/physics"
	"dungeon-raylib-go/internal/worldmap"
)

type snakeState struct {
	stop         int
	biting       bool
	retreatUntil uint64
	side         float64
	lastHP       float64
	recentDamage float64
	forageTicks  int
	forageAfter  uint64
	spitReady    uint64
	slitherReady uint64
	attackAfter  uint64
	surgedTick   uint64
	slitherHit   bool
	foe          string
	foeUntil     uint64
}

func SnakeSegmentBody(segment physics.Vector) physics.Rect {
	offset := (parameters.AISize - parameters.SnakeSegmentSize) / 2
	return physics.Square(segment.X+offset, segment.Y+offset, parameters.SnakeSegmentSize)
}

func (w *World) snakeInput(ai *aiState) Direction {
	ai.trackSnakeDamage()
	if d, busy := w.snakeAttackInput(ai); busy {
		return d
	}
	if ai.Combat.SwordRemaining > 0 {
		ai.snake.biting = true
		return w.meleeGuardInput(ai, nil)
	}
	if ai.snake.biting {
		ai.snake.biting = false
		w.startSnakeRetreat(ai)
	}
	if foe := w.snakeFoe(ai); foe != nil && w.snakeTarget(ai) == nil {
		return w.snakeFightFoe(ai, foe)
	}
	if ai.patrol.returning || ai.alert.aggressorID != "" && w.tick >= ai.alert.until {
		if d, returning := w.investigateAI(ai); returning {
			return d
		}
	}
	if target := w.snakeTarget(ai); target != nil {
		if ai.snake.recentDamage >= parameters.SnakeFlinchDamage*ai.Stats.MaxHP() {
			ai.snake.recentDamage = 0
			w.startSnakeRetreat(ai)
		}
		if w.tick < ai.snake.retreatUntil {
			return w.snakeRetreatInput(ai, target)
		}
		if d, started := w.startSnakeAttack(ai, target); started {
			return d
		}
		return w.snakeWeave(ai, target, w.meleeGuardInput(ai, target))
	}
	if d, investigating := w.investigateAI(ai); investigating {
		return d
	}
	return w.snakePatrolInput(ai)
}

func (w *World) snakeTarget(ai *aiState) *playerState {
	target := w.player
	if !w.aiCanSeePlayer(ai, target, parameters.SnakeDetectionRadius) {
		return nil
	}
	for _, stop := range ai.den.Route {
		x, y := worldmap.CellPosition(stop.X, stop.Y, parameters.AISize)
		if math.Hypot(target.X-x, target.Y-y) <= parameters.SnakeChaseLeash {
			return target
		}
	}
	return nil
}

func (w *World) snakePatrolInput(ai *aiState) Direction {
	if d, feeding := w.snakeForageInput(ai); feeding {
		return d
	}
	route := ai.den.Route
	if len(route) == 0 {
		return ai.patrolInput(w.geometry())
	}
	if ai.patrol.pause > 0 {
		ai.patrol.pause--
		ai.Behavior = BehaviorWatching
		return Direction{}
	}
	geometry := w.geometry()
	if ai.patrol.ticks == 0 {
		ai.chooseSnakeStopPoint(geometry, route[ai.snake.stop%len(route)])
	}
	if math.Hypot(ai.patrol.x-ai.X, ai.patrol.y-ai.Y) <= 18 || ai.patrol.ticks == 1 {
		ai.snake.stop = (ai.snake.stop + 1) % len(route)
		ai.patrol.ticks = 0
		ai.patrol.pause = parameters.SnakeStopPauseTicks/2 + rand.IntN(parameters.SnakeStopPauseTicks)
		ai.Behavior = BehaviorWatching
		return Direction{}
	}
	ai.patrol.ticks--
	ai.Behavior = BehaviorPatrolling
	d := ai.navigateMap(geometry, ai.patrol.x, ai.patrol.y)
	d.X, d.Y = d.X*parameters.SnakePatrolSpeed, d.Y*parameters.SnakePatrolSpeed
	return d
}

func (ai *aiState) chooseSnakeStopPoint(geometry *worldmap.Map, stop worldmap.Cell) {
	ai.patrol.x, ai.patrol.y = worldmap.CellPosition(stop.X, stop.Y, parameters.AISize)
	ai.patrol.ticks = parameters.SnakeLegTimeoutTicks
	radius := float64(ai.den.Patrol.Width/2*parameters.Tile) - parameters.AISize
	for range 16 {
		x := ai.patrol.x + (rand.Float64()*2-1)*radius
		y := ai.patrol.y + (rand.Float64()*2-1)*radius
		if geometry.CanOccupy(MobBody(ai.Type, x, y)) {
			ai.patrol.x, ai.patrol.y = x, y
			return
		}
	}
}

func (ai *aiState) followSegments() {
	lead := physics.Vector{X: ai.X, Y: ai.Y}
	spacing, first := parameters.SnakeSegmentSpacing, parameters.SnakeSegmentSpacing
	if ai.Type == CaveHydra {
		spacing = parameters.HydraTailSpacing * ai.Hydra.Scale
		first = parameters.HydraBodySize * ai.Hydra.Scale * .45
	}
	for i := range ai.Segments {
		segment := &ai.Segments[i]
		gap := spacing
		if i == 0 {
			gap = first
		}
		dx, dy := lead.X-segment.X, lead.Y-segment.Y
		if distance := math.Hypot(dx, dy); distance > gap {
			segment.X += dx / distance * (distance - gap)
			segment.Y += dy / distance * (distance - gap)
		}
		lead = *segment
	}
}

func (ai *aiState) trackSnakeDamage() {
	if ai.snake.lastHP > ai.Stats.HP {
		ai.snake.recentDamage += ai.snake.lastHP - ai.Stats.HP
	}
	ai.snake.lastHP = ai.Stats.HP
	ai.snake.recentDamage = max(0, ai.snake.recentDamage-parameters.SnakeFlinchDecay*ai.Stats.MaxHP()*parameters.StepSeconds)
}

func (w *World) startSnakeRetreat(ai *aiState) {
	ai.snake.retreatUntil = w.tick + uint64(parameters.SnakeRetreatMinTicks+rand.IntN(parameters.SnakeRetreatVariationTicks+1))
	ai.snake.side = 1
	if rand.IntN(2) == 0 {
		ai.snake.side = -1
	}
}

func (w *World) snakeRetreatInput(ai *aiState, target *playerState) Direction {
	ai.Behavior = BehaviorRetreating
	ai.alert.aggressorID, ai.alert.x, ai.alert.y = target.ID, target.X, target.Y
	ai.alert.until = w.tick + parameters.AIAlertSeconds*parameters.TickRate
	dx, dy := ai.X-target.X, ai.Y-target.Y
	distance := max(1, math.Hypot(dx, dy))
	d := Direction{AimX: -dx / distance, AimY: -dy / distance}
	if distance >= parameters.SnakeRetreatDistance {
		ai.snake.retreatUntil = w.tick
		return d
	}
	geometry := w.geometry()
	away := math.Atan2(dy, dx)
	for _, turn := range []float64{0, .5, -.5, 1, -1, 1.6, -1.6} {
		angle := away + turn*ai.snake.side
		mx, my := math.Cos(angle), math.Sin(angle)
		near := MobBody(ai.Type, ai.X+mx*parameters.SnakeRetreatProbe/2, ai.Y+my*parameters.SnakeRetreatProbe/2)
		far := MobBody(ai.Type, ai.X+mx*parameters.SnakeRetreatProbe, ai.Y+my*parameters.SnakeRetreatProbe)
		if geometry.CanOccupy(near) && geometry.CanOccupy(far) {
			d.X, d.Y = mx*parameters.SnakeRetreatSpeed, my*parameters.SnakeRetreatSpeed
			return d
		}
	}
	ai.snake.retreatUntil = w.tick
	return d
}

func (w *World) snakeWeave(ai *aiState, target *playerState, d Direction) Direction {
	if ai.Behavior != BehaviorPursuing || math.Hypot(target.X-ai.X, target.Y-ai.Y) < parameters.SnakeBiteReach*2 {
		return d
	}
	phase := float64(w.tick) * parameters.StepSeconds * 2 * math.Pi / parameters.SnakeWeaveSeconds
	sway := math.Sin(phase+float64(ai.den.At.X*7+ai.den.At.Y*13)) * parameters.SnakeWeaveStrength
	x, y := d.X-d.Y*sway, d.Y+d.X*sway
	if length := math.Hypot(x, y); length > 1 {
		x, y = x/length, y/length
	}
	d.X, d.Y = x, y
	return d
}

func (w *World) snakesFeeding() bool {
	return !w.player.Dead && worldmap.Default().BodyFloor(w.player.X, w.player.Y, parameters.PlayerSize) == worldmap.SecondFloor
}

func (w *World) snakeForageInput(ai *aiState) (Direction, bool) {
	if w.tick < ai.snake.forageAfter || !w.snakesFeeding() {
		return Direction{}, false
	}
	body := MobBody(ai.Type, ai.X, ai.Y)
	headX, headY := body.X+body.W/2, body.Y+body.H/2
	best, bestX, bestY := parameters.SnakeForageRadius, 0.0, 0.0
	for _, fruit := range w.fruits {
		if w.tick < fruit.AvailableTick || fruit.ExpiresTick != 0 && w.tick >= fruit.ExpiresTick {
			continue
		}
		x, y := fruit.X+fruit.Size/2, fruit.Y+fruit.Size/2
		if distance := math.Hypot(x-headX, y-headY); distance < best {
			best, bestX, bestY = distance, x, y
		}
	}
	if best >= parameters.SnakeForageRadius {
		ai.snake.forageTicks = 0
		return Direction{}, false
	}
	if ai.snake.forageTicks++; ai.snake.forageTicks > parameters.SnakeForageTimeoutTicks {
		ai.snake.forageTicks = 0
		ai.snake.forageAfter = w.tick + parameters.SnakeForageTimeoutTicks
		return Direction{}, false
	}
	ai.Behavior = BehaviorFeeding
	d := ai.navigateMap(w.geometry(), bestX-parameters.AISize/2, bestY-parameters.AISize/2)
	d.X, d.Y = d.X*parameters.SnakePatrolSpeed, d.Y*parameters.SnakePatrolSpeed
	return d, true
}

func (ai *aiState) snakeEat(food float64) {
	if len(ai.Segments) >= parameters.SnakeMaxSegments {
		return
	}
	level, maxHP := ai.Stats.Level, ai.Stats.MaxHP()
	ai.Stats.GainXP(food * parameters.SnakeFoodXP)
	if ai.Stats.Level == level {
		return
	}
	if room := parameters.SnakeMaxSegments - len(ai.Segments); ai.Stats.Level-level >= room {
		ai.Stats.Level, ai.Stats.XP = level+room, 0
		ai.Stats.rebuildPerks()
	}
	ai.Stats.HP *= ai.Stats.MaxHP() / maxHP
	ai.spawnMaxHP = max(ai.spawnMaxHP, ai.Stats.MaxHP())
	for range ai.Stats.Level - level {
		tail := physics.Vector{X: ai.X, Y: ai.Y}
		if len(ai.Segments) > 0 {
			tail = ai.Segments[len(ai.Segments)-1]
		}
		ai.Segments = append(ai.Segments, tail)
	}
}

func (w *World) provokeSnake(ai *aiState, foe string) {
	ai.snake.foe, ai.snake.foeUntil = foe, w.tick+parameters.HydraFoeSeconds*parameters.TickRate
}

func (w *World) snakeFoe(ai *aiState) *aiState {
	if ai.snake.foe == "" || w.tick >= ai.snake.foeUntil {
		ai.snake.foe = ""
		return nil
	}
	return w.aiByID(ai.snake.foe)
}

func (w *World) snakeFightFoe(ai *aiState, foe *aiState) Direction {
	center := HydraCenter(foe.AI)
	dx, dy := center.X-ai.X-parameters.AISize/2, center.Y-ai.Y-parameters.AISize/2
	distance := math.Hypot(dx, dy)
	d := Direction{AimX: dx / max(1, distance), AimY: dy / max(1, distance)}
	gap := distance - HydraBodyRadius(foe.Hydra)
	if ai.attackCooldown == 0 && gap <= parameters.SnakeBiteReach-4 {
		attack := MobMelee(ai.Type)
		ai.Combat.SwordRemaining = attack.Duration
		ai.Combat.SwordDirection = Direction{X: d.AimX, Y: d.AimY}
		ai.meleeHits = nil
		ai.attackCooldown = int(math.Ceil((attack.Duration + 1/ai.Stats.AttackRate) * parameters.TickRate))
		ai.Behavior = BehaviorWindingStrike
		return d
	}
	ai.Behavior = BehaviorPursuing
	if gap <= parameters.SnakeBiteReach-4 {
		return d
	}
	movement := ai.navigateMap(w.geometry(), center.X-parameters.AISize/2, center.Y-parameters.AISize/2)
	d.X, d.Y = movement.X, movement.Y
	return d
}

func (w *World) snakeBiteReaches(ai *aiState, target eater) bool {
	if target.ai == nil || target.ai.Type != CaveHydra || ai.snake.foe != target.id {
		return false
	}
	center := HydraCenter(target.ai.AI)
	distance := math.Hypot(center.X-ai.X-parameters.AISize/2, center.Y-ai.Y-parameters.AISize/2)
	return distance <= MobMelee(ai.Type).Reach+HydraBodyRadius(target.ai.Hydra)
}
