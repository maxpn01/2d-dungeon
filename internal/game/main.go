package game

import (
	"raylib-2d-sandbox/internal/gameplay"
	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/rendering"
	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/screens"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Game struct {
	window              *Window
	camera              rl.Camera2D
	settings            *screens.Settings
	menu                *screens.Menu
	gameOver            *screens.GameOver
	multiplayerGameOver *screens.GameOver
	world               *gameplay.World
	serverURL           string
	multiplayer         *gameplay.Multiplayer
	connectionMenu      *screens.Menu
	briefing            *screens.Briefing
	soloBriefed         bool
}

func NewGame() *Game {
	window := NewWindow(
		parameters.WindowWidth,
		parameters.WindowHeight,
		"2D sandbox",
		rl.Black,
	)

	camera := rl.NewCamera2D(window.getWindowCenter(), rl.Vector2{}, 0, 1)

	menu := screens.NewMenu([]*screens.MenuButton{
		screens.NewMenuButton("play", screens.ActionPlay),
		screens.NewMenuButton("multiplayer", screens.ActionMultiplayer),
		screens.NewMenuButton("settings", screens.ActionSettings),
		screens.NewMenuButton("exit", screens.ActionExit),
	}, 30, rl.Black)

	menu.Title = "2D SANDBOX"

	settings := screens.NewSettings()
	world := gameplay.NewWorld()
	world.Configure(settings.PlayerName, settings.Character)

	return &Game{
		settings:            settings,
		window:              window,
		camera:              camera,
		menu:                menu,
		gameOver:            screens.NewGameOver(),
		multiplayerGameOver: screens.NewMultiplayerGameOver(),
		world:               world,
		serverURL:           "ws://localhost:8080/ws",
		connectionMenu:      screens.NewMenu([]*screens.MenuButton{screens.NewMenuButton("retry", screens.ActionMultiplayer), screens.NewMenuButton("back to menu", screens.ActionMenu)}, 20, rl.Black),
	}
}

// Run keeps the default endpoint for callers that don't need configuration.
func Run() { RunWithServer("ws://localhost:8080/ws") }

func RunWithServer(url string) {
	g := NewGame()
	g.serverURL = url
	g.run()
}

type screen int

const (
	screenMenu screen = iota
	screenSinglePlayer
	screenMultiplayer
	screenSettings
	screenBriefing
)

func (g *Game) leaveMultiplayer() {
	if g.multiplayer != nil {
		g.multiplayer.Close()
		g.multiplayer = nil
	}
}

