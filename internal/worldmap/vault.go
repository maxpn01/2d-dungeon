package worldmap

import "raylib-2d-sandbox/internal/physics"

type VaultTrap struct {
	Tile Cell
	Lane physics.Rect
}

func (m *Map) VaultApproach() (Cell, Cell) {
	room, ok := m.Room(OrbVault)
	if !ok {
		return Cell{}, Cell{}
	}
	return vaultApproach(room, m.SecretDoors())
}

func vaultApproach(room Room, doors []SecretDoor) (Cell, Cell) {
	center := room.WorldCell(At(room.Area.Width/2, room.Area.Height/2))
	for _, door := range doors {
		if door.RoomID != OrbVault {
			continue
		}
		if door.Area.Width == 1 {
			x := room.Area.X
			startX := room.Area.X + 2
			if door.Area.X >= room.Area.Right() {
				x = room.Area.Right() - 1
				startX = room.Area.Right() - 3
			}
			entry := At(x, door.Area.Y+door.Area.Height/2)
			return At(startX, entry.Y), entry
		}
		y := room.Area.Y + 2
		entryY := room.Area.Y
		if door.Area.Y >= room.Area.Bottom() {
			y = room.Area.Bottom() - 3
			entryY = room.Area.Bottom() - 1
		}
		return At(center.X, y), At(door.Area.X+door.Area.Width/2, entryY)
	}
	return center, center
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
		x := float64(room.Area.X*Tile) + 12
		if side == 1 {
			x = float64(room.Area.Right()*Tile) - 12
		}
		for row := 3; row < room.Area.Height-2; row += 4 {
			ports = append(ports, physics.Vector{X: x, Y: float64((room.Area.Y+row)*Tile) + Tile/2})
		}
	}
	return ports
}

func (m *Map) VaultTraps() []VaultTrap {
	room, ok := m.Room(OrbVault)
	if !ok {
		return nil
	}
	start, _ := m.VaultApproach()
	center := m.VaultOrbCell()
	if start.X != center.X {
		cell := At(start.X+int(float64(center.X-start.X)*.6), center.Y)
		lane := physics.Rect{X: float64(cell.X*Tile - 8), Y: float64(room.Area.Y * Tile), W: Tile + 16, H: float64(room.Area.Height * Tile)}
		return []VaultTrap{{Tile: cell, Lane: lane}}
	}
	cell := At(center.X, start.Y+int(float64(center.Y-start.Y)*.6))
	lane := physics.Rect{X: float64(room.Area.X * Tile), Y: float64(cell.Y*Tile - 8), W: float64(room.Area.Width * Tile), H: Tile + 16}
	return []VaultTrap{{Tile: cell, Lane: lane}}
}
