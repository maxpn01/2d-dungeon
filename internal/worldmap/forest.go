package worldmap

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"dungeon-raylib-go/internal/parameters"
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
	Basalt    Terrain = 'B'
)

func (t Terrain) Valid() bool {
	return strings.IndexByte(".,\"TF~#:OLRMB", byte(t)) >= 0
}

func (t Terrain) Blocks() bool {
	switch t {
	case Tree, FruitTree, Water, Stump, FallenLog, Ruin, Menhir, Basalt:
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
	ShedSkinMarker   = 's'
	TracksMarker     = 'h'
)

const (
	HydraLair         = "cave-hydra"
	CaveFeedingPrefix = "cave-feeding-"
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
	statues, feeding, caveFeeding, fruitTrees := 0, 0, 0, 0
	for y, row := range grid {
		for x, ch := range row {
			at := At(x, y)
			if !Terrain(ch).Valid() {
				ground, paths := Grass, 0
				for _, step := range steps {
					switch terrainAt(x+step[0], y+step[1]) {
					case byte(CaveFloor):
						ground = CaveFloor
					case byte(DirtPath):
						paths++
					}
				}
				if ground != CaveFloor && paths >= 2 {
					ground = DirtPath
				}
				grid[y][x] = byte(ground)
			}
			switch ch {
			case ReturnMarker:
				room.Portals = append(room.Portals, Portal{ID: ReturnPortal, At: at, Destination: DescentPortal})
			case ForestExitMarker:
				room.Portals = append(room.Portals, Portal{ID: ForestExitPortal, At: at, Destination: EmberArrivalPortal})
				room.FeedingAreas = append(room.FeedingAreas, FeedingArea{ID: HydraLair, At: at, Radius: 1,
					Guard: &Guard{Type: CaveHydra, Level: parameters.HydraLevel, Respawn: parameters.AIRespawnSeconds * time.Second}})
			case StatueMarker:
				statues++
				room.Statues = append(room.Statues, Statue{ID: fmt.Sprintf("forest-statue-%d", statues), At: at})
			case FeedingMarker:
				id := fmt.Sprintf("forest-feeding-%d", feeding+1)
				if grid[y][x] == byte(CaveFloor) {
					caveFeeding++
					id = fmt.Sprintf("%s%d", CaveFeedingPrefix, caveFeeding)
				} else {
					feeding++
				}
				room.FeedingAreas = append(room.FeedingAreas, FeedingArea{ID: id, At: at,
					Radius: parameters.ForestPatchRadius, Style: VioletGarden,
					Food: FoodSource{Capacity: parameters.ForestPatchCapacity, Respawn: parameters.ForestPatchRespawnSecs * time.Second}})
			case ShedSkinMarker:
				room.Props = append(room.Props, Prop{ID: fmt.Sprintf("shed-skin-%d-%d", x, y), Kind: ShedSkinProp, Area: Cells(x, y, 1, 1)})
			case TracksMarker:
				room.Props = append(room.Props, Prop{ID: fmt.Sprintf("tracks-%d-%d", x, y), Kind: TracksProp, Area: Cells(x, y, 1, 1)})
			case byte(FruitTree):
				fruitTrees++
				room.FeedingAreas = append(room.FeedingAreas, FeedingArea{ID: fmt.Sprintf("fruit-tree-%d", fruitTrees), At: at,
					Radius: 1, Style: FruitTreeStyle,
					Food: FoodSource{Capacity: parameters.FruitTreeCapacity, Respawn: parameters.FruitTreeRespawnSecs * time.Second}})
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

func (m *Map) InCave(x, y float64) bool {
	cell := WorldCell(x, y)
	return m.Terrain(cell[0], cell[1]) == CaveFloor
}
