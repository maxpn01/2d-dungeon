package game

import (
	"fmt"
	"log"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"raylib-2d-sandbox/internal/client"
	"raylib-2d-sandbox/internal/protocol"
)

type multiplayer struct {
	motion    client.Motion
	rendered  []protocol.Player
	session   *client.Session
	view      client.Update
	gameMap   *Map
	errorText string
}

func newMultiplayer(url string) *multiplayer {
	return &multiplayer{
		session: client.Connect(url),
		gameMap: NewMap(rl.NewVector2(200, 200), rl.NewVector2(10, 10), rl.Black, rl.Gray),
	}
}

// poll is called only by the render loop; networking never touches game state.
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
		m.view = update
		m.motion.Apply(update)
	default:
	}
}

func (m *multiplayer) readInput(dt float32) {
	direction := protocol.Direction{}
	if m.view.Ready && m.errorText == "" && rl.IsWindowFocused() {
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
	m.motion.Advance(direction, float64(dt), now)
	m.rendered = m.motion.Players(now)
}

func (m *multiplayer) follow(camera *rl.Camera2D) {
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
	for _, player := range m.rendered {
		tint := rl.SkyBlue
		label := player.ID
		if player.ID == m.view.PlayerID {
			tint = rl.Red
			label += " (you)"
		}
		rl.DrawRectangleV(rl.NewVector2(float32(player.X), float32(player.Y)), rl.NewVector2(float32(player.Size), float32(player.Size)), tint)
		rl.DrawText(label, int32(player.X), int32(player.Y)-22, 18, rl.RayWhite)
	}
	rl.EndMode2D()
	rl.DrawRectangle(0, 0, window.width, 88, rl.Fade(rl.Black, .8))
	rl.DrawText(fmt.Sprintf("Multiplayer | %s | Players: %d", m.view.PlayerID, len(m.view.Players)), 24, 20, 22, rl.RayWhite)
	rl.DrawText("WASD to move | Esc to leave | Movement-only world", 24, 52, 20, rl.LightGray)
}
