package simulation

import (
	"math"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/worldmap"
)

const StatueInteractionRadius = 100.0

type StatueState struct {
	Library           LibraryState
	rechargeUntil     uint64
	RechargeRemaining float64
	ID                string
	X                 float64
	Y                 float64
	Cycle             uint64
	ClaimedBy         string
	Cards             []PerkCard
}

func (s StatueState) Near(x, y float64, open ...[]string) bool {
	geometry := worldmap.Default()
	if len(open) > 0 {
		geometry = geometry.WithSecrets(open[0])
	}

	return math.Hypot(x+parameters.PlayerSize/2-s.X, y+parameters.PlayerSize/2-s.Y) <= StatueInteractionRadius && geometry.Visible(x+parameters.PlayerSize/2, y+parameters.PlayerSize/2, s.X, s.Y)
}

func (s StatueState) Forest() bool {
	return worldmap.Default().FloorAt(s.X, s.Y) == worldmap.SecondFloor
}

func (w *World) interactStatue(p *playerState) bool {
	if p.Flight.Flying {
		return false
	}
	nearest, distance := -1, math.Inf(1)
	for i, s := range w.statues {
		if s.ClaimedBy != "" || !s.Near(p.X, p.Y, w.openSecrets) {
			continue
		}
		d := math.Hypot(p.X+parameters.PlayerSize/2-s.X, p.Y+parameters.PlayerSize/2-s.Y)
		if d < distance {
			nearest, distance = i, d
		}
	}
	if nearest < 0 {
		return false
	}
	s := &w.statues[nearest]
	if IsLibraryBook(s.ID) {
		return w.startLibrary(s, p)
	}
	if !worldmap.IsWingStatue(s.ID) {
		s.Cards = refreshPerkOffer(s.Cards, p.Stats, s.Forest())
	}
	p.ActiveStatue = s.ID
	return true
}

func (w *World) CloseOffer() {
	w.player.ActiveStatue = ""
}

func (w *World) ChoosePerk(statueID string, cycle uint64, card int) bool {
	p := w.player
	if p.Dead || p.ActiveStatue != statueID {
		return false
	}
	for i := range w.statues {
		s := &w.statues[i]
		if s.ID != statueID {
			continue
		}
		if s.ClaimedBy != "" || s.Cycle != cycle || !s.Near(p.X, p.Y, w.openSecrets) || card < 0 || (!worldmap.IsWingStatue(s.ID) && card >= len(s.Cards)) {
			return false
		}
		if worldmap.IsWingStatue(s.ID) {
			if card != 0 || p.HasWings {
				return false
			}
			p.HasWings = true
		} else if !p.Stats.ApplyPerk(s.Cards[card]) {
			return false
		}
		s.ClaimedBy = p.ID
		if !worldmap.IsWingStatue(s.ID) {
			s.rechargeUntil = w.tick + parameters.StatueRechargeSeconds*parameters.TickRate
			s.RechargeRemaining = parameters.StatueRechargeSeconds
			s.Cards = nil
		}
		p.ActiveStatue = ""
		return true
	}
	return false
}

func (w *World) releaseStatues(id string) {
	for i := range w.statues {
		s := &w.statues[i]
		s.Cards = nil
		if s.ClaimedBy == id && worldmap.IsWingStatue(s.ID) {
			s.ClaimedBy, s.Cards = "", nil
			s.Cycle++
		}
	}
}

func (w *World) updateStatues() {
	for i := range w.statues {
		s := &w.statues[i]
		if worldmap.IsWingStatue(s.ID) || s.rechargeUntil == 0 {
			continue
		}
		if w.tick < s.rechargeUntil {
			s.RechargeRemaining = float64(s.rechargeUntil-w.tick) / parameters.TickRate
			continue
		}
		s.Cards = nil
		s.ClaimedBy, s.RechargeRemaining, s.rechargeUntil = "", 0, 0
		s.Cycle++
	}

	p := w.player
	if p.Dead {
		if !p.progressionReset {
			w.releaseStatues(p.ID)
			w.releaseOrb(p.ID)
			p.HasWings, p.Flight = false, FlightState{}
			p.Stats = NewStats()
			p.Stats.HP = 0
			p.ActiveStatue = ""
			p.progressionReset = true
		}
		return
	}
	if p.ActiveStatue == "" {
		return
	}
	valid := false
	for _, s := range w.statues {
		if s.ID == p.ActiveStatue && s.ClaimedBy == "" && s.Near(p.X, p.Y, w.openSecrets) {
			valid = true
			break
		}
	}
	if !valid {
		p.ActiveStatue = ""
	}
}

func (w *World) statueSnapshots() []StatueState {
	statues := append([]StatueState(nil), w.statues...)
	for i := range statues {
		statues[i].Cards = append([]PerkCard(nil), statues[i].Cards...)
	}
	return statues
}
