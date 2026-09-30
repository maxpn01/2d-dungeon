package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/terrain"
)

type Projectile struct {
	ID        string    `json:"id"`
	Owner     string    `json:"owner"`
	X         float64   `json:"x"`
	Y         float64   `json:"y"`
	Direction Direction `json:"direction"`
	Remaining float64   `json:"remaining"`
}

func PrepareShot(p *Player, input Direction) bool {
	if p.Dead || !input.Valid() || !input.Shoot || p.ShotCooldown > 1e-9 || p.Stats.MP < parameters.ShotMPCost || math.Hypot(input.AimX, input.AimY) == 0 {
		return false
	}
	p.Stats.MP -= parameters.ShotMPCost
	p.ShotCooldown = parameters.ShotCooldown
	return true
}

func (w *World) fire(p *Player, input Direction) {
	if len(w.projectiles) >= parameters.MaxProjectiles || InSafeZone(p.X, p.Y, PlayerSize, p.Stats.HP, w.safeZones) || !PrepareShot(p, input) {
		return
	}
	length := math.Hypot(input.AimX, input.AimY)
	w.nextProjectile++
	w.projectiles = append(w.projectiles, Projectile{
		ID: fmt.Sprintf("shot-%d", w.nextProjectile), Owner: p.ID,
		X: p.X + (PlayerSize-parameters.ShotSize)/2, Y: p.Y + (PlayerSize-parameters.ShotSize)/2,
		Direction: Direction{X: input.AimX / length, Y: input.AimY / length}, Remaining: parameters.ShotLifetime,
	})
}

func (w *World) advanceProjectiles(actors []eater) {
	remaining := w.projectiles[:0]
	for _, shot := range w.projectiles {
		shot.Remaining -= StepSeconds
		if shot.Remaining <= 0 {
			continue
		}
		alive := true
		steps := int(math.Ceil(parameters.ShotSpeed * StepSeconds / 4))
		for step := range steps {
			start := position{shot.X, shot.Y}
			end := position{shot.X + shot.Direction.X*parameters.ShotSpeed*StepSeconds/float64(steps), shot.Y + shot.Direction.Y*parameters.ShotSpeed*StepSeconds/float64(steps)}
			if !terrain.Walkable(end.x, end.y, parameters.ShotSize, parameters.ShotSize) || terrain.PrivateGateCrossed(start.x, start.y, end.x, end.y, parameters.ShotSize) {
				alive = false
				break
			}
			winner, first := -1, math.Inf(1)
			for i, actor := range actors {
				if actor.id == shot.Owner || actor.dead() || actor.protectedAt(1, w.safeZones) {
					continue
				}
				if t, hit := eatingContact(start, end, parameters.ShotSize, position{actor.start.x + (actor.end.x-actor.start.x)*float64(step)/float64(steps), actor.start.y + (actor.end.y-actor.start.y)*float64(step)/float64(steps)}, position{actor.start.x + (actor.end.x-actor.start.x)*float64(step+1)/float64(steps), actor.start.y + (actor.end.y-actor.start.y)*float64(step+1)/float64(steps)}, actor.size); hit && (t < first || t == first && (winner < 0 || actor.id < actors[winner].id)) {
					winner, first = i, t
				}
			}
			shot.X, shot.Y = end.x, end.y
			if winner >= 0 {
				target := actors[winner]
				hp := target.stats.HP
				target.stats.HP = max(0, hp-parameters.ShotDamage)
				if target.stats.HP == 0 {
					w.dropKillLoot(target, hp, target.ownerStats().MP)
					for _, actor := range actors {
						if actor.id == shot.Owner {
							actor.ownerStats().GainXP(float64(target.stats.Level) * parameters.XPFirstLevelCost)
							break
						}
					}
					target.die(shot.Owner, w.tick)
				}
				alive = false
				break
			}
		}
		if alive {
			remaining = append(remaining, shot)
		}
	}
	w.projectiles = remaining
}

func (w *World) dropKillLoot(target eater, hp, mp float64) {
	drops := 0
	for _, fruit := range w.fruits {
		if fruit.Source != "" {
			drops++
		}
	}
	room := MaxMassDrops - drops
	if room <= 0 {
		return
	}
	hp *= ActorHPFraction
	mp *= ActorMPFraction
	hpCount := min(room, int(math.Ceil(hp)))
	mpCount := min(max(0, room-hpCount), int(math.Ceil(mp/parameters.FruitMP)))
	if mp > 0 && mpCount == 0 && hpCount > 1 {
		hpCount--
		mpCount = 1
	}
	for i := 0; i < hpCount+mpCount; i++ {
		size, _ := naturalFoodType()
		angle := rand.Float64() * 2 * math.Pi
		x, y := terrain.Move(target.end.x+(target.size-size)/2, target.end.y+(target.size-size)/2, math.Cos(angle)*40, math.Sin(angle)*40, size, size)
		if !terrain.Walkable(x, y, size, size) {
			x, y, size = target.end.x, target.end.y, FruitSize
		}
		food := Fruit{X: x, Y: y, Size: size, Source: target.id, Ejected: true, VelocityX: math.Cos(angle) * 180, VelocityY: math.Sin(angle) * 180, AvailableTick: w.tick + TickRate/3, ExpiresTick: w.tick + parameters.MassDropLifetimeTicks}
		if i < hpCount {
			food.HP = hp / float64(hpCount)
		} else {
			food.MP = mp / float64(mpCount)
		}
		w.nextFruit++
		food.ID = fmt.Sprintf("loot-%d", w.nextFruit)
		w.fruits = append(w.fruits, food)
	}
}

func (w *World) rangedTarget(ai *aiState) *playerState {
	var target *playerState
	best := AIDetectionRadius
	for _, p := range w.players {
		distance := math.Hypot(p.X-ai.X, p.Y-ai.Y)
		if p.Dead || distance > best || !ai.sees(p.X, p.Y) || InSafeZone(p.X, p.Y, PlayerSize, p.Stats.HP, w.safeZones) || terrain.PrivateGateCrossed(ai.X, ai.Y, p.X, p.Y, PlayerSize) {
			continue
		}
		if target == nil || distance < best || p.ID < target.ID {
			target, best = p, distance
		}
	}
	return target
}
