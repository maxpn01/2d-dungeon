package simulation

import (
	"math"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/physics"
)

type position struct{ x, y float64 }

const foodGridSize = 128.0

type foodGridCell struct{ x, y int }

func gridCoordinate(v float64) int { return int(math.Floor(v / foodGridSize)) }

func indexEaters(eaters []eater) map[foodGridCell][]int {
	grid := make(map[foodGridCell][]int, len(eaters))
	for i, actor := range eaters {
		reach := 0.0
		if actor.ai != nil && actor.ai.Type == CaveHydra {
			reach = parameters.HydraLungeReach * actor.ai.Hydra.Scale
		}
		minX := gridCoordinate(min(actor.start.x, actor.end.x) - reach)
		maxX := gridCoordinate(max(actor.start.x, actor.end.x) + actor.size + reach)
		minY := gridCoordinate(min(actor.start.y, actor.end.y) - reach)
		maxY := gridCoordinate(max(actor.start.y, actor.end.y) + actor.size + reach)
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				cell := foodGridCell{x, y}
				grid[cell] = append(grid[cell], i)
			}
		}
	}
	return grid
}

type eater struct {
	id         string
	start, end position
	size       float64
	stats      *Stats
	player     *playerState
	ai         *aiState
}

func (e eater) at(step, steps int) position {
	return position{
		e.start.x + (e.end.x-e.start.x)*float64(step)/float64(steps),
		e.start.y + (e.end.y-e.start.y)*float64(step)/float64(steps),
	}
}

func (w *World) eaters(previous map[string]position) []eater {
	eaters := make([]eater, 0, 1+len(w.ais))
	if player := w.player; !player.Dead {
		start := previous[player.ID]
		eaters = append(eaters, eater{id: player.ID, start: start, end: position{player.X, player.Y}, size: parameters.PlayerSize, stats: &player.Stats, player: player})
	}
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick != 0 {
			continue
		}
		ai.refreshParts()
		body := MobBody(ai.Type, ai.X, ai.Y)
		old := previous[ai.ID]
		start := MobBody(ai.Type, old.x, old.y)
		eaters = append(eaters, eater{id: ai.ID, start: position{start.X, start.Y}, end: position{body.X, body.Y}, size: body.W, stats: &ai.Stats, ai: ai})
	}
	return eaters
}

func (e eater) dead() bool {
	if e.player != nil {
		return e.player.Dead || e.stats.HP <= 0
	}
	return e.ai.respawnTick != 0 || e.stats.HP <= 0
}

func (e eater) die(killer string, tick uint64) {
	if e.player != nil && e.player.god {
		e.stats.HP = e.stats.MaxHP()
		return
	}
	status := e.status()
	*status = StatusEffects{Poison: Poison{Dealt: status.Poison.Dealt}, Sheds: status.Sheds}
	if e.player == nil {
		if e.ai.respawnTick != 0 {
			return
		}
		if e.ai.Type == CaveHydra {
			e.ai.respawnTick = tick
			return
		}
		e.ai.deaths++
		e.ai.respawnTick = tick + parameters.AIRespawnSeconds*parameters.TickRate
		if e.ai.den.ID != "" {
			e.ai.respawnTick = tick + uint64(math.Ceil(e.ai.den.Respawn.Seconds()*parameters.TickRate))
		}
		return
	}
	e.player.Dead = true
	e.player.Flight = FlightState{}
	e.player.KilledBy = killer
	e.player.direction = Direction{}
}

func (w *World) consumeFruits(eaters []eater) {
	var grid map[foodGridCell][]int
	var seen []int
	if len(eaters) >= 16 {
		grid = indexEaters(eaters)
		seen = make([]int, len(eaters))
	}
	snakesFeeding := w.snakesFeeding()
	remaining := w.fruits[:0]
	for fruitIndex, fruit := range w.fruits {
		if fruit.ExpiresTick != 0 && w.tick >= fruit.ExpiresTick {
			continue
		}
		if w.tick < fruit.AvailableTick {
			remaining = append(remaining, fruit)
			continue
		}
		winner := -1
		earliest := math.Inf(1)
		tryActor := func(i int) {
			actor := eaters[i]
			if actor.ai != nil && (actor.ai.Type != ForestSnake && actor.ai.Type != CaveHydra || !snakesFeeding) {
				return
			}
			if actor.dead() {
				return
			}
			location := position{fruit.X, fruit.Y}
			if actor.ai != nil && actor.ai.Type == CaveHydra {
				if !w.hydraCanEat(fruit) {
					return
				}
				food := physics.Square(fruit.X, fruit.Y, fruit.Size)
				for _, part := range actor.ai.parts {
					if part.head != hydraTailPart && part.body.Overlaps(food) && (winner < 0 || 0 < earliest) {
						winner, earliest = i, 0
						return
					}
				}
				return
			}
			contact, hit := eatingContact(actor.start, actor.end, actor.size, location, location, fruit.Size)
			if hit && (contact < earliest || contact == earliest && (winner < 0 || actor.id < eaters[winner].id)) {
				winner, earliest = i, contact
			}
		}
		if grid == nil {
			for i := range eaters {
				tryActor(i)
			}
		} else {
			minX, maxX := gridCoordinate(fruit.X), gridCoordinate(fruit.X+fruit.Size)
			minY, maxY := gridCoordinate(fruit.Y), gridCoordinate(fruit.Y+fruit.Size)
			for y := minY; y <= maxY; y++ {
				for x := minX; x <= maxX; x++ {
					for _, i := range grid[foodGridCell{x, y}] {
						if seen[i] == fruitIndex+1 {
							continue
						}
						seen[i] = fruitIndex + 1
						tryActor(i)
					}
				}
			}
		}
		if winner < 0 {
			remaining = append(remaining, fruit)
		} else {
			hp := fruit.rewardHP()
			xp := parameters.FruitXP * hp
			if fruit.Ejected || fruit.Source != "" {
				xp = 0
			}
			if ai := eaters[winner].ai; ai != nil && ai.Type == CaveHydra {
				eaters[winner].gain(hp, 0, fruit.MP)
				if fruit.Source == "" {
					w.hydraEat(ai, hp+fruit.MP/parameters.FruitMP*parameters.FruitHP)
				}
			} else if ai != nil {
				eaters[winner].gain(hp, 0, fruit.MP)
				ai.snakeEat(hp + fruit.MP/parameters.FruitMP*parameters.FruitHP)
			} else {
				eaters[winner].gain(hp, xp, fruit.MP)
			}
		}
	}
	w.fruits = remaining
}

func eatingContact(start, end position, size float64, targetStart, targetEnd position, targetSize float64) (float64, bool) {
	return physics.Square(start.x, start.y, size).CollisionFraction(
		physics.Vector{X: end.x - start.x, Y: end.y - start.y},
		physics.Square(targetStart.x, targetStart.y, targetSize),
		physics.Vector{X: targetEnd.x - targetStart.x, Y: targetEnd.y - targetStart.y},
	)
}

func (e eater) ownerStats() *Stats {
	if e.player != nil {
		return &e.player.Stats
	}
	return &e.ai.Stats
}

func (e eater) gain(hp, xp float64, mana ...float64) {
	for _, mp := range mana {
		e.ownerStats().GainMP(mp)
	}
	e.stats.Eat(hp, xp)
}
