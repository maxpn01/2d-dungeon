package simulation

import (
	"math"

	"2d-dungeon/internal/parameters"
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
	SpectralShots   bool                 `json:"spectral_shots,omitempty"`
	MobType         MobType              `json:"mob_type,omitempty"`
	Mob             bool                 `json:"mob,omitempty"`
	Level           int                  `json:"level"`
	XP              float64              `json:"xp"`
	HP              float64              `json:"hp"`
	MP              float64              `json:"mp"`
	Speed           float64              `json:"speed"`
	Damage          float64              `json:"damage"`
	AttackRate      float64              `json:"attack_rate"`
	BoostMultiplier float64              `json:"boost_multiplier"`
	BoostCost       float64              `json:"boost_cost"`
	Perks           [PerkFamilyCount]int `json:"perks"`
}

func NewStats() Stats {
	return NewStatsAtLevel(1)
}

func NewStatsAtLevel(level int) Stats {
	stats := Stats{Level: max(1, min(MaxLevel, level))}
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
	return hp
}
func (s Stats) MaxMP() float64 {
	return MaxMPForLevel(s.Level) + PerkBonus(ManaPerk, s.Perks[ManaPerk])
}

func XPForNextLevel(level int) float64 {
	if level >= MaxLevel {
		return 0
	}
	levelsGained := float64(max(1, level) - 1)
	return parameters.XPFirstLevelCost + parameters.XPLevelCostGrowth*levelsGained
}

func (s *Stats) GainXP(amount float64) {
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) || s.Level >= MaxLevel {
		return
	}
	previousLevel := s.Level
	s.XP += amount
	for s.Level < MaxLevel && s.XP >= XPForNextLevel(s.Level) {
		s.XP -= XPForNextLevel(s.Level)
		s.Level++
	}
	if s.Level == MaxLevel {
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
	level = max(1, min(MaxLevel, level))
	return parameters.BaseAttackRate*baseMultiplier + parameters.AttackRatePerLevelIncrease*float64(level-1)
}

func MaxHPForLevel(level int) float64 {
	level = max(1, min(MaxLevel, level))
	levelsGained := level - 1
	return StartingMaxHP + float64(parameters.HPPerLevelIncrease*levelsGained)
}

func MaxMPForLevel(level int) float64 {
	level = max(1, min(MaxLevel, level))
	return MaxMP + float64(parameters.MPPerLevelIncrease*(level-1))
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
