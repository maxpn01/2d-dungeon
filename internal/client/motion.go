package client

import (
	"math"
	"time"

	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"
)

const interpolationDelay = 100 * time.Millisecond

// Motion belongs to the render thread. It never alters authoritative snapshots.
type Motion struct {
	frames           []Update
	local            protocol.Player
	offsetX, offsetY float64
}

func (m *Motion) Apply(update Update) {
	if len(m.frames) > 0 {
		// Keep small corrections invisible at arrival, then ease them out over time.
		m.offsetX += m.local.X - update.Local.X
		m.offsetY += m.local.Y - update.Local.Y
		if math.Hypot(m.offsetX, m.offsetY) > 200 {
			m.offsetX, m.offsetY = 0, 0
		}
	}
	m.local = update.Local
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

func (m *Motion) Advance(direction protocol.Direction, dt float64, now time.Time) {
	if len(m.frames) == 0 {
		return
	}
	// Don't predict a huge jump after a paused window or a long network stall.
	dt = max(0, min(dt, .05))
	if now.Sub(m.frames[len(m.frames)-1].ReceivedAt) < 250*time.Millisecond {
		m.local.X, m.local.Y = simulation.Move(m.local.X, m.local.Y, simulation.Direction{X: direction.X, Y: direction.Y}, dt)
	}
	decay := math.Exp(-dt / .08)
	m.offsetX *= decay
	m.offsetY *= decay
}

// Players presents remote entities a little behind real time so two snapshots
// bracket the display time. Missing newer data freezes them rather than guessing.
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
	// Use the newest roster so joins/leaves don't leave stale ghosts in the buffer.
	for _, player := range latest.Players {
		if player.ID == latest.PlayerID {
			player = m.local
			player.X = max(simulation.MinPosition, min(simulation.MaxPosition, player.X+m.offsetX))
			player.Y = max(simulation.MinPosition, min(simulation.MaxPosition, player.Y+m.offsetY))
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
