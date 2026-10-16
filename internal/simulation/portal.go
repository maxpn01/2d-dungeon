package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/worldmap"
)

const PortalSeconds = parameters.PortalSeconds

func (w *World) updatePortals() {
	for _, p := range w.players {
		if p.Dead || p.Flight.Flying {
			p.TeleportCharge = 0
			continue
		}
		if p.portalCooldown > 0 {
			p.portalCooldown--
			p.TeleportCharge = 0
			continue
		}
		if p.TeleportCharge <= 0 {
			continue
		}
		portal, distance, found := nearestPortal(p.X+PlayerSize/2, p.Y+PlayerSize/2)
		if !found {
			p.TeleportCharge = 0
			p.portalCooldown = 0
			continue
		}
		if portal.ID == worldmap.DescentPortal && !w.orb.Consumed && !p.HasOrb {
			p.TeleportCharge = 0
			continue
		}
		destination, _ := worldmap.Default().Portal(portal.Destination)
		landing := destination.Position()
		toX, toY := landing.X, landing.Y
		if distance > 60 {
			p.TeleportCharge = 0
			continue
		}
		p.TeleportCharge += StepSeconds
		if p.TeleportCharge >= PortalSeconds {
			p.X, p.Y = toX-PlayerSize/2, toY-PlayerSize/2
			if portal.ID == worldmap.DescentPortal {
				w.orb.Consumed = true
				w.releaseOrb(w.orb.Carrier)
				p.Escaped = true
				if w.winnerID == "" {
					w.winnerID = p.ID
					w.finishedTick = w.tick
				}
			}
			p.Combat = CombatState{}
			p.TeleportCharge = 0
			p.portalCooldown = 2 * TickRate
			p.direction = Direction{}
		}
	}
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
