package simulation

import (
	"fmt"
	"math/rand/v2"

	"raylib-2d-sandbox/internal/parameters"
)

type PerkFamily int

const (
	DamagePerk PerkFamily = iota
	AttackRatePerk
	ManaPerk
	HealthPerk
	SpeedPerk
	BoostPerk
	HeavyPerk
	ManaRecoveryPerk
	HealthRecoveryPerk
	ShotSizePerk
	PerkFamilyCount
)

const MaxPerkRank = 6

type PerkCard struct {
	Family PerkFamily `json:"family"`
	Rank   int        `json:"rank"`
}

func RankName(rank int) string {
	if rank < 1 || rank > MaxPerkRank {
		return "-"
	}
	return []string{"E", "D", "C", "B", "A", "S"}[rank-1]
}

func (c PerkCard) Valid() bool {
	return c.Family >= 0 && c.Family < PerkFamilyCount && c.Rank >= 1 && c.Rank <= MaxPerkRank
}

func PerkName(family PerkFamily) string {
	names := [...]string{"FORCE", "HASTE", "INSIGHT", "VITALITY", "STRIDE", "OVERDRIVE", "HEAVY HAND", "CHANNELING", "RENEWAL", "MAGNIFY"}
	if family < 0 || family >= PerkFamilyCount {
		return "UNKNOWN"
	}
	return names[family]
}

func PerkBonus(family PerkFamily, rank int) float64 {
	if family < 0 || family >= PerkFamilyCount || rank < 1 || rank > MaxPerkRank {
		return 0
	}
	return parameters.PerkBonuses[family][rank-1]
}

func PerkPenalty(family PerkFamily, rank int) float64 {
	if family < 0 || family >= PerkFamilyCount || rank < 1 || rank > MaxPerkRank {
		return 0
	}
	return parameters.PerkPenalties[family][rank-1]
}

func (c PerkCard) Description() string {
	b := PerkBonus(c.Family, c.Rank)
	switch c.Family {
	case DamagePerk:
		return fmt.Sprintf("+%.0f%% damage per shot", b*100)
	case AttackRatePerk:
		return fmt.Sprintf("+%.0f%% attack rate", b*100)
	case ManaPerk:
		return fmt.Sprintf("+%.0f maximum MP", b)
	case HealthPerk:
		return fmt.Sprintf("+%.0f maximum HP", b)
	case SpeedPerk:
		return fmt.Sprintf("+%.0f%% base speed / no extra MP cost", b*100)
	case BoostPerk:
		return fmt.Sprintf("+%.0f%% Shift boost / +%.0f%% MP drain", b*100, PerkPenalty(c.Family, c.Rank)*100)
	case HeavyPerk:
		return fmt.Sprintf("+%.0f%% damage / -%.0f%% speed", b*100, PerkPenalty(c.Family, c.Rank)*100)
	case ManaRecoveryPerk:
		return fmt.Sprintf("+%.2f MP recovered per second", b)
	case HealthRecoveryPerk:
		return fmt.Sprintf("Recover %.2f HP per second", b)
	case ShotSizePerk:
		return fmt.Sprintf("+%.0f%% shot diameter / +%.0f%% shot MP cost", b*100, b*100)
	}
	return ""
}

func (s *Stats) ApplyPerk(card PerkCard) bool {
	if !card.Valid() || card.Rank <= s.Perks[card.Family] {
		return false
	}
	oldHP, oldMP := s.MaxHP(), s.MaxMP()
	s.Perks[card.Family] = card.Rank
	s.rebuildPerks()
	s.HP = min(s.MaxHP(), s.HP+s.MaxHP()-oldHP)
	s.MP = min(s.MaxMP(), s.MP+s.MaxMP()-oldMP)
	return true
}

func (s *Stats) rebuildPerks() {
	b := func(f PerkFamily) float64 { return PerkBonus(f, s.Perks[f]) }
	damageMultiplier, attackRateMultiplier, movementSpeedMultiplier := 1.0, 1.0, 1.0
	if s.Mob {
		damageMultiplier = parameters.MobDamageMultiplier
		attackRateMultiplier = parameters.MobAttackRateMultiplier
		movementSpeedMultiplier = parameters.MobMovementSpeedMultiplier
	}
	damage := parameters.ShotDamage * damageMultiplier
	damage += parameters.DamagePerLevelIncrease * float64(max(1, min(MaxLevel, s.Level))-1)
	s.Damage = damage * (1 + b(DamagePerk) + b(HeavyPerk))
	s.AttackRate = attackRateForLevel(s.Level, attackRateMultiplier) * (1 + b(AttackRatePerk))
	speed := BaseSpeed + parameters.SpeedPerLevelIncrease*float64(max(1, min(MaxLevel, s.Level))-1)
	s.Speed = speed * movementSpeedMultiplier * (1 + b(SpeedPerk) - PerkPenalty(HeavyPerk, s.Perks[HeavyPerk]))
	if s.Mob {
		switch s.MobType {
		case Sentinel:
			s.Speed *= parameters.SentinelSpeedMultiplier
		case DungeonPursuer:
			s.Speed *= parameters.PursuerSpeedMultiplier
		case DungeonHeavy:
			s.Speed *= parameters.HeavySpeedMultiplier
		}
	}
	s.BoostMultiplier = BoostMultiplier * (1 + b(BoostPerk))
	s.BoostCost = BoostMPPerSecond * (1 + PerkPenalty(BoostPerk, s.Perks[BoostPerk]))
}

func perkOffer() []PerkCard {
	return generatePerkOffer([PerkFamilyCount]int{})
}

func perkUpgradeOffer(stats Stats) []PerkCard {
	return generatePerkOffer(stats.Perks)
}

func generatePerkOffer(minimumRanks [PerkFamilyCount]int) []PerkCard {
	families := rand.Perm(int(PerkFamilyCount))
	cards := make([]PerkCard, 0, 3)
	for _, f := range families {
		owned := minimumRanks[f]
		total := 0
		for rank := owned; rank < MaxPerkRank; rank++ {
			total += max(0, parameters.PerkRankWeights[rank])
		}
		if total == 0 {
			continue
		}
		roll := rand.IntN(total)
		for rank := owned; rank < MaxPerkRank; rank++ {
			roll -= max(0, parameters.PerkRankWeights[rank])
			if roll < 0 {
				cards = append(cards, PerkCard{Family: PerkFamily(f), Rank: rank + 1})
				break
			}
		}
		if len(cards) == 3 {
			break
		}
	}
	return cards
}
