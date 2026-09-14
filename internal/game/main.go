package game

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Game struct {
	window   *Window
	camera   rl.Camera2D
	menu     *Menu
	gameOver *GameOver
	world    *World
	hud      *HUD
}

func NewGame() *Game {
	window := NewWindow(
		1440,
		820,
		"2D sandbox",
		rl.Black,
	)

	camera := rl.NewCamera2D(window.getWindowCenter(), rl.Vector2{}, 0, 1)

	menu := NewMenu([]*MenuButton{
		NewMenuButton("play", ActionPlay),
		NewMenuButton("exit", ActionExit),
	}, 30, rl.Black)

	world := NewWorld()

	hud := NewHUD(world, window)

	return &Game{
		window:   window,
		camera:   camera,
		menu:     menu,
		gameOver: NewGameOver(),
		world:    world,
		hud:      hud,
	}
}

func Run() {
	NewGame().run()
}

func (g *Game) run() {
	g.window.initWindow()
	defer g.window.close()

	g.cameraFollowPlayer()

	menuOpen := true
	exitRequested := false

	for !rl.WindowShouldClose() && !exitRequested {
		dt := rl.GetFrameTime()

		if !g.world.player.isDeadFlag && rl.IsKeyPressed(rl.KeyEscape) {
			menuOpen = !menuOpen
		}

		if !menuOpen && !g.world.player.isDeadFlag {
			g.world.player.readInput()
			g.world.update(dt)

			g.cameraFollowPlayer()

			g.hud.update(dt)
		}

		rl.BeginDrawing()
		rl.ClearBackground(g.window.bgColor)

		if g.world.player.isDeadFlag {
			switch g.gameOver.draw() {
			case ActionPlay:
				g.restart()
				menuOpen = false
			case ActionExit:
				exitRequested = true
			}
		} else if menuOpen {
			switch g.menu.draw() {
			case ActionPlay:
				menuOpen = false
			case ActionExit:
				exitRequested = true
			}
		} else {
			rl.BeginMode2D(g.camera)

			g.world.draw(RenderContext{
				Camera:       g.camera,
				WindowWidth:  g.window.width,
				WindowHeight: g.window.height,
			})

			rl.EndMode2D()

			g.hud.draw()
		}

		rl.EndDrawing()
	}
}

func (g *Game) restart() {
	g.world = NewWorld()
	g.hud = NewHUD(g.world, g.window)
	g.cameraFollowPlayer()
}

func (g *Game) cameraFollowPlayer() {
	g.camera.Target = rl.NewVector2(
		g.world.player.pos.X+g.world.player.size.X/2,
		g.world.player.pos.Y+g.world.player.size.Y/2,
	)
}
