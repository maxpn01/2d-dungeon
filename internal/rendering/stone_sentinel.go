package rendering

import (
	"fmt"
	"math"

	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawStoneSentinel(ai Actor) {
	center := rl.NewVector2(float32(ai.X+ai.Size/2), float32(ai.Y+ai.Size/2))
	progress := float32(0)
	if !ai.Stone {
		progress = 1 - float32(ai.Awakening/simulation.SentinelAwakeningDuration)
	}
	crack := max(float32(0), min(float32(1), (progress-.25)/.45))
	burst := max(float32(0), min(float32(1), (progress-.72)/.28))
	shake := float32(math.Sin(rl.GetTime()*53)) * crack * (1 - burst) * 3
	center.X += shake
	stone := rl.Fade(rl.NewColor(90, 100, 118, 255), 1-burst)
	edge := rl.Fade(rl.NewColor(146, 158, 178, 255), 1-burst)
	glow := rl.NewColor(177, 123, 255, 255)
	if burst > 0 {
		live := ai
		live.Stone, live.Awakening = false, 0
		DrawActor(live.X, live.Y, live.Size, live.HP, live.Combat, primitives.CharacterColor(simulation.Triangle), simulation.Triangle)
	}
	rl.DrawEllipse(int32(center.X), int32(center.Y+24), 32, 12, rl.Fade(rl.Black, .65))
	rl.DrawPoly(rl.NewVector2(center.X, center.Y+18), 6, 28, 30, rl.Fade(rl.NewColor(48, 57, 75, 255), 1-burst))
	rl.DrawPolyLines(rl.NewVector2(center.X, center.Y+18), 6, 28, 30, edge)
	for _, side := range []float32{-1, 1} {
		center.X += side * burst * 24
		a, b, c := rl.NewVector2(center.X+side*12, center.Y-7), rl.NewVector2(center.X+side*32, center.Y-24), rl.NewVector2(center.X+side*29, center.Y+12)
		if side > 0 {
			b, c = c, b
		}
		rl.DrawTriangle(a, b, c, stone)
		rl.DrawLineEx(rl.NewVector2(center.X+side*12, center.Y-7), rl.NewVector2(center.X+side*32, center.Y-24), 2, edge)
		rl.DrawLineEx(rl.NewVector2(center.X+side*16, center.Y+6), rl.NewVector2(center.X+side*27, center.Y+16), 6, stone)
		center.X -= side * burst * 24
	}
	rl.DrawPoly(center, 3, 23, -90, stone)
	rl.DrawPolyLines(center, 3, 23, -90, edge)
	rl.DrawCircleV(rl.NewVector2(center.X, center.Y-14), 10, stone)
	for _, side := range []float32{-1, 1} {
		rl.DrawLineEx(rl.NewVector2(center.X+side*7, center.Y-19), rl.NewVector2(center.X+side*11, center.Y-31), 4, edge)
		eye := rl.NewVector2(center.X+side*4, center.Y-14)
		rl.DrawCircleV(eye, 2+crack, rl.Fade(glow, (.16+.84*progress)*(1-burst)))
	}
	if progress > 0 {
		rl.DrawCircleGradient(center, 35+crack*35+burst*30, rl.Fade(glow, (.12*progress+.3*crack)*(1-burst)), rl.Blank)
		for i := 0; i < 9; i++ {
			if crack < float32(i)/12 {
				continue
			}
			angle := float64(i) * 2.399
			dx, dy := float32(math.Cos(angle)), float32(math.Sin(angle))
			root := rl.NewVector2(center.X+dx*4, center.Y+dy*4-5)
			mid := rl.NewVector2(root.X+dx*9-dy*3, root.Y+dy*9+dx*3)
			end := rl.NewVector2(center.X+dx*24, center.Y+dy*24-5)
			rl.DrawLineEx(root, mid, 3, rl.Fade(rl.NewColor(20, 23, 32, 255), 1-burst))
			rl.DrawLineEx(mid, end, 3, rl.Fade(rl.NewColor(20, 23, 32, 255), 1-burst))
			rl.DrawLineEx(root, mid, 1, rl.Fade(glow, crack*(1-burst)))
			rl.DrawLineEx(mid, end, 1, rl.Fade(glow, crack*(1-burst)))
		}
		if burst > 0 {
			rl.DrawRing(center, 20+burst*70, 23+burst*70, 0, 360, 32, rl.Fade(glow, .5*(1-burst)))
			for i := 0; i < 24; i++ {
				angle := float64(i) * 2.399
				dx, dy := float32(math.Cos(angle)), float32(math.Sin(angle))
				travel := burst * float32(35+i%5*13)
				lift := float32(16+i%4*7) * 4 * burst * (1 - burst)
				at := rl.NewVector2(center.X+dx*(12+travel), center.Y+dy*(12+travel)*.65-lift+burst*burst*20)
				rl.DrawCircleGradient(rl.NewVector2(at.X, at.Y+lift), 10+burst*12, rl.Fade(rl.NewColor(125, 130, 146, 255), .16*(1-burst)), rl.Blank)
				rl.DrawPoly(at, 5, float32(2+i%4), float32(i*47)+burst*230, rl.Fade(rl.NewColor(106, 116, 134, 255), 1-burst*.9))
				rl.DrawLineEx(at, rl.NewVector2(at.X-dx*4, at.Y-dy*4), 1, rl.Fade(rl.RayWhite, .6*(1-burst)))
			}
		}
	}
	primitives.DrawTextAt(fmt.Sprintf("SENTINEL Lv.%d", ai.Level), rl.NewVector2(center.X-43, center.Y-49), 14, edge)
}
