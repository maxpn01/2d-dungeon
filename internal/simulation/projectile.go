package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/physics"
	"dungeon-raylib-go/internal/worldmap"
)

type ProjectileKind string

const (
	ArrowProjectile      ProjectileKind = "arrow"
	ShockwaveProjectile  ProjectileKind = "shockwave"
	SpectralProjectile   ProjectileKind = "spectral"
	VenomProjectile      ProjectileKind = "venom"
	HydraVenomProjectile ProjectileKind = "hydra-venom"
)

type Projectile struct {
	MobOwned         bool
	HitsMobs         bool
	Kind             ProjectileKind
	originX, originY float64
	ID               string
	Owner            string
	X                float64
	Y                float64
	Direction        Direction
	Lifetime         float64
	Remaining        float64
	Damage           float64
	StunDuration     float64
	PoisonDamage     float64
	PoisonSeconds    float64
	Pierces          int
	PierceDamage     float64
	HunterBonus      float64
	hit              []string
	Size             float64
}

func ProjectileWallBody(shot Projectile) physics.Rect {
	size := shot.Size
	if shot.Kind == ShockwaveProjectile {
		size = min(size, parameters.AISize)
	}
	return physics.Square(shot.X+(shot.Size-size)/2, shot.Y+(shot.Size-size)/2, size)
}

func PrepareShot(p *Player, input Direction) bool {
	defer func() { p.ShotCooldown = max(0, p.ShotCooldown) }()
	if p.Combat.StunRemaining > 0 || p.Combat.StompRemaining > 0 || p.Dead || !input.Valid() || !input.Shoot {
		return false
	}
	if p.ShotCooldown > 1e-9 || p.Stats.MP < p.Stats.ShotMPCost() || math.Hypot(input.AimX, input.AimY) == 0 {
		return false
	}
	geometry := p.geometry()
	if !geometry.ShotPassable(p.X, p.Y, parameters.PlayerSize, parameters.PlayerSize) || worldmap.Default().OverlapsCover(physics.Square(p.X, p.Y, parameters.PlayerSize)) {
		return false
	}
	p.Stats.MP -= p.Stats.ShotMPCost()
	p.ShotCooldown += 1 / p.Stats.AttackRate
	return true
}

func (w *World) fire(p *Player, input Direction) {
	defer func() { p.ShotCooldown = max(0, p.ShotCooldown) }()
	slots := 1
	if p.Stats.SpectralShots {
		slots = 2
	}
	if len(w.projectiles)+slots > parameters.MaxProjectiles || !PrepareShot(p, input) {
		return
	}
	p.ShotsFired++
	length := math.Hypot(input.AimX, input.AimY)
	direction := Direction{X: input.AimX / length, Y: input.AimY / length}
	if p.Stats.Mob {
		spread := (1 - parameters.MobAimAccuracy) * math.Pi / 2
		angle := math.Atan2(direction.Y, direction.X) + (rand.Float64()*2-1)*spread
		direction.X, direction.Y = math.Cos(angle), math.Sin(angle)
	}
	size := p.Stats.ShotSize()
	w.nextProjectile++
	w.projectiles = append(w.projectiles, p.Stats.perkShot(Projectile{
		ID: fmt.Sprintf("shot-%d", w.nextProjectile), Owner: p.ID, originX: p.X, originY: p.Y, MobOwned: p.Stats.Mob,
		X: p.X + (parameters.PlayerSize-size)/2, Y: p.Y + (parameters.PlayerSize-size)/2,
		Direction: direction, Remaining: parameters.ShotLifetime, Damage: p.Stats.Damage, Size: size,
	}))
	if p.Stats.SpectralShots && len(w.projectiles) < parameters.MaxProjectiles {
		w.nextProjectile++
		w.projectiles = append(w.projectiles, p.Stats.perkShot(Projectile{
			ID: fmt.Sprintf("echo-%d", w.nextProjectile), Owner: p.ID, Kind: SpectralProjectile, MobOwned: p.Stats.Mob,
			originX: p.X, originY: p.Y, X: p.X + (parameters.PlayerSize-size)/2 - direction.Y*16, Y: p.Y + (parameters.PlayerSize-size)/2 + direction.X*16,
			Direction: direction, Remaining: parameters.ShotLifetime, Damage: p.Stats.Damage * .6, Size: size,
		}))
	}
}

