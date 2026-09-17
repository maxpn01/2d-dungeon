package game

import (
	"fmt"
	"image/color"
	"math/rand"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/terrain"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type AIActionState int

const (
	ActionSearchingFood AIActionState = iota
	ActionEating
	ActionFleeing
	ActionSleeping
)

type AIStats = PlayerStats

type AI struct {
	pos   rl.Vector2
	size  rl.Vector2
	color color.RGBA

	stats AIStats

	actionState AIActionState

	awakeSeconds       float32
	asleepSeconds      float32
	awakeTimerSeconds  float32
	asleepTimerSeconds float32

	detectionRadius float32

	wanderDirection rl.Vector2
	escapeDirection rl.Vector2
	escapeTimer     float32

	isDeadFlag     bool
	respawnSeconds float32
}

func NewAI(
	pos,
	size rl.Vector2,
	color color.RGBA,
	stats AIStats,
	detectionRadius float32,
	awake, asleep float32,
) *AI {
	return &AI{
		pos:                pos,
		size:               size,
		color:              color,
		stats:              stats,
		detectionRadius:    detectionRadius,
		wanderDirection:    rl.Vector2{},
		isDeadFlag:         false,
		awakeSeconds:       awake,
		asleepSeconds:      asleep,
		awakeTimerSeconds:  awake,
		asleepTimerSeconds: asleep,
	}
}

func (ai *AI) update(w *World, dt float32) {
	if ai.isDeadFlag {
		return
	}

	ai.checkDangers(w.player)
	ai.updateAwakenessStatus(dt)

	switch ai.actionState {
	case ActionSearchingFood:
		ai.seekFood(w.fruitSpawner, w.player, dt)
	case ActionEating:
		ai.actionState = ActionSearchingFood
	case ActionFleeing:
		ai.flee(w.player, w.gameMap, dt)
		ai.handleFoodInteractions(w.fruitSpawner, w.player)
	}
}

func (ai *AI) draw() {
	if ai.isDeadFlag {
		return
	}
	rl.DrawRectangleV(ai.pos, ai.size, ai.color)
	drawTextAt(fmt.Sprintf("AI Lv.%d %s", ai.stats.lvl, ai.actionState), rl.NewVector2(ai.pos.X, ai.pos.Y-22), 18, rl.RayWhite)
}

func (ai *AI) checkDangers(p *Player) {
	radius := ai.detectionRadius

	if ai.actionState == ActionFleeing {
		radius *= 1.2
	}

	danger := p.canEatAI(ai) && ai.distanceTo(p.pos, p.size) <= radius*radius

	if danger {
		ai.actionState = ActionFleeing
	} else if ai.actionState == ActionFleeing {
		ai.actionState = ActionSearchingFood
		ai.escapeTimer = 0
	}
}

func (ai *AI) flee(p *Player, gameMap *Map, dt float32) {
	origin := ai.pos
	aiCenter := rectCenter(ai.pos, ai.size)
	playerCenter := rectCenter(p.pos, p.size)

	away := rl.NewVector2(aiCenter.X-playerCenter.X, aiCenter.Y-playerCenter.Y)
	direction := ai.chooseFleeDirection(gameMap, away, dt)

	move := rl.Vector2Normalize(direction)

	ai.pos.X += move.X * ai.stats.speed * simulation.SpeedScale * dt
	ai.pos.Y += move.Y * ai.stats.speed * simulation.SpeedScale * dt

	ai.resolveMovement(origin)
}

func (ai *AI) chooseFleeDirection(gameMap *Map, away rl.Vector2, dt float32) rl.Vector2 {
	direction := away
	if direction.X == 0 && direction.Y == 0 {
		direction = rl.NewVector2(1, 0)
	}

	if ai.escapeTimer > 0 {
		direction = ai.escapeDirection
		ai.escapeTimer = max(0, ai.escapeTimer-dt)
	}

	walls := struct{ left, right, top, bottom bool }{
		left:   ai.pos.X <= gameMap.edgePosX.start,
		right:  ai.pos.X+ai.size.X >= gameMap.edgePosX.end,
		top:    ai.pos.Y <= gameMap.edgePosY.start,
		bottom: ai.pos.Y+ai.size.Y >= gameMap.edgePosY.end,
	}

	if (walls.left && direction.X < 0) || (walls.right && direction.X > 0) {
		direction.X = 0
		ai.escapeTimer = 0
	}
	if (walls.top && direction.Y < 0) || (walls.bottom && direction.Y > 0) {
		direction.Y = 0
		ai.escapeTimer = 0
	}

	if direction.X != 0 || direction.Y != 0 {
		return direction
	}

	inwardX, inwardY := float32(1), float32(1)
	if walls.right {
		inwardX = -1
	}
	if walls.bottom {
		inwardY = -1
	}

	switch {
	case (walls.left || walls.right) && (walls.top || walls.bottom):
		if inwardX*away.X >= inwardY*away.Y {
			direction = rl.NewVector2(inwardX, 0)
		} else {
			direction = rl.NewVector2(0, inwardY)
		}
	case walls.left || walls.right:
		direction = rl.NewVector2(0, 1)
	case walls.top || walls.bottom:
		direction = rl.NewVector2(1, 0)
	}

	ai.escapeDirection = direction
	ai.escapeTimer = 0.3

	return direction
}

func (ai *AI) seekFood(fs *FruitSpawner, p *Player, dt float32) {
	targetPos, targetSize, targetFound := ai.chooseFoodTarget(fs, p)

	if targetFound {
		ai.move(targetPos, targetSize, dt)
	} else {
		ai.wander(dt)
	}

	ai.handleFoodInteractions(fs, p)
}

func (ai *AI) chooseFoodTarget(fs *FruitSpawner, p *Player) (targetPos rl.Vector2, targetSize rl.Vector2, targetFound bool) {
	closestFruit, closestFruitDistance, fruitFound := ai.findClosestFruit(fs, p)

	if ai.canEatPlayer(p) {
		playerDistance := ai.distanceTo(p.pos, p.size)
		radiusSquared := ai.detectionRadius * ai.detectionRadius
		playerIsWithinDetection := playerDistance <= radiusSquared

		if playerIsWithinDetection {
			playerIsCloser := playerDistance <= closestFruitDistance

			if !fruitFound || playerIsCloser {
				return p.pos, p.size, true
			}
		}
	}

	if fruitFound {
		return closestFruit.pos, closestFruit.size, true
	}

	return rl.Vector2{}, rl.Vector2{}, false
}

func (ai *AI) findClosestFruit(fs *FruitSpawner, p *Player) (closestFruit Fruit, closestFruitDistance float32, found bool) {
	radiusSquared := ai.detectionRadius * ai.detectionRadius

	for _, fruit := range fs.fruits {
		distanceToPlayer := findSquaredEuclideanDistance(rectCenter(fruit.pos, fruit.size), rectCenter(p.pos, p.size))

		if p.canEatAI(ai) && distanceToPlayer <= radiusSquared {
			continue
		}

		distanceToFruit := ai.distanceTo(fruit.pos, fruit.size)

		if distanceToFruit > radiusSquared {
			continue
		}

		if !found || distanceToFruit < closestFruitDistance {
			closestFruit = fruit
			closestFruitDistance = distanceToFruit
			found = true
		}
	}

	return closestFruit, closestFruitDistance, found
}

func (ai *AI) distanceTo(pos, size rl.Vector2) float32 {
	return findSquaredEuclideanDistance(rectCenter(ai.pos, ai.size), rectCenter(pos, size))
}

func (ai *AI) canEatPlayer(p *Player) bool {
	return ai.stats.hp > p.stats.hp
}

func (ai *AI) move(targetPos, targetSize rl.Vector2, dt float32) {
	origin := ai.pos
	target := rl.NewVector2(
		targetPos.X+targetSize.X/2-ai.size.X/2,
		targetPos.Y+targetSize.Y/2-ai.size.Y/2,
	)
	ai.pos = rl.Vector2MoveTowards(ai.pos, target, ai.stats.speed*simulation.SpeedScale*dt)

	ai.resolveMovement(origin)
}

func (ai *AI) wander(dt float32) {
	origin := ai.pos
	if ai.wanderDirection.X == 0 && ai.wanderDirection.Y == 0 {
		ai.wanderDirection = rl.NewVector2(
			float32(rand.Intn(3)-1),
			float32(rand.Intn(3)-1),
		)
	}

	direction := ai.wanderDirection

	move := rl.Vector2Normalize(direction)

	ai.pos.X += move.X * ai.stats.speed * simulation.SpeedScale * dt
	ai.pos.Y += move.Y * ai.stats.speed * simulation.SpeedScale * dt

	attemptedPos := ai.pos

	ai.resolveMovement(origin)

	directionShouldReset := ai.pos != attemptedPos

	if directionShouldReset {
		ai.wanderDirection = rl.Vector2{}
	}
}

func (ai *AI) handleFoodInteractions(fs *FruitSpawner, p *Player) {
	hasAICollidedWithFruit, fruitIndex := ai.checkFruitCollision(fs)

	if hasAICollidedWithFruit && fruitIndex >= 0 && fruitIndex < len(fs.fruits) {
		ai.eatFood(simulation.FruitHP, simulation.FruitXP)
		fs.despawnFruit(fruitIndex)
	}

	hasAICollidedWithPlayer := checkCollisions(ai.pos, ai.size, p.pos, p.size)

	if hasAICollidedWithPlayer && ai.canEatPlayer(p) {
		ai.eatFood(p.stats.hp*simulation.ActorHPFraction, p.stats.hp)
		p.markPlayerDeath()
	}
}

func (ai *AI) checkFruitCollision(fs *FruitSpawner) (bool, int) {
	for i, v := range slices.Backward(fs.fruits) {
		hasAICollidedWithFruit := checkCollisions(ai.pos, ai.size, v.pos, v.size)

		if hasAICollidedWithFruit {
			return true, i
		}
	}

	return false, -1
}

func (ai *AI) eatFood(hp, xp float32) {
	ai.stats.eatFood(hp, xp)
	if ai.actionState != ActionFleeing {
		ai.actionState = ActionEating
	}
}

func (ai *AI) markAIDeath() {
	ai.isDeadFlag = true
	ai.respawnSeconds = simulation.AIRespawnSeconds
}

func (ai *AI) updateAwakenessStatus(dt float32) {
	switch ai.actionState {
	case ActionSearchingFood, ActionEating, ActionFleeing:
		ai.awakeTimerSeconds -= dt

		if ai.awakeTimerSeconds <= 0 && ai.actionState != ActionFleeing {
			ai.actionState = ActionSleeping
			ai.asleepTimerSeconds = ai.asleepSeconds
		}
	case ActionSleeping:
		ai.asleepTimerSeconds -= dt

		if ai.asleepTimerSeconds <= 0 {
			ai.actionState = ActionSearchingFood
			ai.awakeTimerSeconds = ai.awakeSeconds
		}
	}
}

func (ai *AI) resolveMovement(origin rl.Vector2) {
	x, y := terrain.Move(float64(origin.X), float64(origin.Y), float64(ai.pos.X-origin.X), float64(ai.pos.Y-origin.Y), float64(ai.size.X), float64(ai.size.Y))
	ai.pos = rl.NewVector2(float32(x), float32(y))
}

func (state AIActionState) String() string {
	switch state {
	case ActionSearchingFood:
		return "searching"
	case ActionEating:
		return "eating"
	case ActionFleeing:
		return "fleeing"
	case ActionSleeping:
		return "sleeping"
	default:
		return "unknown"
	}
}
