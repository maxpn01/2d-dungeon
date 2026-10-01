package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/worldmap"
)

const StatueInteractionRadius = 100.0

type StatueState struct {
	ID        string     `json:"id"`
	X         float64    `json:"x"`
	Y         float64    `json:"y"`
	Cycle     uint64     `json:"cycle"`
	ClaimedBy string     `json:"claimed_by,omitempty"`
	Cards     []PerkCard `json:"cards"`
}

func (s StatueState) Near(x, y float64) bool {
	return math.Hypot(x+PlayerSize/2-s.X, y+PlayerSize/2-s.Y) <= StatueInteractionRadius && worldmap.Default().Visible(x+PlayerSize/2, y+PlayerSize/2, s.X, s.Y)
}

func (w *World) interactStatue(p *playerState) bool {
	nearest, distance := -1, math.Inf(1)
	for i, s := range w.statues {
		if s.ClaimedBy != "" || !s.Near(p.X, p.Y) {
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
	if s.Cards == nil {
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
		if s.ClaimedBy != "" || s.Cycle != cycle || !s.Near(p.X, p.Y) || card < 0 || card >= len(s.Cards) {
			return false
		}
		if !p.Stats.ApplyPerk(s.Cards[card]) {
			return false
		}
		s.ClaimedBy = id
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
		if s.ClaimedBy == id {
			s.ClaimedBy, s.Cards = "", nil
			s.Cycle++
		}
	}
}

func (w *World) updateStatues() {
	for _, p := range w.players {
		if p.Dead {
			if !p.progressionReset {
				w.releaseStatues(p.ID)
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
			if s.ID == p.ActiveStatue && s.ClaimedBy == "" && s.Near(p.X, p.Y) {
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
