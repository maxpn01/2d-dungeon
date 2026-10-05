package simulation

import (
	"fmt"
	"math"
	"sort"

	"raylib-2d-sandbox/internal/physics"
)

type position struct{ x, y float64 }

const foodGridSize = 128.0

type foodGridCell struct{ x, y int }

func gridCoordinate(v float64) int { return int(math.Floor(v / foodGridSize)) }

func indexEaters(eaters []eater) map[foodGridCell][]int {
	grid := make(map[foodGridCell][]int, len(eaters))
	for i, actor := range eaters {
		minX := gridCoordinate(min(actor.start.x, actor.end.x))
		maxX := gridCoordinate(max(actor.start.x, actor.end.x) + actor.size)
		minY := gridCoordinate(min(actor.start.y, actor.end.y))
		maxY := gridCoordinate(max(actor.start.y, actor.end.y) + actor.size)
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
	cell       *SplitCell
	id         string
	start, end position
	size       float64
	stats      *Stats
	player     *playerState
	ai         *aiState
}

func (w *World) eaters(previous map[string]position) []eater {
	capacity := len(w.players) + len(w.ais)
	for _, player := range w.players {
		capacity += len(player.Split)
	}
	for i := range w.ais {
		capacity += len(w.ais[i].Split)
	}
	eaters := make([]eater, 0, capacity)
	for id, player := range w.players {
		if player.Dead {
			continue
		}
		eaters = append(eaters, eater{id: id, start: previous[id], end: position{player.X, player.Y}, size: PlayerSize, stats: &player.Stats, player: player})
		for i := range player.Split {
			cell := &player.Split[i]
			if !cell.Active {
				continue
			}
			cellID := fmt.Sprintf("%s/split/%d", id, i)
			start, ok := previous[cellID]
			if !ok {
				start = position{cell.X, cell.Y}
			}
			eaters = append(eaters, eater{id: cellID, start: start, end: position{cell.X, cell.Y}, size: PlayerSize, stats: &cell.Stats, player: player, cell: cell})
		}
	}
	sort.Slice(eaters, func(i, j int) bool { return eaters[i].id < eaters[j].id })
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick != 0 {
			continue
		}
		eaters = append(eaters, eater{id: ai.ID, start: previous[ai.ID], end: position{ai.X, ai.Y}, size: AISize, stats: &ai.Stats, ai: ai})
		for j := range ai.Split {
			cell := &ai.Split[j]
			if !cell.Active {
				continue
			}
			id := fmt.Sprintf("%s/split/%d", ai.ID, j)
			start, ok := previous[id]
			if !ok {
				start = position{cell.X, cell.Y}
			}
			eaters = append(eaters, eater{id: id, start: start, end: position{cell.X, cell.Y}, size: AISize, stats: &cell.Stats, ai: ai, cell: cell})
		}
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
	if e.player != nil && e.player.Split.Count() > 0 {
		e.stats.HP = 0
		if e.player.Stats.HP > 0 || e.player.Split.Mass() > 0 {
			return
		}
	}
	if e.player == nil {
		if e.ai.respawnTick != 0 {
			return
		}
		if e.ai.Split.Count() > 0 {
			e.stats.HP = 0
			if e.ai.Stats.HP > 0 || e.ai.Split.Mass() > 0 {
				return
			}
		}
		e.ai.deaths++
		e.ai.respawnTick = tick + AIRespawnSeconds*TickRate
		if e.ai.den.ID != "" {
			e.ai.respawnTick = tick + uint64(math.Ceil(e.ai.den.Respawn.Seconds()*TickRate))
		}
		return
	}
	e.player.Dead = true
	e.player.Flight = FlightState{}
	e.player.KilledBy = killer
	e.player.direction = Direction{}
	e.player.commands = nil
	e.player.InputSeq = e.player.lastSequence
}

func (w *World) consumeFruits(eaters []eater) {
	var grid map[foodGridCell][]int
	var seen []int
	if len(eaters) >= 16 {
		grid = indexEaters(eaters)
		seen = make([]int, len(eaters))
	}
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
			if actor.ai != nil && actor.ai.den.ID != "" {
				return
			}
			if fruit.Ejected && actor.player != nil && actor.player.ID == fruit.Source && w.tick < fruit.OwnerPickupTick {
				return
			}
			if actor.dead() || actor.ai != nil && (actor.ai.Behavior == "sleeping" || actor.ai.Behavior == "emerging") {
				return
			}
			location := position{fruit.X, fruit.Y}
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
			xp := FruitXP * hp
			if fruit.Ejected || fruit.Source != "" {
				xp = 0
			}
			eaters[winner].gain(hp, xp, fruit.MP)
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

func (e eater) takeMP() float64 {
	owner := e.ownerStats()
	totalHP := owner.HP
	if e.player != nil {
		totalHP += e.player.Split.Mass()
	} else {
		totalHP += e.ai.Split.Mass()
	}
	amount := owner.MP * min(1, e.stats.HP/max(e.stats.HP, totalHP))
	owner.MP = max(0, owner.MP-amount)
	return amount * ActorMPFraction
}

func (e eater) gain(hp, xp float64, mana ...float64) {
	owner := e.ownerStats()
	for _, mp := range mana {
		owner.GainMP(mp)
	}
	if (e.player == nil || e.player.Split.Count() == 0) && (e.ai == nil || e.ai.Split.Count() == 0) {
		e.stats.Eat(hp, xp)
		return
	}
	var stats *Stats
	var split *SplitCells
	if e.player != nil {
		stats, split = &e.player.Stats, &e.player.Split
	} else {
		stats, split = &e.ai.Stats, &e.ai.Split
	}
	stats.GainXP(xp)
	for i := range *split {
		if (*split)[i].Active {
			(*split)[i].Stats.Level, (*split)[i].Stats.XP = stats.Level, stats.XP
			(*split)[i].Stats.rebuildPerks()
		}
	}
	room := max(0, stats.MaxHP()-stats.HP-split.Mass())
	e.stats.HP += min(hp, room)
}
