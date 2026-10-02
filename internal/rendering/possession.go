package rendering

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

func drawPossession(ai Actor) {
	center := rl.NewVector2(float32(ai.X+ai.Size/2), float32(ai.Y+ai.Size/2))
	phase := rl.GetTime()*2.6 + ai.X*.013 + ai.Y*.007
	pulse := float32(.8 + .2*math.Sin(phase))
	tint := rl.NewColor(166, 112, 245, 255)
	radius := actorRenderSize(ai.Size, ai.HP) * .65
	rl.DrawCircleGradient(center, radius+30, rl.Fade(tint, .35*pulse), rl.Blank)
	rl.DrawRing(center, radius+5, radius+7, float32(phase*45), float32(phase*45)+250, 40, rl.Fade(tint, .55*pulse))
	for i := 0; i < 6; i++ {
		angle := phase + float64(i)*math.Pi/3
		orbit := radius + 10 + float32(math.Sin(phase*1.3+float64(i)))*4
		mote := rl.NewVector2(center.X+float32(math.Cos(angle))*orbit, center.Y+float32(math.Sin(angle))*orbit*.7-8)
		tail := rl.NewVector2(mote.X-float32(math.Cos(angle-.6))*7, mote.Y+6)
		rl.DrawLineEx(tail, mote, 2, rl.Fade(tint, .4*pulse))
		rl.DrawCircleGradient(mote, 7, rl.Fade(tint, .65*pulse), rl.Blank)
		rl.DrawCircleV(mote, 1.5, rl.Fade(rl.RayWhite, .8*pulse))
	}
	if ai.Combat.Boosting {
		direction := ai.Combat.Direction
		for i := 1; i <= 4; i++ {
			trail := rl.NewVector2(center.X-float32(direction.X)*float32(i*13), center.Y-float32(direction.Y)*float32(i*13))
			rl.DrawCircleGradient(trail, radius, rl.Fade(tint, .2/float32(i)), rl.Blank)
		}
	}
}

func drawPossessedEyes(bounds rl.Rectangle) {
	tint := rl.NewColor(196, 92, 255, 255)
	width, height := max(4, bounds.Width*.22), max(2, bounds.Height*.1)
	y := bounds.Y + bounds.Height*.36
	for _, offset := range []float32{.25, .75} {
		x := bounds.X + bounds.Width*offset
		center := rl.NewVector2(x, y)
		rl.DrawRectangleRec(rl.NewRectangle(x-width/2-1, y-height/2-1, width+2, height+2), rl.NewColor(35, 15, 46, 255))
		rl.DrawCircleGradient(center, width*1.2, rl.Fade(tint, .5), rl.Blank)
		rl.DrawRectangleRec(rl.NewRectangle(x-width/2, y-height/2, width, height), tint)
		rl.DrawRectangleRec(rl.NewRectangle(x-width*.32, y-height/2, width*.64, max(1, height*.35)), rl.NewColor(237, 190, 255, 255))
	}
}
