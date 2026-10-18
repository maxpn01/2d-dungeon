package rendering

import (
	"fmt"
	"image/color"
	"time"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/preferences"
	"2d-dungeon/internal/rendering/primitives"
	maprender "2d-dungeon/internal/rendering/worldmap"
	"2d-dungeon/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Scene struct {
	simulation.Snapshot
	PlayerName          string
	Map                 *maprender.Map
	Local               simulation.Player
	RenderedPlayer      simulation.Player
	RenderedAIs         []simulation.AI
	PreviousStatues     []simulation.StatueState
	PreviousTick        uint64
	PreviousProjectiles []simulation.Projectile
	ProjectileAlpha     float64
	Feedback            *Feedback
}

func DrawScene(ctx primitives.Context, scene Scene) {
	scene.Map.PrepareLibraryFrame(scene.PreviousStatues, scene.PreviousTick, scene.Tick, scene.ProjectileAlpha, scene.RenderedPlayer)
	scene.Map.Prepare(scene.Snapshot, scene.Local)
	if scene.Map.InForest() {
		focus := rl.NewVector2(float32(scene.RenderedPlayer.X+parameters.PlayerSize/2), float32(scene.RenderedPlayer.Y+parameters.PlayerSize/2))
		scene.Map.DrawForest3D(ctx, focus, forestActors(scene), func(camera rl.Camera2D) {
			rl.BeginMode2D(camera)
			scene.Map.DrawForestDecals(scene.Snapshot, scene.Local)
			drawSceneEntities(scene, true)
			rl.EndMode2D()
		})
		scene.Feedback.DrawDamageNumbers(time.Now(), &scene.Map.Exploration, func(x, y float32) rl.Vector2 {
			return scene.Map.ForestScreenPoint(ctx, x, y, parameters.DamageNumberLift)
		})
		scene.Map.DrawLibraryHUD(ctx, scene.Snapshot, scene.Local)
		return
	}
	rl.BeginMode2D(ctx.Camera)
	scene.Map.Draw(ctx, scene.Snapshot, scene.Local)
	scene.Map.DrawVaultArrowWarning(scene.Snapshot)
	drawSceneEntities(scene, false)
	rl.EndMode2D()
	scene.Map.DrawFog(ctx)
	rl.BeginMode2D(ctx.Camera)
	scene.Map.DrawVisibleVaultGate(scene.Orb, scene.Local)
	rl.EndMode2D()
	scene.Map.DrawLibraryHUD(ctx, scene.Snapshot, scene.Local)
}

func playerLabel(scene Scene) string {
	return fmt.Sprintf("%s Lv.%d (you)", preferences.NormalizePlayerName(scene.PlayerName), scene.RenderedPlayer.Stats.Level)
}

func forestActors(scene Scene) []maprender.ForestActor {
	player := scene.RenderedPlayer
	if player.Dead {
		return snakeForestActors(scene)
	}
	body := actorBody(player.X, player.Y, parameters.PlayerSize, player.Stats.HP, player.Combat, player.Character)
	height := actorRenderSize(parameters.PlayerSize, player.Stats.HP)
	actors := append(snakeForestActors(scene), poisonForestBubbles(player.ID, rl.NewVector2(body.X+body.Width/2, body.Y+body.Height/2), body.Width, height, player.Status.Poison)...)
	return append(actors, maprender.ForestActor{
		ID: player.ID, Label: playerLabel(scene),
		Center: rl.NewVector2(body.X+body.Width/2, body.Y+body.Height/2),
		Width:  body.Width, Depth: body.Height, Height: height,
		Tint:  poisonTint(primitives.CharacterColor(player.Character), player.Status.Poison),
		Flash: scene.Feedback.HitFlash(player.ID), Spawn: scene.Feedback.SpawnProgress(player.ID),
		Wings: player.HasWings, Flying: player.Flight.Flying,
	})
}

func drawSceneEntities(scene Scene, forest bool) {
	scene.Feedback.DrawWallDebris(time.Now(), &scene.Map.Exploration)
	playerColors := map[string]color.RGBA{scene.Player.ID: primitives.CharacterColor(scene.Player.Character)}
	for _, fruit := range scene.Fruits {
		if !scene.Map.Exploration.EntityVisible(fruit.X, fruit.Y, fruit.Size) {
			continue
		}
		DrawFruit(float32(fruit.X), float32(fruit.Y), float32(fruit.Size), MassColor(fruit.Source, playerColors), fruit.Source != "", fruit.MP > 0)
	}
	for _, ai := range scene.RenderedAIs {
		if !forest {
			drawPoisonBubbles(ai.X, ai.Y, parameters.AISize, ai.Status.Poison, &scene.Map.Exploration)
		}
		if ai.Type == simulation.ForestSnake {
			drawSnakeBite(ai, &scene.Map.Exploration, scene.Feedback)
			continue
		}
		tint := poisonTint(mobTint(ai.Type), ai.Status.Poison)
		DrawAIActor(Actor{
			Type: ai.Type, Level: ai.Stats.Level, Stone: ai.Stone, Awakening: ai.Awakening,
			ID: ai.ID, Label: fmt.Sprintf("%s Lv.%d", mobName(ai.Type), ai.Stats.Level),
			X: ai.X, Y: ai.Y, Size: parameters.AISize, HP: ai.Stats.HP, Character: ai.Character,
			Aiming: ai.Aiming, Frenzied: ai.Frenzied, Combat: ai.Combat, Tint: tint,
		}, &scene.Map.Exploration, scene.Feedback)
	}
	if player := scene.RenderedPlayer; !forest && !player.Dead {
		drawPoisonBubbles(player.X, player.Y, parameters.PlayerSize, player.Status.Poison, &scene.Map.Exploration)
	}
	if player := scene.RenderedPlayer; forest && !player.Dead {
		size := actorRenderSize(parameters.PlayerSize, player.Stats.HP)
		drawCombat(rl.NewVector2(float32(player.X+parameters.PlayerSize/2), float32(player.Y+parameters.PlayerSize/2)), size, player.Combat)
	} else if !player.Dead {
		tint := poisonTint(primitives.CharacterColor(player.Character), player.Status.Poison)
		label := playerLabel(scene)
		lift := float64(0)
		if player.HasWings {
			lift = DrawWings(player.X, player.Y, parameters.PlayerSize, player.Flight.Flying)
		}
		if player.HasOrb {
			center := rl.NewVector2(float32(player.X+parameters.PlayerSize/2), float32(player.Y-30-lift))
			rl.DrawCircleGradient(center, 25, rl.Fade(primitives.RuneColor, .35), rl.Blank)
			rl.DrawCircleV(center, 8, primitives.RuneColor)
			rl.DrawCircleV(rl.NewVector2(center.X-2, center.Y-2), 2, rl.RayWhite)
		}
		DrawPlayerActor(Actor{
			ID: player.ID, Label: label, X: player.X, Y: player.Y - lift, Size: parameters.PlayerSize,
			HP: player.Stats.HP, Dead: player.Dead, Character: player.Character,
			Combat: player.Combat, Tint: tint,
		}, scene.Feedback, &scene.Map.Exploration)
	}
	DrawProjectiles(scene.PreviousProjectiles, scene.Projectiles, scene.ProjectileAlpha, &scene.Map.Exploration)
	DrawPortals(scene.Local, scene.Snapshot.Orb, &scene.Map.Exploration)
	scene.Map.DrawVaultFire(scene.Snapshot)
	DrawTeleportCharge(scene.Player.X, scene.Player.Y, scene.Player.TeleportCharge)
	scene.Feedback.Draw(time.Now(), &scene.Map.Exploration, !forest)
}
