package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/physics"
	"dungeon-raylib-go/internal/rendering"
	"dungeon-raylib-go/internal/rendering/primitives"
	maprender "dungeon-raylib-go/internal/rendering/worldmap"
	"dungeon-raylib-go/internal/simulation"
	"dungeon-raylib-go/internal/ui"
	"dungeon-raylib-go/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	out := os.Args[1]
	width, height := int32(1600), int32(900)
	rl.SetConfigFlags(rl.FlagWindowHidden)
	rl.SetTraceLogLevel(rl.LogError)
	rl.InitWindow(width, height, "forestshot")
	var statues []simulation.StatueState
	for _, s := range worldmap.Default().Statues() {
		p := s.Position()
		statues = append(statues, simulation.StatueState{ID: s.ID, X: p.X, Y: p.Y})
	}
	for i := 2; i+1 < len(os.Args); i += 2 {
		cx, _ := strconv.Atoi(os.Args[i])
		cy, _ := strconv.Atoi(os.Args[i+1])
		x, y := worldmap.CellPosition(parameters.ForestOriginX+cx, parameters.ForestOriginY+cy, parameters.PlayerSize)
		player := simulation.Player{ID: "p", X: x, Y: y, HasWings: os.Getenv("SHOT_WINGS") != ""}
		player.Flight.Flying = os.Getenv("SHOT_WINGS") == "fly"
		player.Combat.Facing = simulation.Direction{X: .8, Y: .6}
		player.Stats.HP = 10
		poison := simulation.Poison{Damage: .3, Remaining: 2.6, Duration: 4}
		if os.Getenv("SHOT_POISON") != "" {
			player.Status.Poison = poison
		}
		m := maprender.NewMap()
		camera := rl.NewCamera2D(rl.NewVector2(float32(width)/2, float32(height)/2), rl.NewVector2(float32(x), float32(y)), 0, 1)
		ctx := primitives.Context{Camera: camera, WindowWidth: width, WindowHeight: height}
		snapshot := simulation.Snapshot{Statues: statues}
		var ais []simulation.AI
		if os.Getenv("SHOT_SNAKE") != "" {
			snake := simulation.AI{ID: "snake", Type: simulation.ForestSnake, X: x + 60, Y: y + 110, Stats: simulation.NewStatsAtLevel(parameters.SnakeLevel)}
			snake.Combat.Facing = simulation.Direction{X: -1}
			if os.Getenv("SHOT_POISON") != "" {
				snake.Status.Poison = poison
			}
			for i := range 6 {
				snake.Segments = append(snake.Segments, physics.Vector{X: snake.X + float64(i+1)*parameters.SnakeSegmentSpacing, Y: snake.Y + float64(i*i)*2})
			}
			if os.Getenv("SHOT_SNAKE") == "bite" {
				attack := simulation.MobMelee(simulation.ForestSnake)
				snake.Combat.SwordDirection = snake.Combat.Facing
				snake.Combat.SwordRemaining = attack.Duration - attack.Windup*.7
			}
			ais = append(ais, snake)
		}
		if steps, err := strconv.Atoi(os.Getenv("SHOT_HYDRA")); err == nil {
			snapshot, player = hydraSnapshot(steps, cx, cy)
			ais = snapshot.AIs
			x, y = player.X, player.Y
			camera.Target = rl.NewVector2(float32(x), float32(y))
			ctx.Camera = camera
		}
		if steps, err := strconv.Atoi(os.Getenv("SHOT_EMBER")); err == nil {
			snapshot, player = emberSnapshot(steps, cx, cy)
			maprender.EmberTimeOffset, _ = strconv.ParseFloat(os.Getenv("SHOT_TIME"), 64)
			center := rl.NewVector2(float32(player.X+parameters.PlayerSize/2), float32(player.Y+parameters.PlayerSize/2))
			camera.Target, _ = maprender.EmberCameraTarget(center, width, height)
			ctx.Camera = camera
		}
		feedback := &rendering.Feedback{}
		if os.Getenv("SHOT_HIT") != "" && len(ais) > 0 {
			feedback.Observe(time.Now(), simulation.Snapshot{Player: player, AIs: ais})
			ais[0].Stats.HP -= 3.4
			player.Status.Poison.Dealt += .4
			player.Stats.HP -= .4
			feedback.Observe(time.Now(), simulation.Snapshot{Player: player, AIs: ais})
		}
		for frame := 0; frame < 4; frame++ {
			started := time.Now()
			rl.BeginDrawing()
			rl.ClearBackground(rl.Black)
			rendering.DrawScene(ctx, rendering.Scene{Snapshot: snapshot, Map: m, Local: player, RenderedPlayer: player, RenderedAIs: ais, Feedback: feedback})
			if hydra, ok := rendering.HydraInView(ais, player, &m.Exploration); ok {
				ui.DrawBossBar(hydra, width)
			}
			if os.Getenv("SHOT_POISON") != "" {
				ui.DrawPlayer(player, "tester", width)
				feedback.DrawCombatOverlay(time.Now(), player, width, height)
			}
			if frame == 0 || frame == 3 {
				fmt.Println(cx, cy, "frame", frame, time.Since(started))
			}
			if frame == 3 {
				rl.DrawRenderBatchActive()
				img := rl.LoadImageFromScreen()
				rl.ExportImage(*img, fmt.Sprintf("%s/shot-%d-%d%s.png", out, cx, cy, os.Getenv("SHOT_WINGS")))
				rl.UnloadImage(img)
			}
			rl.EndDrawing()
		}
	}
	rl.CloseWindow()
}

func hydraSnapshot(steps, approach, sever int) (simulation.Snapshot, simulation.Player) {
	w := simulation.NewWorld()
	w.EnableTesting()
	w.TestingWarpCave()
	for range steps {
		s := w.Snapshot()
		d := simulation.Direction{}
		for _, ai := range s.AIs {
			if ai.Type != simulation.CaveHydra {
				continue
			}
			target := simulation.HydraCenter(ai)
			if sever > 0 {
				head := ai.Hydra.Heads[1]
				if !head.Severed {
					target.X, target.Y = head.X, head.Y
				}
			}
			dx, dy := target.X-s.Player.X-15, target.Y-s.Player.Y-15
			distance := math.Hypot(dx, dy)
			d.AimX, d.AimY, d.Shoot = dx/distance, dy/distance, approach > 0 && sever >= 0
			if distance > float64(approach) {
				d.X, d.Y = dx/distance, dy/distance
			}
		}
		w.Input(d)
		w.Step(time.Now())
	}
	s := w.Snapshot()
	return s, s.Player
}

func emberSnapshot(steps, move, jumpEvery int) (simulation.Snapshot, simulation.Player) {
	w := simulation.NewWorld()
	w.EnableTesting()
	w.TestingWarp(worldmap.EmberHollow)
	for tick := range steps {
		d := simulation.Direction{X: float64(max(-1, min(1, move))), AimX: 1}
		d.Jump = jumpEvery > 0 && tick%jumpEvery < jumpEvery/2
		w.Input(d)
		w.Step(time.Now())
	}
	s := w.Snapshot()
	fmt.Println("player", s.Player.X, s.Player.Y, "grounded", s.Player.Grounded)
	return s, s.Player
}
