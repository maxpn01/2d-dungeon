package worldmap

func (m *Map) CellFloor(x, y int) Floor {
	for _, room := range m.layout.Rooms {
		if room.Floor != FirstFloor && Cells(room.Area.X-1, room.Area.Y-1, room.Area.Width+2, room.Area.Height+2).Contains(At(x, y)) {
			return room.Floor
		}
	}
	return FirstFloor
}

func (m *Map) SanctuaryCell(x, y int) bool {
	return x >= 0 && x < SurfaceCells && y >= 0 && y < SurfaceCells
}
