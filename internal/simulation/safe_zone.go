package simulation

import (
	"math"
	"math/rand/v2"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/terrain"
)

const SafeZoneMaxMass = parameters.SafeZoneMaxMass
const SafeZoneCount = parameters.SafeZoneCount

type SafeZone struct{ X, Y, Radius float64 }

func newSafeZones() []SafeZone {
	zones := make([]SafeZone, 0, SafeZoneCount)
	radius := BodySize(PlayerSize, SafeZoneMaxMass)/2 + parameters.SafeZoneRadiusPadding
	for attempts := 0; attempts < parameters.SafeZonePlacementAttempts && len(zones) < SafeZoneCount; attempts++ {
		margin := radius + parameters.SafeZoneEdgeMargin
		zone := SafeZone{X: MapMin + margin + rand.Float64()*(MapMax-MapMin-2*margin), Y: MapMin + margin + rand.Float64()*(MapMax-MapMin-2*margin), Radius: radius}
		clear := true
		for _, other := range zones {
			if math.Hypot(zone.X-other.X, zone.Y-other.Y) < parameters.SafeZoneMinimumSeparation {
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
	if mass <= 0 || mass > SafeZoneMaxMass {
		return false
	}
	for _, zone := range zones {
		if math.Hypot(x+size/2-zone.X, y+size/2-zone.Y)+BodySize(size, mass)/2 <= zone.Radius {
			return true
		}
	}
	return false
}

func (e eater) protectedAt(t float64, zones []SafeZone) bool {
	mass := e.stats.HP
	if e.player != nil {
		mass = e.player.Stats.HP + e.player.Split.Mass()
	} else if e.ai != nil {
		mass = e.ai.Stats.HP + e.ai.Split.Mass()
	}
	x, y := e.start.x+(e.end.x-e.start.x)*t, e.start.y+(e.end.y-e.start.y)*t
	return InSafeZone(x, y, e.size, mass, zones)
}

func repelBody(x, y, oldX, oldY, hp float64, zones []SafeZone) (float64, float64) {
	for _, zone := range zones {
		radius := zone.Radius + BodySize(PlayerSize, hp)/2 + 2
		sx, sy := oldX+PlayerSize/2-zone.X, oldY+PlayerSize/2-zone.Y
		dx, dy := x-oldX, y-oldY
		a := dx*dx + dy*dy
		t := -1.0
		if sx*sx+sy*sy < radius*radius {
			t = 0
		} else if a > 0 {
			b := 2 * (sx*dx + sy*dy)
			disc := b*b - 4*a*(sx*sx+sy*sy-radius*radius)
			if disc >= 0 {
				hit := (-b - math.Sqrt(disc)) / (2 * a)
				if hit >= 0 && hit <= 1 {
					t = hit
				}
			}
		}
		if t < 0 {
			continue
		}
		nx, ny := sx+dx*t, sy+dy*t
		length := math.Hypot(nx, ny)
		if length < 1e-6 {
			nx, ny, length = 1, 0, 1
		}
		targetX, targetY := zone.X+nx/length*(radius+1)-PlayerSize/2, zone.Y+ny/length*(radius+1)-PlayerSize/2
		x, y = terrain.Move(oldX, oldY, targetX-oldX, targetY-oldY, PlayerSize, PlayerSize)
	}
	return x, y
}

func RepelPlayer(p *Player, before Player, zones []SafeZone) {
	if p.Dead || p.Stats.HP+p.Split.Mass() <= SafeZoneMaxMass {
		return
	}
	p.X, p.Y = repelBody(p.X, p.Y, before.X, before.Y, p.Stats.HP, zones)
	for i := range p.Split {
		cell := &p.Split[i]
		if !cell.Active {
			continue
		}
		oldX, oldY := cell.X, cell.Y
		if before.Split[i].Active {
			oldX, oldY = before.Split[i].X, before.Split[i].Y
		}
		cell.X, cell.Y = repelBody(cell.X, cell.Y, oldX, oldY, cell.Stats.HP, zones)
	}
}
