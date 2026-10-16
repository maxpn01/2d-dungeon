package game

import (
	"raylib-2d-sandbox/internal/display"
	"raylib-2d-sandbox/internal/ui"
)

func (g *Game) enterSinglePlayer() screen {
	if g.world.TestingEnabled() {
		g.startSinglePlayer()
		return screenSinglePlayer
	}
	if !g.soloBriefed {
		g.briefing = ui.NewBriefing()
		return screenBriefing
	}
	g.startSinglePlayer()
	return screenSinglePlayer
}

func (g *Game) startSinglePlayer() {
	testing := g.world.TestingEnabled()
	if g.world.Dead() || g.world.Finished() || g.world.Character() != g.settings.Character {
		g.world = NewWorld()
		g.world.Configure(g.settings.PlayerName, g.settings.Character)
		if testing {
			g.world.EnableTesting()
			g.testWorld = g.world
		}
	}
	g.world.Start(g.settings.PlayerName)
	if !g.world.TestingEnabled() {
		g.normalWorld = g.world
	}
	g.cameraFollowPlayer()
}

func (g *Game) cameraFollowPlayer() {
	g.camera.Target = g.world.Center()
}

func (g *Game) drawFullMap() {
	width, height := int32(display.Width()), int32(display.Height())
	ui.DrawFullMapBackground(g.menu.BackgroundColor, height)
	g.world.DrawMap(width, height)
}
