package game

import (
	"image/color"
	"time"

	"raylib-2d-sandbox/internal/client"
	"raylib-2d-sandbox/internal/parameters"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Game struct {
	window              *Window
	camera              rl.Camera2D
	settings            *Settings
	menu                *Menu
	gameOver            *GameOver
	multiplayerGameOver *GameOver
	world               *World
	hud                 *HUD
	serverURL           string
	multiplayer         *multiplayer
	connectionMenu      *Menu
}

func NewGame() *Game {
	window := NewWindow(
		parameters.WindowWidth,
		parameters.WindowHeight,
		"2D sandbox",
		rl.Black,
	)

	camera := rl.NewCamera2D(window.getWindowCenter(), rl.Vector2{}, 0, 1)

	menu := NewMenu([]*MenuButton{
		NewMenuButton("play", ActionPlay),
		NewMenuButton("multiplayer", ActionMultiplayer),
		NewMenuButton("settings", ActionSettings),
		NewMenuButton("exit", ActionExit),
	}, 30, rl.Black)

	menu.Title = "2D SANDBOX"

	settings := newSettings()
	world := NewWorld()
	world.player.name = settings.PlayerName
	world.simulation.SetCharacter("player", settings.Character)
	world.syncSimulation()

	hud := NewHUD(world, window)

	return &Game{
		settings:            settings,
		window:              window,
		camera:              camera,
		menu:                menu,
		gameOver:            NewGameOver(),
		multiplayerGameOver: NewMultiplayerGameOver(),
		world:               world,
		hud:                 hud,
		serverURL:           "ws://localhost:8080/ws",
		connectionMenu:      NewMenu([]*MenuButton{NewMenuButton("retry", ActionMultiplayer), NewMenuButton("back to menu", ActionMenu)}, 20, rl.Black),
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
)

func (g *Game) leaveMultiplayer() {
	if g.multiplayer != nil {
		g.multiplayer.session.Close()
		g.multiplayer = nil
	}
}

func (g *Game) run() {
	g.window.initWindow()
	defer g.window.close()
	defer unloadTextFonts()
	defer g.leaveMultiplayer()
	g.cameraFollowPlayer()
	current := screenMenu
	mapOpen := false
	exitRequested := false
	for !rl.WindowShouldClose() && !exitRequested {
		dt := rl.GetFrameTime()
		if rl.IsKeyPressed(rl.KeyEscape) {
			if mapOpen {
				mapOpen = false
			} else {
				switch current {
				case screenSettings:
					current = screenMenu
				case screenMultiplayer:
					if g.multiplayer.chat.open {
						g.multiplayer.chat.close()
					} else {
						g.leaveMultiplayer()
						current = screenMenu
					}
				case screenSinglePlayer:
					if !g.world.player.isDeadFlag {
						current = screenMenu
					}
				case screenMenu:
					if !g.world.player.isDeadFlag {
						g.startSinglePlayer()
						current = screenSinglePlayer
					}
				}
			}
		}
		if rl.IsKeyPressed(rl.KeyM) && !rl.IsKeyPressed(rl.KeyEscape) && rl.IsWindowFocused() {
			switch current {
			case screenSinglePlayer:
				if !g.world.player.isDeadFlag {
					mapOpen = !mapOpen
				}
			case screenMultiplayer:
				m := g.multiplayer
				if m.view.Ready && !m.view.Local.Dead && m.errorText == "" && !m.chat.open {
					mapOpen = !mapOpen
				}
			}
		}
		switch current {
		case screenSinglePlayer:
			if !g.world.player.isDeadFlag && !mapOpen {
				g.world.player.readInput(g.camera)
				g.world.update(dt)
				g.cameraFollowPlayer()
			}
		case screenMultiplayer:
			g.multiplayer.poll()
			if g.multiplayer.view.Local.Dead || g.multiplayer.errorText != "" {
				mapOpen = false
			}
			g.multiplayer.readInput(dt, mapOpen, g.camera)
			if g.multiplayer.view.Ready {
				g.multiplayer.follow(&g.camera)
			}
		}

		rl.BeginDrawing()
		rl.ClearBackground(g.window.bgColor)
		action := ActionNone
		switch current {
		case screenSettings:
			action = g.settings.draw()
		case screenMenu:
			action = g.menu.draw()
		case screenSinglePlayer:
			if g.world.player.isDeadFlag && (g.world.eating == nil || g.world.eating.finished(time.Now())) {
				action = g.gameOver.draw()
			} else if mapOpen {
				g.drawFullMap(false)
			} else {
				rl.BeginMode2D(g.camera)
				g.world.draw(RenderContext{Camera: g.camera, WindowWidth: g.window.width, WindowHeight: g.window.height})
				rl.EndMode2D()
				g.hud.draw()
				drawBossSpawnTimer(g.world.snapshot.Boss, g.window.width)
				g.world.drawMinimap(g.window.width)
				drawDungeonLocation(g.world.player.pos.X, g.world.player.pos.Y, g.window.width, g.window.height)
			}
		case screenMultiplayer:
			m := g.multiplayer
			switch {
			case m.view.Local.Dead:
				action = g.multiplayerGameOver.draw()
			case m.errorText != "":
				action = g.connectionMenu.draw()
				title := "Unable to connect"
				if m.view.Ready {
					title = "Disconnected"
				}
				drawCenteredText(title, 150, 32, rl.RayWhite)
				drawCenteredText(m.errorText, 200, 20, rl.LightGray)
			case !m.view.Ready:
				drawCenteredText("Connecting...", g.window.height/2-30, 30, rl.RayWhite)
				drawCenteredText("Esc to cancel", g.window.height/2+20, 20, rl.LightGray)
			case mapOpen:
				g.drawFullMap(true)
			default:
				m.draw(g.camera, g.window)
			}
		}
		rl.EndDrawing()

		if action != ActionNone {
			mapOpen = false
		}

		// Screen transitions happen after drawing, on the same thread as raylib.
		switch action {
		case ActionSettings:
			g.settings.open()
			current = screenSettings
		case ActionSaveSettings:
			g.world.player.name = g.settings.PlayerName
			current = screenMenu
		case ActionPlay:
			g.leaveMultiplayer()
			g.startSinglePlayer()
			current = screenSinglePlayer
		case ActionMultiplayer:
			var history []client.ChatEntry
			if g.multiplayer != nil {
				history = g.multiplayer.session.Chat()
			}
			g.leaveMultiplayer()
			g.multiplayer = newMultiplayer(g.serverURL, g.settings.PlayerName, g.settings.Character)
			g.multiplayer.session.RestoreChat(history)
			current = screenMultiplayer
		case ActionMenu:
			g.leaveMultiplayer()
			current = screenMenu
		case ActionExit:
			exitRequested = true
		}
	}
}

func (g *Game) startSinglePlayer() {
	if g.world.player.isDeadFlag || g.world.player.character != g.settings.Character {
		g.restart()
	}
	if !g.world.started {
		g.world.feedback.spawns["player"] = time.Now()
		g.world.started = true
	}
	g.world.player.name = g.settings.PlayerName
	g.world.player.input = PlayerInput{}
	g.world.player.suppressShootUntilRelease = true
	g.cameraFollowPlayer()
}

func drawCenteredText(text string, y, size int32, tint color.RGBA) {
	drawText(text, (int32(rl.GetScreenWidth())-measureText(text, size))/2, y, size, tint)
}

func (g *Game) restart() {
	g.world = NewWorld()
	g.world.player.name = g.settings.PlayerName
	g.world.simulation.SetCharacter("player", g.settings.Character)
	g.world.syncSimulation()
	g.hud = NewHUD(g.world, g.window)
	g.cameraFollowPlayer()
}

func (g *Game) cameraFollowPlayer() {
	g.camera.Target = rl.NewVector2(
		g.world.player.pos.X+g.world.player.size.X/2,
		g.world.player.pos.Y+g.world.player.size.Y/2,
	)
}
