package simulation

import (
	"math"

	"2d-dungeon/internal/physics"
	"2d-dungeon/internal/worldmap"
)

func (w *World) frenzyPosition(ai *aiState, carrier *playerState) physics.Vector {
	slot := 0
	for i := range w.ais {
		if &w.ais[i] == ai {
			slot = i
			break
		}
	}
	angle := float64(slot%8) * math.Pi / 4
	radius := 160.0 + float64(slot%3)*30
	geometry := worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen)
	for _, turn := range []float64{0, .35, -.35, .7, -.7} {
		point := physics.Vector{X: carrier.X + math.Cos(angle+turn)*radius, Y: carrier.Y + math.Sin(angle+turn)*radius}
		if geometry.CanOccupy(physics.Square(point.X, point.Y, AISize)) && geometry.FloorAt(point.X+AISize/2, point.Y+AISize/2) == worldmap.FirstFloor && !geometry.PrivateGateCrossed(carrier.X, carrier.Y, point.X, point.Y, AISize) {
			return point
		}
	}
	return physics.Vector{X: carrier.X, Y: carrier.Y}
}

func (w *World) clearMobShot(ai *aiState, target *playerState) bool {
	size := ai.Stats.ShotSize()
	body := physics.Square(ai.X+(AISize-size)/2, ai.Y+(AISize-size)/2, size)
	delta := physics.Vector{X: target.X - ai.X, Y: target.Y - ai.Y}
	for i := range w.ais {
		other := &w.ais[i]
		if other == ai || other.respawnTick != 0 || other.Stats.HP <= 0 {
			continue
		}
		if _, hit := body.CollisionFraction(delta, physics.Square(other.X, other.Y, AISize).Expanded(12), physics.Vector{}); hit {
			return false
		}
	}
	return true
}
