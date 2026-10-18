package game

import (
	"time"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/rendering"
	"dungeon-raylib-go/internal/rendering/primitives"
	maprender "dungeon-raylib-go/internal/rendering/worldmap"
	"dungeon-raylib-go/internal/simulation"
	"dungeon-raylib-go/internal/ui"

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
	sceneContext.Camera = p.feedback.ShakeCamera(sceneContext.Camera, time.Now())
	rendering.DrawScene(sceneContext, scene)
	width, height := ctx.WindowWidth, ctx.WindowHeight
	snapshot, local := scene.Snapshot, scene.Local
	ui.DrawPlayer(local, scene.PlayerName, width)
	if hydra, ok := rendering.HydraInView(scene.RenderedAIs, scene.RenderedPlayer, &p.gameMap.Exploration); ok {
		ui.DrawBossBar(hydra, width)
	}
	p.feedback.DrawCombatOverlay(time.Now(), local, width, height)
	drawFrenzyAnnouncement(snapshot, local, width)
	label := ""
	if local.FloorCleared && snapshot.Tick < local.FloorClearedAt+parameters.FloorBannerSeconds*parameters.TickRate {
		label = "FLOOR I COMPLETE"
	}
	if snapshot.WinnerID != "" {
		label = "FLOOR II COMPLETE"
	}
	if label != "" {
		panelWidth := primitives.MeasureText(label, 24) + 40
		rl.DrawRectangle((width-panelWidth)/2, 50, panelWidth, 44, rl.Fade(rl.Black, .85))
		primitives.DrawCenteredText(label, 60, 24, rl.Gold)
	}
	p.drawRadar(width, 0, scene)
	ui.DrawLocation(float32(scene.RenderedPlayer.X), float32(scene.RenderedPlayer.Y), width, height)
}

func (p *presentation) drawRadar(width, height int32, scene rendering.Scene) {
	snapshot, ais, local := scene.Snapshot, scene.RenderedAIs, scene.RenderedPlayer
	if local.Dead {
		return
	}
	markers := make([]maprender.Marker, 0, len(snapshot.Fruits)+len(ais))
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
		marker := maprender.Marker{Center: rl.NewVector2(float32(ai.X+parameters.AISize/2), float32(ai.Y+parameters.AISize/2)), Radius: 3, Tint: rl.Green}
		if ai.Type == simulation.CaveHydra {
			marker.Radius, marker.Tint = 6, rendering.PoisonColor
		}
		markers = append(markers, marker)
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
