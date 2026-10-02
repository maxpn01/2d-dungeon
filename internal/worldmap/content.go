package worldmap

import (
	"fmt"
	"raylib-2d-sandbox/internal/physics"
)

func (m *Map) Cover(x, y int) bool { return m.cover[[2]int{x, y}] }

func (m *Map) Opaque(x, y int) bool { return (y < 0 && m.Cell(x, y) == 0) || m.Cover(x, y) }
func CellPosition(x, y int, size float64) (float64, float64) {
	return float64(x*Tile) + (Tile-size)/2, float64(y*Tile) + (Tile-size)/2
}

func (m *Map) Validate() error {
	seen := map[[2]int]bool{}
	queue := [][2]int{}
	add := func(cell [2]int) bool {
		if seen[cell] {
			return true
		}
		x, y := CellPosition(cell[0], cell[1], ActorSize)
		if !m.Walkable(x, y, ActorSize, ActorSize) {
			return false
		}
		seen[cell] = true
		queue = append(queue, cell)
		return true
	}
	for _, entry := range m.layout.Entrances {
		cell, err := m.layout.resolve(entry)
		if err != nil {
			return err
		}
		if !add(cell.key()) {
			return fmt.Errorf("blocked entrance %v", cell)
		}
	}
	for len(queue) > 0 {
		cell := queue[0]
		queue = queue[1:]
		for _, portal := range m.portals {
			if portal.At.key() == cell {
				destination, _ := m.Portal(portal.Destination)
				if !add(destination.At.key()) {
					return fmt.Errorf("blocked portal destination %s", destination.ID)
				}
			}
		}
		for _, step := range steps {
			next := [2]int{cell[0] + step[0], cell[1] + step[1]}
			if next[1] < 0 {
				add(next)
			}
		}
	}
	for cell := range m.floors {
		if cell[1] < 0 && !seen[cell] {
			return fmt.Errorf("unreachable floor cell %v", cell)
		}
	}
	for _, patch := range m.patches {
		if !seen[patch.At.key()] {
			return fmt.Errorf("unreachable feeding area %s", patch.ID)
		}
	}
	for _, den := range m.dens {
		if !seen[den.At.key()] {
			return fmt.Errorf("unreachable guard %s", den.ID)
		}
	}
	for _, statue := range m.statues {
		if !seen[statue.At.key()] {
			return fmt.Errorf("unreachable statue %s", statue.ID)
		}
	}
	for _, portal := range m.portals {
		if !seen[portal.At.key()] {
			return fmt.Errorf("unreachable portal %s", portal.ID)
		}
		position := portal.Position()
		if !m.CanOccupy(physics.Square(position.X-ActorSize/2, position.Y-ActorSize/2, ActorSize)) {
			return fmt.Errorf("blocked landing at portal %s", portal.ID)
		}
	}
	return nil
}

func (m *Map) FruitCapacity() int {
	total := 0
	for _, patch := range m.patches {
		total += patch.Capacity
	}
	return total
}
