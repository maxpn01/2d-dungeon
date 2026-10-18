package worldmap

import (
	"slices"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
)

func (m *Map) SecretDoors() []SecretDoor { return append([]SecretDoor(nil), m.layout.SecretDoors...) }

func (m *Map) WithSecrets(open []string) *Map {
	var mask uint64
	for i, door := range m.layout.SecretDoors {
		if slices.Contains(open, door.ID) {
			mask |= 1 << i
		}
	}
	if cached, ok := m.secretVariants.Load(mask); ok {
		return cached.(*Map)
	}
	variant := m.variant()
	variant.floors = make(map[[2]int]int, len(m.floors))
	for cell, id := range m.floors {
		variant.floors[cell] = id
	}
	for i, door := range m.layout.SecretDoors {
		if mask&(1<<i) != 0 {
			continue
		}
		if room, ok := m.Room(door.RoomID); ok {
			for y := room.Area.Y; y < room.Area.Bottom(); y++ {
				for x := room.Area.X; x < room.Area.Right(); x++ {
					delete(variant.floors, [2]int{x, y})
				}
			}
		}
		for y := door.Area.Y; y < door.Area.Bottom(); y++ {
			for x := door.Area.X; x < door.Area.Right(); x++ {
				delete(variant.floors, [2]int{x, y})
			}
		}
	}
	variant.collisionWorld = physics.World{CanOccupy: variant.CanOccupy, MaxStep: parameters.Tile / 5}
	cached, _ := m.secretVariants.LoadOrStore(mask, variant)
	return cached.(*Map)
}

func (m *Map) OrbPosition() physics.Vector {
	_, ok := m.Room(OrbVault)
	if !ok {
		return physics.Vector{}
	}
	return m.VaultOrbCell().Center()
}

func (m *Map) SecretHidden(x, y float64) bool {
	for _, door := range m.layout.SecretDoors {
		if m.Cell(door.Area.X, door.Area.Y) != 0 {
			continue
		}
		room, _ := m.Room(door.RoomID)
		if room.Area.ContainsWorld(x, y) || door.Area.ContainsWorld(x, y) {
			return true
		}
	}
	return false
}

func (m *Map) variant() *Map {
	return &Map{
		dungeonLayout: m.dungeonLayout, layout: m.layout, floors: m.floors, flightCells: m.flightCells,
		cover: m.cover, solid: m.solid, passages: m.passages, pools: m.pools, patches: m.patches,
		dens: m.dens, gates: m.gates, portals: m.portals, statues: m.statues,
	}
}
