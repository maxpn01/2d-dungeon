package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
)

const (
	MaxLevel        = parameters.MaxLevel
	StartingMP      = parameters.StartingMP
	MaxMP           = parameters.MaxMP
	ActorMPFraction = parameters.ActorMPFraction
	StartingHP      = parameters.StartingHP
	StartingMaxHP   = parameters.StartingMaxHP
	FruitXP         = parameters.FruitXP
	FruitHP         = parameters.FruitHP
	ActorHPFraction = parameters.ActorHPFraction
)

type Stats struct {
	Level             int
	XP, HP, MP, Speed float64
}

func NewStats() Stats { return Stats{Level: 1, HP: StartingHP, MP: StartingMP, Speed: PlayerSpeed} }

func XPForNextLevel(level int) float64 {
	if level >= MaxLevel {
		return 0
	}
	return parameters.XPFirstLevelCost + parameters.XPLevelCostIncrease*float64(level-1)
}

func (s *Stats) GainXP(amount float64) {
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) || s.Level >= MaxLevel {
		return
	}
	s.XP += amount
	for s.Level < MaxLevel && s.XP >= XPForNextLevel(s.Level) {
		s.XP -= XPForNextLevel(s.Level)
		s.Level++
	}
	if s.Level == MaxLevel {
		s.XP = 0
	}
}

func MaxHPForLevel(level int) float64 {
	level = max(1, min(MaxLevel, level))
	levelsGained := level - 1
	return StartingMaxHP + float64(parameters.HPPerLevelIncrease*levelsGained)
}

func (s *Stats) Eat(hp, xp float64) {
	s.GainXP(xp)
	if hp > 0 && !math.IsNaN(hp) && !math.IsInf(hp, 0) {
		s.HP = min(MaxHPForLevel(s.Level), s.HP+hp)
	}
}

func (s *Stats) GainMP(amount float64) {
	if amount > 0 && !math.IsNaN(amount) && !math.IsInf(amount, 0) {
		s.MP = min(MaxMP, s.MP+amount)
	}
}

func SpeedForMass(base, hp float64) float64 {
	return base
}

func (s Stats) MovementSpeed() float64 { return SpeedForMass(s.Speed, s.HP) }

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

const CircleSplitSpeedMultiplier = parameters.CircleSplitSpeedMultiplier

func PlayerMovementSpeed(character Character, base, hp float64, cells SplitCells) float64 {
	speed := SpeedForMass(base, MovementMass(hp, cells))
	if character == Circle && cells.Count() > 0 {
		speed *= CircleSplitSpeedMultiplier
	}
	return speed
}

func (p Player) MovementSpeed() float64 {
	return PlayerMovementSpeed(p.Character, p.Stats.Speed, p.Stats.HP, p.Split)
}
