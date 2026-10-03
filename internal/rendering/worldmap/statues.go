package worldmap

import (
	"fmt"
	"image/color"
	"math"
	"raylib-2d-sandbox/internal/parameters"

	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"
	mapdata "raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawStatues(statues []simulation.StatueState, active string, px, py float64, dead bool, exploration *Exploration) {
	for _, s := range statues {
		if !exploration.EntityVisible(s.X, s.Y, 0) {
			continue
		}
		x, y := int32(s.X), int32(s.Y)
		tint := primitives.RuneColor
		winged := mapdata.IsWingStatue(s.ID)
		if winged {
			tint = color.RGBA{192, 220, 244, 255}
		}
		if s.ClaimedBy != "" {
			tint = color.RGBA{73, 83, 88, 255}
		}
		if !winged {
			drawStatueCharge(s, tint)
		}
		if winged {
			drawStatueWings(x, y, tint)
		}
		rl.DrawEllipse(x, y+19, 31, 12, rl.Fade(rl.Black, .5))
		rl.DrawRectangle(x-27, y+10, 54, 12, stone)
		rl.DrawRectangle(x-22, y+6, 44, 7, paleStone)
		rl.DrawRectangle(x-13, y-25, 26, 31, stone)
		rl.DrawRectangle(x-9, y-41, 18, 17, paleStone)
		rl.DrawRectangle(x-23, y-20, 9, 17, paleStone)
		rl.DrawRectangle(x+14, y-20, 9, 17, paleStone)
		rl.DrawRectangle(x-20, y-23, 40, 29, color.RGBA{29, 39, 47, 255})
		rl.DrawRectangleLines(x-20, y-23, 40, 29, paleStone)
		for i := int32(0); i < 3; i++ {
			rl.DrawLine(x-12, y-16+i*6, x+12-i*4, y-16+i*6, tint)
		}
		if winged {
			drawStatueWisps(x, y)
		}
		if s.ClaimedBy == "" {
			if !winged {
				rl.DrawCircleLines(x, y-9, 37, rl.Fade(tint, .3))
			}
			if !dead && active == "" && s.Near(px, py) {
				label := "[E] READ INSCRIPTION"
				if winged {
					label = "[E] LISTEN TO THE GUARDIAN"
				}
				width := primitives.MeasureText(label, 16)
				rl.DrawRectangle(x-width/2-6, y-71, width+12, 24, rl.Fade(rl.Black, .8))
				primitives.DrawText(label, x-width/2, y-68, 16, tint)
			}
		} else if !dead && !winged && s.Near(px, py) {
			label := fmt.Sprintf("RECHARGING %.0fs", math.Ceil(s.RechargeRemaining))
			primitives.DrawText(label, x-primitives.MeasureText(label, 14)/2, y-68, 14, rl.Gray)
		}
	}
}

func drawStatueWings(x, y int32, tint color.RGBA) {
	center := rl.NewVector2(float32(x), float32(y-16))
	pulse := float32(.65 + .15*math.Sin(rl.GetTime()*1.8))
	rl.DrawCircleGradient(center, 88, rl.Fade(rl.NewColor(158, 115, 225, 255), .1*pulse), rl.Blank)
	for _, side := range []int32{-1, 1} {
		for i := int32(0); i < 5; i++ {
			root := rl.NewVector2(float32(x+side*12), float32(y-18+i*3))
			tip := rl.NewVector2(float32(x+side*(44-i*4)), float32(y-53+i*10))
			end := rl.NewVector2(float32(x+side*20), float32(y-5+i*2))
			if side < 0 {
				rl.DrawTriangle(root, tip, end, paleStone)
			} else {
				rl.DrawTriangle(root, end, tip, paleStone)
			}
			rl.DrawLineEx(root, tip, 1, rl.Fade(tint, .6*pulse))
		}
	}
}

func statueWispPosition(center rl.Vector2, time, phase, speed float64) rl.Vector2 {
	angle := time*speed + phase
	radius := 43 + 11*math.Sin(time*.6+phase)
	return rl.NewVector2(
		center.X+float32(math.Cos(angle)*radius+8*math.Sin(time*.9+phase)),
		center.Y+float32(math.Sin(angle)*radius*.65+12*math.Sin(time*.75+phase*1.6)),
	)
}

func drawStatueWisps(x, y int32) {
	center := rl.NewVector2(float32(x), float32(y-18))
	time := rl.GetTime()
	for i := 0; i < 4; i++ {
		phase := float64(i)*2.37 + float64(x)*.007 + float64(y)*.003
		speed := .3 + float64(i)*.04
		at := statueWispPosition(center, time, phase, speed)
		tint := rl.NewColor(176, 132, 241, 255)
		if i%2 == 1 {
			tint = rl.NewColor(178, 213, 249, 255)
		}
		brightness := float32(.7 + .15*math.Sin(time*2+phase))
		for trail := 1; trail <= 4; trail++ {
			previous := statueWispPosition(center, time-float64(trail)*.1, phase, speed)
			rl.DrawCircleV(previous, .7, rl.Fade(tint, brightness*float32(5-trail)*.065))
		}
		rl.DrawCircleGradient(at, 13, rl.Fade(tint, .2*brightness), rl.Blank)
		span := float32(1.2 + 3.8*math.Abs(math.Sin(time*9+phase)))
		bob := float32(math.Sin(time*3 + phase))
		left := rl.NewVector2(at.X-span, at.Y-3+bob)
		right := rl.NewVector2(at.X+span, at.Y-3-bob)
		leftLower := rl.NewVector2(at.X-span*.7, at.Y+2)
		rightLower := rl.NewVector2(at.X+span*.7, at.Y+2)
		rl.DrawTriangle(at, left, leftLower, rl.Fade(tint, brightness))
		rl.DrawTriangle(at, rightLower, right, rl.Fade(tint, brightness))
		rl.DrawLineV(rl.NewVector2(at.X, at.Y-1.5), rl.NewVector2(at.X, at.Y+1.5), rl.Fade(rl.RayWhite, brightness))
	}
}

func drawStatueCharge(s simulation.StatueState, tint color.RGBA) {
	center := rl.NewVector2(float32(s.X), float32(s.Y-9))
	if s.ClaimedBy != "" {
		progress := float32(1 - s.RechargeRemaining/parameters.StatueRechargeSeconds)
		rl.DrawCircleLinesV(center, 37, rl.Fade(tint, .3))
		if progress > 0 {
			rl.DrawRing(center, 36, 38, -90, -90+360*progress, 64, rl.Fade(primitives.RuneColor, .4))
		}
		return
	}
	phase := rl.GetTime()*1.5 + s.X*.01
	pulse := float32(.7 + .2*math.Sin(phase))
	rl.DrawCircleGradient(center, 65, rl.Fade(tint, .18*pulse), rl.Blank)
	for i := 0; i < 4; i++ {
		angle := phase + float64(i)*math.Pi/2
		mote := rl.NewVector2(center.X+float32(math.Cos(angle))*39, center.Y+float32(math.Sin(angle))*30)
		rl.DrawCircleGradient(mote, 7, rl.Fade(tint, .5*pulse), rl.Blank)
		rl.DrawCircleV(mote, 1.5, rl.Fade(rl.RayWhite, .7*pulse))
	}
}
