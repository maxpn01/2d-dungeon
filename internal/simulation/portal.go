package simulation

import (
	"math"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/worldmap"
)

func (w *World) updatePortals() {
	p := w.player
	if p.Dead || p.Flight.Flying {
		p.TeleportCharge = 0
		return
	}
	if p.portalCooldown > 0 {
		p.portalCooldown--
		p.TeleportCharge = 0
		return
	}
	if p.TeleportCharge <= 0 {
		return
	}
	portal, distance, found := nearestPortal(p.X+parameters.PlayerSize/2, p.Y+parameters.PlayerSize/2)
	if !found {
		p.TeleportCharge = 0
		p.portalCooldown = 0
		return
	}
	if !w.portalActive(portal.ID, p) {
		p.TeleportCharge = 0
		return
	}
	destination, _ := worldmap.Default().Portal(portal.Destination)
	landing := destination.Position()
	toX, toY := landing.X, landing.Y
	if distance > 60 {
		p.TeleportCharge = 0
		return
	}
	p.TeleportCharge += parameters.StepSeconds
	if p.TeleportCharge >= parameters.PortalSeconds {
		p.X, p.Y = toX-parameters.PlayerSize/2, toY-parameters.PlayerSize/2
		if portal.ID == worldmap.DescentPortal && !p.FloorCleared {
			w.orb.Consumed = true
			w.releaseOrb(w.orb.Carrier)
			p.FloorCleared = true
			p.FloorClearedAt = w.tick
		}
		if destination.Floor == worldmap.ThirdFloor && w.winnerID == "" {
			w.winnerID = p.ID
			w.finishedTick = w.tick
		}
		p.Combat = CombatState{}
		p.TeleportCharge = 0
		p.portalCooldown = 2 * parameters.TickRate
		p.direction = Direction{}
	}
}

func (w *World) portalActive(id string, p *playerState) bool {
	switch id {
	case worldmap.DescentPortal:
		return w.orb.Consumed || p.HasOrb
	case worldmap.ForestExitPortal:
		return w.hydraDefeated()
	}
	return true
}

func nearestPortal(x, y float64) (worldmap.PortalPoint, float64, bool) {
	world := worldmap.Default()
	floor := world.FloorAt(x, y)
	nearest := worldmap.PortalPoint{}
	distance := math.Inf(1)
	for _, portal := range world.Portals() {
		if portal.Floor != floor {
			continue
		}
		position := portal.Position()
		candidate := math.Hypot(x-position.X, y-position.Y)
		if candidate < distance {
			nearest, distance = portal, candidate
		}
	}
	return nearest, distance, nearest.ID != ""
}
