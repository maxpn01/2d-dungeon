package simulation

import (
	"fmt"
	"math/rand/v2"
	"slices"

	"2d-dungeon/internal/parameters"
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
	VenomPerk
	PiercePerk
	ShedskinPerk
	HunterPerk
	PerkFamilyCount
)

const FirstForestPerk = VenomPerk

func (f PerkFamily) Forest() bool { return f >= FirstForestPerk && f < PerkFamilyCount }

const MaxPerkRank = 6

type PerkCard struct {
	Family PerkFamily
	Rank   int
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
	names := [...]string{"FORCE", "HASTE", "INSIGHT", "VITALITY", "STRIDE", "OVERDRIVE", "HEAVY HAND", "CHANNELING", "RENEWAL", "MAGNIFY", "VENOM", "PIERCE", "SHEDSKIN", "HUNTER'S INSTINCT"}
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
	return c.DescriptionFor(false)
}

func (c PerkCard) DescriptionFor(hasWings bool) string {
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
		if hasWings {
			return fmt.Sprintf("+%.0f%% flight speed / +%.0f%% MP drain", b*100, PerkPenalty(c.Family, c.Rank)*100)
		}
		return fmt.Sprintf("+%.0f%% Shift boost / +%.0f%% MP drain", b*100, PerkPenalty(c.Family, c.Rank)*100)
	case HeavyPerk:
		return fmt.Sprintf("+%.0f%% damage / -%.0f%% speed", b*100, PerkPenalty(c.Family, c.Rank)*100)
	case ManaRecoveryPerk:
		return fmt.Sprintf("+%.2f MP recovered per second", b)
	case HealthRecoveryPerk:
		return fmt.Sprintf("Recover %.2f HP per second", b)
	case ShotSizePerk:
		return fmt.Sprintf("+%.0f%% shot diameter / +%.0f%% shot MP cost", b*100, b*100)
	case VenomPerk:
		return fmt.Sprintf("Hits poison for %.0f%% damage every %.1fs for %.0fs", b*100, parameters.PoisonTickSeconds, parameters.VenomPoisonSeconds)
	case PiercePerk:
		targets := parameters.PiercePerkTargets[c.Rank-1]
		enemies := "enemy"
		if targets > 1 {
			enemies = "enemies"
		}
		return fmt.Sprintf("Shots pass through %d more %s / %.0f%% damage after each", targets, enemies, b*100)
	case ShedskinPerk:
		return fmt.Sprintf("Shed poison every %.0fs / immune for %.2gs after", b, parameters.ShedskinWardSeconds[c.Rank-1])
	case HunterPerk:
		return fmt.Sprintf("+%.0f%% damage to enemies above %.0f%% HP", b*100, parameters.HunterHPThreshold*100)
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
	damage += parameters.DamagePerLevelIncrease * float64(max(1, min(parameters.MaxLevel, s.Level))-1)
	s.Damage = damage * (1 + b(DamagePerk) + b(HeavyPerk))
	s.AttackRate = attackRateForLevel(s.Level, attackRateMultiplier) * (1 + b(AttackRatePerk))
	speed := parameters.BaseSpeed + parameters.SpeedPerLevelIncrease*float64(max(1, min(parameters.MaxLevel, s.Level))-1)
	s.Speed = speed * movementSpeedMultiplier * (1 + b(SpeedPerk) - PerkPenalty(HeavyPerk, s.Perks[HeavyPerk]))
	if s.Mob {
		switch s.MobType {
		case Sentinel:
			s.Speed *= parameters.SentinelSpeedMultiplier
		case DungeonPursuer:
			s.Speed *= parameters.PursuerSpeedMultiplier
		case ForestSnake:
			s.Speed *= parameters.SnakeSpeedMultiplier
		case CaveHydra:
			s.Speed *= parameters.HydraSpeedMultiplier
			s.Damage = (parameters.HydraBaseDamage + parameters.HydraDamagePerLevel*float64(max(1, s.Level)-1)) * (1 + b(DamagePerk) + b(HeavyPerk))
		case DungeonHeavy, GateHeavy:
			s.Speed *= parameters.HeavySpeedMultiplier
		}
	}
	if s.Mob && s.MobType == GateHeavy {
		s.Damage *= parameters.GateHeavyDamageMultiplier
	}
	s.BoostMultiplier = parameters.BoostMultiplier * (1 + b(BoostPerk))
	s.BoostCost = parameters.BoostMPPerSecond * (1 + PerkPenalty(BoostPerk, s.Perks[BoostPerk]))
}

func perkPool(forest bool) []PerkFamily {
	count := FirstForestPerk
	if forest {
		count = PerkFamilyCount
	}
	pool := make([]PerkFamily, count)
	for i := range pool {
		pool[i] = PerkFamily(i)
	}
	return pool
}

func perkUpgradeOffer(stats Stats) []PerkCard {
	return generatePerkOffer(stats.Perks, perkPool(false), 3)
}

func refreshPerkOffer(previous []PerkCard, stats Stats, forest bool) []PerkCard {
	cards := make([]PerkCard, 0, 3)
	excluded := make([]PerkFamily, 0, 3)
	for _, card := range previous {
		if card.Valid() && card.Rank > stats.Perks[card.Family] && (forest || !card.Family.Forest()) {
			cards = append(cards, card)
			excluded = append(excluded, card.Family)
		}
	}
	kept := len(cards)
	if forest && len(cards) < 3 && !slices.ContainsFunc(cards, func(c PerkCard) bool { return c.Family.Forest() }) {
		for _, card := range generatePerkOffer(stats.Perks, perkPool(true)[FirstForestPerk:], 1, excluded...) {
			cards = append(cards, card)
			excluded = append(excluded, card.Family)
		}
	}
	cards = append(cards, generatePerkOffer(stats.Perks, perkPool(forest), 3-len(cards), excluded...)...)
	if kept == 0 {
		rand.Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
	}
	return cards
}

func generatePerkOffer(minimumRanks [PerkFamilyCount]int, pool []PerkFamily, count int, excluded ...PerkFamily) []PerkCard {
	cards := make([]PerkCard, 0, max(0, count))
	for _, i := range rand.Perm(len(pool)) {
		if len(cards) >= count {
			break
		}
		f := pool[i]
		if slices.Contains(excluded, f) {
			continue
		}
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
				cards = append(cards, PerkCard{Family: f, Rank: rank + 1})
				break
			}
		}
	}
	return cards
}

func (s Stats) VenomDamage() float64 {
	return s.Damage * PerkBonus(VenomPerk, s.Perks[VenomPerk])
}

func (s Stats) PierceTargets() int {
	if rank := s.Perks[PiercePerk]; rank > 0 {
		return parameters.PiercePerkTargets[rank-1]
	}
	return 0
}

func (s Stats) HunterDamage() float64 {
	return s.Damage * (1 + PerkBonus(HunterPerk, s.Perks[HunterPerk]))
}

func (s Stats) perkShot(shot Projectile) Projectile {
	if rank := s.Perks[VenomPerk]; rank > 0 {
		shot.PoisonDamage = shot.Damage * PerkBonus(VenomPerk, rank)
		shot.PoisonSeconds = parameters.VenomPoisonSeconds
	}
	shot.Pierces = s.PierceTargets()
	shot.PierceDamage = PerkBonus(PiercePerk, s.Perks[PiercePerk])
	shot.HunterBonus = PerkBonus(HunterPerk, s.Perks[HunterPerk])
	return shot
}
