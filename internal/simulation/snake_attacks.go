package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/physics"
)

type SnakeAttack struct {
	SpitRemaining    float64
	SpitDirection    Direction
	SlitherPhase     DashPhase
	SlitherRemaining float64
	SlitherDirection Direction
}

func (s *SnakeAttack) endSlither() {
	s.SlitherPhase, s.SlitherRemaining = DashRecovering, parameters.SnakeSlitherRecovery
}

func aimFrom(ai *aiState, x, y float64) Direction {
	dx, dy := x-ai.X, y-ai.Y
	length := max(1, math.Hypot(dx, dy))
	return Direction{X: dx / length, Y: dy / length}
}

func (w *World) snakeAttackInput(ai *aiState) (Direction, bool) {
	s := &ai.Snake
	target := w.player
	sight := parameters.SnakeDetectionRadius + parameters.SnakeSpitRange
	if s.SpitRemaining > 0 {
		before := s.SpitRemaining
		s.SpitRemaining = max(0, before-parameters.StepSeconds)
		ai.Behavior = BehaviorWindingSpit
		if before > parameters.SnakeSpitAimLock && w.aiCanSeePlayer(ai, target, sight) {
			travel := math.Hypot(target.X-ai.X, target.Y-ai.Y) / parameters.SnakeSpitSpeed
			intercept := meleeIntercept(target, before+travel)
			s.SpitDirection = aimFrom(ai, intercept.X, intercept.Y)
		}
		if s.SpitRemaining <= 1e-9 {
			s.SpitRemaining = 0
			w.spitVenom(ai)
		}
		return Direction{AimX: s.SpitDirection.X, AimY: s.SpitDirection.Y}, true
	}
	switch s.SlitherPhase {
	case DashWinding:
		ai.Behavior = BehaviorCoiling
		if s.SlitherRemaining > parameters.SnakeSlitherAimLock && w.aiCanSeePlayer(ai, target, sight) {
			s.SlitherDirection = aimFrom(ai, target.X, target.Y)
		}
		if s.SlitherRemaining -= parameters.StepSeconds; s.SlitherRemaining <= 1e-9 {
			s.SlitherPhase, s.SlitherRemaining = DashRushing, parameters.SnakeSlitherSurge
			ai.snake.slitherHit = false
		}
	case DashRushing:
		ai.Behavior = BehaviorSlithering
		if s.SlitherRemaining -= parameters.StepSeconds; s.SlitherRemaining <= 1e-9 {
			s.endSlither()
		}
	case DashRecovering:
		ai.Behavior = BehaviorRecoveringStrike
		if s.SlitherRemaining -= parameters.StepSeconds; s.SlitherRemaining <= 1e-9 {
			s.SlitherPhase, s.SlitherRemaining = DashIdle, 0
			w.startSnakeRetreat(ai)
		}
	default:
		return Direction{}, false
	}
	return Direction{AimX: s.SlitherDirection.X, AimY: s.SlitherDirection.Y}, true
}

func (w *World) startSnakeAttack(ai *aiState, target *playerState) (Direction, bool) {
	if ai.attackCooldown > 0 || w.tick < ai.snake.attackAfter {
		return Direction{}, false
	}
	geometry := w.geometry()
	dx, dy := target.X-ai.X, target.Y-ai.Y
	distance := math.Hypot(dx, dy)
	canSlither := w.tick >= ai.snake.slitherReady && distance >= parameters.SnakeSlitherMinRange && distance <= parameters.SnakeSlitherRange
	if canSlither {
		body := MobBody(ai.Type, ai.X, ai.Y)
		limit, _ := geometry.MovementLimit(body.X, body.Y, dx, dy, body.W, body.H)
		canSlither = limit >= 1-1e-6
	}
	canSpit := w.tick >= ai.snake.spitReady && distance >= parameters.SnakeSpitMinRange && distance <= parameters.SnakeSpitRange && ai.Stats.MP >= parameters.SnakeSpitCost
	if canSpit {
		size := parameters.SnakeSpitSize
		_, blocked := geometry.ShotLimit(ai.X+(parameters.AISize-size)/2, ai.Y+(parameters.AISize-size)/2, dx, dy, size, size)
		canSpit = !blocked
	}
	if canSlither && canSpit {
		canSlither = rand.IntN(2) == 0
		canSpit = !canSlither
	}
	s := &ai.Snake
	var busy float64
	switch {
	case canSlither:
		s.SlitherPhase, s.SlitherRemaining = DashWinding, parameters.SnakeSlitherCoil
		s.SlitherDirection = aimFrom(ai, target.X, target.Y)
		ai.snake.slitherReady = w.tick + parameters.SnakeSlitherCooldownTicks + uint64(rand.IntN(parameters.TickRate))
		busy = parameters.SnakeSlitherCoil + parameters.SnakeSlitherSurge + parameters.SnakeSlitherRecovery
		ai.Behavior = BehaviorCoiling
	case canSpit:
		s.SpitRemaining = parameters.SnakeSpitWindup
		s.SpitDirection = aimFrom(ai, target.X, target.Y)
		ai.snake.spitReady = w.tick + parameters.SnakeSpitCooldownTicks + uint64(rand.IntN(parameters.TickRate))
		busy = parameters.SnakeSpitWindup
		ai.Behavior = BehaviorWindingSpit
	default:
		return Direction{}, false
	}
	ai.snake.attackAfter = w.tick + uint64(math.Ceil(busy*parameters.TickRate)) + parameters.SnakeSpecialGapTicks
	ai.alert.aggressorID, ai.alert.x, ai.alert.y = target.ID, target.X, target.Y
	ai.alert.until = w.tick + parameters.AIAlertSeconds*parameters.TickRate
	return Direction{AimX: dx / max(1, distance), AimY: dy / max(1, distance)}, true
}

