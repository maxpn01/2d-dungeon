package game

import (
	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Game struct {
	window      *Window
	camera      rl.Camera2D
	settings    *ui.Settings
	menu        *ui.Menu
	gameOver    *ui.GameOver
	world       *World
	normalWorld *World
	testWorld   *World
	briefing    *ui.Briefing
	soloBriefed bool
}

func NewGame() *Game {
	window := NewWindow(
		parameters.WindowWidth,
		parameters.WindowHeight,
		"2D Dungeon",
		rl.Black,
	)

	camera := rl.NewCamera2D(window.getWindowCenter(), rl.Vector2{}, 0, 1)

	menu := ui.NewMenu([]*ui.MenuButton{
		ui.NewMenuButton("play", ui.ActionPlay),
		ui.NewMenuButton("settings", ui.ActionSettings),
		ui.NewMenuButton("exit", ui.ActionExit),
	}, 30, rl.Black)

	menu.Title = "2D DUNGEON"

	settings := ui.NewSettings()
	world := NewWorld()
	world.Configure(settings.PlayerName, settings.Character)

	return &Game{
		settings:    settings,
		window:      window,
		camera:      camera,
		menu:        menu,
		gameOver:    ui.NewGameOver(),
		world:       world,
		normalWorld: world,
	}
}

func Run(testing bool) {
	g := NewGame()
	g.run(testing)
}
