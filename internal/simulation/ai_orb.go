package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

const OrbFrenzyDelayTicks = 3 * TickRate

func (w *World) orbTarget(ai *aiState) *playerState {
	if w.tick < w.orb.FrenzyAt {
		return nil
	}
	carrier := w.players[w.orb.Carrier]
	if carrier == nil || carrier.Dead || !carrier.HasOrb || carrier.Y >= 0 {
		return nil
	}
	geometry := worldmap.Default()
	room, ok := geometry.Room(ai.den.RoomID)
	if !ok || room.Floor != worldmap.FirstFloor || room.HasRole(worldmap.Private) || geometry.FloorAt(carrier.X+PlayerSize/2, carrier.Y+PlayerSize/2) != worldmap.FirstFloor {
		return nil
	}
	return carrier
}

func (w *World) huntOrb(ai *aiState, carrier *playerState) Direction {
	if ai.orbTargetID != carrier.ID {
		ai.orbTargetID = carrier.ID
		ai.orbPathUntil = 0
		ai.orbPathFound = false
		ai.patrolTicks, ai.patrolPause = 0, 0
	}
	ai.returning = true
	d := Direction{}
	if target := w.orbShotTarget(ai, carrier); target != nil {
		ai.shotTargetID = target.ID
		sx, sy := target.X-ai.X, target.Y-ai.Y
		length := math.Hypot(sx, sy)
		d.AimX, d.AimY, d.Shoot = sx/max(1, length), sy/max(1, length), true
	}
	ai.Behavior = "hunting orb"
	geometry := worldmap.Default().WithSecrets(w.openSecrets)
	targetX, targetY := carrier.X, carrier.Y
	if carrier.Flight.Flying && !geometry.CanOccupy(physics.Square(targetX, targetY, AISize)) {
		targetX, targetY = carrier.Flight.LandingX, carrier.Flight.LandingY
	} else {
		goal := w.frenzyPosition(ai, carrier)
		targetX, targetY = goal.X, goal.Y
	}
	distance := math.Hypot(targetX-ai.X, targetY-ai.Y)
	if distance < 12 {
		ai.Behavior = "covering orb"
		return d
	}
	mx, my := geometry.Move(ai.X, ai.Y, targetX-ai.X, targetY-ai.Y, AISize, AISize)
	if math.Hypot(mx-targetX, my-targetY) < .1 {
		movement := ai.approach(targetX, targetY)
		speed := min(1, distance/60)
		d.X, d.Y = movement.X*speed, movement.Y*speed
		return d
	}
	if w.tick >= ai.orbPathUntil || ai.orbPathFound && math.Hypot(ai.X-ai.orbPathX, ai.Y-ai.orbPathY) < 10 {
		ai.orbPathUntil = w.tick + TickRate/2
		nx, ny, found := geometry.NextStep(ai.X, ai.Y, targetX, targetY, AISize)
		ai.orbPathFound = found
		if !found {
			return d
		}
		ai.orbPathX, ai.orbPathY = nx, ny
	}
	if !ai.orbPathFound {
		return d
	}
	movement := ai.approach(ai.orbPathX, ai.orbPathY)
	d.X, d.Y = movement.X, movement.Y
	return d
}

func (ai *aiState) setFrenzy(active bool) {
	if ai.Frenzied == active && (!active || ai.Stats.Level == parameters.AIFrenzyLevel) {
		return
	}
	hpFraction, mpFraction := ai.Stats.HP/ai.Stats.MaxHP(), ai.Stats.MP/ai.Stats.MaxMP()
	if active {
		if !ai.Frenzied {
			ai.frenzyBase = ai.Stats
		}
		ai.Stats.Level, ai.Stats.XP = parameters.AIFrenzyLevel, 0
	} else {
		ai.Stats.Level, ai.Stats.XP = ai.frenzyBase.Level, ai.frenzyBase.XP
		ai.Combat.Boosting = false
	}
	ai.Stats.rebuildPerks()
	ai.Stats.HP, ai.Stats.MP = hpFraction*ai.Stats.MaxHP(), mpFraction*ai.Stats.MaxMP()
	ai.shotCooldown = min(ai.shotCooldown, 1/ai.Stats.AttackRate)
	ai.Frenzied = active
}

func (w *World) orbShotTarget(ai *aiState, carrier *playerState) *playerState {
	var target *playerState
	best := math.Inf(1)
	for _, p := range w.players {
		radius := AIDetectionRadius
		if p.ID == carrier.ID {
			radius = parameters.ShotSpeed * parameters.ShotLifetime
		}
		if !w.aiCanSeePlayer(ai, p, radius) {
			continue
		}
		if !w.clearMobShot(ai, p) {
			continue
		}
		distance := math.Hypot(p.X-ai.X, p.Y-ai.Y)
		if target == nil || distance < best || distance == best && p.ID < target.ID {
			target, best = p, distance
		}
	}
	return target
}
