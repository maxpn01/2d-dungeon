package simulation

import (
	"2d-dungeon/internal/parameters"
)

const CircleSplitSpeedMultiplier = parameters.CircleSplitSpeedMultiplier

func SpeedForMass(base, hp float64) float64 {
	return base
}

func MovementMass(hp float64, cells SplitCells) float64 {
	largest, total := hp, hp
	for _, cell := range cells {
		if cell.Active {
			largest = max(largest, cell.Stats.HP)
			total += cell.Stats.HP
		}
	}
	return largest + parameters.SplitMassMovementFraction*(total-largest)
}

func PlayerMovementSpeed(character Character, base, hp float64, cells SplitCells) float64 {
	speed := SpeedForMass(base, MovementMass(hp, cells))
	if character == Circle && cells.Count() > 0 {
		speed *= CircleSplitSpeedMultiplier
	}
	return speed
}
