package simulation

import (
	"math"
	"raylib-2d-sandbox/internal/worldmap"
)

func (w *World) spawnPosition(slot uint64) (float64, float64) {
	centerX, centerY := worldmap.SpawnCenter(slot)
	bestX, bestY, bestDistance := centerX-PlayerSize/2, centerY-80-PlayerSize/2, -1.0
	for offset := uint64(0); offset < 400; offset++ {
		cell := offset % 400
		index := int(cell % 40)
		spread := (index + 1) / 2
		if index%2 != 0 {
			spread = -spread
		}
		angle := -math.Pi/2 + float64(spread)*math.Pi/120
		if slot == 1 {
			angle = -angle
		}
		radius := 80 + float64(cell/40)*8
		x := centerX + math.Cos(angle)*radius - PlayerSize/2
		y := centerY + math.Sin(angle)*radius - PlayerSize/2
		distance := math.Inf(1)
		for _, ai := range w.ais {
			if ai.respawnTick == 0 {
				distance = min(distance, math.Hypot(x-ai.X, y-ai.Y))
				for _, cell := range ai.Split {
					if cell.Active {
						distance = min(distance, math.Hypot(x-cell.X, y-cell.Y))
					}
				}
			}
		}
		for _, player := range w.players {
			if !player.Dead {
				distance = min(distance, math.Hypot(x-player.X, y-player.Y))
			}
		}
		if distance > bestDistance {
			bestX, bestY, bestDistance = x, y, distance
		}
		if distance > AIDetectionRadius+PlayerSize {
			return x, y
		}
	}
	return bestX, bestY
}

func (ai *aiState) approach(x, y float64) Direction {
	ai.wanderTicks = 0
	dx, dy := x-ai.X, y-ai.Y
	if distance := math.Hypot(dx, dy); distance > 0 {
		return Direction{X: dx / distance, Y: dy / distance}
	}
	return Direction{}
}

func (ai *aiState) fleeFrom(threatX, threatY float64) Direction {
	dx, dy := ai.X-threatX, ai.Y-threatY
	length := math.Hypot(dx, dy)
	d := Direction{X: 1}
	if length > 0 {
		d = Direction{X: dx / length, Y: dy / length}
	}
	if ai.escapeTicks > 0 {
		d = ai.escapeDirection
		ai.escapeTicks--
	}
	nx, ny := Move(ai.X, ai.Y, d, .15, ai.Stats.MovementSpeed())
	blocked := math.Hypot(nx-ai.X, ny-ai.Y) < ai.Stats.MovementSpeed()*SpeedScale*.1
	if blocked {
		best := -1.0
		for _, candidate := range [8]Direction{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}, {X: 1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: 1}, {X: -1, Y: -1}} {
			x, y := Move(ai.X, ai.Y, candidate, StepSeconds, ai.Stats.MovementSpeed())
			if x == ai.X && y == ai.Y {
				continue
			}
			score := (x-threatX)*(x-threatX) + (y-threatY)*(y-threatY)
			if score > best {
				best = score
				d = candidate
			}
		}
		ai.escapeDirection, ai.escapeTicks = d, 9
	}
	ai.wanderTicks = 0
	return d
}

func (ai *aiState) sees(x, y float64) bool {
	return worldmap.Default().Visible(ai.X+AISize/2, ai.Y+AISize/2, x+AISize/2, y+AISize/2)
}
func (ai *aiState) inTerritory(x, y float64) bool {
	return x >= MapMin && y >= MapMin && x <= MaxPosition && y <= MaxPosition
}
func (ai *aiState) navigate(x, y float64) Direction {
	dx, dy := x-ai.X, y-ai.Y
	mx, my := worldmap.Default().Move(ai.X, ai.Y, dx, dy, AISize, AISize)
	if math.Hypot(mx-x, my-y) < .1 {
		return ai.approach(x, y)
	}
	nx, ny, ok := worldmap.Default().NextStep(ai.X, ai.Y, x, y, AISize)
	if !ok {
		return Direction{}
	}
	return ai.approach(nx, ny)
}
