package simulation

import (
	"math"
	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

func (w *World) alertAI(ai *aiState, shot Projectile) {
	if w.players[shot.Owner] == nil {
		return
	}
	ai.aggressorID = shot.Owner
	ai.alertX, ai.alertY = shot.originX, shot.originY
	ai.alertUntil = w.tick + parameters.AIAlertSeconds*TickRate
	ai.patrolPause, ai.patrolTicks = 0, 0
	ai.returning = true
	ai.Behavior = "investigating"
}

func (w *World) aiCanSeePlayer(ai *aiState, p *playerState, radius float64) bool {
	geometry := worldmap.Default().WithSecrets(w.openSecrets)
	return !p.Dead && math.Hypot(p.X-ai.X, p.Y-ai.Y) <= radius &&
		!geometry.OverlapsCover(physics.Square(p.X, p.Y, PlayerSize)) &&
		geometry.Visible(ai.X+AISize/2, ai.Y+AISize/2, p.X+PlayerSize/2, p.Y+PlayerSize/2) &&
		!InSafeZone(p.X, p.Y, PlayerSize, p.Stats.HP, w.safeZones) &&
		!geometry.PrivateGateCrossed(ai.X, ai.Y, p.X, p.Y, PlayerSize)
}

func (w *World) aiWithinTerritory(ai *aiState) bool {
	geometry := worldmap.Default()
	body := physics.Square(ai.X, ai.Y, AISize)
	if room, ok := geometry.Room(ai.den.RoomID); ok && room.HasRole(worldmap.Private) {
		return room.Bounds().Contains(body)
	}
	if w.tick < ai.alertUntil || ai.returning {
		x, y := worldmap.CellPosition(ai.den.At.X, ai.den.At.Y, AISize)
		return ai.Y < 0 && geometry.FloorAt(ai.X+AISize/2, ai.Y+AISize/2) == geometry.FloorAt(x, y) && (ai.returning || math.Hypot(ai.X-x, ai.Y-y) <= parameters.AIChaseLeash)
	}
	return geometry.RoomContains(ai.den.RoomID, body)
}

func (w *World) investigateAI(ai *aiState) (Direction, bool) {
	geometry := worldmap.Default().WithSecrets(w.openSecrets)
	if ai.aggressorID != "" && w.tick < ai.alertUntil {
		ai.Behavior = "investigating"
		if math.Hypot(ai.X-ai.alertX, ai.Y-ai.alertY) > 24 {
			return ai.navigateMap(geometry, ai.alertX, ai.alertY), true
		}
		ai.Behavior = "searching"
		return Direction{}, true
	}
	ai.aggressorID = ""
	if ai.returning && ai.den.ID != "" {
		x, y := worldmap.CellPosition(ai.den.At.X, ai.den.At.Y, AISize)
		if math.Hypot(ai.X-x, ai.Y-y) > 24 {
			ai.Behavior = "returning"
			return ai.navigateMap(geometry, x, y), true
		}
		ai.returning = false
	}
	return Direction{}, false
}

func (ai *aiState) navigateMap(geometry *worldmap.Map, x, y float64) Direction {
	dx, dy := x-ai.X, y-ai.Y
	mx, my := geometry.Move(ai.X, ai.Y, dx, dy, AISize, AISize)
	if math.Hypot(mx-x, my-y) < .1 {
		return ai.approach(x, y)
	}
	nx, ny, ok := geometry.NextStep(ai.X, ai.Y, x, y, AISize)
	if !ok {
		return Direction{}
	}
	return ai.approach(nx, ny)
}
