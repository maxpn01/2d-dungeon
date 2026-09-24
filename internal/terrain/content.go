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

var Patches []FeedingPatch
var Dens []Den
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
	second := [2]int{ReturnCircleX / Tile, ReturnCircleY / Tile}
	seen[second] = true
	queue := [][2]int{start, second}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, d := range steps {
			n := [2]int{c[0] + d[0], c[1] + d[1]}
			if n[1] >= 0 || seen[n] {
				continue
			}
			x, y := CellPosition(n[0], n[1], ActorSize)
			if walkableLayout(x, y, ActorSize, ActorSize) {
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
