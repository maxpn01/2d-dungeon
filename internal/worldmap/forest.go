package worldmap

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"2d-dungeon/internal/parameters"
)

//go:embed forest.txt
var forestMap string

type Terrain byte

const (
	NoTerrain Terrain = 0
	Grass     Terrain = '.'
	DirtPath  Terrain = ','
	Bush      Terrain = '"'
	Tree      Terrain = 'T'
	FruitTree Terrain = 'F'
	Water     Terrain = '~'
	Rock      Terrain = '#'
	CaveFloor Terrain = ':'
	Stump     Terrain = 'O'
	FallenLog Terrain = 'L'
	Ruin      Terrain = 'R'
	Menhir    Terrain = 'M'
)

func (t Terrain) Valid() bool {
	return strings.IndexByte(".,\"TF~#:OLRM", byte(t)) >= 0
}

func (t Terrain) Blocks() bool {
	switch t {
	case Tree, FruitTree, Water, Stump, FallenLog, Ruin, Menhir:
		return true
	}
	return false
}

func (t Terrain) StopsShots() bool { return t.Blocks() && t != Water }

const (
	ReturnMarker     = 'P'
	ForestExitMarker = 'X'
	StatueMarker     = 'S'
	FeedingMarker    = 'f'
)

func forest() Room {
	rows := strings.Split(strings.TrimRight(strings.ReplaceAll(forestMap, "\r", ""), "\n"), "\n")
	room := Room{ID: Forest, Name: "THE WILDWOOD", Floor: SecondFloor,
		Area: Cells(parameters.ForestOriginX, parameters.ForestOriginY, len(rows[0]), len(rows))}
	grid := make([][]byte, len(rows))
	for y, row := range rows {
		grid[y] = []byte(row)
	}
	terrainAt := func(x, y int) byte {
		if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[y]) {
			return byte(Tree)
		}
		return grid[y][x]
	}
	statues, feeding, fruitTrees := 0, 0, 0
	for y, row := range grid {
		for x, ch := range row {
			at := At(x, y)
			switch ch {
			case ReturnMarker:
				room.Portals = append(room.Portals, Portal{ID: ReturnPortal, At: at, Destination: DescentPortal})
			case ForestExitMarker:
				room.Portals = append(room.Portals, Portal{ID: ForestExitPortal, At: at, Destination: EmberArrivalPortal})
			case StatueMarker:
				statues++
				room.Statues = append(room.Statues, Statue{ID: fmt.Sprintf("forest-statue-%d", statues), At: at})
			case FeedingMarker:
				feeding++
				room.FeedingAreas = append(room.FeedingAreas, FeedingArea{ID: fmt.Sprintf("forest-feeding-%d", feeding), At: at,
					Radius: parameters.ForestPatchRadius, Style: VioletGarden,
					Food: FoodSource{Capacity: parameters.ForestPatchCapacity, Respawn: parameters.ForestPatchRespawnSecs * time.Second}})
			case byte(FruitTree):
				fruitTrees++
				room.FeedingAreas = append(room.FeedingAreas, FeedingArea{ID: fmt.Sprintf("fruit-tree-%d", fruitTrees), At: at,
					Radius: 1, Style: FruitTreeStyle,
					Food: FoodSource{Capacity: parameters.FruitTreeCapacity, Respawn: parameters.FruitTreeRespawnSecs * time.Second}})
			}
			if !Terrain(ch).Valid() {
				ground := Grass
				for _, step := range steps {
					if terrainAt(x+step[0], y+step[1]) == byte(CaveFloor) {
						ground = CaveFloor
					}
				}
				grid[y][x] = byte(ground)
			}
		}
	}
	assignForestSnakes(&room, grid)
	seen := map[[2]int]bool{}
	for y, row := range grid {
		for x, ch := range row {
			kind, ok := map[Terrain]PropKind{Stump: StumpProp, FallenLog: LogProp}[Terrain(ch)]
			if !ok || seen[[2]int{x, y}] {
				continue
			}
			area := Cells(x, y, 1, 1)
			for area.Right() < len(row) && row[area.Right()] == ch {
				area.Width++
			}
			for area.Bottom() < len(grid) && grid[area.Bottom()][x] == ch {
				area.Height++
			}
			for cy := area.Y; cy < area.Bottom(); cy++ {
				for cx := area.X; cx < area.Right(); cx++ {
					seen[[2]int{cx, cy}] = true
				}
			}
			room.Props = append(room.Props, Prop{ID: fmt.Sprintf("%s-%d-%d", kind, x, y), Kind: kind, Area: area})
		}
	}
	room.Terrain = make([]string, len(grid))
	for y, row := range grid {
		room.Terrain[y] = string(row)
	}
	return room
}
