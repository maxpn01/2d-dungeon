package gameplay

import (
	"time"

	"raylib-2d-sandbox/internal/hud"
	"raylib-2d-sandbox/internal/rendering"
	"raylib-2d-sandbox/internal/rendering/primitives"
	maprender "raylib-2d-sandbox/internal/rendering/worldmap"
	"raylib-2d-sandbox/internal/screens"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type presentation struct {
	gameMap           *maprender.Map
	feedback          rendering.Feedback
	perks             hud.PerkPanel
	statsOpen         bool
	rematch           *screens.MenuButton
	resultsMenu       *screens.MenuButton
	continueExploring *screens.MenuButton
	showResults       *screens.MenuButton
	resultsDismissed  bool
}

func newPresentation() presentation {
	return presentation{gameMap: maprender.NewMap(), rematch: screens.NewMenuButton("Rematch", screens.ActionRematch), resultsMenu: screens.NewMenuButton("Back to menu", screens.ActionMenu), continueExploring: screens.NewMenuButton("Continue exploring", screens.ActionContinueExploring), showResults: screens.NewMenuButton("Round results", screens.ActionShowResults)}
}

func (p *presentation) draw(ctx primitives.Context, scene rendering.Scene, multiplayer bool) {
	scene.Map, scene.Feedback = p.gameMap, &p.feedback
	rendering.DrawScene(ctx, scene)
	width, height := ctx.WindowWidth, ctx.WindowHeight
	snapshot, local, players := scene.Snapshot, scene.Local, scene.RenderedPlayers
	hud.DrawPlayer(local, width)
	p.feedback.DrawCombatOverlay(time.Now(), local, width, height)
	hud.DrawBossTimer(snapshot.Boss, width)
	drawFrenzyAnnouncement(snapshot, local, width)
	if snapshot.WinnerID != "" {
		label := "FLOOR I COMPLETE"
		if snapshot.WinnerID != local.ID {
			label = "THE OTHER PLAYER ESCAPED FIRST"
		}
		panelWidth := primitives.MeasureText(label, 24) + 40
		rl.DrawRectangle((width-panelWidth)/2, 50, panelWidth, 44, rl.Fade(rl.Black, .85))
		primitives.DrawCenteredText(label, 60, 24, rl.Gold)
	}
	for _, player := range players {
		if player.ID != local.ID {
			continue
		}
		p.drawRadar(width, 0, scene)
		x, y := float32(player.X), float32(player.Y)
		if multiplayer {
			x, y = ctx.Camera.Target.X, ctx.Camera.Target.Y
		}
		hud.DrawLocation(x, y, width, height)
		break
	}
}

func (p *presentation) drawRadar(width, height int32, scene rendering.Scene) {
	snapshot, players, ais := scene.Snapshot, scene.RenderedPlayers, scene.RenderedAIs
	local := scene.Local
	found := false
	for _, player := range players {
		if player.ID == local.ID {
			local = player
			found = true
			break
		}
	}
	if !found || local.Dead {
		return
	}
	markers := make([]maprender.Marker, 0, len(snapshot.Fruits)+len(ais)+len(players))
	for _, fruit := range snapshot.Fruits {
		markers = append(markers, maprender.Marker{
			Center: rl.NewVector2(float32(fruit.X+fruit.Size/2), float32(fruit.Y+fruit.Size/2)), Radius: 2, Tint: rendering.EnergyColor,
		})
	}
	for _, ai := range ais {
		markers = append(markers, maprender.Marker{
			Center: rl.NewVector2(float32(ai.X+ai.Size/2), float32(ai.Y+ai.Size/2)), Radius: 3, Tint: rl.Green,
		})
	}
	for _, player := range players {
		if player.ID != local.ID && !player.Dead && p.gameMap.Exploration.PlayerVisible(player.X, player.Y, player.Size) {
			markers = append(markers, maprender.Marker{Center: playerCenter(player), Radius: 3, Tint: rl.SkyBlue})
		}
	}
	p.gameMap.DrawRadar(width, height, playerCenter(local), markers, snapshot)
}

func (p *presentation) StatsOpen() bool { return p.statsOpen }
func (p *presentation) CloseStats()     { p.statsOpen = false }
func (p *presentation) ToggleStats()    { p.statsOpen = !p.statsOpen }
