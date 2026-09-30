package terrain

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
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
	{X: 14, Y: -74, W: 8, H: 18, Name: "THE GATE HALL", ID: "gate-hall", Tags: []string{"entrance"}},
	{X: 11, Y: -82, W: 14, H: 8, Name: "THE COLUMN HALL", ID: "the-column-hall", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 12, -79, 3, 4}}},
	{X: -4, Y: -82, W: 11, H: 8, Name: "WEST CLOISTER", ID: "west-cloister", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", -3, -80, 3, 5}}},
	{X: 29, Y: -82, W: 11, H: 8, Name: "THE BROKEN GALLERY", ID: "the-broken-gallery", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 30, -80, 3, 5}}},
	{X: 9, Y: -100, W: 18, H: 18, Name: "THE STILLWATER BRIDGE", ID: "the-stillwater-bridge", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 10, -98, 3, 5}}},
	{X: -5, Y: -106, W: 11, H: 15, Name: "HALL OF NICHES", ID: "hall-of-niches", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", -4, -104, 3, 5}}},
	{X: 30, Y: -106, W: 11, H: 15, Name: "EAST ROTUNDA", ID: "east-rotunda", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 31, -104, 3, 5}}},
	{X: 7, Y: -126, W: 22, H: 19, Name: "THE THRESHOLD HALL", ID: "the-threshold-hall", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 8, -124, 3, 5}}},
	{X: 13, Y: -136, W: 10, H: 7, Name: "THE WATCH", ID: "the-watch", Tags: []string{"exploration"}},
	{X: -23, Y: -86, W: 15, H: 10, Name: "THE VEILED COURT", ID: "fern-garden", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", -22, -84, 3, 5}}},
	{X: 44, Y: -86, W: 15, H: 10, Name: "THE RESONANT GALLERY", ID: "root-gallery", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 45, -84, 3, 5}}},
	{X: -23, Y: -107, W: 15, H: 14, Name: "WEST GRANARY", ID: "west-granary", Tags: []string{"exploration"}},
	{X: 44, Y: -107, W: 15, H: 14, Name: "EAST GRANARY", ID: "east-granary", Tags: []string{"exploration"}},
	{X: -21, Y: -124, W: 11, H: 10, Name: "SEALED STORE", ID: "sealed-store", Tags: []string{"exploration"}},
	{X: 46, Y: -124, W: 11, H: 10, Name: "ASH VAULT", ID: "ash-vault", Tags: []string{"exploration"}},
	{X: 8, Y: -156, W: 20, H: 14, Name: "LONG SANCTUM", ID: "long-sanctum", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 9, -154, 3, 5}}},
	{X: -7, Y: -159, W: 10, H: 20, Name: "THE WHISPER WALK", ID: "moss-walk", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", -6, -157, 3, 5}}},
	{X: 33, Y: -159, W: 10, H: 20, Name: "RELIQUARY WALK", ID: "reliquary-walk", Tags: []string{"exploration"}},
	{X: 5, Y: -176, W: 26, H: 16, Name: "THE CROWN", ID: "the-crown-boss-arena", Tags: []string{"exploration"}},
	{X: 13, Y: -186, W: 10, H: 8, Name: "THE DESCENT", ID: "the-descent", Tags: []string{"descent"}},
	{X: 6, Y: -232, W: 24, H: 24, Name: "FLOOR II / THE FEAST", ID: "floor-two", Tags: []string{"floor-two"}},
}

func FloorTwo(x, y float64) bool { return y < -202*Tile }

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
	return x >= 2*Tile && x+size <= 34*Tile && y >= -56*Tile && y+size <= -24*Tile
}

