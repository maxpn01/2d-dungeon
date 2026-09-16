package simulation

import "math"

type position struct{ x, y float64 }

type eater struct {
	id         string
	start, end position
	size       float64
	stats      *Stats
}

func (w *World) consumeFruits(previous map[string]position, previousAIs []position) {
	eaters := make([]eater, 0, len(w.players)+len(w.ais))
	for id, player := range w.players {
		if player.Dead {
			continue
		}
		eaters = append(eaters, eater{id: id, start: previous[id], end: position{player.X, player.Y}, size: PlayerSize, stats: &player.Stats})
	}
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick != 0 {
			continue
		}
		eaters = append(eaters, eater{id: ai.ID, start: previousAIs[i], end: position{ai.X, ai.Y}, size: AISize, stats: &ai.Stats})
	}
	remaining := w.fruits[:0]
	for _, fruit := range w.fruits {
		winner := -1
		earliest := math.Inf(1)
		for i, actor := range eaters {
			contact, hit := fruitContact(actor.start.x, actor.start.y, actor.end.x, actor.end.y, actor.size, fruit)
			if hit && (contact < earliest || contact == earliest && (winner < 0 || actor.id < eaters[winner].id)) {
				winner, earliest = i, contact
			}
		}
		if winner < 0 {
			remaining = append(remaining, fruit)
		} else {
			eaters[winner].stats.GainXP(FruitNutrition)
		}
	}
	w.fruits = remaining
}

func fruitContact(x, y, endX, endY, size float64, fruit Fruit) (float64, bool) {
	enter, exit := 0.0, 1.0
	for _, axis := range [][4]float64{{x, endX - x, fruit.X - size, fruit.X + fruit.Size}, {y, endY - y, fruit.Y - size, fruit.Y + fruit.Size}} {
		start, delta, lo, hi := axis[0], axis[1], axis[2], axis[3]
		if delta == 0 {
			if start <= lo || start >= hi {
				return 0, false
			}
			continue
		}
		a, b := (lo-start)/delta, (hi-start)/delta
		if a > b {
			a, b = b, a
		}
		enter = max(enter, a)
		exit = min(exit, b)
		if enter >= exit {
			return 0, false
		}
	}
	return enter, enter < 1 && exit > 0
}
