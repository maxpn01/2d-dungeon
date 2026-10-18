package simulation

import (
	"math"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
	"2d-dungeon/internal/worldmap"
)

func (w *World) alertAI(ai *aiState, shot Projectile) {
	attacker := w.playerWithID(shot.Owner)
	if attacker == nil {
		return
	}
	if ai.patrol.returning {
		if ai.den.ID == "" || !withinLeash(ai, attacker.X, attacker.Y) {
			return
		}
		ai.patrol.returning = false
	}
	ai.alert.aggressorID = shot.Owner
	ai.alert.x, ai.alert.y = shot.originX, shot.originY
	ai.alert.until = w.tick + parameters.AIAlertSeconds*parameters.TickRate
	ai.patrol.pause, ai.patrol.ticks = 0, 0
	ai.Behavior = BehaviorInvestigating
}

func withinLeash(ai *aiState, x, y float64) bool {
	homeX, homeY := worldmap.CellPosition(ai.den.At.X, ai.den.At.Y, parameters.AISize)
	return math.Hypot(x-homeX, y-homeY) <= parameters.AIChaseLeash
}

func (w *World) aiCanSeePlayer(ai *aiState, p *playerState, radius float64) bool {
	geometry := w.geometry()
	return !p.Dead && math.Hypot(p.X-ai.X, p.Y-ai.Y) <= radius &&
		!geometry.OverlapsCover(physics.Square(p.X, p.Y, parameters.PlayerSize)) &&
		geometry.Visible(ai.X+parameters.AISize/2, ai.Y+parameters.AISize/2, p.X+parameters.PlayerSize/2, p.Y+parameters.PlayerSize/2) &&
		!geometry.PrivateGateCrossed(ai.X, ai.Y, p.X, p.Y, parameters.PlayerSize)
}

func (w *World) aiWithinTerritory(ai *aiState) bool {
	geometry := worldmap.Default()
	body := MobBody(ai.Type, ai.X, ai.Y)
	if w.orbTarget(ai) != nil {
		return ai.Y < 0 && geometry.BodyFloor(ai.X, ai.Y, parameters.AISize) == worldmap.OrbFloor
	}
	if room, ok := geometry.Room(ai.den.RoomID); ok && room.HasRole(worldmap.Private) {
		return room.Bounds().Contains(body)
	}
	if ai.Type == Sentinel || ai.Type == ForestSnake {
		return geometry.RoomContains(ai.den.RoomID, body)
	}
	if w.tick < ai.alert.until || ai.patrol.returning {
		x, y := worldmap.CellPosition(ai.den.At.X, ai.den.At.Y, parameters.AISize)
		sameFloor := geometry.BodyFloor(ai.X, ai.Y, parameters.AISize) == geometry.FloorAt(x, y)
		withinLeash := ai.patrol.returning && ai.Behavior == BehaviorReturning || math.Hypot(ai.X-x, ai.Y-y) <= parameters.AIChaseLeash
		return ai.Y < 0 && sameFloor && withinLeash
	}
	return geometry.RoomContains(ai.den.RoomID, body)
}

func (w *World) investigateAI(ai *aiState) (Direction, bool) {
	geometry := w.geometry()
	if !ai.patrol.returning && ai.alert.aggressorID != "" && w.tick < ai.alert.until {
		ai.Behavior = BehaviorInvestigating
		if math.Hypot(ai.X-ai.alert.x, ai.Y-ai.alert.y) > 24 {
			return ai.navigateMap(geometry, ai.alert.x, ai.alert.y), true
		}
		ai.Behavior = BehaviorSearching
		return Direction{}, true
	}
	if ai.alert.aggressorID != "" {
		ai.patrol.returning = ai.den.ID != ""
	}
	ai.alert.aggressorID = ""
	ai.alert.until = 0
	if ai.patrol.returning && ai.den.ID != "" {
		x, y := worldmap.CellPosition(ai.den.At.X, ai.den.At.Y, parameters.AISize)
		if math.Hypot(ai.X-x, ai.Y-y) > 24 {
			ai.Behavior = BehaviorReturning
			return ai.navigateMap(geometry, x, y), true
		}
		ai.patrol.returning = false
		ai.Stats.HP = ai.Stats.MaxHP()
	}
	return Direction{}, false
}

func (ai *aiState) navigateMap(geometry *worldmap.Map, x, y float64) Direction {
	avoidCover := ai.Type == DungeonShooter && !geometry.OverlapsCover(physics.Square(ai.X, ai.Y, parameters.AISize))
	terrain := physics.World{CanOccupy: func(body physics.Rect) bool {
		return geometry.CanOccupy(body) && (!avoidCover || !geometry.OverlapsCover(body))
	}, MaxStep: parameters.Tile / 5}
	clear := func(tx, ty float64) bool {
		_, blocked := terrain.MovementLimit(MobBody(ai.Type, ai.X, ai.Y), physics.Vector{X: tx - ai.X, Y: ty - ai.Y})
		return !blocked
	}
	if clear(x, y) {
		ai.navigation.active = false
		return ai.approachWaypoint(x, y)
	}
	goal := worldmap.WorldCell(x+parameters.AISize/2, y+parameters.AISize/2)
	if ai.navigation.active && ai.navigation.geometry == geometry && ai.navigation.goal == goal && math.Hypot(ai.navigation.x-ai.X, ai.navigation.y-ai.Y) > 1 && clear(ai.navigation.x, ai.navigation.y) {
		return ai.approachWaypoint(ai.navigation.x, ai.navigation.y)
	}
	ai.navigation.active = false
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
		cell := worldmap.WorldCell(ai.X+parameters.AISize/2, ai.Y+parameters.AISize/2)
		nx, ny = worldmap.CellPosition(cell[0], cell[1], parameters.AISize)
		if !clear(nx, ny) || math.Hypot(nx-ai.X, ny-ai.Y) <= 1 {
			return Direction{}
		}
	}
	ai.navigation.active = true
	ai.navigation.geometry, ai.navigation.goal = geometry, goal
	ai.navigation.x, ai.navigation.y = nx, ny
	return ai.approachWaypoint(nx, ny)
}

func (ai *aiState) approachWaypoint(x, y float64) Direction {
	dx, dy := x-ai.X, y-ai.Y
	distance := math.Hypot(dx, dy)
	if distance <= .01 {
		return Direction{}
	}
	scale := min(1, distance/(ai.Stats.MovementSpeed()*parameters.SpeedScale*parameters.StepSeconds))
	return Direction{X: dx / distance * scale, Y: dy / distance * scale}
}
