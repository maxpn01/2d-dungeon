package simulation

import (
	"math"
	"math/rand/v2"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/worldmap"
)

func (w *World) respawnDenAI(previous aiState) aiState {
	ai := w.newDenAI(previous.den)
	ai.deaths, ai.perkMilestones = previous.deaths, previous.perkMilestones
	if ai.den.RoomID == worldmap.OrbVault {
		return ai
	}
	level := min(parameters.MaxLevel, ai.den.Level+ai.deaths/parameters.AIDeathsPerLevel)
	ai.Stats = newMobStats(level, ai.Type)
	ai.Stats.Perks = previous.Stats.Perks
	ai.Stats.rebuildPerks()
	ai.grantPerkMilestones(level / parameters.AIPerkLevelInterval)
	ai.Stats.HP, ai.Stats.MP = ai.Stats.MaxHP(), ai.Stats.MaxMP()
	ai.spawnMaxHP = ai.Stats.MaxHP()
	return ai
}

func (ai *aiState) grantPerkMilestones(milestones int) {
	for ; ai.perkMilestones < milestones; ai.perkMilestones++ {
		cards := perkUpgradeOffer(ai.Stats)
		if len(cards) > 0 {
			ai.Stats.ApplyPerk(cards[rand.IntN(len(cards))])
		}
	}
}

func (w *World) averageDungeonGuardLevel() int {
	total, count := 0, 0
	geometry := w.geometry()
	for _, ai := range w.ais {
		if ai.den.ID == "" || ai.den.RoomID == worldmap.OrbVault || ai.Type == GateHeavy || geometry.Cell(ai.den.At.X, ai.den.At.Y) == 0 {
			continue
		}
		room, ok := geometry.Room(ai.den.RoomID)
		if !ok || room.Floor != worldmap.OrbFloor || room.HasRole(worldmap.Private) {
			continue
		}
		level := ai.Stats.Level
		if ai.Frenzied {
			level = ai.frenzyBase.Level
		}
		total, count = total+level, count+1
	}
	if count == 0 {
		return 2
	}
	return int(math.Round(float64(total) / float64(count)))
}
