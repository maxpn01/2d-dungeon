package simulation

import "math"

type position struct{ x, y float64 }

func (w *World) consumeFruits(previous map[string]position) {
	remaining := w.fruits[:0]
	for _, fruit := range w.fruits {
		var winner *playerState
		earliest := math.Inf(1)
		for id, player := range w.players {
			start := previous[id]
			contact, hit := fruitContact(start.x, start.y, player.X, player.Y, fruit)
			if hit && (contact < earliest || contact == earliest && (winner == nil || id < winner.ID)) {
				winner, earliest = player, contact
			}
		}
		if winner == nil {
			remaining = append(remaining, fruit)
		} else {
			winner.Stats.GainXP(FruitNutrition)
		}
	}
	w.fruits = remaining
}

func fruitContact(x, y, endX, endY float64, fruit Fruit) (float64, bool) {
	enter, exit := 0.0, 1.0
	for _, axis := range [][4]float64{{x, endX - x, fruit.X - PlayerSize, fruit.X + fruit.Size}, {y, endY - y, fruit.Y - PlayerSize, fruit.Y + fruit.Size}} {
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
