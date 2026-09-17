package simulation

import "math"

const (
	MaxLevel       = 100
	MaxHP          = 100.0
	HPPerLevel     = 0.25
	SpeedPerLevel  = 0.25 / SpeedScale
	FruitNutrition = 0.25
)

type Stats struct {
	Level         int
	XP, HP, Speed float64
}

func NewStats() Stats { return Stats{Level: 1, HP: 1, Speed: PlayerSpeed} }

func XPForNextLevel(level int) float64 {
	if level >= MaxLevel {
		return 0
	}
	return float64(level * level)
}

func (s *Stats) GainXP(amount float64) {
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) || s.Level >= MaxLevel {
		return
	}
	s.XP += amount
	for s.Level < MaxLevel && s.XP >= XPForNextLevel(s.Level) {
		s.XP -= XPForNextLevel(s.Level)
		s.Level++
		s.HP = min(MaxHP, s.HP+HPPerLevel)
		s.Speed += SpeedPerLevel
	}
	if s.Level == MaxLevel {
		s.XP = 0
	}
}
