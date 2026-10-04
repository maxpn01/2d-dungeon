package worldmap

func EntranceID(slot uint64) string {
	if slot == 1 {
		return NorthID(EntranceGate)
	}
	return EntranceGate
}

func (m *Map) CanAccessEntrance(id string, x, y, size float64) bool {
	if id == "" || m.FloorAt(x+size/2, y+size/2) != FirstFloor {
		return true
	}
	north, ok := m.Gate(NorthID(EntranceGate))
	if !ok {
		return true
	}
	south := m.EntranceGate()
	if id == EntranceGate {
		return y >= float64(north.Y)
	}
	if id == NorthID(EntranceGate) {
		return y+size <= float64(south.Y)
	}
	return false
}

func (m *Map) CellFloor(x, y int) Floor {
	for _, room := range m.layout.Rooms {
		if room.Floor != FirstFloor && Cells(room.Area.X-1, room.Area.Y-1, room.Area.Width+2, room.Area.Height+2).Contains(At(x, y)) {
			return room.Floor
		}
	}
	return FirstFloor
}

func (m *Map) SanctuaryCell(x, y int) bool {
	if x >= 0 && x < SurfaceCells && y >= 0 && y < SurfaceCells {
		return true
	}
	for _, room := range m.layout.Rooms {
		if room.ID == NorthSanctuary {
			return room.Area.Contains(At(x, y))
		}
	}
	return false
}
