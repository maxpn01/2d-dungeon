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
	renderedAIs []protocol.AI
	motion      client.Motion
	rendered    []protocol.Player
	session     *client.Session
	view        client.Update
	gameMap     *Map
	errorText   string
	chat        chatBox
	eating      *eatingScene
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
		if update.Local.Dead && m.eating == nil {
			m.eating = newEatingScene(time.Now(), update, m.rendered, m.renderedAIs)
		}
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
		direction.Sprint = rl.IsKeyDown(rl.KeyLeftShift)
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
		m.motion.Advance(predicted, float64(dt), direction.Sprint)
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
		rl.DrawRectangleV(rl.NewVector2(float32(fruit.X), float32(fruit.Y)), rl.NewVector2(float32(fruit.Size), float32(fruit.Size)), rl.Yellow)
	}
	for _, ai := range m.renderedAIs {
		rl.DrawRectangleV(rl.NewVector2(float32(ai.X), float32(ai.Y)), rl.NewVector2(float32(ai.Size), float32(ai.Size)), rl.Green)
		rl.DrawText(fmt.Sprintf("%s Lv.%d %s", ai.ID, ai.Stats.Level, ai.Behavior), int32(ai.X), int32(ai.Y)-22, 18, rl.RayWhite)
	}
	for _, player := range m.rendered {
		tint := rl.SkyBlue
		label := fmt.Sprintf("%s Lv.%d", protocol.NormalizePlayerName(player.Name), player.Stats.Level)
		if player.ID == m.view.PlayerID {
			tint = rl.Red
			label += " (you)"
		}
		rl.DrawRectangleV(rl.NewVector2(float32(player.X), float32(player.Y)), rl.NewVector2(float32(player.Size), float32(player.Size)), tint)
		if !player.Dead {
			rl.DrawText(label, int32(player.X), int32(player.Y)-22, 18, rl.RayWhite)
		}
	}
	rl.EndMode2D()
	drawDungeonLocation(camera.Target.X, camera.Target.Y, window.width, window.height)
	alive := 0
	for _, player := range m.view.Players {
		if !player.Dead {
			alive++
		}
	}
	title := fmt.Sprintf("Multiplayer | %s | Players: %d | Fruit: %d | AIs: %d", protocol.NormalizePlayerName(m.view.Local.Name), alive, len(m.view.Fruits), len(m.view.AIs))
	statsLine := ""
	for _, player := range m.view.Players {
		if player.ID != m.view.PlayerID {
			continue
		}
		stats := player.Stats
		progress := fmt.Sprintf("XP %.2f / %.2f", stats.XP, simulation.XPForNextLevel(stats.Level))
		if stats.Level == simulation.MaxLevel {
			progress = "MAX LEVEL"
		}
		statsLine = fmt.Sprintf("Level %d | %s | HP %.2f | Speed %.2f", stats.Level, progress, stats.HP, stats.Speed)
		break
	}
	drawStatusPanel(title, statsLine, "WASD to move | Shift to sprint | M for map | T to chat | Eat fruit, weaker players and AIs to grow | Esc to leave", window.width)
	m.drawMinimap(window.width)
	m.chat.draw(window, m.session.Chat())
}
