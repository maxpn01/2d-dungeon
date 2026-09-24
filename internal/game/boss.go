package game

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/terrain"
)

func drawEncounter(b simulation.Boss, e *mapExploration) {
	for _, pad := range []rl.Vector2{{X: terrain.DescentCircleX, Y: terrain.DescentCircleY}, {X: terrain.ReturnCircleX, Y: terrain.ReturnCircleY}} {
		if !e.entityVisible(float64(pad.X), float64(pad.Y), 0) {
			continue
		}
		active := b.Phase == "" || b.Phase == "defeated"
		tint := rl.Gray
		label := "DEFEAT THE CROWN DEVOURER"
		if active {
			tint = runeColor
			label = "STAND IN THE CIRCLE TO DESCEND"
			if pad.Y == terrain.ReturnCircleY {
				label = "STAND IN THE CIRCLE TO RETURN"
			}
		}
		pulse := float32(.65 + .2*math.Sin(rl.GetTime()*3))
		if active {
			rl.DrawCircleGradient(pad, 110, rl.Fade(tint, .22*pulse), rl.Blank)
		}
		for _, radius := range []float32{60, 82, 90} {
			rl.DrawCircleLines(int32(pad.X), int32(pad.Y), radius, rl.Fade(tint, pulse))
		}
		rotation := float32(30)
		if active {
			rotation += float32(rl.GetTime() * 25)
		}
		rl.DrawPolyLines(pad, 6, 78, rotation, tint)
		drawTextAt(label, rl.NewVector2(pad.X-float32(measureText(label, 14))/2, pad.Y+105), 14, tint)
	}
	if b.Phase == "fighting" || b.Phase == "gathering" {
		tint := rl.Orange
		if b.Phase == "fighting" {
			tint = rl.Red
		}
		// Mark the whole sealed boundary, including both side entrances.
		r := rl.NewRectangle(simulation.ArenaLeft, simulation.ArenaTop, simulation.ArenaRight-simulation.ArenaLeft, simulation.ArenaBottom-simulation.ArenaTop)
		rl.DrawRectangleLinesEx(r, 6, rl.Fade(tint, .65))
		rl.DrawRectangleLinesEx(rl.NewRectangle(r.X+7, r.Y+7, r.Width-14, r.Height-14), 2, rl.Fade(tint, .25))
	}
	if b.Phase == "" || b.Phase == "defeated" || !e.entityVisible(b.X, b.Y, simulation.BossSize) {
		return
	}
	if b.Phase == "fighting" && b.Behavior == "winding" {
		start := rl.NewVector2(float32(b.X+simulation.BossSize/2), float32(b.Y+simulation.BossSize/2))
		target := rl.NewVector2(float32(b.AttackTarget.X+simulation.BossSize/2), float32(b.AttackTarget.Y+simulation.BossSize/2))
		rl.DrawLineEx(start, target, simulation.BossSize, rl.Fade(rl.Orange, .12))
		rl.DrawLineEx(start, target, 4, rl.Fade(rl.Orange, .85))
		rl.DrawCircleLines(int32(target.X), int32(target.Y), 24, rl.Orange)
	}
	tint := rl.NewColor(161, 97, 191, 255)
	if b.Enraged {
		tint = rl.NewColor(225, 60, 85, 255)
	}
	if b.Accelerating {
		tint = rl.Orange
	}
	for i := len(b.Tail) - 1; i >= 0; i-- {
		s := b.Tail[i]
		r := rl.NewRectangle(float32(s.X), float32(s.Y), simulation.BossSize, simulation.BossSize)
		rl.DrawRectangleRec(r, rl.Fade(tint, .85))
		rl.DrawRectangleLinesEx(r, 2, rl.Fade(rl.Black, .5))
		rl.DrawCircleV(rl.NewVector2(r.X+18, r.Y+18), 4, rl.Fade(rl.Gold, .5))
	}
	center := rl.NewVector2(float32(b.X+simulation.BossSize/2), float32(b.Y+simulation.BossSize/2))
	if b.Enraged {
		rl.DrawRing(center, 43, 46, 0, 360, 40, rl.Fade(rl.Red, float32(.5+.2*math.Sin(rl.GetTime()*12))))
	}
	if b.Accelerating {
		direction := rl.Vector2Normalize(rl.Vector2Subtract(rl.NewVector2(float32(b.AttackTarget.X+simulation.BossSize/2), float32(b.AttackTarget.Y+simulation.BossSize/2)), center))
		for i := 1; i <= 4; i++ {
			trail := rl.Vector2Subtract(center, rl.Vector2Scale(direction, float32(i)*18))
			rl.DrawCircleV(trail, float32(18-i*2), rl.Fade(tint, float32(5-i)*.08))
		}
	}
	rl.DrawCircleGradient(center, 48, rl.Fade(tint, .35), rl.Blank)
	rl.DrawRectangleRec(rl.NewRectangle(float32(b.X), float32(b.Y), simulation.BossSize, simulation.BossSize), tint)
	rl.DrawRectangleLinesEx(rl.NewRectangle(float32(b.X), float32(b.Y), simulation.BossSize, simulation.BossSize), 2, rl.Gold)
	rl.DrawCircleV(rl.NewVector2(center.X-8, center.Y-5), 4, rl.RayWhite)
	rl.DrawCircleV(rl.NewVector2(center.X+8, center.Y-5), 4, rl.RayWhite)
	label := fmt.Sprintf("BOSS Lv.%d", b.Level)
	if b.Phase == "waiting" || b.Phase == "gathering" {
		label += " / " + b.Behavior
	}
	drawTextAt(label, rl.NewVector2(center.X-float32(measureText(label, 16))/2, center.Y-58), 16, rl.Gold)
}

