package game

import (
	"fmt"

	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"
	"2d-dungeon/internal/ui"
	"2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (w *World) EnableTesting()       { w.simulation.EnableTesting(localPlayerID); w.syncSimulation() }
func (w *World) TestingEnabled() bool { return w.simulation.TestingEnabled(localPlayerID) }

func (w *World) ReadTestingInput() bool {
	if !w.TestingEnabled() || !rl.IsWindowFocused() {
		return false
	}
	changed := true
	switch {
	case rl.IsKeyPressed(rl.KeyF1):
		w.simulation.ToggleTestingGod(localPlayerID)
	case rl.IsKeyPressed(rl.KeyF2):
		w.simulation.TestingWarp(localPlayerID, worldmap.OrbVault)
	case rl.IsKeyPressed(rl.KeyF3):
		w.simulation.TestingWarp(localPlayerID, worldmap.ColumnHall)
	case rl.IsKeyPressed(rl.KeyF4):
		w.simulation.TestingWarp(localPlayerID, worldmap.TeleportSanctum)
	case rl.IsKeyPressed(rl.KeyF5):
		w.simulation.TestingWarp(localPlayerID, worldmap.FeastHall)
	case rl.IsKeyPressed(rl.KeyF6):
		w.simulation.CycleTestingLevel(localPlayerID)
	case rl.IsKeyPressed(rl.KeyF8):
		w.simulation.TestingWarp(localPlayerID, worldmap.VaultCourt)
	case rl.IsKeyPressed(rl.KeyF9):
		w.simulation.TestingWarp(localPlayerID, worldmap.SecretLibrary)
	case rl.IsKeyPressed(rl.KeyF7):
		w.simulation.ToggleTestingWings(localPlayerID)
	default:
		changed = false
	}
	if changed {
		w.ClosePerks()
		w.CloseStats()
		w.perks = ui.PerkPanel{}
		w.direction = simulation.Direction{}
		w.accumulator = 0
		w.syncSimulation()
		w.previousSnapshot = w.snapshot
		if w.snapshot.WinnerID == "" {
			w.resultsDismissed = false
		}
		w.gameMap.Exploration.SetSecrets(w.snapshot.OpenSecrets, w.snapshot.Orb.GateOpen)
		w.gameMap.Exploration.Reveal(w.Center(), w.player.HasWings, w.player.Flight.Flying)
	}
	return changed
}

func (w *World) drawTesting(width, height int32) {
	if !w.TestingEnabled() {
		return
	}
	state := "OFF"
	if w.simulation.GodMode(localPlayerID) {
		state = "ON"
	}
	x, y := int32(18), height-223
	rl.DrawRectangle(x-8, y-8, 405, 179, rl.Fade(rl.NewColor(12, 17, 27, 255), .9))
	primitives.DrawText(fmt.Sprintf("TESTING / F1 God: %s / Lv.%d", state, w.player.Stats.Level), x, y, 18, rl.Gold)
	lines := []string{
		"F2 Reset + enter orb vault",
		"F8 Reset + enter gate courtyard",
		"F9 Reset + enter haunted library",
		"F3 Dungeon   F4 Descent + orb   F5 Floor 2",
		"F6 Cycle level   F7 Toggle wings",
		"God mode: no damage, unlimited mana",
	}
	for i, line := range lines {
		primitives.DrawText(line, x, y+29+int32(i)*22, 16, rl.LightGray)
	}
}
