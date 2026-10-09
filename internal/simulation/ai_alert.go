package simulation

import (
	"math"
	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

func (w *World) alertAI(ai *aiState, shot Projectile) {
	if w.players[shot.Owner] == nil || ai.returning {
		return
	}
	ai.aggressorID = shot.Owner
	ai.alertX, ai.alertY = shot.originX, shot.originY
	ai.alertUntil = w.tick + parameters.AIAlertSeconds*TickRate
	ai.patrolPause, ai.patrolTicks = 0, 0
	ai.Behavior = "investigating"
}

func (w *World) aiCanSeePlayer(ai *aiState, p *playerState, radius float64) bool {
	geometry := worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen)
	return !p.Dead && math.Hypot(p.X-ai.X, p.Y-ai.Y) <= radius &&
		!geometry.OverlapsCover(physics.Square(p.X, p.Y, PlayerSize)) &&
		geometry.Visible(ai.X+AISize/2, ai.Y+AISize/2, p.X+PlayerSize/2, p.Y+PlayerSize/2) &&
		!InSafeZone(p.X, p.Y, PlayerSize, p.Stats.HP, w.safeZones) &&
		!geometry.PrivateGateCrossed(ai.X, ai.Y, p.X, p.Y, PlayerSize)
}

func (w *World) aiWithinTerritory(ai *aiState) bool {
	geometry := worldmap.Default()
	body := MobBody(ai.Type, ai.X, ai.Y)
	if w.orbTarget(ai) != nil {
		return ai.Y < 0 && geometry.FloorAt(ai.X+AISize/2, ai.Y+AISize/2) == worldmap.FirstFloor
	}
	if room, ok := geometry.Room(ai.den.RoomID); ok && room.HasRole(worldmap.Private) {
		return room.Bounds().Contains(body)
	}
	if w.tick < ai.alertUntil || ai.returning {
		x, y := worldmap.CellPosition(ai.den.At.X, ai.den.At.Y, AISize)
		return ai.Y < 0 && geometry.FloorAt(ai.X+AISize/2, ai.Y+AISize/2) == geometry.FloorAt(x, y) && ((ai.returning && ai.Behavior == "returning") || math.Hypot(ai.X-x, ai.Y-y) <= parameters.AIChaseLeash)
	}
	return geometry.RoomContains(ai.den.RoomID, body)
}

func (w *World) investigateAI(ai *aiState) (Direction, bool) {
	geometry := worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen)
	if !ai.returning && ai.aggressorID != "" && w.tick < ai.alertUntil {
		ai.Behavior = "investigating"
		if math.Hypot(ai.X-ai.alertX, ai.Y-ai.alertY) > 24 {
			return ai.navigateMap(geometry, ai.alertX, ai.alertY), true
		}
		ai.Behavior = "searching"
		return Direction{}, true
	}
	if ai.aggressorID != "" {
		ai.returning = ai.den.ID != ""
	}
	ai.aggressorID = ""
	ai.alertUntil = 0
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
	avoidCover := ai.Type == DungeonShooter && !geometry.OverlapsCover(physics.Square(ai.X, ai.Y, AISize))
	terrain := physics.World{CanOccupy: func(body physics.Rect) bool {
		return geometry.CanOccupy(body) && (!avoidCover || !geometry.OverlapsCover(body))
	}, MaxStep: worldmap.Tile / 5}
	clear := func(tx, ty float64) bool {
		_, blocked := terrain.MovementLimit(MobBody(ai.Type, ai.X, ai.Y), physics.Vector{X: tx - ai.X, Y: ty - ai.Y})
		return !blocked
	}
	if clear(x, y) {
		ai.navigationActive = false
		return ai.approachWaypoint(x, y)
	}
	goal := worldmap.WorldCell(x+AISize/2, y+AISize/2)
	if ai.navigationActive && ai.navigationMap == geometry && ai.navigationGoal == goal && math.Hypot(ai.navigationX-ai.X, ai.navigationY-ai.Y) > 1 && clear(ai.navigationX, ai.navigationY) {
		return ai.approachWaypoint(ai.navigationX, ai.navigationY)
	}
	ai.navigationActive = false
	nextStep := geometry.NextStep
	if avoidCover {
		nextStep = geometry.NextStepOutsideCover
	}
	body := MobBody(ai.Type, ai.X, ai.Y)
	offset := ai.X - body.X
	nx, ny, ok := nextStep(body.X, body.Y, x-offset, y-offset, body.W)
	nx, ny = nx+offset, ny+offset
	if !ok {
		return Direction{}
	}
	if !clear(nx, ny) {
		cell := worldmap.WorldCell(ai.X+AISize/2, ai.Y+AISize/2)
		nx, ny = worldmap.CellPosition(cell[0], cell[1], AISize)
		if !clear(nx, ny) || math.Hypot(nx-ai.X, ny-ai.Y) <= 1 {
			return Direction{}
		}
	}
	ai.navigationActive = true
	ai.navigationMap, ai.navigationGoal = geometry, goal
	ai.navigationX, ai.navigationY = nx, ny
	return ai.approachWaypoint(nx, ny)
}

func (ai *aiState) approachWaypoint(x, y float64) Direction {
	dx, dy := x-ai.X, y-ai.Y
	distance := math.Hypot(dx, dy)
	if distance <= .01 {
		return Direction{}
	}
	scale := min(1, distance/(ai.Stats.MovementSpeed()*SpeedScale*StepSeconds))
	return Direction{X: dx / distance * scale, Y: dy / distance * scale}
}
