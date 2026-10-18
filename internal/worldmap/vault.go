package worldmap

import (
	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/physics"
)

const VaultFireTrapCount = 3

type VaultTrap struct {
	Tile Cell
	Lane physics.Rect
}

func (m *Map) VaultApproach() (Cell, Cell) {
	room, ok := m.Room(OrbVault)
	if !ok {
		return Cell{}, Cell{}
	}
	return vaultApproach(room)
}

func vaultApproach(room Room) (Cell, Cell) {
	center := room.Area.X + room.Area.Width/2
	return At(center, room.Area.Bottom()-3), At(center, room.Area.Bottom()-1)
}

func vaultOrbCell(room Room) Cell {
	return room.WorldCell(At(room.Area.Width/2, room.Area.Height/2))
}

func (m *Map) VaultOrbCell() Cell {
	room, ok := m.Room(OrbVault)
	if !ok {
		return Cell{}
	}
	return vaultOrbCell(room)
}

func vaultSentinels(room Room, start Cell) []Cell {
	orb := vaultOrbCell(room)
	if start.X != orb.X {
		direction := 1
		if start.X > orb.X {
			direction = -1
		}
		return []Cell{At(orb.X+direction*3, orb.Y-4), At(orb.X+direction*5, orb.Y-7), At(orb.X+direction*3, orb.Y+4), At(orb.X+direction*5, orb.Y+7)}
	}
	direction := 1
	if start.Y > orb.Y {
		direction = -1
	}
	return []Cell{At(orb.X-4, orb.Y+direction*3), At(orb.X-7, orb.Y+direction*5), At(orb.X+4, orb.Y+direction*3), At(orb.X+7, orb.Y+direction*5)}
}

func (m *Map) VaultArrowPorts() []physics.Vector {
	room, ok := m.Room(OrbVault)
	if !ok {
		return nil
	}
	ports := make([]physics.Vector, 0, 12)
	for _, side := range []int{0, 1} {
		x := float64(room.Area.X*parameters.Tile) + 12
		if side == 1 {
			x = float64(room.Area.Right()*parameters.Tile) - 12
		}
		for row := 3; row < room.Area.Height-2; row += 4 {
			ports = append(ports, physics.Vector{X: x, Y: float64((room.Area.Y+row)*parameters.Tile) + parameters.Tile/2})
		}
	}
	return ports
}

func (m *Map) VaultTraps() []VaultTrap {
	passage, ok := m.Passage(VaultProcessional)
	if !ok {
		return nil
	}
	traps := make([]VaultTrap, 0, VaultFireTrapCount)
	for i, offset := range []int{8, 18, 28} {
		cell := At(passage.X+1+i, passage.Y+offset)
		lane := physics.Rect{X: float64(passage.X * parameters.Tile), Y: float64(cell.Y*parameters.Tile - 8), W: float64(passage.Width * parameters.Tile), H: parameters.Tile + 16}
		traps = append(traps, VaultTrap{Tile: cell, Lane: lane})
	}
	return traps
}

func (m *Map) VaultGateBarrier() physics.Rect {
	gate, ok := m.Gate(VaultGate)
	if !ok {
		return physics.Rect{}
	}
	return physics.Rect{X: float64(gate.Left), Y: float64(gate.Y - 20), W: float64(gate.Right - gate.Left), H: 40}
}

func (m *Map) WithVaultGate(open float64) *Map {
	if open >= .95 {
		return m
	}
	m.vaultGateOnce.Do(func() {
		variant := m.variant()
		variant.vaultGateClosed = true
		variant.bindCollision()
		m.closedVaultGate = variant
	})
	return m.closedVaultGate
}
