package hud

import (
	"fmt"
	"math"

	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawBossTimer(b simulation.Boss, width int32) {
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
	for fontSize > 10 && float32(primitives.MeasureText(label, fontSize)) > barWidth {
		fontSize--
	}
	primitives.DrawTextAt(label, rl.NewVector2(left+(barWidth-float32(primitives.MeasureText(label, fontSize)))/2, y), fontSize, rl.Gold)
	if waiting {
		rl.DrawRectangleRec(rl.NewRectangle(left, y+27, barWidth, 9), rl.NewColor(49, 29, 42, 255))
		progress := 1 - float32(b.SpawnTicks)/float32(max(1, b.SpawnDuration))
		rl.DrawRectangleRec(rl.NewRectangle(left, y+27, barWidth*max(0, min(1, progress)), 9), rl.Gold)
	} else {
		rl.DrawRectangleRec(rl.NewRectangle(left, y+27, barWidth, 17), rl.NewColor(49, 29, 42, 255))
		progress := float32(max(0, min(1, b.HP/max(1, b.MaxHP))))
		rl.DrawRectangleRec(rl.NewRectangle(left, y+27, barWidth*progress, 17), rl.NewColor(190, 78, 115, 255))
		hp := fmt.Sprintf("%.0f / %.0f HP", b.HP, b.MaxHP)
		primitives.DrawTextAt(hp, rl.NewVector2(left+(barWidth-float32(primitives.MeasureText(hp, 16)))/2, y+49), 16, rl.RayWhite)
	}
}
