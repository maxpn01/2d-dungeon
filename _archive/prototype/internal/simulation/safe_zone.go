package simulation

import (
	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
)

const (
	SafeZoneMaxMass = parameters.SafeZoneMaxMass
	SafeZoneCount   = parameters.SafeZoneCount
)

type SafeZone struct{ X, Y, Radius float64 }

func newSafeZones() []SafeZone {
	zones := make([]SafeZone, 0, SafeZoneCount)
	radius := BodySize(PlayerSize, SafeZoneMaxMass)/2 + parameters.SafeZoneRadiusPadding
	for attempts := 0; attempts < parameters.SafeZonePlacementAttempts && len(zones) < SafeZoneCount; attempts++ {
		margin := radius + parameters.SafeZoneEdgeMargin
		zone := SafeZone{X: MapMin + margin + rand.Float64()*(MapMax-MapMin-2*margin), Y: MapMin + margin + rand.Float64()*(MapMax-MapMin-2*margin), Radius: radius}
		clear := true
		for _, other := range zones {
			if (physics.Vector{X: zone.X, Y: zone.Y}).Distance(physics.Vector{X: other.X, Y: other.Y}) < parameters.SafeZoneMinimumSeparation {
				clear = false
				break
			}
		}
		if clear {
			zones = append(zones, zone)
		}
	}
	return zones
}

func InSafeZone(x, y, size, mass float64, zones []SafeZone) bool {
	if mass <= 0 {
		return false
	}
	body := physics.Circle{Center: physics.Square(x, y, size).Center(), Radius: BodySize(size, mass) / 2}
	for _, zone := range zones {
		if (physics.Circle{Center: physics.Vector{X: zone.X, Y: zone.Y}, Radius: zone.Radius}).Contains(body) {
			return true
		}
	}
	return false
}

func (e eater) protectedAt(t float64, zones []SafeZone) bool {
	x, y := e.start.x+(e.end.x-e.start.x)*t, e.start.y+(e.end.y-e.start.y)*t
	return InSafeZone(x, y, e.size, e.stats.HP, zones)
}
