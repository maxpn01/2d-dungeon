package simulation

import (
	"math"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/worldmap"
)

func NearVaultGate(x, y, size float64) bool {
	gate, ok := worldmap.Default().Gate(worldmap.VaultGate)
	return ok && x+size >= float64(gate.Left)-40 && x <= float64(gate.Right)+40 && math.Abs(y+size/2-float64(gate.Y)) <= 115
}

func (w *World) interactVaultGate(p *playerState) bool {
	if p.Dead || p.Flight.Flying || !NearVaultGate(p.X, p.Y, parameters.PlayerSize) || w.orb.GateActivated {
		return false
	}
	if w.orb.GateUnlocked {
		w.orb.GateActivated = true
	}
	return true
}
