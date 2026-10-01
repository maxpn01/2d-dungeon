package game

import (
	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const entityDetectionRadius = parameters.EntityDetectionRadius

type mapExploration struct {
	state *worldmap.Exploration
	focus rl.Vector2
}

func explorationCell(p rl.Vector2) [2]int {
	return worldmap.WorldCell(float64(p.X), float64(p.Y))
}

func (e *mapExploration) reveal(p rl.Vector2) {
	e.focus = p
	if e.state == nil {
		e.state = worldmap.Default().NewExploration(parameters.ExplorationRadius)
	}
	e.state.Reveal(physics.Vector{X: float64(p.X), Y: float64(p.Y)})
}

func (e *mapExploration) actorVisible(x, y, size float64) bool {
	return e.state != nil && e.state.ActorVisible(physics.Square(x, y, size))
}

func (e *mapExploration) explored(cell [2]int) bool {
	return e.state != nil && e.state.Explored(cell)
}

func (e *mapExploration) visible(cell [2]int) bool {
	return e.state != nil && e.state.Visible(cell)
}

func (e *mapExploration) entityVisible(x, y, size float64) bool {
	center := physics.Square(x, y, size).Center()
	return center.DistanceSquared(physics.Vector{X: float64(e.focus.X), Y: float64(e.focus.Y)}) <= entityDetectionRadius*entityDetectionRadius
}
