package client

import (
	"math"
	"time"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/worldmap"
)

const (
	maxExtrapolation  = parameters.MaxExtrapolation
	correctionSeconds = parameters.CorrectionSeconds
	teleportDistance  = parameters.TeleportDistance
	playerBufferTicks = parameters.PlayerBufferTicks
)

type motionCorrection struct {
	x, y float64
	at   time.Time
}

func (c motionCorrection) offset(now time.Time) (float64, float64) {
	decay := math.Exp(-max(0, now.Sub(c.at).Seconds()) / correctionSeconds)
	return c.x * decay, c.y * decay
}

type Motion struct {
	frames        []Update
	local         protocol.Player
	playerTick    float64
	playerClock   time.Time
	aiCorrections map[string]motionCorrection
}

func (m *Motion) Apply(update Update) {
	now := update.ReceivedAt
	previousAIs := indexAIs(m.AIs(now))
	if len(m.frames) == 0 {
		m.local = update.Local
		m.local.Split = update.Local.Split.Clone()
		m.playerTick = float64(update.Tick) - playerBufferTicks
		m.playerClock = now
	}

	n := len(m.frames)
	if n > 0 && m.frames[n-1].Tick == update.Tick {
		update.ReceivedAt = m.frames[n-1].ReceivedAt
		m.frames[n-1] = update
	} else {
		m.frames = append(m.frames, update)
		if len(m.frames) > 32 {
			m.frames = m.frames[len(m.frames)-32:]
		}
	}
	m.aiCorrections = nil
	ais := m.AIs(now)
	m.aiCorrections = make(map[string]motionCorrection, len(ais))
	for _, ai := range ais {
		if previous, ok := previousAIs[ai.ID]; ok {
			m.aiCorrections[ai.ID] = newCorrection(previous.X, previous.Y, ai.X, ai.Y, now)
		}
	}
}

func newCorrection(oldX, oldY, x, y float64, now time.Time) motionCorrection {
	dx, dy := oldX-x, oldY-y
	if math.Hypot(dx, dy) > teleportDistance {
		return motionCorrection{}
	}
	return motionCorrection{x: dx, y: dy, at: now}
}

func (m *Motion) Advance(target protocol.Player, dt float64) {
	if len(m.frames) == 0 {
		return
	}
	defer constrainPlayerBody(&m.local)
	dt = max(0, min(dt, .05))
	previous := m.local
	defer func() {
		if math.Hypot(target.X-previous.X, target.Y-previous.Y) > teleportDistance {
			return
		}
		for i, cell := range target.Split {
			if !cell.Active {
				continue
			}
			offsetX, offsetY := cell.X-target.X, cell.Y-target.Y
			if i < len(previous.Split) && previous.Split[i].Active {
				oldX, oldY := previous.Split[i].X-previous.X, previous.Split[i].Y-previous.Y
				blend := 1 - math.Exp(-simulation.SplitFollowResponse*dt)
				offsetX, offsetY = oldX+(offsetX-oldX)*blend, oldY+(offsetY-oldY)*blend
			}
			if i < len(m.local.Split) {
				m.local.Split[i].X, m.local.Split[i].Y = worldmap.Default().Move(m.local.X, m.local.Y, offsetX, offsetY, simulation.PlayerSize, simulation.PlayerSize)
			}
		}
	}()
	dx, dy := target.X-m.local.X, target.Y-m.local.Y
	distance := math.Hypot(dx, dy)
	if distance > 200 {
		m.local = target
		m.local.Split = target.Split.Clone()
		return
	}
	speed := playerMovementSpeed(target) * simulation.SpeedScale
	if target.Flight.Flying || m.local.Flight.Flying {
		speed *= simulation.WingSpeedMultiplier
		if target.Flight.Returning {
			speed *= 2
		}
	}
	if target.Combat.Boosting || m.local.Combat.Boosting {
		speed *= target.Stats.BoostMultiplier
	}
	if target.Combat.Phase == simulation.DashRushing || m.local.Combat.Phase == simulation.DashRushing {
		speed *= simulation.DashSpeedMultiplier
	}
	if target.Combat.Phase == simulation.DashRecoiling {
		speed = max(speed, 2*simulation.BossRecoilDistance/(simulation.BossRecoilDuration-simulation.BossImpactHold))
	}
	if distance > 2*speed*simulation.StepSeconds {
		speed *= 2
	}
	if distance <= speed*dt || distance == 0 {
		m.local = target
		m.local.Split = target.Split.Clone()
		return
	}
	geometry := worldmap.Default().WithSecrets(target.OpenSecrets).WithVaultGate(target.VaultGateOpen)
	x, y := geometry.Move(m.local.X, m.local.Y, dx/distance*speed*dt, dy/distance*speed*dt, simulation.PlayerSize, simulation.PlayerSize)
	if target.Flight.Flying || m.local.Flight.Flying {
		x, y = geometry.MoveFlying(m.local.X, m.local.Y, dx/distance*speed*dt, dy/distance*speed*dt, simulation.PlayerSize)
	}
	m.local = target
	m.local.Split = target.Split.Clone()
	m.local.X, m.local.Y = x, y
}

