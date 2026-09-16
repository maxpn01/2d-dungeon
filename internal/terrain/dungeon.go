package terrain

import "math"

const Tile = 40
const North = -3800

type Room struct {
	X, Y, W, H int
	Name       string
}

var Rooms = []Room{
	{18, -28, 14, 18, "THE COLUMN HALL"},
	{3, -40, 11, 20, "WEST CLOISTER"},
	{36, -40, 11, 20, "THE BROKEN GALLERY"},
	{16, -58, 18, 18, "THE STILLWATER BRIDGE"},
	{2, -64, 11, 15, "HALL OF NICHES"},
	{37, -64, 11, 15, "EAST ROTUNDA"},
	{14, -84, 22, 19, "THE THRESHOLD HALL"},
	{20, -94, 10, 7, "THE DESCENT"},
}

var passages = []Room{
	{22, -10, 6, 11, ""},
	{12, -25, 8, 4, ""}, {30, -25, 8, 4, ""},
	{7, -51, 4, 13, ""}, {39, -51, 4, 13, ""},
	{10, -55, 8, 4, ""}, {32, -55, 9, 4, ""},
	{7, -72, 4, 10, ""}, {9, -72, 7, 4, ""},
	{39, -72, 4, 10, ""}, {34, -72, 7, 4, ""},
	{23, -67, 4, 11, ""}, {23, -89, 4, 7, ""},
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
	if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) || x < 10 || x+w > 1990 || y < North || y+h > 1990 {
		return false
	}
	if y >= 10 {
		return true
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
