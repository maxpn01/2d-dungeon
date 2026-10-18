package simulation

import (
	"math"

	"2d-dungeon/internal/worldmap"
)

type EntranceGateState struct {
	ID        string  `json:"id,omitempty"`
	Open      float64 `json:"open"`
	Remaining float64 `json:"remaining"`
	Side      int     `json:"side"`
	Passed    bool    `json:"passed"`
}

func NearEntranceGateFor(id string, x, y float64) bool {
	gate, ok := worldmap.Default().Gate(id)
	return ok && x+PlayerSize >= float64(gate.Left)-48 && x <= float64(gate.Right)+48 && math.Abs(y+PlayerSize/2-float64(gate.Y)) <= 110
}

func advanceEntranceGate(p *Player, input Direction, dt float64) {
	id := p.EntranceID
	gate, ok := worldmap.Default().Gate(id)
	if !ok {
		return
	}
	p.EntranceGate.ID = id
	left, right, yGate := float64(gate.Left), float64(gate.Right), float64(gate.Y)
	g := &p.EntranceGate
	if input.Interact && NearEntranceGateFor(id, p.X, p.Y) && g.Remaining == 0 {
		g.Remaining = 6
		g.Passed = false
		g.Side = 1
		if p.Y+PlayerSize/2 < yGate {
			g.Side = -1
		}
	}
	occupied := p.X+PlayerSize > left && p.X < right && math.Abs(p.Y+PlayerSize/2-yGate) < PlayerSize/2+38
	if g.Remaining > 0 {
		g.Remaining = max(0, g.Remaining-dt)
		if g.Passed && !occupied {
			g.Remaining = 0
		}
	}
	if occupied && g.Open >= .95 {
		g.Remaining = max(g.Remaining, .25)
	}
	if g.Remaining > 0 {
		g.Open = min(1, g.Open+dt*2.5)
	} else {
		g.Open = max(0, g.Open-dt*2)
	}
}

func constrainEntranceGate(p *Player, oldX, oldY float64) {
	if p.Flight.Flying {
		return
	}
	for _, id := range []string{worldmap.EntranceGate} {
		gate, _ := worldmap.Default().Gate(id)
		left, right, yGate := float64(gate.Left), float64(gate.Right), float64(gate.Y)
		g := &p.EntranceGate
		open := 0.0
		if p.EntranceID == id && g.ID == id {
			open = g.Open
		}
		if open < .95 && p.X+PlayerSize > left && p.X < right {
			if oldY+PlayerSize/2 >= yGate && p.Y < yGate+30 && oldY >= yGate {
				p.Y = yGate + 30
			} else if oldY+PlayerSize/2 < yGate && p.Y+PlayerSize > yGate-30 && oldY+PlayerSize <= yGate {
				p.Y = yGate - 30 - PlayerSize
			}
		}
		if g.ID == id && g.Remaining > 0 && ((g.Side == 1 && p.Y+PlayerSize < yGate-38) || (g.Side == -1 && p.Y > yGate+38)) {
			g.Passed = true
		}
	}
}
