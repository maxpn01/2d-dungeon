package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

type Projectile struct {
	mobOwned         bool
	Kind             string `json:"kind,omitempty"`
	originX, originY float64
	ID               string    `json:"id"`
	Owner            string    `json:"owner"`
	X                float64   `json:"x"`
	Y                float64   `json:"y"`
	Direction        Direction `json:"direction"`
	Remaining        float64   `json:"remaining"`
	Damage           float64   `json:"damage"`
	Size             float64   `json:"size"`
}

func ProjectileWallBody(shot Projectile) physics.Rect {
	size := shot.Size
	if shot.Kind == "shockwave" {
		size = min(size, AISize)
	}
	return physics.Square(shot.X+(shot.Size-size)/2, shot.Y+(shot.Size-size)/2, size)
}

func PrepareShot(p *Player, input Direction) bool {
	defer func() { p.ShotCooldown = max(0, p.ShotCooldown) }()
	if p.Combat.StunRemaining > 0 || p.Combat.StompRemaining > 0 || p.Dead || !worldmap.Default().WithSecrets(p.OpenSecrets).Walkable(p.X, p.Y, PlayerSize, PlayerSize) || worldmap.Default().OverlapsCover(physics.Square(p.X, p.Y, PlayerSize)) || !input.Valid() || !input.Shoot || p.ShotCooldown > 1e-9 || p.Stats.MP < parameters.ShotMPCost || math.Hypot(input.AimX, input.AimY) == 0 {
		return false
	}
	p.Stats.MP -= parameters.ShotMPCost
	p.ShotCooldown += 1 / p.Stats.AttackRate
	return true
}

func (w *World) fire(p *Player, input Direction) {
	defer func() { p.ShotCooldown = max(0, p.ShotCooldown) }()
	if len(w.projectiles) >= parameters.MaxProjectiles || InSafeZone(p.X, p.Y, PlayerSize, p.Stats.HP, w.safeZones) || !PrepareShot(p, input) {
		return
	}
	length := math.Hypot(input.AimX, input.AimY)
	direction := Direction{X: input.AimX / length, Y: input.AimY / length}
	if p.Stats.Mob {
		spread := (1 - parameters.MobAimAccuracy) * math.Pi / 2
		angle := math.Atan2(direction.Y, direction.X) + (rand.Float64()*2-1)*spread
		direction.X, direction.Y = math.Cos(angle), math.Sin(angle)
	}
	size := p.Stats.ShotSize()
	w.nextProjectile++
	w.projectiles = append(w.projectiles, Projectile{
		ID: fmt.Sprintf("shot-%d", w.nextProjectile), Owner: p.ID, originX: p.X, originY: p.Y, mobOwned: p.Stats.Mob,
		X: p.X + (PlayerSize-size)/2, Y: p.Y + (PlayerSize-size)/2,
		Direction: direction, Remaining: parameters.ShotLifetime, Damage: p.Stats.Damage, Size: size,
	})
}

func (w *World) advanceProjectiles(actors []eater) {
	remaining := w.projectiles[:0]
	for _, shot := range w.projectiles {
		if shot.Kind == "arrow" && (w.vaultArrowStarted == 0 || w.tick-w.vaultArrowStarted >= VaultArrowDurationTicks) {
			continue
		}
		shot.Remaining -= StepSeconds
		if shot.Remaining <= 0 {
			continue
		}
		alive := true
		speed := ProjectileSpeed(shot)
		steps := int(math.Ceil(speed * StepSeconds / 4))
		for step := range steps {
			start := position{shot.X, shot.Y}
			end := position{shot.X + shot.Direction.X*speed*StepSeconds/float64(steps), shot.Y + shot.Direction.Y*speed*StepSeconds/float64(steps)}
			if shot.Kind == "arrow" && !worldmap.Default().RoomContains(worldmap.OrbVault, physics.Square(end.x, end.y, shot.Size)) {
				alive = false
				break
			}
			body := ProjectileWallBody(shot)
			wallFraction, wallBlocked := worldmap.Default().WithSecrets(w.openSecrets).MovementLimit(body.X, body.Y, end.x-start.x, end.y-start.y, body.W, body.H)
			if worldmap.Default().PrivateGateCrossed(body.X, body.Y, body.X+end.x-start.x, body.Y+end.y-start.y, body.W) {
				alive = false
				break
			}
			winner, first := -1, math.Inf(1)
			for i, actor := range actors {
				if actor.id == shot.Owner || actor.dead() || actor.protectedAt(1, w.safeZones) || (shot.Kind == "arrow" || shot.Kind == "shockwave") && actor.player == nil {
					continue
				}
				if t, hit := eatingContact(start, end, shot.Size, position{actor.start.x + (actor.end.x-actor.start.x)*float64(step)/float64(steps), actor.start.y + (actor.end.y-actor.start.y)*float64(step)/float64(steps)}, position{actor.start.x + (actor.end.x-actor.start.x)*float64(step+1)/float64(steps), actor.start.y + (actor.end.y-actor.start.y)*float64(step+1)/float64(steps)}, actor.size); hit && (t < first || t == first && (winner < 0 || actor.id < actors[winner].id)) {
					winner, first = i, t
				}
			}
			if wallBlocked && (winner < 0 || wallFraction <= first) {
				if w.players[shot.Owner] != nil {
					w.revealSecret(start.x, start.y, end.x-start.x, end.y-start.y, shot.Size)
				}
				alive = false
				break
			}
			shot.X, shot.Y = end.x, end.y
			if winner >= 0 {
				target := actors[winner]
				if shot.mobOwned && target.ai != nil {
					alive = false
					break
				}
				hp := target.stats.HP
				target.stats.HP = max(0, hp-shot.Damage)
				if shot.Kind == "shockwave" {
					target.player.Combat = CombatState{Facing: target.player.Combat.Facing, StunRemaining: parameters.HeavyStompStun}
				}
				if target.ai != nil && target.stats.HP > 0 {
					w.alertAI(target.ai, shot)
				}
				if target.stats.HP == 0 {
					w.dropKillLoot(target, hp, target.ownerStats().MP)
					for _, actor := range actors {
						if actor.id == shot.Owner {
							actor.ownerStats().GainXP(float64(target.stats.Level) * parameters.KillXPPerLevel)
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
	if target.ai != nil && target.cell == nil {
		hp = target.ai.spawnMaxHP
	}
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
	hp = min(hp, float64(hpCount))
	mp = min(mp, float64(mpCount)*parameters.FruitMP)
	for i := 0; i < hpCount+mpCount; i++ {
		size, _ := naturalFoodType()
		angle := rand.Float64() * 2 * math.Pi
		x, y := worldmap.Default().Move(target.end.x+(target.size-size)/2, target.end.y+(target.size-size)/2, math.Cos(angle)*40, math.Sin(angle)*40, size, size)
		if !worldmap.Default().Walkable(x, y, size, size) {
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
		alerted := p.ID == ai.aggressorID && w.tick < ai.alertUntil
		if !w.aiCanSeePlayer(ai, p, AIDetectionRadius) {
			continue
		}
		if alerted {
			return p
		}

		if target == nil || distance < best || distance == best && p.ID < target.ID {
			target, best = p, distance
		}
	}
	return target
}
