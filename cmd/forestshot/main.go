package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/rendering"
	"2d-dungeon/internal/rendering/primitives"
	maprender "2d-dungeon/internal/rendering/worldmap"
	"2d-dungeon/internal/simulation"
	"2d-dungeon/internal/worldmap"

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
		m := maprender.NewMap()
		camera := rl.NewCamera2D(rl.NewVector2(float32(width)/2, float32(height)/2), rl.NewVector2(float32(x), float32(y)), 0, 1)
		ctx := primitives.Context{Camera: camera, WindowWidth: width, WindowHeight: height}
		snapshot := simulation.Snapshot{Statues: statues}
		for frame := 0; frame < 4; frame++ {
			started := time.Now()
			rl.BeginDrawing()
			rl.ClearBackground(rl.Black)
			rendering.DrawScene(ctx, rendering.Scene{Snapshot: snapshot, Map: m, Local: player, RenderedPlayer: player, Feedback: &rendering.Feedback{}})
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
