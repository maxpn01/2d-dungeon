package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/worldmap"
)

func (w *World) spawnPosition() (float64, float64) {
	centerX, centerY := float64(worldmap.SpawnAltarCenter), float64(worldmap.SpawnAltarCenter)
	bestX, bestY, bestDistance := centerX-PlayerSize/2, centerY-80-PlayerSize/2, -1.0
	for offset := uint64(0); offset < 400; offset++ {
		cell := offset % 400
		index := int(cell % 40)
		spread := (index + 1) / 2
		if index%2 != 0 {
			spread = -spread
		}
		angle := -math.Pi/2 + float64(spread)*math.Pi/120
		radius := 80 + float64(cell/40)*8
		x := centerX + math.Cos(angle)*radius - PlayerSize/2
		y := centerY + math.Sin(angle)*radius - PlayerSize/2
		distance := math.Inf(1)
		for _, ai := range w.ais {
			if ai.respawnTick == 0 {
				distance = min(distance, math.Hypot(x-ai.X, y-ai.Y))
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