func (g *Game) run() {
	g.window.initWindow()
	defer g.window.close()
	rendering.InitFeedbackAudio()
	defer rendering.CloseFeedbackAudio()
	defer primitives.UnloadTextFonts()
	defer g.leaveMultiplayer()
	g.cameraFollowPlayer()
	current := screenMenu
	mapOpen := false
	exitRequested := false
	for !rl.WindowShouldClose() && !exitRequested {
		dt := rl.GetFrameTime()
		if rl.IsKeyPressed(rl.KeyF8) {
			rendering.ToggleFeedbackAudio()
		}
		if rl.IsKeyPressed(rl.KeyEscape) {
			if current == screenSinglePlayer && g.world.StatsOpen() {
				g.world.CloseStats()
			} else if current == screenMultiplayer && g.multiplayer.StatsOpen() {
				g.multiplayer.CloseStats()
			} else if current == screenSinglePlayer && g.world.PerksOpen() {
				g.world.ClosePerks()
			} else if current == screenMultiplayer && g.multiplayer.PerksOpen() {
				g.multiplayer.ClosePerks()
			} else if mapOpen {
				mapOpen = false
			} else {
				switch current {
				case screenSettings:
					current = screenMenu
				case screenBriefing:
					if g.briefing.Multiplayer {
						g.leaveMultiplayer()
					}
					current = screenMenu
				case screenMultiplayer:
					if g.multiplayer.ChatOpen() {
						g.multiplayer.CloseChat()
					} else {
						g.leaveMultiplayer()
						current = screenMenu
					}
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
			case screenMultiplayer:
				m := g.multiplayer
				if m.Ready() && m.Started() && !m.Dead() && m.Error() == "" && !m.ChatOpen() && !m.PerksOpen() {
					m.ToggleStats()
				}
			}
		}
		if rl.IsKeyPressed(rl.KeyM) && !rl.IsKeyPressed(rl.KeyEscape) && rl.IsWindowFocused() {
			switch current {
			case screenSinglePlayer:
				if !g.world.Dead() && !g.world.ResultsOpen() && !g.world.PerksOpen() && !g.world.StatsOpen() {
					mapOpen = !mapOpen
				}
			case screenMultiplayer:
				m := g.multiplayer
				if m.CanOpenMap() {
					mapOpen = !mapOpen
				}
			}
		}
		switch current {
		case screenBriefing:
			if g.briefing.Multiplayer {
				m := g.multiplayer
				m.Poll()
				m.Follow(&g.camera)
				g.briefing.Countdown = m.StartCountdown()
				if m.Error() != "" || m.Started() || !m.BriefingActive() && m.StartCountdown() == 0 {
					current = screenMultiplayer
				}
			}
		case screenSinglePlayer:
			if !g.world.Dead() && !g.world.ResultsOpen() && !mapOpen {
				g.world.ReadInput(g.camera, g.window.width, g.window.height)
				g.world.Update(dt)
				g.cameraFollowPlayer()
			}
		case screenMultiplayer:
			g.multiplayer.Poll()
			if g.multiplayer.Ready() && g.multiplayer.Error() == "" && g.multiplayer.BriefingActive() {
				g.briefing = screens.NewBriefing(true)
				g.multiplayer.CloseChat()
				g.multiplayer.Follow(&g.camera)
				current = screenBriefing
				break
			}
			if g.multiplayer.Dead() || g.multiplayer.ResultsOpen() || g.multiplayer.Error() != "" {
				mapOpen = false
				g.multiplayer.CloseStats()
			}
			g.multiplayer.ReadInput(dt, mapOpen, g.camera)
			if g.multiplayer.Ready() {
				g.multiplayer.Follow(&g.camera)
			}
		}

		rl.BeginDrawing()
		rl.ClearBackground(g.window.bgColor)
		action := screens.ActionNone
		switch current {
		case screenSettings:
			action = g.settings.Draw()
		case screenMenu:
			action = g.menu.Draw()
		case screenBriefing:
			if g.briefing.Multiplayer {
				g.multiplayer.Draw(g.camera, g.window.width, g.window.height)
			}
			action = g.briefing.Draw(g.window.width, g.window.height)
		case screenSinglePlayer:
			if g.world.ResultsOpen() {
				g.world.Draw(g.camera, g.window.width, g.window.height)
				action = g.world.DrawResults(g.window.width, g.window.height)
			} else if g.world.Dead() {
				action = g.gameOver.Draw()
			} else if mapOpen {
				g.drawFullMap(false)
			} else {
				g.world.Draw(g.camera, g.window.width, g.window.height)
				if g.world.Finished() {
					action = g.world.DrawResultsButton(g.window.width)
				}
			}
		case screenMultiplayer:
			m := g.multiplayer
			switch {
			case m.Error() != "":
				action = g.connectionMenu.Draw()
				title := "Unable to connect"
				if m.Ready() {
					title = "Disconnected"
				}
				primitives.DrawCenteredText(title, 150, 32, rl.RayWhite)
				primitives.DrawCenteredText(m.Error(), 200, 20, rl.LightGray)
			case !m.Ready():
				primitives.DrawCenteredText("Connecting...", g.window.height/2-30, 30, rl.RayWhite)
				primitives.DrawCenteredText("Esc to cancel", g.window.height/2+20, 20, rl.LightGray)
			case m.ResultsOpen():
				m.Draw(g.camera, g.window.width, g.window.height)
				action = m.DrawResults(g.window.width, g.window.height)
			case m.Dead():
				action = g.multiplayerGameOver.Draw()
			case !m.Started():
				action = m.DrawLobby(g.window.width, g.window.height)
			case mapOpen:
				g.drawFullMap(true)
			default:
				m.Draw(g.camera, g.window.width, g.window.height)
				if m.Finished() {
					action = m.DrawResultsButton(g.window.width)
				}
			}
		}
		rl.EndDrawing()

		if action != screens.ActionNone {
			mapOpen = false
		}

		// Screen transitions happen after drawing, on the same thread as raylib.
		switch action {
		case screens.ActionEnterDungeon:
			if g.briefing.Multiplayer {
				g.briefing.Waiting = true
				g.multiplayer.BriefingDone()
			} else {
				g.soloBriefed = true
				g.startSinglePlayer()
				current = screenSinglePlayer
			}
		case screens.ActionContinueExploring:
			if current == screenMultiplayer {
				g.multiplayer.ContinueExploring()
			} else {
				g.world.ContinueExploring()
			}
		case screens.ActionShowResults:
			if current == screenMultiplayer {
				g.multiplayer.ShowResults()
			} else {
				g.world.ShowResults()
			}
		case screens.ActionRematch:
			if current == screenMultiplayer {
				g.multiplayer.Rematch()
			} else {
				g.startSinglePlayer()
			}
		case screens.ActionStartGame:
			g.multiplayer.StartGame()
		case screens.ActionSettings:
			g.settings.Open()
			current = screenSettings
		case screens.ActionSaveSettings:
			g.world.SetName(g.settings.PlayerName)
			current = screenMenu
		case screens.ActionPlay:
			g.leaveMultiplayer()
			current = g.enterSinglePlayer()
		case screens.ActionMultiplayer:
			if g.multiplayer != nil && g.multiplayer.Ready() && g.multiplayer.Dead() && g.multiplayer.Error() == "" {
				g.multiplayer.Respawn()
				current = screenMultiplayer
				break
			}
			g.startMultiplayer()
			current = screenMultiplayer
		case screens.ActionMenu:
			g.leaveMultiplayer()
			current = screenMenu
		case screens.ActionExit:
			exitRequested = true
		}
	}
}

func (g *Game) enterSinglePlayer() screen {
	if !g.soloBriefed {
		g.briefing = screens.NewBriefing(false)
		return screenBriefing
	}
	g.startSinglePlayer()
	return screenSinglePlayer
}

func (g *Game) startMultiplayer() {
	var history []protocol.ChatEntry
	if g.multiplayer != nil {
		history = g.multiplayer.ChatHistory()
	}
	g.leaveMultiplayer()
	g.multiplayer = gameplay.NewMultiplayer(g.serverURL, g.settings.PlayerName, g.settings.Character)
	g.multiplayer.RestoreChat(history)
}

func (g *Game) startSinglePlayer() {
	if g.world.Finished() || g.world.Character() != g.settings.Character {
		g.world = gameplay.NewWorld()
		g.world.Configure(g.settings.PlayerName, g.settings.Character)
	}
	g.world.Start(g.settings.PlayerName)
	g.cameraFollowPlayer()
}

func (g *Game) cameraFollowPlayer() {
	g.camera.Target = g.world.Center()
}

func (g *Game) drawFullMap(multiplayer bool) {
	width, height := int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight())
	screens.DrawFullMapBackground(g.menu.BackgroundColor, height)
	if multiplayer {
		g.multiplayer.DrawMap(width, height)
	} else {
		g.world.DrawMap(width, height)
	}
}
