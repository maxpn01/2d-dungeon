package terrain

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
)

const Tile = parameters.Tile
const (
	MinCellX     = parameters.MinCellX
	MaxCellX     = parameters.MaxCellX
	MinCellY     = parameters.MinCellY
	SurfaceCells = parameters.SurfaceCells
	SurfaceMin   = parameters.SurfaceMin
	SurfaceMax   = parameters.SurfaceMax
	North        = parameters.North
	ActorSize    = parameters.ActorSize
)

const SpawnAltarCenter = SurfaceCells * Tile / 2
const SpawnAltarRadius = 49

const DescentCircleX = parameters.DescentCircleX
const DescentCircleY = parameters.DescentCircleY

type Room struct {
	X, Y, W, H int
	Name       string
	ID         string
	Tags       []string
	Props      []Prop
}

var Rooms = []Room{
	{X: 2, Y: -56, W: 32, H: 32, Name: "THE PROVISION COURT", ID: "the-provision-hall", Tags: []string{"entrance", "private"}},
	{X: 14, Y: -64, W: 8, H: 8, Name: "THE GATE HALL", ID: "gate-hall", Tags: []string{"entrance"}},
	{X: 9, Y: -72, W: 18, H: 8, Name: "THE COLUMN HALL", ID: "the-column-hall", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 10, -70, 3, 4}, {"cover", 23, -70, 3, 4}}},
	{X: 0, Y: -92, W: 36, H: 20, Name: "THE STILLWATER COURT", ID: "the-stillwater-bridge", Tags: []string{"exploration"}},
	{X: 9, Y: -112, W: 18, H: 20, Name: "THE CROWN", ID: "the-crown-boss-arena", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 10, -102, 3, 4}, {"cover", 23, -102, 3, 4}}},
	{X: 12, Y: -128, W: 12, H: 12, Name: "THE TELEPORT SANCTUM", ID: "the-descent", Tags: []string{"descent"}},
	{X: 10, Y: -156, W: 16, H: 16, Name: "FLOOR II / THE FEAST", ID: "floor-two", Tags: []string{"floor-two"}},
}

func FloorTwo(x, y float64) bool { return y < -133*Tile }

const ReturnCircleX = parameters.ReturnCircleX
const ReturnCircleY = parameters.ReturnCircleY

const EntranceGateY = -56 * Tile
const EntranceGateLeft = 14 * Tile
const EntranceGateRight = 22 * Tile

func PrivateGateCrossed(fromX, fromY, toX, toY, size float64) bool {
	fromCenter, toCenter := fromY+size/2, toY+size/2
	if (fromCenter < EntranceGateY) == (toCenter < EntranceGateY) {
		return false
	}
	left, right := min(fromX, toX), max(fromX, toX)+size
	return left < EntranceGateRight && right > EntranceGateLeft
}

func InEntranceRoom(x, y, size float64) bool {
	return (physics.Rect{X: 2 * Tile, Y: -56 * Tile, W: 32 * Tile, H: 32 * Tile}).Contains(physics.Square(x, y, size))
}

var passages = []Room{
	{X: 15, Y: -24, W: 6, H: 25},
	{X: 16, Y: -116, W: 4, H: 4},
}

var Pools = []Room{
	{X: 11, Y: -87, W: 5, H: 4}, {X: 20, Y: -87, W: 5, H: 4},
	{X: 11, Y: -79, W: 5, H: 4}, {X: 20, Y: -79, W: 5, H: 4},
}

func Pool(x, y int) bool {
	for _, pool := range Pools {
		if x >= pool.X && x < pool.X+pool.W && y >= pool.Y && y < pool.Y+pool.H {
			return true
		}
	}
	return false
}

var floors = build()

func build() map[[2]int]int {
	m := make(map[[2]int]int)
	carve := func(r Room, id int) {
		for y := r.Y; y < r.Y+r.H; y++ {
			for x := r.X; x < r.X+r.W; x++ {
				m[[2]int{x, y}] = id
			}
		}
	}
	for _, r := range passages {
		carve(r, 1)
	}
	for i, r := range Rooms {
		carve(r, i+2)
	}
	for _, pool := range Pools {
		for y := pool.Y; y < pool.Y+pool.H; y++ {
			for x := pool.X; x < pool.X+pool.W; x++ {
				delete(m, [2]int{x, y})
			}
		}
	}
	return m
}

func Cell(x, y int) int { return floors[[2]int{x, y}] }
func Bounds() (left, top, right, bottom int) {
	left, right = 0, SurfaceCells
	for c := range floors {
		left, top = min(left, c[0]-1), min(top, c[1]-1)
		right = max(right, c[0]+2)
	}
	return left * Tile, top * Tile, right * Tile, SurfaceCells * Tile
}

func Walkable(x, y, w, h float64) bool {
	return walkableLayout(x, y, w, h)
}

func walkableLayout(x, y, w, h float64) bool {
	if !(physics.Rect{X: x, Y: y, W: w, H: h}).Valid() || x < MinCellX*Tile || x+w > MaxCellX*Tile || y < North || y+h > SurfaceMax {
		return false
	}
	if y >= SurfaceMin {
		return x >= SurfaceMin && x+w <= SurfaceMax && !overlapsSpawnAltar(x, y, w, h)
	}
	for cy := int(math.Floor(y / Tile)); cy <= int(math.Floor(math.Nextafter((y+h)/Tile, math.Inf(-1)))); cy++ {
		for cx := int(math.Floor(x / Tile)); cx <= int(math.Floor(math.Nextafter((x+w)/Tile, math.Inf(-1)))); cx++ {
			if cy >= 0 {
				if y < 10 && (cx < 15 || cx >= 21) {
					return false
				}
				continue
			}
			if Cell(cx, cy) == 0 {
				return false
			}
		}
	}
	return true
}

var collisionWorld = physics.World{
	CanOccupy: func(body physics.Rect) bool {
		return Walkable(body.X, body.Y, body.W, body.H)
	},
	MaxStep: Tile / 5,
}

func Move(x, y, dx, dy, w, h float64) (float64, float64) {
	motion := MoveBody(physics.Rect{X: x, Y: y, W: w, H: h}, physics.Vector{X: dx, Y: dy})
	return motion.Bounds.X, motion.Bounds.Y
}

func MoveBody(body physics.Rect, delta physics.Vector) physics.Motion {
	return collisionWorld.MoveAndSlide(body, delta)
}

func MovementLimit(x, y, dx, dy, w, h float64) (float64, bool) {
	return collisionWorld.MovementLimit(physics.Rect{X: x, Y: y, W: w, H: h}, physics.Vector{X: dx, Y: dy})
}

func overlapsSpawnAltar(x, y, w, h float64) bool {
	return (physics.Rect{X: x, Y: y, W: w, H: h}).OverlapsOctagon(
		physics.Vector{X: SpawnAltarCenter, Y: SpawnAltarCenter}, SpawnAltarRadius,
	)
}
