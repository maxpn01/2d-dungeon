package simulation

import (
	"math"
	"raylib-2d-sandbox/internal/worldmap"
)

type EntranceGateState struct {
	Open      float64 `json:"open"`
	Remaining float64 `json:"remaining"`
	Side      int     `json:"side"`
	Passed    bool    `json:"passed"`
}

func NearEntranceGate(x, y float64) bool {
	gate := worldmap.Default().EntranceGate()
	left, right, yGate := float64(gate.Left), float64(gate.Right), float64(gate.Y)
	return right > left && x+PlayerSize >= left-48 && x <= right+48 && math.Abs(y+PlayerSize/2-yGate) <= 110
}

func advanceEntranceGate(p *Player, input Direction, dt float64) {
	gate := worldmap.Default().EntranceGate()
	left, right, yGate := float64(gate.Left), float64(gate.Right), float64(gate.Y)
	g := &p.EntranceGate
	if input.Interact && NearEntranceGate(p.X, p.Y) && g.Remaining == 0 {
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
	gate := worldmap.Default().EntranceGate()
	if gate.Right <= gate.Left {
		return
	}
	left, right, yGate := float64(gate.Left), float64(gate.Right), float64(gate.Y)
	g := &p.EntranceGate
	if g.Open < .95 && p.X+PlayerSize > left && p.X < right {
		if oldY+PlayerSize/2 >= yGate && p.Y < yGate+30 {
			p.Y = yGate + 30
		} else if oldY+PlayerSize/2 < yGate && p.Y+PlayerSize > yGate-30 {
			p.Y = yGate - 30 - PlayerSize
		}
	}
	if g.Remaining > 0 && ((g.Side == 1 && p.Y+PlayerSize < yGate-38) || (g.Side == -1 && p.Y > yGate+38)) {
		g.Passed = true
	}
}
