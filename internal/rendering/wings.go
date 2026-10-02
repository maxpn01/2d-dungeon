package rendering

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawWings(x, y, size float64, flying bool) float64 {
	lift := float64(0)
	spread := float32(12)
	flap := float32(0)
	if flying {
		lift = 18 + math.Sin(rl.GetTime()*5)*3
		spread = 34
		flap = float32(math.Sin(rl.GetTime()*11)) * 10
		rl.DrawEllipse(int32(x+size/2), int32(y+size), float32(size*.65), 6, rl.Fade(rl.Black, .45))
	}
	center := rl.NewVector2(float32(x+size/2), float32(y+size/2-lift))
	tint := rl.NewColor(192, 220, 244, 255)
	if flying {
		rl.DrawCircleGradient(center, 55, rl.Fade(tint, .1), rl.Blank)
	}
	for _, side := range []float32{-1, 1} {
		root := rl.NewVector2(center.X+side*8, center.Y+2)
		for i := 0; i < 5; i++ {
			f := float32(i)
			tip := rl.NewVector2(center.X+side*(spread+10-f*3), center.Y-13+flap+f*7)
			end := rl.NewVector2(center.X+side*(14+f), center.Y+14-f)
			if side < 0 {
				rl.DrawTriangle(root, tip, end, rl.Fade(tint, .78-f*.08))
			} else {
				rl.DrawTriangle(root, end, tip, rl.Fade(tint, .78-f*.08))
			}
			rl.DrawLineEx(root, tip, 1, rl.Fade(rl.RayWhite, .65))
		}
	}
	return lift
}
