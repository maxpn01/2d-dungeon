package client

import (
	"math"
	"time"

	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"
)

const interpolationDelay = 100 * time.Millisecond

type Motion struct {
	frames []Update
	local  protocol.Player
}

func (m *Motion) Apply(update Update) {
	if len(m.frames) == 0 {
		m.local = update.Local
	}

	n := len(m.frames)
	if n > 0 && m.frames[n-1].Tick == update.Tick {
		update.ReceivedAt = m.frames[n-1].ReceivedAt
		m.frames[n-1] = update
	} else {
		m.frames = append(m.frames, update)
		if len(m.frames) > 16 {
			m.frames = m.frames[len(m.frames)-16:]
		}
	}
}

func (m *Motion) Advance(target protocol.Player, dt float64) {
	if len(m.frames) == 0 {
		return
	}
	dt = max(0, min(dt, .05))
	dx, dy := target.X-m.local.X, target.Y-m.local.Y
	distance := math.Hypot(dx, dy)
	if distance > 200 {
		m.local = target
		return
	}
	speed := target.Stats.Speed
	if distance > 2*speed*simulation.StepSeconds {
		speed *= 2
	}
	if distance <= speed*dt || distance == 0 {
		m.local = target
		return
	}
	m.local.X += dx / distance * speed * dt
	m.local.Y += dy / distance * speed * dt
	m.local.ID, m.local.Size, m.local.InputSeq, m.local.Stats = target.ID, target.Size, target.InputSeq, target.Stats
}

func (m *Motion) Players(now time.Time) []protocol.Player {
	if len(m.frames) == 0 {
		return nil
	}
	latest := m.frames[len(m.frames)-1]
	target := now.Add(-interpolationDelay)
	before, after := m.frames[0], m.frames[0]
	for _, frame := range m.frames {
		after = frame
		if !frame.ReceivedAt.Before(target) {
			break
		}
		before = frame
	}
	alpha := 0.0
	if span := after.ReceivedAt.Sub(before.ReceivedAt); span > 0 {
		alpha = max(0, min(1, float64(target.Sub(before.ReceivedAt))/float64(span)))
	}
	beforeByID, afterByID := indexPlayers(before.Players), indexPlayers(after.Players)
	positions := make([]protocol.Player, 0, len(latest.Players))
	for _, player := range latest.Players {
		if player.ID == latest.PlayerID {
			player = m.local
			player.X = max(simulation.MinPosition, min(simulation.MaxPosition, player.X))
			player.Y = max(simulation.MinPosition, min(simulation.MaxPosition, player.Y))
		} else {
			a, okA := beforeByID[player.ID]
			b, okB := afterByID[player.ID]
			if okA && okB {
				player.X = a.X + (b.X-a.X)*alpha
				player.Y = a.Y + (b.Y-a.Y)*alpha
			}
		}
		positions = append(positions, player)
	}
	return positions
}

func indexPlayers(players []protocol.Player) map[string]protocol.Player {
	byID := make(map[string]protocol.Player, len(players))
	for _, player := range players {
		byID[player.ID] = player
	}
	return byID
}
