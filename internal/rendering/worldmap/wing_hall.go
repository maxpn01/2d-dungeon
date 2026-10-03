package worldmap

import (
	"math"

	mapdata "raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawWingHall(room mapdata.Room) {
	bounds := room.Bounds()
	tint := rl.NewColor(158, 115, 225, 255)
	vertical := bounds.H > bounds.W
	centerX, centerY := float32(bounds.X+bounds.W/2), float32(bounds.Y+bounds.H/2)
	carpet := rl.NewRectangle(float32(bounds.X+48), centerY-34, float32(bounds.W-96), 68)
	if vertical {
		carpet = rl.NewRectangle(centerX-34, float32(bounds.Y+48), 68, float32(bounds.H-96))
	}
	rl.DrawRectangleRec(carpet, rl.Fade(rl.NewColor(52, 42, 73, 255), .22))
	for _, offset := range []float32{-34, 34} {
		from, to := rl.NewVector2(float32(bounds.X+48), centerY+offset), rl.NewVector2(float32(bounds.X+bounds.W-48), centerY+offset)
		if vertical {
			from, to = rl.NewVector2(centerX+offset, float32(bounds.Y+48)), rl.NewVector2(centerX+offset, float32(bounds.Y+bounds.H-48))
		}
		rl.DrawLineEx(from, to, 1, rl.Fade(tint, .15))
	}
	for _, tile := range []int{3, 6, 9, 12} {
		along := float32((room.Area.X+tile)*mapdata.Tile + mapdata.Tile/2)
		positions := []rl.Vector2{rl.NewVector2(along, float32(bounds.Y-17)), rl.NewVector2(along, float32(bounds.Y+bounds.H+19))}
		mark := rl.NewVector2(along, centerY)
		if vertical {
			along = float32((room.Area.Y+tile)*mapdata.Tile + mapdata.Tile/2)
			positions = []rl.Vector2{rl.NewVector2(float32(bounds.X-20), along), rl.NewVector2(float32(bounds.X+bounds.W+20), along)}
			mark = rl.NewVector2(centerX, along)
		}
		for i, at := range positions {
			drawMagicTorch(at, float64(tile+i*7))
		}
		rl.DrawPolyLines(mark, 4, 4, 45, rl.Fade(tint, .18))
	}
}

func drawMagicTorch(at rl.Vector2, phase float64) {
	t := rl.GetTime()
	pulse := float32(.75 + .12*math.Sin(t*4+phase) + .06*math.Sin(t*9+phase*2))
	tint := rl.NewColor(164, 117, 237, 255)
	flame := rl.NewVector2(at.X, at.Y-13)
	rl.DrawCircleGradient(flame, 65, rl.Fade(tint, .16*pulse), rl.Blank)
	rl.DrawRectangleRec(rl.NewRectangle(at.X-7, at.Y-7, 14, 19), rl.NewColor(22, 28, 39, 255))
	rl.DrawRectangleLinesEx(rl.NewRectangle(at.X-6, at.Y-6, 12, 17), 1, paleStone)
	rl.DrawRectangleRec(rl.NewRectangle(at.X-3, at.Y-5, 6, 13), stone)
	rl.DrawRectangleRec(rl.NewRectangle(at.X-9, at.Y-3, 18, 4), paleStone)
	rl.DrawCircleGradient(flame, 25, rl.Fade(tint, .35*pulse), rl.Blank)
	rl.DrawEllipse(int32(flame.X), int32(flame.Y), 5*pulse, 10*pulse, rl.Fade(tint, .9))
	rl.DrawEllipse(int32(flame.X), int32(flame.Y+2), 2.5, 5*pulse, rl.NewColor(187, 229, 244, 255))
	for i := 0; i < 2; i++ {
		life := math.Mod(t*.6+phase*.17+float64(i)/2, 1)
		mote := rl.NewVector2(flame.X+float32(math.Sin(t*2+phase+float64(i)))*4, flame.Y-8-float32(life)*20)
		rl.DrawCircleV(mote, 1, rl.Fade(tint, float32(1-life)*.6))
	}
}
