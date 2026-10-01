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
	OverdrivePerk
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
	names := [...]string{"FORCE", "HASTE", "INSIGHT", "VITALITY", "STRIDE", "SURGE", "HEAVY HAND", "OVERDRIVE"}
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
		return fmt.Sprintf("+%.0f%% movement speed", b*100)
	case BoostPerk:
		return fmt.Sprintf("+%.0f%% Shift boost strength", b*100)
	case HeavyPerk:
		return fmt.Sprintf("+%.0f%% damage / -%.0f%% speed", b*100, PerkPenalty(c.Family, c.Rank)*100)
	case OverdrivePerk:
		return fmt.Sprintf("+%.0f%% boost / +%.0f%% MP cost", b*100, PerkPenalty(c.Family, c.Rank)*100)
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
	s.Damage = parameters.ShotDamage * (1 + b(DamagePerk) + b(HeavyPerk))
	s.AttackRate = parameters.PlayerAttackRate * (1 + b(AttackRatePerk))
	s.Speed = PlayerSpeed * (1 + b(SpeedPerk) - PerkPenalty(HeavyPerk, s.Perks[HeavyPerk]))
	s.BoostMultiplier = BoostMultiplier * (1 + b(BoostPerk) + b(OverdrivePerk))
	s.BoostCost = BoostMPPerSecond * (1 + PerkPenalty(OverdrivePerk, s.Perks[OverdrivePerk]))
}

func perkOffer(stats Stats) []PerkCard {
	families := rand.Perm(int(PerkFamilyCount))
	cards := make([]PerkCard, 0, 3)
	for _, f := range families {
		owned := stats.Perks[f]
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
