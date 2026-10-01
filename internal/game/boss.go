package game

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/worldmap"
)

func drawEncounter(b simulation.Boss, e *mapExploration) {
	for _, portal := range worldmap.Default().Portals() {
		position := portal.Position()
		pad := rl.NewVector2(float32(position.X), float32(position.Y))
		if !e.entityVisible(float64(pad.X), float64(pad.Y), 0) {
			continue
		}
		active := b.Phase == "" || b.Defeats > 0
		tint := rl.Gray
		label := "DEFEAT THE CROWN DEVOURER"
		if active {
			tint = runeColor
			label = "STAND IN THE CIRCLE TO DESCEND"
			destination, _ := worldmap.Default().Portal(portal.Destination)
			if destination.Floor < portal.Floor {
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
	if b.Phase == "" || b.Phase == "waiting" || b.Phase == "defeated" || !e.entityVisible(b.X, b.Y, simulation.BossSize) {
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

func drawBossSpawnTimer(b simulation.Boss, width int32) {
	if b.Phase == "" {
		return
	}
	waiting := b.Phase == "waiting" || b.Phase == "defeated"
	maxWidth := float32(520)
	if waiting {
		maxWidth = 370
	}
	barWidth := min(float32(width)-32, maxWidth)
	if barWidth <= 0 {
		return
	}
	left := (float32(width) - barWidth) / 2
	y := float32(12)
	if width < 1100 {
		y = 182 // below the player HUD on smaller windows
	}
	label := fmt.Sprintf("%s   Lv.%d", b.Name, b.Level)
	if waiting {
		seconds := int(math.Ceil(float64(b.SpawnTicks) / simulation.TickRate))
		label = fmt.Sprintf("NEXT BOSS  %02d:%02d", seconds/60, seconds%60)
	}
	height := float32(70)
	if waiting {
		height = 47
	}
	rl.DrawRectangleRec(rl.NewRectangle(left-10, y-5, barWidth+20, height), rl.Fade(rl.Black, .85))
	fontSize := int32(18)
	for fontSize > 10 && float32(measureText(label, fontSize)) > barWidth {
		fontSize--
	}
	drawTextAt(label, rl.NewVector2(left+(barWidth-float32(measureText(label, fontSize)))/2, y), fontSize, rl.Gold)
	if waiting {
		rl.DrawRectangleRec(rl.NewRectangle(left, y+27, barWidth, 9), rl.NewColor(49, 29, 42, 255))
		progress := 1 - float32(b.SpawnTicks)/float32(max(1, b.SpawnDuration))
		rl.DrawRectangleRec(rl.NewRectangle(left, y+27, barWidth*max(0, min(1, progress)), 9), rl.Gold)
	} else {
		rl.DrawRectangleRec(rl.NewRectangle(left, y+27, barWidth, 17), rl.NewColor(49, 29, 42, 255))
		progress := float32(max(0, min(1, b.HP/max(1, b.MaxHP))))
		rl.DrawRectangleRec(rl.NewRectangle(left, y+27, barWidth*progress, 17), rl.NewColor(190, 78, 115, 255))
		hp := fmt.Sprintf("%.0f / %.0f HP", b.HP, b.MaxHP)
		drawTextAt(hp, rl.NewVector2(left+(barWidth-float32(measureText(hp, 16)))/2, y+49), 16, rl.RayWhite)
	}
}
