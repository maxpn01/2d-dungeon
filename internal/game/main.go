package game

import (
	"image/color"

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
		1440,
		1000,
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
	defer g.leaveMultiplayer()
	g.cameraFollowPlayer()
	current := screenMenu
	exitRequested := false
	for !rl.WindowShouldClose() && !exitRequested {
		dt := rl.GetFrameTime()
		if rl.IsKeyPressed(rl.KeyEscape) {
			switch current {
			case screenSettings:
				current = screenMenu
			case screenMultiplayer:
				g.leaveMultiplayer()
				current = screenMenu
			case screenSinglePlayer:
				if !g.world.player.isDeadFlag {
					current = screenMenu
				}
			case screenMenu:
				if !g.world.player.isDeadFlag {
					current = screenSinglePlayer
				}
			}
		}
		switch current {
		case screenSinglePlayer:
			if !g.world.player.isDeadFlag {
				g.world.player.readInput()
				g.world.update(dt)
				g.cameraFollowPlayer()
				g.hud.update(dt)
			}
		case screenMultiplayer:
			g.multiplayer.poll()
			g.multiplayer.readInput(dt)
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
			if g.world.player.isDeadFlag {
				action = g.gameOver.draw()
			} else {
				rl.BeginMode2D(g.camera)
				g.world.draw(RenderContext{Camera: g.camera, WindowWidth: g.window.width, WindowHeight: g.window.height})
				rl.EndMode2D()
				g.hud.draw()
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
			default:
				m.draw(g.camera, g.window)
			}
		}
		rl.EndDrawing()

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
			if g.world.player.isDeadFlag {
				g.restart()
			}
			g.cameraFollowPlayer()
			current = screenSinglePlayer
		case ActionMultiplayer:
			g.leaveMultiplayer()
			g.multiplayer = newMultiplayer(g.serverURL, g.settings.PlayerName)
			current = screenMultiplayer
		case ActionMenu:
			g.leaveMultiplayer()
			current = screenMenu
		case ActionExit:
			exitRequested = true
		}
	}
}

func drawCenteredText(text string, y, size int32, tint color.RGBA) {
	rl.DrawText(text, (int32(rl.GetScreenWidth())-rl.MeasureText(text, size))/2, y, size, tint)
}

func (g *Game) restart() {
	g.world = NewWorld()
	g.world.player.name = g.settings.PlayerName
	g.hud = NewHUD(g.world, g.window)
	g.cameraFollowPlayer()
}

func (g *Game) cameraFollowPlayer() {
	g.camera.Target = rl.NewVector2(
		g.world.player.pos.X+g.world.player.size.X/2,
		g.world.player.pos.Y+g.world.player.size.Y/2,
	)
}
