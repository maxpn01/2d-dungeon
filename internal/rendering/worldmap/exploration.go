package worldmap

import (
	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	mapdata "raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const entityDetectionRadius = parameters.EntityDetectionRadius

type Exploration struct {
	world *mapdata.Map
	state *mapdata.Exploration
	focus rl.Vector2
}

func explorationCell(p rl.Vector2) [2]int {
	return mapdata.WorldCell(float64(p.X), float64(p.Y))
}

func (e *Exploration) SetSecrets(open []string) {
	e.world = mapdata.Default().WithSecrets(open)
	if e.state != nil {
		e.state.SetWorld(e.world)
	}
}

func (e *Exploration) Reveal(p rl.Vector2) {
	e.focus = p
	if e.state == nil {
		if e.world == nil {
			e.world = mapdata.Default().WithSecrets(nil)
		}
		e.state = e.world.NewExploration(parameters.ExplorationRadius)
	}
	e.state.Reveal(physics.Vector{X: float64(p.X), Y: float64(p.Y)})
}

func (e *Exploration) ActorVisible(x, y, size float64) bool {
	return e.state != nil && e.state.ActorVisible(physics.Square(x, y, size))
}

func (e *Exploration) Explored(cell [2]int) bool {
	return e.state != nil && e.state.Explored(cell)
}

func (e *Exploration) Visible(cell [2]int) bool {
	return e.state != nil && e.state.Visible(cell)
}

func (e *Exploration) EntityVisible(x, y, size float64) bool {
	center := physics.Square(x, y, size).Center()
	return e.world != nil && !e.world.SecretHidden(center.X, center.Y) && center.DistanceSquared(physics.Vector{X: float64(e.focus.X), Y: float64(e.focus.Y)}) <= entityDetectionRadius*entityDetectionRadius
}

func (e *Exploration) PlayerVisible(x, y, size float64) bool {
	return !mapdata.Default().OverlapsCover(physics.Square(x, y, size)) && e.EntityVisible(x, y, size)
}