func drawTeleportCharge(x, y, charge float64) {
	if charge <= 0 {
		return
	}
	center := rl.NewVector2(float32(x+simulation.PlayerSize/2), float32(y+simulation.PlayerSize/2))
	progress := float32(min(1, charge/simulation.PortalSeconds))
	rl.DrawCircleGradient(center, 80, rl.Fade(runeColor, progress*.65), rl.Blank)
	rl.DrawRing(center, 35, 40, -90, -90+360*progress, 48, rl.RayWhite)
}

func drawBossHUD(b simulation.Boss, x, y float64, width int32) {
	if terrain.FloorTwo(x, y) || y > simulation.ArenaBottom+160 || y < simulation.ArenaTop-160 {
		return
	}
	if b.Phase == "defeated" {
		drawCenteredText("DEVOURER DEFEATED - THE DESCENT IS OPEN", 235, 20, runeColor)
		return
	}
	if b.Phase != "gathering" && b.Phase != "fighting" {
		return
	}
	barWidth := min(float32(width)-48, float32(580))
	left := (float32(width) - barWidth) / 2
	rl.DrawRectangleRec(rl.NewRectangle(left-12, 224, barWidth+24, 104), rl.Fade(rl.Black, .85))
	label := fmt.Sprintf("%s   Lv.%d", b.Name, b.Level)
	if b.Enraged {
		label += "  /  ENRAGED"
	}
	drawCenteredText(label, 230, 22, rl.Gold)
	rl.DrawRectangleRec(rl.NewRectangle(left, 261, barWidth, 17), rl.NewColor(49, 29, 42, 255))
	rl.DrawRectangleRec(rl.NewRectangle(left, 261, barWidth*float32(b.HP/max(1, b.MaxHP)), 17), rl.NewColor(190, 78, 115, 255))
	drawCenteredText(fmt.Sprintf("%.0f / %.0f HP   |   %d tail segments", b.HP, b.MaxHP, len(b.Tail)), 282, 16, rl.RayWhite)
	hint := "DASH INTO THE SNAKE - CONTEST ITS FOOD"
	if b.Phase == "gathering" {
		hint = fmt.Sprintf("ENTER TOGETHER - EXITS SEAL IN %.0f", math.Ceil(b.Countdown))
	} else {
		switch b.Behavior {
		case "winding":
			total := 2
			if b.Enraged {
				total = 3
			}
			hint = fmt.Sprintf("RUSH %d/%d - MOVE OFF THE LINE", b.Chain, total)
		case "rushing":
			hint = "DODGE THROUGH THE GAP - WATCH FOR THE FOLLOW-UP"
		case "recovering":
			hint = "COMBO FINISHED - STRIKE NOW"
		case "feeding":
			hint = "FEEDING - CONTEST ITS FOOD"
		}
	}
	drawCenteredText(hint, 305, 14, rl.Orange)
}
