package game

import (
	"raylib-2d-sandbox/internal/rendering"
	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type screen int

const (
	screenMenu screen = iota
	screenSinglePlayer
	screenSettings
	screenBriefing
)

func (g *Game) run(testing bool) {
	g.window.initWindow()
	defer g.window.close()
	rendering.InitFeedbackAudio()
	defer rendering.CloseFeedbackAudio()
	defer primitives.UnloadTextFonts()
	g.cameraFollowPlayer()
	current := screenMenu
	if testing {
		g.testWorld = NewWorld()
		g.testWorld.Configure(g.settings.PlayerName, g.settings.Character)
		g.testWorld.EnableTesting()
		g.world = g.testWorld
		g.startSinglePlayer()
		current = screenSinglePlayer
	}
	mapOpen := false
	exitRequested := false
	for !rl.WindowShouldClose() && !exitRequested {
		dt := rl.GetFrameTime()
		singlePlayerDeadAtFrameStart := current == screenSinglePlayer && g.world.Dead()
		if current == screenSinglePlayer && g.world.ReadTestingInput() {
			mapOpen = false
			g.cameraFollowPlayer()
		}
		if rl.IsKeyPressed(rl.KeyF8) {
			rendering.ToggleFeedbackAudio()
		}
		if rl.IsKeyPressed(rl.KeyEscape) {
			if current == screenSinglePlayer && g.world.StatsOpen() {
				g.world.CloseStats()
			} else if current == screenSinglePlayer && g.world.PerksOpen() {
				g.world.ClosePerks()
			} else if mapOpen {
				mapOpen = false
			} else {
				switch current {
				case screenSettings:
					current = screenMenu
				case screenBriefing:
					current = screenMenu
				case screenSinglePlayer:
					if !g.world.Dead() {
						current = screenMenu
					}
				case screenMenu:
					if !g.world.Dead() {
						current = g.enterSinglePlayer()
					}
				}
			}
		}
		if rl.IsKeyPressed(rl.KeyP) && !rl.IsKeyPressed(rl.KeyEscape) && rl.IsWindowFocused() && !mapOpen {
			switch current {
			case screenSinglePlayer:
				if !g.world.Dead() && !g.world.ResultsOpen() && !g.world.PerksOpen() {
					g.world.ToggleStats()
				}
			}
		}
		if rl.IsKeyPressed(rl.KeyM) && !rl.IsKeyPressed(rl.KeyEscape) && rl.IsWindowFocused() {
			switch current {
			case screenSinglePlayer:
				if !g.world.Dead() && !g.world.ResultsOpen() && !g.world.PerksOpen() && !g.world.StatsOpen() {
					mapOpen = !mapOpen
				}
			}
		}
		switch current {
		case screenSinglePlayer:
			if !g.world.Dead() && !g.world.ResultsOpen() && !mapOpen {
				g.world.ReadInput(g.camera, g.window.width, g.window.height)
				g.world.Update(dt)
				g.cameraFollowPlayer()
			}
		}

		rl.BeginDrawing()
		rl.ClearBackground(g.window.bgColor)
		action := ui.ActionNone
		switch current {
		case screenSettings:
			action = g.settings.Draw()
		case screenMenu:
			action = g.menu.Draw()
		case screenBriefing:
			action = g.briefing.Draw(g.window.width, g.window.height)
		case screenSinglePlayer:
			if g.world.ResultsOpen() {
				g.world.Draw(g.camera, g.window.width, g.window.height)
				action = g.world.DrawResults(g.window.width, g.window.height)
			} else if g.world.Dead() {
				action = g.gameOver.Draw()
				if !singlePlayerDeadAtFrameStart {
					action = ui.ActionNone
				}
			} else if mapOpen {
				g.drawFullMap()
			} else {
				g.world.Draw(g.camera, g.window.width, g.window.height)
				if g.world.Finished() {
					action = g.world.DrawResultsButton(g.window.width)
				}
			}
		}
		rl.EndDrawing()

		if action != ui.ActionNone {
			mapOpen = false
		}

		switch action {
		case ui.ActionEnterDungeon:
			g.soloBriefed = true
			g.startSinglePlayer()
			current = screenSinglePlayer
		case ui.ActionContinueExploring:
			g.world.ContinueExploring()
		case ui.ActionShowResults:
			g.world.ShowResults()
		case ui.ActionPlayAgain:
			g.startSinglePlayer()
		case ui.ActionSettings:
			g.settings.Open()
			current = screenSettings
		case ui.ActionSaveSettings:
			g.world.SetName(g.settings.PlayerName)
			current = screenMenu
		case ui.ActionPlay:
			g.world = g.normalWorld
			current = g.enterSinglePlayer()
		case ui.ActionRestart:
			g.startSinglePlayer()
			current = screenSinglePlayer
		case ui.ActionMenu:
			current = screenMenu
		case ui.ActionExit:
			exitRequested = true
		}
	}
}
