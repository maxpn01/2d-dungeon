package rendering

import (
	"image/color"
	"math"

	maprender "2d-dungeon/internal/rendering/worldmap"
	"2d-dungeon/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	PoisonColor     = rl.NewColor(178, 92, 230, 255)
	poisonDarkColor = rl.NewColor(74, 30, 104, 255)
)

func DrawPoisonIcon(center rl.Vector2, size float32, tint color.RGBA) {
	radius := size * .32
	body := rl.NewVector2(center.X, center.Y+size*.14)
	top := rl.NewVector2(center.X, center.Y-size*.5)
	rl.DrawCircleV(body, radius+1.5, poisonDarkColor)
	rl.DrawTriangle(rl.NewVector2(body.X-radius-1.5, body.Y-1), rl.NewVector2(body.X+radius+1.5, body.Y-1), rl.NewVector2(top.X, top.Y-2), poisonDarkColor)
	rl.DrawCircleV(body, radius, tint)
	rl.DrawTriangle(rl.NewVector2(body.X-radius, body.Y), rl.NewVector2(body.X+radius, body.Y), top, tint)
	rl.DrawCircleV(rl.NewVector2(body.X-radius*.35, body.Y-radius*.2), radius*.28, rl.Fade(rl.RayWhite, .75))
}

func poisonTint(base color.RGBA, poison simulation.Poison) color.RGBA {
	if !poison.Active() {
		return base
	}
	t := float32(.45 + .15*math.Sin(rl.GetTime()*6))
	return color.RGBA{
		R: uint8(float32(base.R) + (float32(PoisonColor.R)-float32(base.R))*t),
		G: uint8(float32(base.G) + (float32(PoisonColor.G)-float32(base.G))*t),
		B: uint8(float32(base.B) + (float32(PoisonColor.B)-float32(base.B))*t),
		A: base.A,
	}
}

func drawPoisonBubbles(x, y, size float64, poison simulation.Poison, exploration Visibility) {
	if !poison.Active() || !exploration.EntityVisible(x, y, size) {
		return
	}
	clock := rl.GetTime()
	center := rl.NewVector2(float32(x+size/2), float32(y+size/2))
	seed := float64(int(x*7+y*13) % 97)
	for i := range 5 {
		phase := math.Mod(clock*.9+float64(i)/5+seed*.01, 1)
		angle := seed + float64(i)*2.4
		radius := float32(size) * (.25 + .2*float32(math.Sin(angle*3)))
		bubble := rl.NewVector2(
			center.X+float32(math.Cos(angle))*radius+float32(math.Sin(clock*3+float64(i)))*3,
			center.Y-float32(size)*.2-float32(phase)*float32(size)*.9,
		)
		fade := float32(math.Sin(phase * math.Pi))
		r := 2.5 + 2*float32(1-phase)
		rl.DrawCircleV(bubble, r+1, rl.Fade(poisonDarkColor, .6*fade))
		rl.DrawCircleV(bubble, r, rl.Fade(PoisonColor, .9*fade))
	}
}

func poisonForestBubbles(id string, center rl.Vector2, width, height float32, poison simulation.Poison) []maprender.ForestActor {
	if !poison.Active() {
		return nil
	}
	clock := rl.GetTime()
	var bubbles []maprender.ForestActor
	for i := range 4 {
		phase := float32(math.Mod(clock*.8+float64(i)/4, 1))
		angle := float64(i)*2.4 + clock*.7
		radius := width * .35
		size := 10 * (1 - phase*.6)
		tint := PoisonColor
		if i%2 == 1 {
			tint = rl.NewColor(214, 160, 255, 255)
		}
		bubbles = append(bubbles, maprender.ForestActor{
			ID: id, Center: rl.NewVector2(center.X+float32(math.Cos(angle))*radius, center.Y+float32(math.Sin(angle))*radius),
			Width: size, Depth: size, Height: size, Lift: height + 6 + phase*44,
			Tint: tint, Spawn: 1, NoShadow: true,
		})
	}
	return bubbles
}
