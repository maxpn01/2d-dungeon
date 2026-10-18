package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
)

type HydraHead struct {
	X, Y           float64
	HP, MaxHP      float64
	Severed        bool
	Regrow         float64
	Growing        float64
	SpitDelay      float64
	SpitRemaining  float64
	SpitDirection  Direction
	LungePhase     DashPhase
	LungeRemaining float64
	LungeDirection Direction
}

type bodyPart struct {
	body physics.Rect
	head int
}

const (
	hydraBodyPart = -1
	hydraTailPart = -2
)

func (h HydraHead) ready() bool {
	return !h.Severed && h.LungePhase == DashIdle && h.SpitRemaining <= 0 && h.SpitDelay <= 0
}

func (h Hydra) aliveHeads() int {
	count := 0
	for _, head := range h.Heads {
		if !head.Severed {
			count++
		}
	}
	return count
}

func (h Hydra) lunging() bool {
	for _, head := range h.Heads {
		if head.LungePhase == DashWinding || head.LungePhase == DashRushing {
			return true
		}
	}
	return false
}

func (h Hydra) spitting() bool {
	for _, head := range h.Heads {
		if head.SpitRemaining > 0 {
			return true
		}
	}
	return false
}

func (h Hydra) regrowPending() bool {
	for _, head := range h.Heads {
		if head.Severed && head.Growing == 0 && head.Regrow < parameters.HydraRegrowReserveSecs {
			return true
		}
	}
	return false
}

func HydraHeadBody(h Hydra, i int) physics.Rect {
	size := parameters.HydraHeadSize * h.Scale
	return physics.Square(h.Heads[i].X-size/2, h.Heads[i].Y-size/2, size)
}

func HydraTailBody(h Hydra, i, count int, segment physics.Vector) physics.Rect {
	size := parameters.HydraTailSize * h.Scale * (1 - .45*float64(i)/float64(max(1, count)))
	return physics.Square(segment.X+parameters.AISize/2-size/2, segment.Y+parameters.AISize/2-size/2, size)
}

func (h Hydra) restPoint(i int, center physics.Vector, tick uint64) (float64, float64) {
	angle := math.Atan2(h.Facing.Y, h.Facing.X)
	offset := float64(i - 1)
	length := parameters.HydraNeckLength * h.Scale
	t := float64(tick) * parameters.StepSeconds
	switch {
	case h.Phase == HydraSleeping && !h.Engaged:
		angle += offset * parameters.HydraHeadSpread * 1.7
		length *= .55 + .03*math.Sin(t*.8+offset)
	default:
		angle += offset*parameters.HydraHeadSpread + math.Sin(t*1.6+offset*2.1)*.1
		length += math.Sin(t*2.2+offset) * 3 * h.Scale
	}
	if h.Heads[i].Severed {
		length *= .4
	}
	return center.X + math.Cos(angle)*length, center.Y + math.Sin(angle)*length
}

func (ai *aiState) refreshParts() {
	ai.parts = ai.parts[:0]
	if ai.Type != CaveHydra {
		for _, segment := range ai.Segments {
			ai.parts = append(ai.parts, bodyPart{body: SnakeSegmentBody(segment), head: hydraBodyPart})
		}
		return
	}
	h := ai.Hydra
	center := HydraCenter(ai.AI)
	size := parameters.HydraBodySize * h.Scale
	ai.parts = append(ai.parts, bodyPart{body: physics.Square(center.X-size/2, center.Y-size/2, size), head: hydraBodyPart})
	for i, segment := range ai.Segments {
		ai.parts = append(ai.parts, bodyPart{body: HydraTailBody(h, i, len(ai.Segments), segment), head: hydraTailPart})
	}
	for i, head := range h.Heads {
		if !head.Severed {
			ai.parts = append(ai.parts, bodyPart{body: HydraHeadBody(h, i), head: i})
		}
	}
}

func (w *World) hydraAim(from physics.Vector, foe hydraTarget, seconds float64) Direction {
	x, y := foe.x, foe.y
	if foe.player != nil {
		intercept := meleeIntercept(foe.player, seconds)
		x, y = intercept.X+parameters.PlayerSize/2, intercept.Y+parameters.PlayerSize/2
	}
	dx, dy := x-from.X, y-from.Y
	length := max(1, math.Hypot(dx, dy))
	return Direction{X: dx / length, Y: dy / length}
}

func (w *World) hydraClearShot(ai *aiState, foe hydraTarget) bool {
	center := HydraCenter(ai.AI)
	size := parameters.HydraSpitSize
	_, blocked := w.geometry().ShotLimit(center.X-size/2, center.Y-size/2, foe.x-center.X, foe.y-center.Y, size, size)
	return !blocked
}

