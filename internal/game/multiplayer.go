package game

import (
	"fmt"
	"log"
	"time"

	"raylib-2d-sandbox/internal/client"
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type multiplayer struct {
	feedback    combatFeedback
	renderedAIs []protocol.AI
	motion      client.Motion
	rendered    []protocol.Player
	session     *client.Session
	view        client.Update
	gameMap     *Map
	errorText   string
	chat        chatBox
	eating      *eatingScene
	otherEating []*eatingScene
}

func newMultiplayer(url, name string) *multiplayer {
	return &multiplayer{
		session: client.ConnectWithName(url, name),
		gameMap: NewMap(rl.NewVector2(200, 200), rl.NewVector2(10, 10), rl.Black, rl.Gray),
	}
}

func (m *multiplayer) poll() {
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
			}
			return
		}
		m.feedback.multiplayer(time.Now(), update.Players, update.AIs, update.Fruits, update.PlayerID)
		m.startEatingScenes(time.Now(), update)
		m.view = update
		m.motion.Apply(update)
	default:
	}
}

func (m *multiplayer) readInput(dt float32, mapOpen bool) {
	if !mapOpen {
		m.chat.update(m)
	}
	direction := protocol.Direction{}
	if m.view.Ready && !m.view.Local.Dead && m.errorText == "" && rl.IsWindowFocused() && !m.chat.open && !mapOpen {
		direction.Dash = rl.IsKeyPressed(rl.KeyLeftShift)
		if rl.IsKeyDown(rl.KeyW) {
			direction.Y--
		}
		if rl.IsKeyDown(rl.KeyS) {
			direction.Y++
		}
		if rl.IsKeyDown(rl.KeyA) {
			direction.X--
		}
		if rl.IsKeyDown(rl.KeyD) {
			direction.X++
		}
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
				m.gameMap.exploration.reveal(rl.NewVector2(float32(player.X+player.Size/2), float32(player.Y+player.Size/2)))
				break
			}
		}
	}
	active := m.otherEating[:0]
	for _, scene := range m.otherEating {
		if !scene.finished(now) {
			m.rendered, m.renderedAIs = scene.apply(now, m.rendered, m.renderedAIs)
			active = append(active, scene)
		}
	}
	m.otherEating = active
	if m.eating != nil {
		m.rendered, m.renderedAIs = m.eating.apply(now, m.rendered, m.renderedAIs)
	}
}

func (m *multiplayer) follow(camera *rl.Camera2D) {
	if m.eating != nil {
		return
	}
	for _, player := range m.rendered {
		if player.ID == m.view.PlayerID {
			camera.Target = rl.NewVector2(float32(player.X+player.Size/2), float32(player.Y+player.Size/2))
			return
		}
	}
}

func (m *multiplayer) draw(camera rl.Camera2D, window *Window) {
	rl.BeginMode2D(camera)
	m.gameMap.draw(RenderContext{Camera: camera, WindowWidth: window.width, WindowHeight: window.height})
	for _, fruit := range m.view.Fruits {
		if !m.gameMap.exploration.entityVisible(fruit.X, fruit.Y, fruit.Size) {
			continue
		}
		drawFruit(float32(fruit.X), float32(fruit.Y), float32(fruit.Size), massColor(fruit.Source, m.view.PlayerID), fruit.Source != "")
	}
	for _, ai := range m.renderedAIs {
		if !m.gameMap.exploration.entityVisible(ai.X, ai.Y, ai.Size) {
			continue
		}
		bounds := drawActor(ai.X, ai.Y, ai.Size, ai.Stats.HP, ai.Combat, rl.Green)
		drawTextAt(fmt.Sprintf("%s Lv.%d %s", ai.ID, ai.Stats.Level, ai.Behavior), rl.NewVector2(bounds.X, bounds.Y-22), 18, rl.RayWhite)
	}
	for _, player := range m.rendered {
		if player.ID != m.view.PlayerID && !m.gameMap.exploration.entityVisible(player.X, player.Y, player.Size) {
			continue
		}
		tint := rl.SkyBlue
		label := fmt.Sprintf("%s Lv.%d", protocol.NormalizePlayerName(player.Name), player.Stats.Level)
		if player.ID == m.view.PlayerID {
			tint = rl.Red
			label += " (you)"
		}
		bounds := drawActor(player.X, player.Y, player.Size, player.Stats.HP, player.Combat, tint)
		if !player.Dead {
			drawTextAt(label, rl.NewVector2(bounds.X, bounds.Y-22), 18, rl.RayWhite)
		}
	}
	m.feedback.draw(time.Now(), &m.gameMap.exploration)
	rl.EndMode2D()
	drawDungeonLocation(camera.Target.X, camera.Target.Y, window.width, window.height)
	alive := 0
	for _, player := range m.view.Players {
		if !player.Dead {
			alive++
		}
	}
	stats := m.view.Local.Stats
	hp := float32(stats.HP)
	if m.view.Local.Dead {
		hp = 0
	}
	drawPlayerHUD(playerHUD{
		name: m.view.Local.Name, avatar: rl.Red,
		level: stats.Level, maxLevel: simulation.MaxLevel,
		hp: hp, maxHP: float32(simulation.MaxHPForLevel(stats.Level)),
		xp: float32(stats.XP), nextXP: float32(simulation.XPForNextLevel(stats.Level)),
		speed: float32(stats.Speed), fruit: len(m.view.Fruits), ais: len(m.view.AIs), players: alive,
	}, window.width)
	m.drawMinimap(window.width)
	m.chat.draw(window, m.session.Chat())
}

func (m *multiplayer) startEatingScenes(now time.Time, update client.Update) {
	if update.Local.Dead && m.eating == nil {
		m.eating = newEatingScene(now, update, m.rendered, m.renderedAIs)
	}
	for _, prey := range update.Players {
		if !prey.Dead || prey.ID == update.PlayerID {
			continue
		}
		for _, previous := range m.view.Players {
			if previous.ID == prey.ID && !previous.Dead {
				death := update
				death.Local, death.PlayerID = prey, prey.ID
				m.otherEating = append(m.otherEating, newEatingScene(now, death, m.rendered, m.renderedAIs))
				break
			}
		}
	}
}
