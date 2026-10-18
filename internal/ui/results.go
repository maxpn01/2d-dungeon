package ui

import (
	"fmt"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/preferences"
	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ResultsPanel struct {
	playAgain         *MenuButton
	resultsMenu       *MenuButton
	continueExploring *MenuButton
	showResults       *MenuButton
}

func NewResultsPanel() *ResultsPanel {
	return &ResultsPanel{
		playAgain:         NewMenuButton("Play again", ActionPlayAgain),
		resultsMenu:       NewMenuButton("Back to menu", ActionMenu),
		continueExploring: NewMenuButton("Continue exploring", ActionContinueExploring),
		showResults:       NewMenuButton("Run results", ActionShowResults),
	}
}

func (p *ResultsPanel) Draw(snapshot simulation.Snapshot, local simulation.Player, playerName string, width, height int32) Action {
	y := height/2 - 245
	rl.DrawRectangle(0, 0, width, height, rl.Fade(rl.Black, .65))
	rl.DrawRectangle((width-640)/2, y-24, 640, 510, rl.NewColor(18, 22, 28, 255))
	title, tint := "FLOOR I COMPLETE", rl.Gold

	primitives.DrawCenteredText(title, y, 36, tint)
	primitives.DrawCenteredText("You discovered the way to the next floor.", y+52, 20, rl.RayWhite)
	seconds := snapshot.FinishedTick / parameters.TickRate
	primitives.DrawCenteredText(fmt.Sprintf("Run time  %d:%02d", seconds/60, seconds%60), y+88, 20, rl.Gray)
	name := preferences.NormalizePlayerName(playerName) + " (you)"
	status := ""
	rowTint := rl.LightGray
	if local.ID == snapshot.WinnerID {
		status = " / Escaped"
		rowTint = rl.Gold
	}
	primitives.DrawCenteredText(fmt.Sprintf("%s / Lv.%d%s", name, local.Stats.Level, status), y+138, 20, rowTint)
	action := ActionNone
	p.playAgain.Bounds = rl.NewRectangle(float32(width-260)/2, float32(y+228), 260, 56)
	action = p.playAgain.Draw()
	p.continueExploring.Bounds = rl.NewRectangle(float32(width-300)/2, float32(y+300), 300, 56)
	if clicked := p.continueExploring.Draw(); clicked != ActionNone {
		action = clicked
	}
	p.resultsMenu.Bounds = rl.NewRectangle(float32(width-260)/2, float32(y+372), 260, 56)
	if clicked := p.resultsMenu.Draw(); clicked != ActionNone {
		action = clicked
	}
	return action
}

func (p *ResultsPanel) DrawButton(width int32) Action {
	p.showResults.Bounds = rl.NewRectangle(float32(width-220)/2, 110, 220, 44)
	p.showResults.FontSize = 20
	return p.showResults.Draw()
}
