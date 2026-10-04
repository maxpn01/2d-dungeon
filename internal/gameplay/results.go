package gameplay

import (
	"fmt"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/screens"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (p *presentation) drawResults(snapshot protocol.Snapshot, local protocol.Player, width, height int32, multiplayer bool) screens.Action {
	y := height/2 - 245
	rl.DrawRectangle(0, 0, width, height, rl.Fade(rl.Black, .65))
	rl.DrawRectangle((width-640)/2, y-24, 640, 510, rl.NewColor(18, 22, 28, 255))
	title, tint := "ROUND COMPLETE", rl.Gold
	if multiplayer {
		title, tint = "YOU WON", rl.Gold
		if snapshot.WinnerID != local.ID {
			title, tint = "ROUND LOST", rl.LightGray
		}
	}
	primitives.DrawCenteredText(title, y, 36, tint)
	winner := "The other player"
	for _, player := range snapshot.Players {
		if player.ID == snapshot.WinnerID {
			winner = protocol.NormalizePlayerName(player.Name)
			break
		}
	}
	primitives.DrawCenteredText(winner+" reached the next floor first.", y+52, 20, rl.RayWhite)
	seconds := snapshot.FinishedTick / simulation.TickRate
	primitives.DrawCenteredText(fmt.Sprintf("Round time  %d:%02d", seconds/60, seconds%60), y+88, 20, rl.Gray)
	for i, player := range snapshot.Players {
		name := protocol.NormalizePlayerName(player.Name)
		if player.ID == local.ID {
			name += " (you)"
		}
		status := ""
		rowTint := rl.LightGray
		if player.ID == snapshot.WinnerID {
			status = " / Winner"
			rowTint = rl.Gold
		}
		if multiplayer && player.LobbyReady {
			status += " / Rematch ready"
			rowTint = rl.Green
		}
		primitives.DrawCenteredText(fmt.Sprintf("%s / Lv.%d%s", name, player.Stats.Level, status), y+138+int32(i)*36, 20, rowTint)
	}
	action := screens.ActionNone
	switch {
	case multiplayer && snapshot.StartCountdown > 0:
		primitives.DrawCenteredText(fmt.Sprintf("Next round in %d...", snapshot.StartCountdown), y+238, 28, rl.Gold)
	case multiplayer && local.LobbyReady:
		label := "Waiting for the other player to rematch..."
		if len(snapshot.Players) < simulation.MaxPlayers {
			label = "Waiting for another player to join..."
		}
		primitives.DrawCenteredText(label, y+246, 20, rl.LightGray)
	default:
		p.rematch.Text = "Rematch"
		if !multiplayer {
			p.rematch.Text = "Play again"
		}
		p.rematch.Bounds = rl.NewRectangle(float32(width-260)/2, float32(y+228), 260, 56)
		action = p.rematch.Draw()
	}
	if snapshot.StartCountdown == 0 {
		p.continueExploring.Bounds = rl.NewRectangle(float32(width-300)/2, float32(y+300), 300, 56)
		if clicked := p.continueExploring.Draw(); clicked != screens.ActionNone {
			action = clicked
		}
	}
	p.resultsMenu.Bounds = rl.NewRectangle(float32(width-260)/2, float32(y+372), 260, 56)
	if clicked := p.resultsMenu.Draw(); clicked != screens.ActionNone {
		action = clicked
	}
	if multiplayer {
		primitives.DrawCenteredText("Both players must press Rematch. Esc to leave.", y+450, 18, rl.Gray)
	}
	return action
}

func (m *Multiplayer) Finished() bool { return m.view.WinnerID != "" }
func (m *Multiplayer) Rematch()       { m.session.Rematch() }
func (m *Multiplayer) DrawResults(width, height int32) screens.Action {
	action := m.presentation.drawResults(m.view.Snapshot, m.view.Local, width, height, true)
	m.chat.Draw(width, height, m.session.Chat())
	return action
}
func (w *World) Finished() bool { return w.snapshot.WinnerID != "" }
func (w *World) DrawResults(width, height int32) screens.Action {
	return w.presentation.drawResults(w.snapshot, w.snapshot.Players[0], width, height, false)
}

func (p *presentation) resultsOpen(snapshot protocol.Snapshot) bool {
	return snapshot.WinnerID != "" && (snapshot.StartCountdown > 0 || !p.resultsDismissed && snapshot.Tick >= snapshot.FinishedTick+parameters.ResultsDelaySeconds*simulation.TickRate)
}
func (p *presentation) ContinueExploring() { p.resultsDismissed = true }
func (p *presentation) ShowResults()       { p.resultsDismissed = false }
func (p *presentation) drawResultsButton(width int32) screens.Action {
	p.showResults.Bounds = rl.NewRectangle(float32(width-220)/2, 110, 220, 44)
	p.showResults.FontSize = 20
	return p.showResults.Draw()
}
func (m *Multiplayer) ResultsOpen() bool { return m.resultsOpen(m.view.Snapshot) }
func (w *World) ResultsOpen() bool       { return w.resultsOpen(w.snapshot) }
func (m *Multiplayer) DrawResultsButton(width int32) screens.Action {
	if m.view.Tick < m.view.FinishedTick+parameters.ResultsDelaySeconds*simulation.TickRate {
		return screens.ActionNone
	}
	return m.drawResultsButton(width)
}
func (w *World) DrawResultsButton(width int32) screens.Action {
	if w.snapshot.Tick < w.snapshot.FinishedTick+parameters.ResultsDelaySeconds*simulation.TickRate {
		return screens.ActionNone
	}
	return w.drawResultsButton(width)
}
