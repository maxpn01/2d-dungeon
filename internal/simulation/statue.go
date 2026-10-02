package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/worldmap"
)

const StatueInteractionRadius = 100.0

type StatueState struct {
	rechargeUntil     uint64
	RechargeRemaining float64    `json:"recharge_remaining,omitempty"`
	ID                string     `json:"id"`
	X                 float64    `json:"x"`
	Y                 float64    `json:"y"`
	Cycle             uint64     `json:"cycle"`
	ClaimedBy         string     `json:"claimed_by,omitempty"`
	Cards             []PerkCard `json:"cards"`
}

func (s StatueState) Near(x, y float64, open ...[]string) bool {
	geometry := worldmap.Default()
	if len(open) > 0 {
		geometry = geometry.WithSecrets(open[0])
	}

	return math.Hypot(x+PlayerSize/2-s.X, y+PlayerSize/2-s.Y) <= StatueInteractionRadius && geometry.Visible(x+PlayerSize/2, y+PlayerSize/2, s.X, s.Y)
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
		d := math.Hypot(p.X+PlayerSize/2-s.X, p.Y+PlayerSize/2-s.Y)
		if d < distance {
			nearest, distance = i, d
		}
	}
	if nearest < 0 {
		return false
	}
	s := &w.statues[nearest]
	if s.ID != worldmap.WingStatue && s.Cards == nil {
		s.Cards = perkOffer(p.Stats)
		if len(s.Cards) == 0 {
			s.Cards = perkOffer(NewStats())
		}
		if s.Cards == nil {
			s.Cards = []PerkCard{}
		}
	}
	p.ActiveStatue = s.ID
	return true
}

func (w *World) CloseOffer(id string) {
	if p := w.players[id]; p != nil {
		p.ActiveStatue = ""
	}
}

func (w *World) ChoosePerk(id, statueID string, cycle uint64, card int) bool {
	p := w.players[id]
	if p == nil || p.Dead || p.ActiveStatue != statueID {
		return false
	}
	for i := range w.statues {
		s := &w.statues[i]
		if s.ID != statueID {
			continue
		}
		if s.ClaimedBy != "" || s.Cycle != cycle || !s.Near(p.X, p.Y, w.openSecrets) || card < 0 || (s.ID != worldmap.WingStatue && card >= len(s.Cards)) {
			return false
		}
		if s.ID == worldmap.WingStatue {
			if card != 0 || p.HasWings {
				return false
			}
			p.HasWings = true
		} else if !p.Stats.ApplyPerk(s.Cards[card]) {
			return false
		}
		s.ClaimedBy = id
		if s.ID != worldmap.WingStatue {
			s.rechargeUntil = w.tick + parameters.StatueRechargeSeconds*TickRate
			s.RechargeRemaining = parameters.StatueRechargeSeconds
			s.Cards = nil
		}
		for _, viewer := range w.players {
			if viewer.ActiveStatue == statueID {
				viewer.ActiveStatue = ""
			}
		}
		return true
	}
	return false
}

func (w *World) releaseStatues(id string) {
	for i := range w.statues {
		s := &w.statues[i]
		if s.ClaimedBy == id && s.ID == worldmap.WingStatue {
			s.ClaimedBy, s.Cards = "", nil
			s.Cycle++
		}
	}
}

func (w *World) updateStatues() {
	for i := range w.statues {
		s := &w.statues[i]
		if s.ID == worldmap.WingStatue || s.rechargeUntil == 0 {
			continue
		}
		if w.tick < s.rechargeUntil {
			s.RechargeRemaining = float64(s.rechargeUntil-w.tick) / TickRate
			continue
		}
		stats := NewStats()
		if p := w.players[s.ClaimedBy]; p != nil && !p.Dead {
			stats = p.Stats
		}
		s.Cards = perkOffer(stats)
		if len(s.Cards) == 0 {
			s.Cards = perkOffer(NewStats())
		}
		s.ClaimedBy, s.RechargeRemaining, s.rechargeUntil = "", 0, 0
		s.Cycle++
	}

	for _, p := range w.players {
		if p.Dead {
			if !p.progressionReset {
				w.releaseStatues(p.ID)
				w.releaseOrb(p.ID)
				p.HasWings, p.Flight = false, FlightState{}
				p.Stats = NewStats()
				p.Stats.HP = 0
				p.Split = nil
				p.ActiveStatue = ""
				p.progressionReset = true
			}
			continue
		}
		if p.ActiveStatue == "" {
			continue
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
}

func (w *World) Respawn(id string) bool {
	p := w.players[id]
	if p == nil || !p.Dead {
		return false
	}
	w.releaseStatues(id)
	x, y := w.spawnPosition(p.spawnSlot)
	p.Player = Player{ID: id, X: x, Y: y, Character: p.Character, Stats: NewStats(), InputSeq: p.lastSequence}
	p.direction, p.commands, p.progressionReset = Direction{}, nil, false
	return true
}