func (w *World) spitVenom(ai *aiState) {
	if len(w.projectiles) >= parameters.MaxProjectiles || ai.Stats.MP < parameters.SnakeSpitCost {
		return
	}
	ai.Stats.MP -= parameters.SnakeSpitCost
	ai.ShotsFired++
	size := parameters.SnakeSpitSize
	x, y := ai.X+(parameters.AISize-size)/2, ai.Y+(parameters.AISize-size)/2
	lifetime := parameters.SnakeSpitRange * 1.25 / parameters.SnakeSpitSpeed
	w.nextProjectile++
	w.projectiles = append(w.projectiles, Projectile{
		Kind: VenomProjectile, ID: fmt.Sprintf("venom-%d", w.nextProjectile), Owner: ai.ID, MobOwned: true,
		X: x, Y: y, originX: x, originY: y, Direction: ai.Snake.SpitDirection, Size: size,
		Remaining: lifetime, Lifetime: lifetime, Damage: ai.Stats.Damage * parameters.SnakeSpitDamage,
		PoisonDamage: PoisonStrength(ai.Stats, len(ai.Segments)), PoisonSeconds: parameters.SnakeSpitPoisonSeconds,
	})
}

func (w *World) surgeSnake(ai *aiState) {
	s := &ai.Snake
	ai.snake.surgedTick = w.tick
	if target := w.player; w.aiCanSeePlayer(ai, target, parameters.SnakeDetectionRadius+parameters.SnakeSlitherRange) {
		want := aimFrom(ai, target.X, target.Y)
		facing := s.SlitherDirection
		turn := math.Atan2(facing.X*want.Y-facing.Y*want.X, facing.X*want.X+facing.Y*want.Y)
		limit := parameters.SnakeSlitherTurnRate * parameters.StepSeconds
		angle := math.Atan2(facing.Y, facing.X) + max(-limit, min(limit, turn))
		s.SlitherDirection = Direction{X: math.Cos(angle), Y: math.Sin(angle)}
	}
	ai.Combat.Facing = s.SlitherDirection
	geometry := w.geometry()
	distance := ai.Stats.MovementSpeed() * parameters.SpeedScale * parameters.SnakeSlitherSpeed * parameters.StepSeconds
	for distance > 1e-9 {
		travel := min(distance, 4)
		body := MobBody(ai.Type, ai.X, ai.Y)
		motion := geometry.MoveBody(body, physics.Vector{X: s.SlitherDirection.X * travel, Y: s.SlitherDirection.Y * travel})
		ai.X += motion.Bounds.X - body.X
		ai.Y += motion.Bounds.Y - body.Y
		if motion.Blocked() || motion.Rejected {
			s.endSlither()
			return
		}
		distance -= travel
	}
}

func (w *World) advanceSnakeSlithers(actors []eater) {
	for _, actor := range actors {
		ai := actor.ai
		if ai == nil || ai.Type != ForestSnake || ai.snake.surgedTick != w.tick || ai.snake.slitherHit || actor.dead() {
			continue
		}
		for _, target := range actors {
			if target.player == nil || target.dead() {
				continue
			}
			if _, hit := eatingContact(actor.start, actor.end, actor.size, target.start, target.end, target.size); !hit {
				continue
			}
			ai.snake.slitherHit = true
			w.mobStrike(ai, target, parameters.SnakeSlitherDamage)
			ai.Snake.endSlither()
			break
		}
	}
}
