package terrain

import "math"

const Tile = 40
const (
	MinCellX     = -20
	MaxCellX     = 100
	MinCellY     = -192
	SurfaceCells = 100
	SurfaceMin   = 10
	SurfaceMax   = SurfaceCells*Tile - 10
	North        = MinCellY * Tile
	ActorSize    = 30
)

const DescentCircleX = 1000
const DescentCircleY = -5580

type Room struct {
	X, Y, W, H int
	Name       string
	ID         string
	Tags       []string
	Props      []Prop
}

var Rooms = []Room{
	{X: 18, Y: -28, W: 14, H: 18, Name: "THE COLUMN HALL", ID: "the-column-hall", Tags: []string{"entrance", "refuge"}, Props: []Prop{{"cover", 19, -26, 3, 5}}},
	{X: 3, Y: -40, W: 11, H: 20, Name: "WEST CLOISTER", ID: "west-cloister", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 4, -38, 3, 5}}},
	{X: 36, Y: -40, W: 11, H: 20, Name: "THE BROKEN GALLERY", ID: "the-broken-gallery", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 37, -38, 3, 5}}},
	{X: 16, Y: -58, W: 18, H: 18, Name: "THE STILLWATER BRIDGE", ID: "the-stillwater-bridge", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 17, -56, 3, 5}}},
	{X: 2, Y: -64, W: 11, H: 15, Name: "HALL OF NICHES", ID: "hall-of-niches", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 3, -62, 3, 5}}},
	{X: 37, Y: -64, W: 11, H: 15, Name: "EAST ROTUNDA", ID: "east-rotunda", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 38, -62, 3, 5}}},
	{X: 14, Y: -84, W: 22, H: 19, Name: "THE THRESHOLD HALL", ID: "the-threshold-hall", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 15, -82, 3, 5}}},
	{X: 20, Y: -94, W: 10, H: 7, Name: "THE WATCH", ID: "the-watch", Tags: []string{"exploration"}},
	{X: -16, Y: -44, W: 15, H: 22, Name: "THE VEILED COURT", ID: "fern-garden", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", -15, -42, 3, 5}}},
	{X: 51, Y: -44, W: 15, H: 22, Name: "THE RESONANT GALLERY", ID: "root-gallery", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 52, -42, 3, 5}}},
	{X: -16, Y: -65, W: 15, H: 14, Name: "WEST GRANARY", ID: "west-granary", Tags: []string{"exploration"}},
	{X: 51, Y: -65, W: 15, H: 14, Name: "EAST GRANARY", ID: "east-granary", Tags: []string{"exploration"}},
	{X: -14, Y: -82, W: 11, H: 10, Name: "SEALED STORE", ID: "sealed-store", Tags: []string{"exploration"}},
	{X: 53, Y: -82, W: 11, H: 10, Name: "ASH VAULT", ID: "ash-vault", Tags: []string{"exploration"}},
	{X: 15, Y: -114, W: 20, H: 14, Name: "LONG SANCTUM", ID: "long-sanctum", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 16, -112, 3, 5}}},
	{X: 0, Y: -117, W: 10, H: 20, Name: "THE WHISPER WALK", ID: "moss-walk", Tags: []string{"exploration", "refuge"}, Props: []Prop{{"cover", 1, -115, 3, 5}}},
	{X: 40, Y: -117, W: 10, H: 20, Name: "RELIQUARY WALK", ID: "reliquary-walk", Tags: []string{"exploration"}},
	{X: 12, Y: -134, W: 26, H: 16, Name: "THE CROWN", ID: "the-crown-boss-arena", Tags: []string{"exploration"}},
	{X: 20, Y: -144, W: 10, H: 8, Name: "THE DESCENT", ID: "the-descent", Tags: []string{"descent"}},
	{X: 13, Y: -190, W: 24, H: 24, Name: "FLOOR II / THE FEAST", ID: "floor-two", Tags: []string{"floor-two"}},
}

func FloorTwo(x, y float64) bool { return y < -160*Tile }

const ReturnCircleX = 1000
const ReturnCircleY = -6760

var passages = []Room{
	{X: -2, Y: -28, W: 7, H: 4}, {X: -2, Y: -58, W: 7, H: 4},
	{X: 45, Y: -28, W: 8, H: 4}, {X: 45, Y: -58, W: 8, H: 4},
	{X: -11, Y: -53, W: 4, H: 11}, {X: 57, Y: -53, W: 4, H: 11},
	{X: -10, Y: -74, W: 4, H: 11}, {X: 57, Y: -74, W: 4, H: 11},
	{X: 23, Y: -102, W: 4, H: 10},
	{X: 8, Y: -105, W: 9, H: 4}, {X: 33, Y: -105, W: 9, H: 4},
	{X: 5, Y: -122, W: 10, H: 4}, {X: 5, Y: -122, W: 4, H: 7},
	{X: 36, Y: -122, W: 10, H: 4}, {X: 42, Y: -122, W: 4, H: 7},
	{X: 23, Y: -120, W: 4, H: 8}, {X: 23, Y: -138, W: 4, H: 6},
	{X: 22, Y: -10, W: 6, H: 11, Name: ""},
	{X: 12, Y: -25, W: 8, H: 4, Name: ""}, {X: 30, Y: -25, W: 8, H: 4, Name: ""},
	{X: 7, Y: -51, W: 4, H: 13, Name: ""}, {X: 39, Y: -51, W: 4, H: 13, Name: ""},
	{X: 10, Y: -55, W: 8, H: 4, Name: ""}, {X: 32, Y: -55, W: 9, H: 4, Name: ""},
	{X: 7, Y: -72, W: 4, H: 10, Name: ""}, {X: 9, Y: -72, W: 7, H: 4, Name: ""},
	{X: 39, Y: -72, W: 4, H: 10, Name: ""}, {X: 34, Y: -72, W: 7, H: 4, Name: ""},
	{X: 23, Y: -67, W: 4, H: 11, Name: ""}, {X: 23, Y: -89, W: 4, H: 7, Name: ""},
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
	for _, y := range []int{-24, -19, -14} {
		for _, x := range []int{20, 28} {
			for dy := 0; dy < 2; dy++ {
				for dx := 0; dx < 2; dx++ {
					delete(m, [2]int{x + dx, y + dy})
				}
			}
		}
	}
	for y := -55; y < -43; y++ {
		for x := 18; x < 32; x++ {
			if (x < 23 || x >= 27) && (y < -51 || y >= -47) {
				delete(m, [2]int{x, y})
			}
		}
	}
	for _, p := range [][2]int{{5, -36}, {5, -29}, {11, -36}, {11, -29}, {38, -36}, {43, -31}, {39, -25}, {4, -61}, {4, -55}, {10, -61}, {10, -55}, {39, -61}, {44, -61}, {39, -53}, {44, -53}, {17, -79}, {31, -79}, {17, -70}, {31, -70}} {
		delete(m, p)
		delete(m, [2]int{p[0] + 1, p[1]})
	}
	return m
}

func Cell(x, y int) int  { return floors[[2]int{x, y}] }
func Pool(x, y int) bool { return x >= 18 && x < 32 && y >= -55 && y < -43 && Cell(x, y) == 0 }

func Walkable(x, y, w, h float64) bool {
	if y < SurfaceMin && y+h > 0 {
		return false
	}
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
				if y < 10 && (cx < 22 || cx >= 28) {
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