func (m *Motion) Players(now time.Time) []protocol.Player {
	if len(m.frames) == 0 {
		return nil
	}
	latest := m.frames[len(m.frames)-1]
	before, after, alpha := m.playerFramesAt(now)
	afterByID := indexPlayers(after.Players)
	positions := make([]protocol.Player, 0, len(before.Players))
	for _, player := range before.Players {
		player.Split = player.Split.Clone()
		if player.ID == latest.PlayerID {
			if latest.Local.Dead {
				continue
			}
			player = m.local
			constrainPlayerBody(&player)
			positions = append(positions, player)
			continue
		}
		if player.Dead {
			continue
		}
		if next, ok := afterByID[player.ID]; ok {
			if math.Hypot(next.X-player.X, next.Y-player.Y) <= teleportDistance {
				oldX, oldY := player.X, player.Y
				oldSplit := player.Split
				if player.Flight.Flying || next.Flight.Flying {
					player.X, player.Y = worldmap.Default().WithSecrets(next.OpenSecrets).WithVaultGate(next.VaultGateOpen).MoveFlying(player.X, player.Y, (next.X-player.X)*alpha, (next.Y-player.Y)*alpha, simulation.PlayerSize)
				} else {
					player.X, player.Y = interpolatePosition(player.X, player.Y, next.X, next.Y, alpha)
				}
				player.HasWings, player.Flight = next.HasWings, next.Flight
				player.Split = next.Split.Clone()
				for i, cell := range next.Split {
					if !cell.Active {
						continue
					}
					startX, startY := cell.X-next.X, cell.Y-next.Y
					if i < len(oldSplit) && oldSplit[i].Active {
						startX, startY = oldSplit[i].X-oldX, oldSplit[i].Y-oldY
					}
					player.Split[i].X = player.X + startX + (cell.X-next.X-startX)*alpha
					player.Split[i].Y = player.Y + startY + (cell.Y-next.Y-startY)*alpha
				}
			}
		}
		constrainPlayerBody(&player)
		positions = append(positions, player)
	}
	return positions
}

func (m *Motion) playerFramesAt(now time.Time) (Update, Update, float64) {
	latest := m.frames[len(m.frames)-1]
	if now.After(m.playerClock) {
		dt := now.Sub(m.playerClock).Seconds()
		buffered := float64(latest.Tick) - m.playerTick
		rate := 1.0
		if buffered > playerBufferTicks+1 {
			rate = 1.05
		}
		if buffered < playerBufferTicks-1 {
			rate = 0.95
		}
		m.playerTick = min(float64(latest.Tick), m.playerTick+dt*simulation.TickRate*rate)
		m.playerClock = now
	}
	before := m.frames[0]
	for _, after := range m.frames {
		if float64(after.Tick) > m.playerTick {
			span := float64(after.Tick - before.Tick)
			if span == 0 {
				return before, before, 0
			}
			return before, after, max(0, min(1, (m.playerTick-float64(before.Tick))/span))
		}
		before = after
	}
	return before, before, 0
}

