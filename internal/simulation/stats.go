package simulation

import "math"

const (
	MaxLevel        = 100
	StartingHP      = 5.0
	StartingMaxHP   = 10.0
	FruitXP         = 0.25
	FruitHP         = 1.0
	ActorHPFraction = 0.5
)

type Stats struct {
	Level         int
	XP, HP, Speed float64
}

func NewStats() Stats { return Stats{Level: 1, HP: StartingHP, Speed: PlayerSpeed} }

func XPForNextLevel(level int) float64 {
	if level >= MaxLevel {
		return 0
	}
	// Each next level costs 2 XP more: 5, 7, 9, 11, ...
	return 5 + 2.0*float64(level-1)
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
	return StartingMaxHP + float64((level-1)*(level-1))
}

func (s *Stats) Eat(hp, xp float64) {
	s.GainXP(xp)
	if hp > 0 && !math.IsNaN(hp) && !math.IsInf(hp, 0) {
		s.HP = min(MaxHPForLevel(s.Level), s.HP+hp)
	}
}
