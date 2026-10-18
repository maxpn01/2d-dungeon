package game

import (
	"fmt"

	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"
	"2d-dungeon/internal/ui"
	"2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (w *World) EnableTesting()       { w.simulation.EnableTesting(); w.syncSimulation() }
func (w *World) TestingEnabled() bool { return w.simulation.TestingEnabled() }

func (w *World) ReadTestingInput() bool {
	if !w.TestingEnabled() || !rl.IsWindowFocused() {
		return false
	}
	changed := true
	switch {
	case rl.IsKeyPressed(rl.KeyF1):
		w.simulation.ToggleTestingGod()
	case rl.IsKeyPressed(rl.KeyF2):
		w.simulation.TestingWarp(worldmap.OrbVault)
	case rl.IsKeyPressed(rl.KeyF3):
		w.simulation.TestingWarp(worldmap.ColumnHall)
	case rl.IsKeyPressed(rl.KeyF4):
		w.simulation.TestingWarp(worldmap.TeleportSanctum)
	case rl.IsKeyPressed(rl.KeyF5):
		w.simulation.TestingWarp(worldmap.Forest)
	case rl.IsKeyPressed(rl.KeyF6):
		w.simulation.CycleTestingLevel()
	case rl.IsKeyPressed(rl.KeyF8):
		w.simulation.TestingWarp(worldmap.VaultCourt)
	case rl.IsKeyPressed(rl.KeyF9):
		w.simulation.TestingWarp(worldmap.SecretLibrary)
	case rl.IsKeyPressed(rl.KeyF7):
		w.simulation.ToggleTestingWings()
	case rl.IsKeyPressed(rl.KeyF10):
		w.simulation.TestingWarpCave()
	case rl.IsKeyPressed(rl.KeyF11):
		w.simulation.GrowTestingHydra()
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
	if w.simulation.GodMode() {
		state = "ON"
	}
	x, y := int32(18), height-245
	rl.DrawRectangle(x-8, y-8, 405, 201, rl.Fade(rl.NewColor(12, 17, 27, 255), .9))
	primitives.DrawText(fmt.Sprintf("TESTING / F1 God: %s / Lv.%d", state, w.player.Stats.Level), x, y, 18, rl.Gold)
	lines := []string{
		"F2 Reset + enter orb vault",
		"F8 Reset + enter gate courtyard",
		"F9 Reset + enter haunted library",
		"F3 Dungeon   F4 Descent + orb   F5 Floor 2",
		"F6 Cycle level   F7 Toggle wings",
		"F10 Reset hydra + enter cave   F11 Hydra +3 Lv",
		"God mode: no damage, unlimited mana",
	}
	for i, line := range lines {
		primitives.DrawText(line, x, y+29+int32(i)*22, 16, rl.LightGray)
	}
}
