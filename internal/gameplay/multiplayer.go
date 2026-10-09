package gameplay

import (
	"log"
	"strings"
	"time"

	"raylib-2d-sandbox/internal/client"
	"raylib-2d-sandbox/internal/hud"
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/rendering"
	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/screens"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Multiplayer struct {
	presentation
	lobbyStart          *screens.MenuButton
	controls            controls
	previousProjectiles []simulation.Projectile
	projectileInterval  float64
	renderedAIs         []protocol.AI
	motion              client.Motion
	rendered            []protocol.Player
	session             *client.Session
	view                client.Update
	errorText           string
	chat                hud.ChatBox
}

func NewMultiplayer(url, name string, character simulation.Character) *Multiplayer {
	return &Multiplayer{
		lobbyStart:   screens.NewMenuButton("Start game", screens.ActionStartGame),
		controls:     controls{suppressShootUntilRelease: true},
		session:      client.ConnectWithCharacter(url, name, character),
		presentation: newPresentation(),
	}
}

func (m *Multiplayer) Poll() {
	if m.errorText != "" {
		return
	}
	select {
	case update, open := <-m.session.Updates():
		if !open {
			m.errorText = "Connection closed. Please try again."
			return
		}
		if update.Err != nil {
			log.Printf("multiplayer: %v", update.Err)
			if m.view.Ready {
				m.errorText = "Connection lost. The server may be unavailable."
			} else {
				m.errorText = "Could not join the server. Check that it is running."
				if strings.Contains(update.Err.Error(), "dungeon is full") {
					m.errorText = "The dungeon is full (2 players). Try again when a spot opens."
				}
			}
			return
		}
		if m.view.Ready && update.RoundID != m.view.RoundID {
			m.presentation = newPresentation()
			m.motion = client.Motion{}
			m.previousProjectiles = nil
			m.controls = controls{suppressShootUntilRelease: true}
		}
		m.feedback.Observe(time.Now(), update.Snapshot, update.PlayerID)
		if update.RoundID == m.view.RoundID {
			m.previousProjectiles = m.view.Projectiles
		}
		m.projectileInterval = simulation.StepSeconds
		if update.Tick > m.view.Tick && m.view.Ready {
			m.projectileInterval = float64(update.Tick-m.view.Tick) / simulation.TickRate
		}
		m.view = update
		m.motion.Apply(update)
	default:
	}
}

func (m *Multiplayer) ReadInput(dt float32, mapOpen bool, camera rl.Camera2D) {
	m.controls.releaseShoot()
	if m.ResultsOpen() {
		m.CloseStats()
		m.perks = hud.PerkPanel{}
		m.controls.suppressShootUntilRelease = true
		m.session.SetInput(protocol.Direction{})
		m.chat.Update(m.view.Ready, false, m.errorText, m.session.SendChat)
		return
	}
	s := m.perks.Sync(m.view.Local.ActiveStatue, m.view.Statues)
	if !mapOpen && !m.StatsOpen() && s == nil {
		m.chat.Update(m.view.Ready, m.view.Local.Dead, m.errorText, m.session.SendChat)
	}
	if s != nil {
		m.CloseStats()
		m.chat.Close()
		m.controls.suppressShootUntilRelease = true
		if card := m.perks.Read(*s, m.view.Local.Stats, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight())); card >= 0 {
			if worldmap.IsWingStatue(s.ID) && card == 1 {
				m.ClosePerks()
			} else {
				m.session.ChoosePerk(s.ID, s.Cycle, card)
			}
		}
	}
	if m.StatsOpen() {
		m.controls.suppressShootUntilRelease = true
	}
	direction := protocol.Direction{}
	if m.view.Ready && m.view.MatchStarted && !m.view.Local.Dead && m.errorText == "" && rl.IsWindowFocused() && !m.chat.Open && !mapOpen && !m.StatsOpen() && s == nil {

		center := rl.NewVector2(float32(m.view.Local.X+m.view.Local.Size/2), float32(m.view.Local.Y+m.view.Local.Size/2))
		for _, p := range m.rendered {
			if p.ID == m.view.PlayerID {
				center = rl.NewVector2(float32(p.X+p.Size/2), float32(p.Y+p.Size/2))
				break
			}
		}
		direction = m.controls.read(camera, center, m.view.Local.Combat.Facing)
		resourcePlayer := m.view.Local
		if predicted, ready := m.session.PredictedPlayer(); ready {
			resourcePlayer = predicted
		}
		m.feedback.CheckCombatInput(time.Now(), direction, resourcePlayer)
	}
	m.session.SetInput(direction)
	now := time.Now()
	if predicted, ready := m.session.PredictedPlayer(); ready {
		m.motion.Advance(predicted, float64(dt))
	}
	m.rendered = m.motion.Players(now)
	m.renderedAIs = m.motion.AIs(now)
	if m.view.Ready && !m.view.Local.Dead && m.errorText == "" {
		for _, player := range m.rendered {
			if player.ID == m.view.PlayerID {
				m.gameMap.Exploration.SetSecrets(m.view.OpenSecrets, m.view.Orb.GateOpen)
				m.gameMap.Exploration.Reveal(rl.NewVector2(float32(player.X+player.Size/2), float32(player.Y+player.Size/2)))
				break
			}
		}
	}

}

func (m *Multiplayer) Follow(camera *rl.Camera2D) {
	for _, player := range m.rendered {
		if player.ID == m.view.PlayerID {
			camera.Target = rl.NewVector2(float32(player.X+player.Size/2), float32(player.Y+player.Size/2))
			return
		}
	}
}

func (m *Multiplayer) scene() rendering.Scene {
	return rendering.Scene{
		Snapshot: m.view.Snapshot, Local: m.view.Local, RenderedPlayers: m.rendered, RenderedAIs: m.renderedAIs,
		PreviousProjectiles: m.previousProjectiles,
		ProjectileAlpha:     time.Since(m.view.ReceivedAt).Seconds() / max(simulation.StepSeconds, m.projectileInterval),
	}
}

func (m *Multiplayer) Draw(camera rl.Camera2D, width, height int32) {
	m.presentation.draw(primitives.Context{Camera: camera, WindowWidth: width, WindowHeight: height}, m.scene(), true)
	ping, measured := m.session.Ping()
	hud.DrawPing(ping, measured, width)
	m.chat.Draw(width, height, m.session.Chat())
	m.perks.Draw(m.view.Local.ActiveStatue, m.view.Statues, m.view.Local.Stats, width, height)
	if m.StatsOpen() {
		hud.DrawStats(m.view.Local, width, height)
	}
}

func (m *Multiplayer) DrawMap(width, height int32) {
	m.presentation.drawRadar(width, height, m.scene())
}

func (m *Multiplayer) ClosePerks() {
	m.perks.Closing = true
	m.controls.suppressShootUntilRelease = true
	m.session.CloseOffer()
}
