package ui

import (
	"image/color"

	"2d-dungeon/internal/rendering/primitives"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var wingTint = color.RGBA{192, 220, 244, 255}

func wingOfferRect(width, height int32) rl.Rectangle {
	w, h := min(float32(620), float32(width)-32), min(float32(620), float32(height)-32)
	return rl.NewRectangle((float32(width)-w)/2, (float32(height)-h)/2, w, h)
}

func wingOfferButtons(width, height int32) [2]rl.Rectangle {
	panel := wingOfferRect(width, height)
	w := (panel.Width - 72) / 2
	return [2]rl.Rectangle{
		rl.NewRectangle(panel.X+28, panel.Y+panel.Height-88, w, 48),
		rl.NewRectangle(panel.X+44+w, panel.Y+panel.Height-88, w, 48),
	}
}

func (p *PerkPanel) readWingOffer(width, height int32) int {
	keys := [2]int32{rl.KeyOne, rl.KeyTwo}
	for i, rect := range wingOfferButtons(width, height) {
		if rl.IsKeyPressed(keys[i]) || rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(rl.GetMousePosition(), rect) {
			p.pending = true
			return i
		}
	}
	return -1
}

func (p *PerkPanel) drawWingOffer(width, height int32) {
	panel := wingOfferRect(width, height)
	rl.DrawRectangle(0, 0, width, height, color.RGBA{6, 10, 17, 220})
	rl.DrawRectangleRec(panel, color.RGBA{19, 27, 38, 255})
	rl.DrawRectangleLinesEx(panel, 1, rl.Fade(wingTint, .5))
	rl.DrawRectangleRec(rl.NewRectangle(panel.X, panel.Y, panel.Width, 3), wingTint)
	x, y, w := int32(panel.X)+28, int32(panel.Y)+28, int32(panel.Width)-56
	primitives.DrawText("THE GUARDIAN WITHOUT A SKY", x, y, 24, wingTint)
	y += 50
	y = wrappedText("Long before these halls sank beneath the earth, a winged guardian carried the lost home. When the gates were sealed, it chose to stay with those who could not escape. Its wings turned to stone, but their light never went out.", x, y, w, 19, rl.LightGray)
	y += 12
	y = wrappedText("A whisper stirs inside the statue: 'I have watched these walls long enough. Take what remains of my sky.'", x, y, w, 19, wingTint)
	y += 18
	primitives.DrawText("UNIQUE BLESSING: STONEWING", x, y, 18, wingTint)
	wrappedText("Replaces Shift boost with flight. Hold Shift to fly over interior walls and pools; release to land. Flight uses 2 MP/sec before blessings. Overdrive improves flight speed and increases its MP drain. If you release above stone, your wings glide back to safe footing. The outer walls still bind you.", x, y+30, w, 17, rl.LightGray)
	labels := [2]string{"[1] ACCEPT THE WINGS", "[2] LEAVE THEM HERE"}
	for i, rect := range wingOfferButtons(width, height) {
		tint := wingTint
		if i == 1 {
			tint = rl.Gray
		}
		hover := !p.pending && rl.CheckCollisionPointRec(rl.GetMousePosition(), rect)
		rl.DrawRectangleRec(rect, rl.Fade(tint, .08))
		if hover {
			rl.DrawRectangleRec(rect, rl.Fade(tint, .12))
		}
		rl.DrawRectangleLinesEx(rect, 1, rl.Fade(tint, .7))
		primitives.DrawText(labels[i], int32(rect.X)+(int32(rect.Width)-primitives.MeasureText(labels[i], 16))/2, int32(rect.Y)+16, 16, tint)
	}
	footer := "Esc to leave / World paused"
	if p.pending {
		footer = "Listening to the guardian..."
	}
	primitives.DrawText(footer, int32(panel.X)+(int32(panel.Width)-primitives.MeasureText(footer, 15))/2, int32(panel.Y+panel.Height)-26, 15, rl.Gray)
}
