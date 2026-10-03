package gameplay

import (
	"fmt"

	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/screens"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (m *Multiplayer) DrawLobby(width, height int32) screens.Action {
	y := height/2 - 180
	primitives.DrawCenteredText("DUNGEON RACE", y, 36, rl.Gold)
	primitives.DrawCenteredText("Explore the dungeon and reach the next floor first.", y+52, 20, rl.LightGray)
	for i := range simulation.MaxPlayers {
		label := "Waiting for a player..."
		tint := rl.Gray
		if i < len(m.view.Players) {
			player := m.view.Players[i]
			name := player.Name
			if player.ID == m.view.PlayerID {
				name += " (you)"
			}
			status := "Not ready"
			tint = rl.RayWhite
			if player.LobbyReady {
				status, tint = "Ready", rl.Green
			}
			label = name + " / " + status
		}
		primitives.DrawCenteredText(label, y+110+int32(i)*42, 24, tint)
	}
	action := screens.ActionNone
	if m.view.StartCountdown > 0 {
		primitives.DrawCenteredText(fmt.Sprintf("Starting in %d...", m.view.StartCountdown), y+224, 30, rl.Gold)
	} else if m.view.Local.LobbyReady {
		primitives.DrawCenteredText("Waiting for the other player to start...", y+224, 20, rl.LightGray)
	} else {
		m.lobbyStart.Bounds = rl.NewRectangle(float32(width-240)/2, float32(y+210), 240, 60)
		action = m.lobbyStart.Draw()
	}
	primitives.DrawCenteredText("Both players must press Start game. Esc to leave.", y+298, 18, rl.Gray)
	m.chat.Draw(width, height, m.session.Chat())
	return action
}
