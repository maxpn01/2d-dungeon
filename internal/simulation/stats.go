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
	levelsGained := level - 1
	return StartingMaxHP + float64(4*levelsGained+levelsGained*(levelsGained-1)/2)
}

func (s *Stats) Eat(hp, xp float64) {
	s.GainXP(xp)
	if hp > 0 && !math.IsNaN(hp) && !math.IsInf(hp, 0) {
		s.HP = min(MaxHPForLevel(s.Level), s.HP+hp)
	}
}

func SpeedForMass(base, hp float64) float64 {
	return base * max(1.0/3, min(4.0/3, math.Pow(StartingHP/max(1, hp), .18)))
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
	return largest + .25*(total-largest)
}

const CircleSplitSpeedMultiplier = 1.2

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
