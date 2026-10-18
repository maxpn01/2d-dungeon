package rendering

import (
	"fmt"
	"math"

	"dungeon-raylib-go/internal/parameters"
	maprender "dungeon-raylib-go/internal/rendering/worldmap"
	"dungeon-raylib-go/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	snakeTint     = rl.NewColor(96, 178, 74, 255)
	snakeBandTint = rl.NewColor(52, 112, 52, 255)
	snakeEyeTint  = rl.NewColor(246, 214, 92, 255)
	venomTint     = rl.NewColor(176, 236, 92, 255)
)

func snakePullback(ai simulation.AI) float32 {
	if ai.Snake.SpitRemaining > 0 {
		p := float32(1 - ai.Snake.SpitRemaining/parameters.SnakeSpitWindup)
		return parameters.SnakeSpitPullback * p * p * (3 - 2*p)
	}
	c := ai.Combat
	attack := simulation.MobMelee(simulation.ForestSnake)
	elapsed := attack.Duration - c.SwordRemaining
	if c.SwordRemaining <= 0 || elapsed >= attack.Windup {
		return 0
	}
	p := float32(elapsed / attack.Windup)
	return parameters.SnakeBitePullback * p * p * (3 - 2*p)
}

func snakeCoilProgress(ai simulation.AI) float32 {
	if ai.Snake.SlitherPhase != simulation.DashWinding {
		return 0
	}
	return float32(1 - ai.Snake.SlitherRemaining/parameters.SnakeSlitherCoil)
}

func snakeFacing(ai simulation.AI) rl.Vector2 {
	c := ai.Combat
	facing := c.Facing
	switch {
	case ai.Snake.SpitRemaining > 0:
		facing = ai.Snake.SpitDirection
	case ai.Snake.SlitherPhase != simulation.DashIdle:
		facing = ai.Snake.SlitherDirection
	case c.SwordRemaining > 0:
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
		facing := snakeFacing(ai)
		pullback := snakePullback(ai)
		size := float32(parameters.AISize)
		center := rl.NewVector2(float32(ai.X)+size/2-facing.X*pullback, float32(ai.Y)+size/2-facing.Y*pullback)
		if coil := snakeCoilProgress(ai); coil > 0 {
			shake := float32(math.Sin(rl.GetTime()*45)) * 2.5 * coil
			center.X -= facing.Y * shake
			center.Y += facing.X * shake
		}
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

func drawSnakeAttacks(ai simulation.AI, exploration Visibility, feedback *Feedback) {
	if !exploration.EntityVisible(ai.X, ai.Y, parameters.AISize) {
		return
	}
	size := float32(parameters.AISize)
	center := rl.NewVector2(float32(ai.X)+size/2, float32(ai.Y)+size/2)
	feedback.DrawAlert(ai.ID, rl.NewRectangle(float32(ai.X), float32(ai.Y), size, size))
	drawSnakeSpitWindup(ai, center)
	drawSnakeSlither(ai, center)
	c := ai.Combat
	if c.SwordRemaining <= 0 {
		return
	}
	attack := simulation.MobMelee(simulation.ForestSnake)
	facing := snakeFacing(ai)
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

func drawSnakeSpitWindup(ai simulation.AI, center rl.Vector2) {
	if ai.Snake.SpitRemaining <= 0 {
		return
	}
	p := float32(1 - ai.Snake.SpitRemaining/parameters.SnakeSpitWindup)
	facing := snakeFacing(ai)
	size := float32(parameters.AISize)
	mouth := rl.NewVector2(center.X+facing.X*size*.45, center.Y+facing.Y*size*.45)
	line := rl.Fade(PoisonColor, .15+.35*p)
	if ai.Snake.SpitRemaining <= parameters.SnakeSpitAimLock {
		line = rl.Fade(venomTint, .75)
	}
	reach := float32(parameters.SnakeSpitRange) * (.35 + .65*p)
	for d := size * .6; d < reach; d += 22 {
		from := rl.NewVector2(mouth.X+facing.X*d, mouth.Y+facing.Y*d)
		to := rl.NewVector2(mouth.X+facing.X*min(reach, d+11), mouth.Y+facing.Y*min(reach, d+11))
		rl.DrawLineEx(from, to, 2, line)
	}
	radius := 3 + 6*p
	rl.BeginBlendMode(rl.BlendAdditive)
	rl.DrawCircleGradient(mouth, radius*2.6, rl.Fade(PoisonColor, .55*p), rl.Fade(PoisonColor, 0))
	rl.EndBlendMode()
	rl.DrawCircleV(mouth, radius, venomTint)
	rl.DrawCircleV(rl.NewVector2(mouth.X-radius*.3, mouth.Y-radius*.3), radius*.35, rl.Fade(rl.RayWhite, .8))
}

func drawSnakeSlither(ai simulation.AI, center rl.Vector2) {
	s := ai.Snake
	if s.SlitherPhase != simulation.DashWinding && s.SlitherPhase != simulation.DashRushing {
		return
	}
	facing := rl.NewVector2(float32(s.SlitherDirection.X), float32(s.SlitherDirection.Y))
	width := float32(parameters.AISize)
	if s.SlitherPhase == simulation.DashRushing {
		for i := range 6 {
			side := (float32(i) - 2.5) * width * .18
			trail := 26 + float32(i%3)*12
			from := rl.NewVector2(center.X-facing.X*width*.6-facing.Y*side, center.Y-facing.Y*width*.6+facing.X*side)
			to := rl.NewVector2(from.X-facing.X*trail, from.Y-facing.Y*trail)
			rl.DrawLineEx(from, to, 2, rl.Fade(rl.RayWhite, .45))
		}
		return
	}
	p := snakeCoilProgress(ai)
	angle := float32(math.Atan2(float64(facing.Y), float64(facing.X)) * 180 / math.Pi)
	length := float32(ai.Stats.MovementSpeed() * parameters.SpeedScale * parameters.SnakeSlitherSpeed * parameters.SnakeSlitherSurge)
	rl.DrawRectanglePro(rl.NewRectangle(center.X, center.Y, length, width), rl.NewVector2(0, width/2), angle, rl.Fade(rl.Orange, .06+.16*p))
	for _, side := range []float32{-1, 1} {
		edge := rl.NewVector2(center.X-facing.Y*side*width/2, center.Y+facing.X*side*width/2)
		rl.DrawLineEx(edge, rl.NewVector2(edge.X+facing.X*length, edge.Y+facing.Y*length), 2, rl.Fade(rl.Orange, .3+.5*p))
	}
	shift := float32(math.Mod(rl.GetTime()*2.5, 1)) * 30
	for d := width*.6 + shift; d < length-8; d += 30 {
		tip := rl.NewVector2(center.X+facing.X*(d+8), center.Y+facing.Y*(d+8))
		for _, side := range []float32{-1, 1} {
			root := rl.NewVector2(center.X+facing.X*d-facing.Y*side*9, center.Y+facing.Y*d+facing.X*side*9)
			rl.DrawLineEx(root, tip, 3, rl.Fade(rl.Gold, (.25+.6*p)*(1-d/length)))
		}
	}
}
