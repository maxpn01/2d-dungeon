package worldmap

import (
	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
	mapdata "2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Exploration struct {
	world       *mapdata.Map
	state       *mapdata.Exploration
	focus       rl.Vector2
	hiddenAreas []physics.Rect
	fogLight    map[[2]int]float64
	fogFocus    rl.Vector2
	radius      float64
}

func explorationCell(p rl.Vector2) [2]int {
	return mapdata.WorldCell(float64(p.X), float64(p.Y))
}

func (e *Exploration) SetSecrets(open []string, gateOpen ...float64) {
	openGate := 0.0
	if len(gateOpen) > 0 {
		openGate = gateOpen[0]
	}
	world := mapdata.Default().WithSecrets(open).WithVaultGate(openGate)
	if e.world != world {
		e.hiddenAreas = nil
		for _, door := range world.SecretDoors() {
			if world.Cell(door.Area.X, door.Area.Y) == 0 {
				if room, ok := world.Room(door.RoomID); ok {
					e.hiddenAreas = append(e.hiddenAreas, room.Bounds())
				}
			}
		}
		if openGate < .95 {
			if room, ok := world.Room(mapdata.OrbVault); ok {
				e.hiddenAreas = append(e.hiddenAreas, room.Bounds())
			}
		}
	}
	e.world = world
	if e.state != nil {
		e.state.SetWorld(e.world)
	}
}

func (e *Exploration) Reveal(p rl.Vector2, hasWings, flying bool) {
	e.radius = parameters.ExplorationRadius
	if hasWings {
		e.radius = parameters.WingExplorationRadius
	}
	e.focus = p
	if e.state == nil {
		if e.world == nil {
			e.world = mapdata.Default().WithSecrets(nil)
		}
		e.state = e.world.NewExploration(parameters.ExplorationRadius)
	}
	e.state.SetVision(e.radius, flying, parameters.WingOverviewMargin)
	e.state.Reveal(physics.Vector{X: float64(p.X), Y: float64(p.Y)})
}

func (e *Exploration) ActorVisible(x, y, size float64) bool {
	for _, area := range e.hiddenAreas {
		if area.Overlaps(physics.Square(x, y, max(size, 1))) {
			return false
		}
	}
	return e.state != nil && e.state.ActorVisible(physics.Square(x, y, size))
}

func (e *Exploration) Explored(cell [2]int) bool {
	return e.state != nil && e.state.Explored(cell)
}

func (e *Exploration) Visible(cell [2]int) bool {
	return e.state != nil && e.state.Visible(cell)
}

func (e *Exploration) EntityVisible(x, y, size float64) bool {
	return e.ActorVisible(x, y, size)
}

func (e *Exploration) PlayerVisible(x, y, size float64) bool {
	return !mapdata.Default().OverlapsCover(physics.Square(x, y, size)) && e.EntityVisible(x, y, size)
}
