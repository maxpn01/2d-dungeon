package game

import (
	"image/color"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type PlayerStats struct {
	lvl int
	exp float32

	hp               float32
	speed            float32
	sprintMultiplier float32

	maxLvl int
	maxHP  float32
}

type PlayerGrowthStats struct {
	HPPerLevel    float32
	SpeedPerLevel float32
}

type PlayerInput struct {
	Direction rl.Vector2
	Sprint    bool
}

type Player struct {
	name  string
	input PlayerInput
	pos   rl.Vector2
	size  rl.Vector2
	color color.RGBA

	stats PlayerStats

	growth PlayerGrowthStats

	nutritionalFactor float32

	isDeadFlag bool
}

func NewPlayer(
	pos, size rl.Vector2,
	color color.RGBA,
	stats PlayerStats,
	growth PlayerGrowthStats,
	nutritionalFactor float32,
) *Player {
	return &Player{
		pos:               pos,
		size:              size,
		color:             color,
		stats:             stats,
		growth:            growth,
		nutritionalFactor: nutritionalFactor,
		isDeadFlag:        false,
	}
}

func (p *Player) update(w *World, dt float32) {
	p.move(w.gameMap, dt)
	p.handleFoodCollision(w.fruitSpawner, w.ai)
	p.handleLevelUp()
}

func (p *Player) draw() {
	rl.DrawRectangleV(p.pos, p.size, p.color)
	rl.DrawText(p.name, int32(p.pos.X), int32(p.pos.Y)-22, 18, rl.RayWhite)
}

func (p *Player) readInput() {
	input := PlayerInput{Sprint: rl.IsKeyDown(rl.KeyLeftShift)}
	if rl.IsKeyDown(rl.KeyW) {
		input.Direction.Y -= 1
	}
	if rl.IsKeyDown(rl.KeyS) {
		input.Direction.Y += 1
	}
	if rl.IsKeyDown(rl.KeyA) {
		input.Direction.X -= 1
	}
	if rl.IsKeyDown(rl.KeyD) {
		input.Direction.X += 1
	}
	p.input = input
}

func (p *Player) move(gameMap *Map, dt float32) {
	move := p.input.Direction
	speed := p.stats.speed

	if p.input.Sprint {
		speed *= p.stats.sprintMultiplier
	}

	// normalize the diagonal speed
	if move.X != 0 || move.Y != 0 {
		move = rl.Vector2Normalize(move)

		p.pos.X += move.X * speed * dt
		p.pos.Y += move.Y * speed * dt
	}

	clamp(gameMap.edgePosX.start, &p.pos.X, &p.size.X, gameMap.edgePosX.end)
	clamp(gameMap.edgePosY.start, &p.pos.Y, &p.size.Y, gameMap.edgePosY.end)
}

func (p *Player) handleFoodCollision(fs *FruitSpawner, ai *AI) {
	p.handleFruitCollision(fs)

	hasPlayerCollidedWithAI := checkCollisions(p.pos, p.size, ai.pos, ai.size)

	if hasPlayerCollidedWithAI && p.canEatAI(ai) {
		p.eatFood(ai.nutritionalValue())
		ai.markAIDeath()
	}
}

func (p *Player) canEatAI(ai *AI) bool {
	return !ai.isDeadFlag && p.stats.hp > ai.stats.hp
}

func (p *Player) handleFruitCollision(fs *FruitSpawner) {
	for i, v := range slices.Backward(fs.fruits) {
		hasPlayerCollidedWithFruit := checkCollisions(p.pos, p.size, v.pos, v.size)

		if hasPlayerCollidedWithFruit {
			p.eatFood(v.nutritionalValue)
			fs.despawnFruit(i)
		}
	}
}

func (p *Player) nutritionalValue() float32 {
	return p.nutritionalFactor * float32(p.stats.lvl)
}

func (p *Player) eatFood(nutrition float32) {
	p.stats.exp += nutrition
}

func (p *Player) markPlayerDeath() {
	p.isDeadFlag = true
}

func (p *Player) handleLevelUp() {
	for p.stats.lvl < p.stats.maxLvl && p.stats.exp >= calcExpForNextLvl(p.stats.lvl) {
		p.stats.exp -= calcExpForNextLvl(p.stats.lvl)
		p.stats.lvl++

		p.stats.hp = min(p.stats.hp+p.growth.HPPerLevel, p.stats.maxHP)
		p.stats.speed += p.growth.SpeedPerLevel
	}
}
