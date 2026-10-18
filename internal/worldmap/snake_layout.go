package worldmap

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"2d-dungeon/internal/parameters"
)

func assignForestSnakes(room *Room, grid [][]byte) {
	var spots []Cell
	for _, area := range room.FeedingAreas {
		if strings.HasPrefix(area.ID, "forest-feeding-") && grid[area.At.Y][area.At.X] != byte(CaveFloor) {
			spots = append(spots, area.At)
		}
	}
	if len(spots) == 0 {
		return
	}
	distance := func(a, b Cell) int { return (a.X-b.X)*(a.X-b.X) + (a.Y-b.Y)*(a.Y-b.Y) }
	for n := range parameters.ForestSnakeCount {
		home := spots[n%len(spots)]
		others := slices.DeleteFunc(slices.Clone(spots), func(c Cell) bool { return c == home })
		slices.SortFunc(others, func(a, b Cell) int { return distance(home, a) - distance(home, b) })
		route := append([]Cell{home}, others[:min(len(others), parameters.SnakeRouteStops-1)]...)
		if n/len(spots)%2 == 1 {
			slices.Reverse(route)
		}
		room.FeedingAreas = append(room.FeedingAreas, FeedingArea{ID: fmt.Sprintf("forest-snake-%d", n+1), At: route[0], Radius: parameters.ForestPatchRadius,
			Guard: &Guard{Type: ForestSnake, PatrolRadius: parameters.ForestPatchRadius, Level: parameters.SnakeLevel,
				Respawn: parameters.AIRespawnSeconds * time.Second, Route: route}})
	}
}