var passages = []Room{
	{X: -9, Y: -78, W: 7, H: 4}, {X: -9, Y: -100, W: 7, H: 4},
	{X: 38, Y: -78, W: 8, H: 4}, {X: 38, Y: -100, W: 8, H: 4},
	{X: -18, Y: -95, W: 4, H: 11}, {X: 50, Y: -95, W: 4, H: 11},
	{X: -17, Y: -116, W: 4, H: 11}, {X: 50, Y: -116, W: 4, H: 11},
	{X: 16, Y: -144, W: 4, H: 10},
	{X: 1, Y: -147, W: 9, H: 4}, {X: 26, Y: -147, W: 9, H: 4},
	{X: -2, Y: -164, W: 10, H: 4}, {X: -2, Y: -164, W: 4, H: 7},
	{X: 29, Y: -164, W: 10, H: 4}, {X: 35, Y: -164, W: 4, H: 7},
	{X: 16, Y: -162, W: 4, H: 8}, {X: 16, Y: -180, W: 4, H: 6},
	{X: 15, Y: -24, W: 6, H: 25},
	{X: 16, Y: -82, W: 4, H: 7, Name: ""},
	{X: 0, Y: -93, W: 4, H: 13, Name: ""}, {X: 32, Y: -93, W: 4, H: 13, Name: ""},
	{X: 3, Y: -97, W: 8, H: 4, Name: ""}, {X: 25, Y: -97, W: 9, H: 4, Name: ""},
	{X: 0, Y: -114, W: 4, H: 10, Name: ""}, {X: 2, Y: -114, W: 7, H: 4, Name: ""},
	{X: 32, Y: -114, W: 4, H: 10, Name: ""}, {X: 27, Y: -114, W: 7, H: 4, Name: ""},
	{X: 16, Y: -109, W: 4, H: 11, Name: ""}, {X: 16, Y: -131, W: 4, H: 7, Name: ""},
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
	for y := -97; y < -85; y++ {
		for x := 11; x < 25; x++ {
			if (x < 16 || x >= 20) && (y < -93 || y >= -89) {
				delete(m, [2]int{x, y})
			}
		}
	}
	for _, p := range [][2]int{{-2, -78}, {-2, -71}, {4, -78}, {4, -71}, {31, -78}, {36, -73}, {32, -67}, {-3, -103}, {-3, -97}, {3, -103}, {3, -97}, {32, -103}, {37, -103}, {32, -95}, {37, -95}, {10, -121}, {24, -121}, {10, -112}, {24, -112}} {
		delete(m, p)
		delete(m, [2]int{p[0] + 1, p[1]})
	}
	return m
}

func Cell(x, y int) int  { return floors[[2]int{x, y}] }
func Pool(x, y int) bool { return x >= 11 && x < 25 && y >= -97 && y < -85 && Cell(x, y) == 0 }

func Walkable(x, y, w, h float64) bool {
	return walkableLayout(x, y, w, h)
}

func walkableLayout(x, y, w, h float64) bool {
	if w <= 0 || h <= 0 || math.IsNaN(w) || math.IsNaN(h) || math.IsInf(w, 0) || math.IsInf(h, 0) || math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) || x < MinCellX*Tile || x+w > MaxCellX*Tile || y < North || y+h > SurfaceMax {
		return false
	}
	if y >= SurfaceMin {
		return x >= SurfaceMin && x+w <= SurfaceMax
	}
	for cy := int(math.Floor(y / Tile)); cy <= int(math.Floor((y+h-.001)/Tile)); cy++ {
		for cx := int(math.Floor(x / Tile)); cx <= int(math.Floor((x+w-.001)/Tile)); cx++ {
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

func Move(x, y, dx, dy, w, h float64) (float64, float64) {
	values := x + y + dx + dy + w + h
	if y >= SurfaceMin && y+dy >= SurfaceMin && w > 0 && h > 0 && !math.IsNaN(values) && !math.IsInf(values, 0) {
		return max(SurfaceMin, min(SurfaceMax-w, x+dx)), max(SurfaceMin, min(SurfaceMax-h, y+dy))
	}
	steps := max(1, int(math.Ceil(max(math.Abs(dx), math.Abs(dy))/8)))
	for range steps {
		nx := x + dx/float64(steps)
		if Walkable(nx, y, w, h) {
			x = nx
		}
		ny := y + dy/float64(steps)
		if Walkable(x, ny, w, h) {
			y = ny
		}
	}
	return x, y
}