func (w *World) startHydraAttacks(ai *aiState, foe hydraTarget, distance float64) {
	h := &ai.Hydra
	center := HydraCenter(ai.AI)
	if w.tick >= ai.hydra.lungeReady && !h.lunging() && distance <= parameters.HydraLungeReach*h.Scale+foe.size/2-6 && w.hydraClearShot(ai, foe) {
		aim := w.hydraAim(center, foe, parameters.HydraLungeWindup)
		best, bestDot := -1, -2.0
		for i, head := range h.Heads {
			if !head.ready() {
				continue
			}
			dx, dy := head.X-center.X, head.Y-center.Y
			dot := (dx*aim.X + dy*aim.Y) / max(1, math.Hypot(dx, dy))
			if dot > bestDot {
				best, bestDot = i, dot
			}
		}
		if best >= 0 {
			head := &h.Heads[best]
			head.LungePhase, head.LungeRemaining, head.LungeDirection = DashWinding, parameters.HydraLungeWindup, aim
			busy := parameters.HydraLungeWindup + parameters.HydraLungeStrike + parameters.HydraLungeRecover
			ai.hydra.lungeReady = w.tick + uint64(math.Ceil(busy*parameters.TickRate)) + parameters.HydraLungeCooldownTicks
			return
		}
	}
	if w.tick < ai.hydra.volleyReady || distance < parameters.HydraSpitMinRange || distance > parameters.HydraSpitRange || !w.hydraClearShot(ai, foe) {
		return
	}
	reserve := 0.0
	if h.regrowPending() {
		reserve = parameters.HydraRegrowCost
	}
	count := 0
	for _, i := range rand.Perm(len(h.Heads)) {
		head := &h.Heads[i]
		if !head.ready() || ai.Stats.MP-float64(count+1)*parameters.HydraSpitCost < reserve {
			continue
		}
		head.SpitDelay = float64(count) * parameters.HydraSpitStagger
		head.SpitRemaining = parameters.HydraSpitWindup
		head.SpitDirection = w.hydraAim(physics.Vector{X: head.X, Y: head.Y}, foe, 0)
		count++
	}
	if count > 0 {
		busy := parameters.HydraSpitWindup + parameters.HydraSpitStagger*float64(count-1)
		ai.hydra.volleyReady = w.tick + uint64(math.Ceil(busy*parameters.TickRate)) + parameters.HydraVolleyCooldownTicks
	}
}

func (w *World) advanceHydraHeads(ai *aiState) {
	h := &ai.Hydra
	center := HydraCenter(ai.AI)
	foe, hasFoe := w.hydraTargetByID(h.FoeID)
	ease := 1 - math.Exp(-parameters.StepSeconds/.07)
	length := parameters.HydraNeckLength * h.Scale
	grazer := -1
	if ai.hydra.grazing {
		best := math.Inf(1)
		for i, head := range h.Heads {
			if distance := math.Hypot(head.X-ai.hydra.grazeX, head.Y-ai.hydra.grazeY); head.ready() && distance < best {
				grazer, best = i, distance
			}
		}
	}
	for i := range h.Heads {
		head := &h.Heads[i]
		tx, ty := h.restPoint(i, center, w.tick)
		snap := false
		if !head.Severed {
			w.advanceHydraSpit(ai, i, foe, hasFoe)
			switch head.LungePhase {
			case DashWinding:
				if head.LungeRemaining > parameters.HydraLungeAimLock && hasFoe {
					head.LungeDirection = w.hydraAim(center, foe, head.LungeRemaining)
				}
				d := head.LungeDirection
				tx, ty = center.X+d.X*length*.55, center.Y+d.Y*length*.55
				if head.LungeRemaining -= parameters.StepSeconds; head.LungeRemaining <= 1e-9 {
					head.LungePhase, head.LungeRemaining = DashRushing, parameters.HydraLungeStrike
					ai.hydra.lungeHit[i] = false
				}
			case DashRushing:
				p := 1 - head.LungeRemaining/parameters.HydraLungeStrike
				p = 1 - (1-p)*(1-p)
				reach := parameters.HydraLungeReach * h.Scale
				d := head.LungeDirection
				tx, ty = center.X+d.X*(length+(reach-length)*p), center.Y+d.Y*(length+(reach-length)*p)
				snap = true
				if head.LungeRemaining -= parameters.StepSeconds; head.LungeRemaining <= 1e-9 {
					head.LungePhase, head.LungeRemaining = DashRecovering, parameters.HydraLungeRecover
				}
			case DashRecovering:
				if head.LungeRemaining -= parameters.StepSeconds; head.LungeRemaining <= 1e-9 {
					head.LungePhase, head.LungeRemaining = DashIdle, 0
				}
			}
			if head.SpitRemaining > 0 && head.SpitDelay <= 0 {
				p := 1 - head.SpitRemaining/parameters.HydraSpitWindup
				d := head.SpitDirection
				tx, ty = center.X+d.X*length*(1-.3*p), center.Y+d.Y*length*(1-.3*p)
			}
			if i == grazer {
				tx, ty = ai.hydra.grazeX, ai.hydra.grazeY
			}
		}
		if snap {
			head.X, head.Y = tx, ty
			continue
		}
		head.X += (tx - head.X) * ease
		head.Y += (ty - head.Y) * ease
	}
}

