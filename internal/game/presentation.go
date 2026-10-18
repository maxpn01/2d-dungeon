package game

import (
	"time"

	"2d-dungeon/internal/rendering"
	"2d-dungeon/internal/rendering/primitives"
	maprender "2d-dungeon/internal/rendering/worldmap"
	"2d-dungeon/internal/simulation"
	"2d-dungeon/internal/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type presentation struct {
	results             *ui.ResultsPanel
	dungeonShakeAt      uint64
	dungeonShakeStarted time.Time
	gameMap             *maprender.Map
	feedback            rendering.Feedback
	perks               ui.PerkPanel
	stats               ui.StatsPanel
	statsOpen           bool
	resultsDismissed    bool
}

func newPresentation() presentation {
	return presentation{gameMap: maprender.NewMap(), results: ui.NewResultsPanel()}
}

func (p *presentation) draw(ctx primitives.Context, scene rendering.Scene) {
	scene.Map, scene.Feedback = p.gameMap, &p.feedback
	sceneContext := ctx
	sceneContext.Camera = p.dungeonAwakeningCamera(ctx.Camera, scene.Snapshot, scene.Local, time.Now())
	rendering.DrawScene(sceneContext, scene)
	width, height := ctx.WindowWidth, ctx.WindowHeight
	snapshot, local, players := scene.Snapshot, scene.Local, scene.RenderedPlayers
	ui.DrawPlayer(local, scene.PlayerName, width)
	p.feedback.DrawCombatOverlay(time.Now(), local, width, height)
	drawFrenzyAnnouncement(snapshot, local, width)
	if snapshot.WinnerID != "" {
		label := "FLOOR I COMPLETE"

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

		ui.DrawLocation(x, y, width, height)
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
		tint := rendering.HealthFoodColor
		if fruit.MP > 0 {
			tint = rendering.ManaFoodColor
		}
		markers = append(markers, maprender.Marker{
			Center: rl.NewVector2(float32(fruit.X+fruit.Size/2), float32(fruit.Y+fruit.Size/2)), Radius: 2, Tint: tint,
		})
	}
	for _, ai := range ais {
		markers = append(markers, maprender.Marker{
			Center: rl.NewVector2(float32(ai.X+simulation.AISize/2), float32(ai.Y+simulation.AISize/2)), Radius: 3, Tint: rl.Green,
		})
	}
	for _, player := range players {
		if player.ID != local.ID && !player.Dead && p.gameMap.Exploration.PlayerVisible(player.X, player.Y, simulation.PlayerSize) {
			markers = append(markers, maprender.Marker{Center: playerCenter(player), Radius: 3, Tint: rl.SkyBlue})
		}
	}
	p.gameMap.DrawRadar(width, height, playerCenter(local), markers, snapshot)
}

func (p *presentation) StatsOpen() bool { return p.statsOpen }
func (p *presentation) CloseStats() {
	p.statsOpen = false
	p.stats = ui.StatsPanel{}
}
func (p *presentation) ToggleStats() {
	if p.statsOpen {
		p.CloseStats()
	} else {
		p.statsOpen = true
	}
}