func (w *World) advanceProjectiles(actors []eater) {
	remaining := w.projectiles[:0]
	for _, shot := range w.projectiles {
		if shot.Kind == ArrowProjectile && (w.vaultArrowStarted == 0 || w.tick-w.vaultArrowStarted >= VaultArrowDurationTicks) {
			continue
		}
		shot.Remaining -= parameters.StepSeconds
		if shot.Remaining <= 0 {
			continue
		}
		alive := true
		speed := ProjectileSpeed(shot)
		steps := int(math.Ceil(speed * parameters.StepSeconds / 4))
		for step := range steps {
			start := position{shot.X, shot.Y}
			end := position{shot.X + shot.Direction.X*speed*parameters.StepSeconds/float64(steps), shot.Y + shot.Direction.Y*speed*parameters.StepSeconds/float64(steps)}
			if shot.Kind == ArrowProjectile && !worldmap.Default().RoomContains(worldmap.OrbVault, physics.Square(end.x, end.y, shot.Size)) {
				alive = false
				break
			}
			body := ProjectileWallBody(shot)
			wallFraction, wallBlocked := w.geometry().ShotLimit(body.X, body.Y, end.x-start.x, end.y-start.y, body.W, body.H)
			if worldmap.Default().PrivateGateCrossed(body.X, body.Y, body.X+end.x-start.x, body.Y+end.y-start.y, body.W) {
				alive = false
				break
			}
			limit := math.Inf(1)
			if wallBlocked {
				limit = wallFraction
			}
			if w.hitLibraryTome(shot, start, end, limit) {
				alive = false
				break
			}
			winner, first, winnerPart := -1, math.Inf(1), hydraBodyPart
			for i, actor := range actors {
				trapShot := shot.Kind == ArrowProjectile || shot.Kind == ShockwaveProjectile
				shieldPierced := shot.Kind == ShockwaveProjectile && actor.player != nil
				if actor.id == shot.Owner || actor.dead() || slices.Contains(shot.hit, actor.id) || actor.protected() && !shieldPierced || trapShot && actor.player == nil {
					continue
				}
				from := actor.at(step, steps)
				to := actor.at(step+1, steps)
				if t, hit := eatingContact(start, end, shot.Size, from, to, actor.size); hit && (t < first || t == first && (winner < 0 || actor.id < actors[winner].id)) {
					winner, first, winnerPart = i, t, hydraBodyPart
				}
				if actor.ai == nil {
					continue
				}
				for _, part := range actor.ai.parts {
					at := position{part.body.X, part.body.Y}
					if t, hit := eatingContact(start, end, shot.Size, at, at, part.body.W); hit && t < first {
						winner, first, winnerPart = i, t, part.head
					}
				}
			}
			if wallBlocked && (winner < 0 || wallFraction <= first) {
				if w.playerWithID(shot.Owner) != nil {
					w.revealSecret(start.x, start.y, end.x-start.x, end.y-start.y, shot.Size)
				}
				alive = false
				break
			}
			shot.X, shot.Y = end.x, end.y
			if winner >= 0 {
				target := actors[winner]
				if shot.MobOwned && target.ai != nil && !shot.HitsMobs {
					alive = false
					break
				}
				if shot.HitsMobs && target.ai != nil && target.ai.Type == ForestSnake {
					w.provokeSnake(target.ai, shot.Owner)
				}
				if shot.PoisonDamage > 0 {
					target.status().ApplyPoison(shot.PoisonDamage, shot.PoisonSeconds)
				}
				hp := target.stats.HP
				damage := shot.Damage
				if shot.HunterBonus > 0 && hp >= target.stats.MaxHP()*parameters.HunterHPThreshold {
					damage *= 1 + shot.HunterBonus
				}
				if !target.protected() {
					target.stats.HP = max(0, hp-damage)
					if target.ai != nil && target.ai.Type == CaveHydra {
						w.hydraPartHit(target.ai, winnerPart, damage)
					}
				}
				if shot.Kind == ShockwaveProjectile {
					target.player.Combat = CombatState{Facing: target.player.Combat.Facing, StunRemaining: max(target.player.Combat.StunRemaining, shot.StunDuration)}
				}
				if target.ai != nil && target.ai.Type != CaveHydra && target.stats.HP > 0 {
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
				if shot.Pierces > 0 && target.ai != nil && !wallBlocked {
					shot.Pierces--
					shot.hit = append(slices.Clip(shot.hit), target.id)
					shot.Damage *= shot.PierceDamage
					shot.PoisonDamage *= shot.PierceDamage
					continue
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
	if target.ai != nil {
		hp = target.ai.spawnMaxHP
	}
	drops := 0
	for _, fruit := range w.fruits {
		if fruit.Source != "" {
			drops++
		}
	}
	room := parameters.MaxMassDrops - drops
	if room <= 0 {
		return
	}
	hp *= parameters.ActorHPFraction
	mp *= parameters.ActorMPFraction
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
			x, y, size = target.end.x, target.end.y, parameters.FruitSize
		}
		food := Fruit{
			X: x, Y: y, Size: size, Source: target.id, Ejected: true,
			VelocityX: math.Cos(angle) * 180, VelocityY: math.Sin(angle) * 180,
			AvailableTick: w.tick + parameters.TickRate/3, ExpiresTick: w.tick + parameters.MassDropLifetimeTicks,
		}
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
	if !w.aiCanSeePlayer(ai, w.player, parameters.AIDetectionRadius) {
		return nil
	}
	return w.player
}
