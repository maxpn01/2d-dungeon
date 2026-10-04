package rendering

import (
	"fmt"
	"image/color"
	"time"

	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/rendering/primitives"
	maprender "raylib-2d-sandbox/internal/rendering/worldmap"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Scene struct {
	protocol.Snapshot
	Map                 *maprender.Map
	Local               protocol.Player
	RenderedPlayers     []protocol.Player
	RenderedAIs         []protocol.AI
	PreviousProjectiles []simulation.Projectile
	ProjectileAlpha     float64
	Feedback            *Feedback
}

func DrawScene(ctx primitives.Context, scene Scene) {
	rl.BeginMode2D(ctx.Camera)
	scene.Map.Draw(ctx, scene.Snapshot, scene.Local)
	scene.Map.DrawVaultArrowWarning(scene.Snapshot)
	scene.Feedback.DrawWallDebris(time.Now(), &scene.Map.Exploration)
	playerColors := make(map[string]color.RGBA, len(scene.Players))
	for _, player := range scene.Players {
		playerColors[player.ID] = primitives.CharacterColor(player.Character)
	}
	for _, fruit := range scene.Fruits {
		if !scene.Map.Exploration.EntityVisible(fruit.X, fruit.Y, fruit.Size) {
			continue
		}
		DrawFruit(float32(fruit.X), float32(fruit.Y), float32(fruit.Size), MassColor(fruit.Source, playerColors), fruit.Source != "", fruit.MP > 0)
	}
	for _, ai := range scene.RenderedAIs {
		tint := rl.Green
		if ai.Character == simulation.Triangle {
			tint = primitives.CharacterColor(ai.Character)
		}
		DrawAIActor(Actor{
			Stone: ai.Stone, Awakening: ai.Awakening,
			ID: ai.ID, Label: fmt.Sprintf("%s Lv.%d %s", ai.ID, ai.Stats.Level, ai.Behavior),
			X: ai.X, Y: ai.Y, Size: ai.Size, HP: ai.Stats.HP, Character: ai.Character,
			Aiming: ai.Aiming, Frenzied: ai.Frenzied, Combat: ai.Combat, Split: ai.Split, Tint: tint,
		}, &scene.Map.Exploration, scene.Feedback)
	}
	for _, player := range scene.RenderedPlayers {
		if player.Dead {
			continue
		}
		if player.ID != scene.Local.ID && !scene.Map.Exploration.PlayerVisible(player.X, player.Y, player.Size) {
			continue
		}
		tint := primitives.CharacterColor(player.Character)
		label := fmt.Sprintf("%s Lv.%d", protocol.NormalizePlayerName(player.Name), player.Stats.Level)
		if player.ID == scene.Local.ID {
			label += " (you)"
		}
		lift := float64(0)
		if player.HasWings {
			lift = DrawWings(player.X, player.Y, player.Size, player.Flight.Flying)
		}
		if player.HasOrb {
			center := rl.NewVector2(float32(player.X+player.Size/2), float32(player.Y-30-lift))
			rl.DrawCircleGradient(center, 25, rl.Fade(primitives.RuneColor, .35), rl.Blank)
			rl.DrawCircleV(center, 8, primitives.RuneColor)
			rl.DrawCircleV(rl.NewVector2(center.X-2, center.Y-2), 2, rl.RayWhite)
		}
		DrawPlayerActor(Actor{
			ID: player.ID, Label: label, X: player.X, Y: player.Y - lift, Size: player.Size,
			HP: player.Stats.HP, Dead: player.Dead, Character: player.Character,
			Combat: player.Combat, Split: player.Split, Tint: tint,
		}, scene.Feedback, &scene.Map.Exploration)
	}
	DrawProjectiles(scene.PreviousProjectiles, scene.Projectiles, scene.ProjectileAlpha, &scene.Map.Exploration)
	DrawEncounter(scene.Boss, scene.Local, &scene.Map.Exploration)
	scene.Map.DrawVaultFire(scene.Snapshot)
	for _, p := range scene.Players {
		if p.ID != scene.Local.ID && !scene.Map.Exploration.PlayerVisible(p.X, p.Y, p.Size) {
			continue
		}
		DrawTeleportCharge(p.X, p.Y, p.TeleportCharge)
	}
	scene.Feedback.Draw(time.Now(), &scene.Map.Exploration)
	rl.EndMode2D()
}
