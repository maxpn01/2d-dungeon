package game

import (
	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/simulation"
	"2d-dungeon/internal/ui"
)

func (w *World) Finished() bool { return w.snapshot.WinnerID != "" }
func (w *World) DrawResults(width, height int32) ui.Action {
	return w.results.Draw(w.snapshot, w.snapshot.Players[0], w.name, width, height)
}

func (p *presentation) resultsOpen(snapshot simulation.Snapshot) bool {
	return snapshot.WinnerID != "" && (!p.resultsDismissed && snapshot.Tick >= snapshot.FinishedTick+parameters.ResultsDelaySeconds*simulation.TickRate)
}
func (p *presentation) ContinueExploring() { p.resultsDismissed = true }
func (p *presentation) ShowResults()       { p.resultsDismissed = false }

func (w *World) ResultsOpen() bool { return w.resultsOpen(w.snapshot) }

func (w *World) DrawResultsButton(width int32) ui.Action {
	if !w.resultsDismissed || !w.Finished() || w.ResultsOpen() {
		return ui.ActionNone
	}
	return w.results.DrawButton(width)
}