func (w *World) advanceHydraSpit(ai *aiState, i int, foe hydraTarget, hasFoe bool) {
	head := &ai.Hydra.Heads[i]
	if head.SpitDelay > 0 {
		head.SpitDelay = max(0, head.SpitDelay-parameters.StepSeconds)
		return
	}
	if head.SpitRemaining <= 0 {
		return
	}
	before := head.SpitRemaining
	head.SpitRemaining -= parameters.StepSeconds
	if before > parameters.HydraSpitAimLock && hasFoe {
		from := physics.Vector{X: head.X, Y: head.Y}
		travel := math.Hypot(foe.x-head.X, foe.y-head.Y) / parameters.HydraSpitSpeed
		aim := w.hydraAim(from, foe, before+travel)
		angle := math.Atan2(aim.Y, aim.X) + float64(i-1)*parameters.HydraSpitSpread
		head.SpitDirection = Direction{X: math.Cos(angle), Y: math.Sin(angle)}
	}
	if head.SpitRemaining <= 1e-9 {
		head.SpitRemaining = 0
		w.hydraSpit(ai, i)
	}
}

func (w *World) hydraSpit(ai *aiState, i int) {
	if len(w.projectiles) >= parameters.MaxProjectiles || ai.Stats.MP < parameters.HydraSpitCost {
		return
	}
	head := ai.Hydra.Heads[i]
	ai.Stats.MP -= parameters.HydraSpitCost
	ai.ShotsFired++
	size := parameters.HydraSpitSize * min(1.4, ai.Hydra.Scale)
	x, y := head.X-size/2, head.Y-size/2
	lifetime := parameters.HydraSpitRange * 1.25 / parameters.HydraSpitSpeed
	w.nextProjectile++
	w.projectiles = append(w.projectiles, Projectile{
		Kind: HydraVenomProjectile, ID: fmt.Sprintf("hydra-venom-%d", w.nextProjectile), Owner: ai.ID, MobOwned: true, HitsMobs: true,
		X: x, Y: y, originX: x, originY: y, Direction: head.SpitDirection, Size: size,
		Remaining: lifetime, Lifetime: lifetime, Damage: ai.Stats.Damage * parameters.HydraSpitDamage,
		PoisonDamage: ai.Stats.Damage * parameters.PoisonDamageRatio * parameters.HydraSpitPoison, PoisonSeconds: parameters.HydraSpitPoisonSeconds,
	})
}

func (w *World) advanceHydraLunges(actors []eater) {
	for _, actor := range actors {
		ai := actor.ai
		if ai == nil || ai.Type != CaveHydra || actor.dead() {
			continue
		}
		for i, head := range ai.Hydra.Heads {
			if head.Severed || head.LungePhase != DashRushing || ai.hydra.lungeHit[i] {
				continue
			}
			bite := HydraHeadBody(ai.Hydra, i).Expanded(6)
			for _, target := range actors {
				if target.ai == ai || target.dead() || target.ai != nil && target.ai.Type != ForestSnake {
					continue
				}
				if !bite.Overlaps(physics.Square(target.end.x, target.end.y, target.size)) {
					continue
				}
				ai.hydra.lungeHit[i] = true
				if target.ai != nil {
					w.provokeSnake(target.ai, ai.ID)
				}
				w.mobStrike(ai, target, parameters.HydraLungeDamage)
				break
			}
		}
	}
}

func (w *World) hydraPartHit(ai *aiState, part int, damage float64) {
	if part < 0 || part >= len(ai.Hydra.Heads) {
		return
	}
	head := &ai.Hydra.Heads[part]
	if head.Severed {
		return
	}
	if head.HP -= damage; head.HP > 1e-9 {
		return
	}
	*head = HydraHead{X: head.X, Y: head.Y, MaxHP: head.MaxHP, Severed: true, Regrow: parameters.HydraRegrowSeconds}
	ai.Hydra.Severed++
}

func (w *World) advanceHydraRegrowth(ai *aiState) {
	for i := range ai.Hydra.Heads {
		head := &ai.Hydra.Heads[i]
		if !head.Severed {
			continue
		}
		if head.Growing > 0 {
			if head.Growing += parameters.StepSeconds / parameters.HydraRegrowGrowSeconds; head.Growing >= 1 {
				head.Severed, head.Growing, head.HP = false, 0, head.MaxHP
				ai.Hydra.Regrown++
			}
			continue
		}
		head.Regrow = max(0, head.Regrow-parameters.StepSeconds)
		if head.Regrow == 0 && ai.Stats.MP >= parameters.HydraRegrowCost {
			ai.Stats.MP -= parameters.HydraRegrowCost
			head.Growing = 1e-6
		}
	}
}
