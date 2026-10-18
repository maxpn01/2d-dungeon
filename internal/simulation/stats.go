package simulation

import (
	"math"

	"2d-dungeon/internal/parameters"
)

type Stats struct {
	SpectralShots   bool
	MobType         MobType
	Mob             bool
	Level           int
	XP              float64
	HP              float64
	MP              float64
	Speed           float64
	Damage          float64
	AttackRate      float64
	BoostMultiplier float64
	BoostCost       float64
	Perks           [PerkFamilyCount]int
}

func NewStats() Stats {
	return NewStatsAtLevel(1)
}

func NewStatsAtLevel(level int) Stats {
	stats := Stats{Level: max(1, min(parameters.MaxLevel, level))}
	stats.rebuildPerks()
	stats.HP, stats.MP = stats.MaxHP(), stats.MaxMP()
	return stats
}

func NewMobStatsAtLevel(level int) Stats {
	stats := NewStatsAtLevel(level)
	stats.Mob = true
	stats.rebuildPerks()
	return stats
}

func (s Stats) MaxHP() float64 {
	hp := MaxHPForLevel(s.Level) + PerkBonus(HealthPerk, s.Perks[HealthPerk])
	if s.Mob && s.MobType.IsHeavy() {
		hp *= parameters.HeavyHealthMultiplier
	}
	if s.Mob && s.MobType == GateHeavy {
		hp *= parameters.GateHeavyHealthMultiplier
	}
	if s.Mob && s.MobType == ForestSnake {
		hp *= parameters.SnakeHealthMultiplier
	}
	return hp
}
func (s Stats) MaxMP() float64 {
	return MaxMPForLevel(s.Level) + PerkBonus(ManaPerk, s.Perks[ManaPerk])
}

func XPForNextLevel(level int) float64 {
	if level >= parameters.MaxLevel {
		return 0
	}
	levelsGained := float64(max(1, level) - 1)
	return parameters.XPFirstLevelCost + parameters.XPLevelCostGrowth*levelsGained
}

func (s *Stats) GainXP(amount float64) {
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) || s.Level >= parameters.MaxLevel {
		return
	}
	previousLevel := s.Level
	s.XP += amount
	for s.Level < parameters.MaxLevel && s.XP >= XPForNextLevel(s.Level) {
		s.XP -= XPForNextLevel(s.Level)
		s.Level++
	}
	if s.Level == parameters.MaxLevel {
		s.XP = 0
	}
	if s.Level != previousLevel {
		s.rebuildPerks()
	}
}

func AttackRateForLevel(level int) float64 {
	return attackRateForLevel(level, 1)
}

func attackRateForLevel(level int, baseMultiplier float64) float64 {
	level = max(1, min(parameters.MaxLevel, level))
	return parameters.BaseAttackRate*baseMultiplier + parameters.AttackRatePerLevelIncrease*float64(level-1)
}

func MaxHPForLevel(level int) float64 {
	level = max(1, min(parameters.MaxLevel, level))
	levelsGained := level - 1
	return parameters.StartingMaxHP + float64(parameters.HPPerLevelIncrease*levelsGained)
}

func MaxMPForLevel(level int) float64 {
	level = max(1, min(parameters.MaxLevel, level))
	return parameters.MaxMP + float64(parameters.MPPerLevelIncrease*(level-1))
}

func (s *Stats) Eat(hp, xp float64) {
	s.GainXP(xp)
	if hp > 0 && !math.IsNaN(hp) && !math.IsInf(hp, 0) {
		s.HP = min(s.MaxHP(), s.HP+hp)
	}
}

func (s *Stats) GainMP(amount float64) {
	if amount > 0 && !math.IsNaN(amount) && !math.IsInf(amount, 0) {
		s.MP = min(s.MaxMP(), s.MP+amount)
	}
}

func (s Stats) MovementSpeed() float64 { return s.Speed }

func (p Player) MovementSpeed() float64 { return p.Stats.MovementSpeed() }

func (s Stats) MPRecoveryPerSecond() float64 {
	return parameters.MPRecoveryPerSecond + PerkBonus(ManaRecoveryPerk, s.Perks[ManaRecoveryPerk])
}

func (s Stats) HPRecoveryPerSecond() float64 {
	return PerkBonus(HealthRecoveryPerk, s.Perks[HealthRecoveryPerk])
}

func (s Stats) ShotSize() float64 {
	return parameters.ShotSize * (1 + PerkBonus(ShotSizePerk, s.Perks[ShotSizePerk]))
}

func (s Stats) ShotMPCost() float64 {
	cost := parameters.ShotMPCost * s.ShotSize() / parameters.ShotSize
	if s.SpectralShots {
		cost *= 1.25
	}
	return cost
}
