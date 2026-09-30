package simulation

import (
	"math"
	"raylib-2d-sandbox/internal/terrain"
)

type EntranceGateState struct {
	Open      float64 `json:"open"`
	Remaining float64 `json:"remaining"`
	Side      int     `json:"side"`
	Passed    bool    `json:"passed"`
}

func NearEntranceGate(x, y float64) bool {
	return x+PlayerSize >= terrain.EntranceGateLeft-48 && x <= terrain.EntranceGateRight+48 && math.Abs(y+PlayerSize/2-terrain.EntranceGateY) <= 110
}

func advanceEntranceGate(p *Player, input Direction, dt float64) {
	g := &p.EntranceGate
	if input.Interact && NearEntranceGate(p.X, p.Y) && g.Remaining == 0 {
		g.Remaining = 6
		g.Passed = false
		g.Side = 1
		if p.Y+PlayerSize/2 < terrain.EntranceGateY {
			g.Side = -1
		}
	}
	occupied := p.X+PlayerSize > terrain.EntranceGateLeft && p.X < terrain.EntranceGateRight && math.Abs(p.Y+PlayerSize/2-terrain.EntranceGateY) < PlayerSize/2+38
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
	g := &p.EntranceGate
	if g.Open < .95 && p.X+PlayerSize > terrain.EntranceGateLeft && p.X < terrain.EntranceGateRight {
		if oldY+PlayerSize/2 >= terrain.EntranceGateY && p.Y < terrain.EntranceGateY+30 {
			p.Y = terrain.EntranceGateY + 30
		} else if oldY+PlayerSize/2 < terrain.EntranceGateY && p.Y+PlayerSize > terrain.EntranceGateY-30 {
			p.Y = terrain.EntranceGateY - 30 - PlayerSize
		}
	}
	if g.Remaining > 0 && ((g.Side == 1 && p.Y+PlayerSize < terrain.EntranceGateY-38) || (g.Side == -1 && p.Y > terrain.EntranceGateY+38)) {
		g.Passed = true
	}
}