func interpolatePosition(ax, ay, bx, by, alpha float64) (float64, float64) {
	dx, dy := (bx-ax)*alpha, (by-ay)*alpha
	distance := math.Hypot(dx, dy)
	if distance == 0 {
		return ax, ay
	}
	return simulation.MoveBy(ax, ay, simulation.Direction{X: dx / distance, Y: dy / distance}, distance)
}

func indexPlayers(players []protocol.Player) map[string]protocol.Player {
	byID := make(map[string]protocol.Player, len(players))
	for _, player := range players {
		byID[player.ID] = player
	}
	return byID
}

func (m *Motion) aiFramesAt(now time.Time) (Update, Update, float64) {
	after := m.frames[len(m.frames)-1]
	if len(m.frames) < 2 {
		return after, after, 1
	}
	before := m.frames[len(m.frames)-2]
	span := float64(after.Tick-before.Tick) * simulation.StepSeconds
	if span <= 0 {
		return after, after, 1
	}
	elapsed := max(0, min(now.Sub(after.ReceivedAt).Seconds(), maxExtrapolation.Seconds()))
	return before, after, 1 + elapsed/span
}

func (m *Motion) AIs(now time.Time) []protocol.AI {
	if len(m.frames) == 0 {
		return nil
	}
	latest := m.frames[len(m.frames)-1]
	before, after, alpha := m.aiFramesAt(now)
	beforeByID := indexAIs(before.AIs)
	afterByID := indexAIs(after.AIs)
	ais := make([]protocol.AI, 0, len(latest.AIs))
	for _, ai := range latest.AIs {
		ai.Split = ai.Split.Clone()
		a, okA := beforeByID[ai.ID]
		b, okB := afterByID[ai.ID]
		if okA && okB {
			ai.X, ai.Y = projectPosition(a.X, a.Y, b.X, b.Y, alpha)
		}
		dx, dy := m.aiCorrections[ai.ID].offset(now)
		ai.X += dx
		ai.Y += dy
		for i := range ai.Split {
			cell := &ai.Split[i]
			if !cell.Active {
				continue
			}
			if okA && okB && i < len(a.Split) && i < len(b.Split) && a.Split[i].Active && b.Split[i].Active {
				cell.X, cell.Y = projectPosition(a.Split[i].X, a.Split[i].Y, b.Split[i].X, b.Split[i].Y, alpha)
			}
			cell.X += dx
			cell.Y += dy
			simulation.ClampSurfaceBody(&cell.X, &cell.Y, ai.Size, cell.Stats.HP)
		}
		simulation.ClampSurfaceBody(&ai.X, &ai.Y, ai.Size, ai.Stats.HP)
		ais = append(ais, ai)
	}
	return ais
}

func constrainPlayerBody(p *protocol.Player) {
	simulation.ClampSurfaceBody(&p.X, &p.Y, p.Size, p.Stats.HP)
	for i := range p.Split {
		if p.Split[i].Active {
			cell := &p.Split[i]
			simulation.ClampSurfaceBody(&cell.X, &cell.Y, p.Size, cell.Stats.HP)
		}
	}
}

func indexAIs(ais []protocol.AI) map[string]protocol.AI {
	byID := make(map[string]protocol.AI, len(ais))
	for _, ai := range ais {
		byID[ai.ID] = ai
	}
	return byID
}

func projectPosition(ax, ay, bx, by, alpha float64) (float64, float64) {
	dx, dy := (bx-ax)*(alpha-1), (by-ay)*(alpha-1)
	distance := math.Hypot(dx, dy)
	if distance == 0 {
		return bx, by
	}
	return simulation.MoveBy(bx, by, simulation.Direction{X: dx / distance, Y: dy / distance}, distance)
}

func playerMovementSpeed(p protocol.Player) float64 {
	return simulation.PlayerMovementSpeed(p.Character, p.Stats.Speed, p.Stats.HP, p.Split)
}
