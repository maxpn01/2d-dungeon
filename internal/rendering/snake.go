package rendering

import (
	"fmt"
	"math"

	"2d-dungeon/internal/parameters"
	maprender "2d-dungeon/internal/rendering/worldmap"
	"2d-dungeon/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	snakeTint     = rl.NewColor(96, 178, 74, 255)
	snakeBandTint = rl.NewColor(52, 112, 52, 255)
	snakeEyeTint  = rl.NewColor(246, 214, 92, 255)
)

func snakeBitePullback(c simulation.CombatState) float32 {
	attack := simulation.MobMelee(simulation.ForestSnake)
	elapsed := attack.Duration - c.SwordRemaining
	if c.SwordRemaining <= 0 || elapsed >= attack.Windup {
		return 0
	}
	p := float32(elapsed / attack.Windup)
	return parameters.SnakeBitePullback * p * p * (3 - 2*p)
}

func snakeFacing(c simulation.CombatState) rl.Vector2 {
	facing := c.Facing
	if c.SwordRemaining > 0 {
		facing = c.SwordDirection
	}
	if facing.X == 0 && facing.Y == 0 {
		facing.X = 1
	}
	return rl.NewVector2(float32(facing.X), float32(facing.Y))
}

func snakeForestActors(scene Scene) []maprender.ForestActor {
	var actors []maprender.ForestActor
	for _, ai := range scene.RenderedAIs {
		if ai.Type != simulation.ForestSnake {
			continue
		}
		flash, spawn := scene.Feedback.HitFlash(ai.ID), scene.Feedback.SpawnProgress(ai.ID)
		for i := len(ai.Segments) - 1; i >= 0; i-- {
			body := simulation.SnakeSegmentBody(ai.Segments[i])
			if !scene.Map.Exploration.EntityVisible(body.X, body.Y, body.W) {
				continue
			}
			taper := 1 - .35*float32(i+1)/float32(len(ai.Segments)+1)
			tint := snakeTint
			if i%2 == 0 {
				tint = snakeBandTint
			}
			tint = poisonTint(tint, ai.Status.Poison)
			actors = append(actors, maprender.ForestActor{
				ID: ai.ID, Center: rl.NewVector2(float32(body.X+body.W/2), float32(body.Y+body.H/2)),
				Width: float32(body.W) * taper, Depth: float32(body.H) * taper, Height: 16 * taper,
				Tint: tint, Flash: flash, Spawn: spawn,
			})
		}
		if !scene.Map.Exploration.EntityVisible(ai.X, ai.Y, parameters.AISize) {
			continue
		}
		facing := snakeFacing(ai.Combat)
		pullback := snakeBitePullback(ai.Combat)
		size := float32(parameters.AISize)
		center := rl.NewVector2(float32(ai.X)+size/2-facing.X*pullback, float32(ai.Y)+size/2-facing.Y*pullback)
		height := float32(20)
		actors = append(actors, maprender.ForestActor{
			ID: ai.ID, Label: fmt.Sprintf("%s Lv.%d", mobName(ai.Type), ai.Stats.Level),
			Center: center, Width: size, Depth: size, Height: height,
			Tint: poisonTint(snakeTint, ai.Status.Poison), Flash: flash, Spawn: spawn,
		})
		actors = append(actors, poisonForestBubbles(ai.ID, center, size, height, ai.Status.Poison)...)
		for _, side := range []float32{-1, 1} {
			eye := rl.NewVector2(center.X+facing.X*size*.3-facing.Y*side*size*.24, center.Y+facing.Y*size*.3+facing.X*side*size*.24)
			actors = append(actors, maprender.ForestActor{
				ID: ai.ID, Center: eye, Width: 6, Depth: 6, Height: 4, Lift: height + 2,
				Tint: snakeEyeTint, Flash: flash, Spawn: spawn, NoShadow: true,
			})
		}
	}
	return actors
}

func drawSnakeBite(ai simulation.AI, exploration Visibility, feedback *Feedback) {
	if !exploration.EntityVisible(ai.X, ai.Y, parameters.AISize) {
		return
	}
	size := float32(parameters.AISize)
	center := rl.NewVector2(float32(ai.X)+size/2, float32(ai.Y)+size/2)
	feedback.DrawAlert(ai.ID, rl.NewRectangle(float32(ai.X), float32(ai.Y), size, size))
	c := ai.Combat
	if c.SwordRemaining <= 0 {
		return
	}
	attack := simulation.MobMelee(simulation.ForestSnake)
	facing := snakeFacing(c)
	base := float32(math.Atan2(float64(facing.Y), float64(facing.X)) * 180 / math.Pi)
	reach := float32(attack.Reach)
	elapsed := attack.Duration - c.SwordRemaining
	if elapsed < attack.Windup {
		p := float32(elapsed / attack.Windup)
		rl.DrawCircleSector(center, reach, base-35, base+35, 20, rl.Fade(rl.Orange, .1+.2*p))
		rl.DrawCircleSectorLines(center, reach, base-35, base+35, 20, rl.Fade(rl.Orange, .45+.55*p))
		rl.DrawRing(center, reach*p, reach*p+2, base-35, base+35, 20, rl.Fade(rl.Gold, .8))
		return
	}
	p := float32(min(1, (elapsed-attack.Windup)/attack.Active))
	fade := float32(max(0, 1-(elapsed-attack.Windup)/(attack.Duration-attack.Windup)))
	tip := rl.NewVector2(center.X+facing.X*reach, center.Y+facing.Y*reach)
	for _, side := range []float32{-1, 1} {
		spread := (1 - p) * 14
		root := rl.NewVector2(center.X+facing.X*size*.4-facing.Y*side*spread, center.Y+facing.Y*size*.4+facing.X*side*spread)
		rl.DrawLineEx(root, tip, 3, rl.Fade(rl.RayWhite, fade))
	}
	rl.DrawCircleV(tip, 4*fade+1, rl.Fade(rl.RayWhite, fade))
}
