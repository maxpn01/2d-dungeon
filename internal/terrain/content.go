package terrain

import (
	"fmt"
)

type Prop struct {
	Kind       string
	X, Y, W, H int
}
type FeedingPatch struct {
	ID                              string
	X, Y, Radius, Capacity, Seconds int
}
type Den struct {
	ID                  string
	X, Y, Radius, Level int
	HP                  float64
}

var Patches = []FeedingPatch{
	{"entry", 24, -14, 2, 6, 5}, {"entry-east", 28, -20, 2, 5, 6},
	{"west-cloister", 8, -32, 2, 7, 9}, {"east-cloister", 41, -33, 2, 7, 9},
	{"garden", -8, -32, 3, 4, 12}, {"roots", 58, -32, 3, 4, 12},
	{"west-granary", -8, -58, 2, 9, 8}, {"east-granary", 58, -58, 2, 9, 8},
	{"niches", 7, -58, 2, 6, 10}, {"rotunda", 42, -58, 2, 6, 10},
	{"store", -8, -77, 2, 10, 16}, {"vault", 58, -77, 2, 10, 16},
	{"threshold", 25, -75, 3, 8, 10}, {"sanctum", 25, -107, 3, 10, 9},
	{"moss", 5, -110, 2, 4, 12}, {"reliquary", 45, -110, 2, 7, 12},
}
var Dens = []Den{
	{"cloister", 8, -33, 9, 1, 7}, {"gallery", 41, -34, 9, 1, 7},
	{"west-granary", -8, -59, 8, 2, 10}, {"east-granary", 58, -59, 8, 2, 10},
	{"niches", 7, -59, 7, 2, 10}, {"rotunda", 42, -59, 7, 2, 10},
	{"threshold", 25, -77, 9, 3, 13}, {"sanctum", 25, -108, 8, 4, 17},
}
var cover = map[[2]int]bool{}

func init() {
	for i := range Rooms {
		r := &Rooms[i]
		for _, d := range Dens {
			if d.X >= r.X && d.X < r.X+r.W && d.Y >= r.Y && d.Y < r.Y+r.H {
				r.Tags = append(r.Tags, "guarded")
			}
		}
		for _, p := range r.Props {
			if p.Kind != "cover" {
				panic("unknown terrain prop: " + p.Kind)
			}
			for y := p.Y; y < p.Y+p.H; y++ {
				for x := p.X; x < p.X+p.W; x++ {
					if Cell(x, y) != 0 {
						cover[[2]int{x, y}] = true
					}
				}
			}
		}
	}
	if err := Validate(); err != nil {
		panic(err)
	}
}
func Cover(x, y int) bool  { return cover[[2]int{x, y}] }
func Opaque(x, y int) bool { return (y < 0 && Cell(x, y) == 0) || Cover(x, y) }
func CellPosition(x, y int, size float64) (float64, float64) {
	return float64(x*Tile) + (Tile-size)/2, float64(y*Tile) + (Tile-size)/2
}

func Validate() error {
	start := [2]int{24, -1}
	seen := map[[2]int]bool{start: true}
	queue := [][2]int{start}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, d := range steps {
			n := [2]int{c[0] + d[0], c[1] + d[1]}
			if n[1] >= 0 || seen[n] {
				continue
			}
			x, y := CellPosition(n[0], n[1], ActorSize)
			if Walkable(x, y, ActorSize, ActorSize) {
				seen[n] = true
				queue = append(queue, n)
			}
		}
	}
	for c := range floors {
		if c[1] < 0 && !seen[c] {
			return fmt.Errorf("unreachable floor cell %v", c)
		}
	}
	for _, p := range Patches {
		if !seen[[2]int{p.X, p.Y}] {
			return fmt.Errorf("unreachable feeding patch %s", p.ID)
		}
	}
	for _, d := range Dens {
		if !seen[[2]int{d.X, d.Y}] {
			return fmt.Errorf("unreachable den %s", d.ID)
		}
	}
	if !seen[[2]int{DescentCircleX / Tile, DescentCircleY / Tile}] {
		return fmt.Errorf("unreachable descent")
	}
	return nil
}

func FruitCapacity() int {
	total := 0
	for _, p := range Patches {
		total += p.Capacity
	}
	return total
}
