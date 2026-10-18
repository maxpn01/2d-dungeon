package worldmap

import (
	"fmt"
	"math"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/physics"
)

func (m *Map) Cover(x, y int) bool { return m.cover[[2]int{x, y}] }

func (m *Map) OverlapsCover(body physics.Rect) bool {
	if !body.Valid() {
		return false
	}
	first := WorldCell(body.X, body.Y)
	last := WorldCell(math.Nextafter(body.X+body.W, math.Inf(-1)), math.Nextafter(body.Y+body.H, math.Inf(-1)))
	for y := first[1]; y <= last[1]; y++ {
		for x := first[0]; x <= last[0]; x++ {
			if m.Cover(x, y) {
				return true
			}
		}
	}
	return false
}
func (m *Map) Opaque(x, y int) bool {
	if m.vaultGateClosed && Cells(x, y, 1, 1).Bounds().Overlaps(m.VaultGateBarrier()) {
		return true
	}
	return (!m.SanctuaryCell(x, y) && m.Cell(x, y) == 0) || m.Cover(x, y) || m.solid[[2]int{x, y}]
}

func (m *Map) fogOpaque(x, y int) bool {
	if m.vaultGateClosed && Cells(x, y, 1, 1).Bounds().Overlaps(m.VaultGateBarrier()) {
		return true
	}
	return m.Cell(x, y) == 0 && !m.SanctuaryCell(x, y) && !m.Pool(x, y)
}
func CellPosition(x, y int, size float64) (float64, float64) {
	return float64(x*parameters.Tile) + (parameters.Tile-size)/2, float64(y*parameters.Tile) + (parameters.Tile-size)/2
}

func (m *Map) Validate() error {
	seen := map[[2]int]bool{}
	queue := [][2]int{}
	add := func(cell [2]int) bool {
		if seen[cell] {
			return true
		}
		x, y := CellPosition(cell[0], cell[1], parameters.ActorSize)
		if !m.Walkable(x, y, parameters.ActorSize, parameters.ActorSize) {
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
			add(next)
		}
	}
	for cell := range m.floors {
		x, y := CellPosition(cell[0], cell[1], parameters.ActorSize)
		sanctuaryWall := m.SanctuaryCell(cell[0], cell[1]) && !m.Walkable(x, y, parameters.ActorSize, parameters.ActorSize)
		if cell[1] < 0 && !seen[cell] && !sanctuaryWall && !m.solid[cell] && !m.TerrainBlocks(cell[0], cell[1]) {
			return fmt.Errorf("unreachable floor cell %v", cell)
		}
	}
	for _, patch := range m.patches {
		if !reachableAround(seen, patch.At, patch.Radius) {
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
		if !m.CanOccupy(physics.Square(position.X-parameters.ActorSize/2, position.Y-parameters.ActorSize/2, parameters.ActorSize)) {
			return fmt.Errorf("blocked landing at portal %s", portal.ID)
		}
	}
	return nil
}

func reachableAround(seen map[[2]int]bool, at Cell, radius int) bool {
	area := Around(at, radius)
	for y := area.Y; y < area.Bottom(); y++ {
		for x := area.X; x < area.Right(); x++ {
			if seen[[2]int{x, y}] {
				return true
			}
		}
	}
	return false
}

func (m *Map) FruitCapacity() int {
	total := 0
	for _, patch := range m.patches {
		total += patch.Capacity
	}
	return total
}
