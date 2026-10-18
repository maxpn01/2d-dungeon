package ui

import (
	"math"

	"dungeon-raylib-go/internal/rendering"
	"dungeon-raylib-go/internal/rendering/primitives"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Briefing struct {
	elapsed     float64
	soundPlayed bool
}

func NewBriefing() *Briefing {
	return &Briefing{}
}

func (b *Briefing) Draw(width, height int32) Action {
	if !b.soundPlayed {
		rendering.PlayDungeonWelcomeSound()
		b.soundPlayed = true
	}
	b.elapsed += float64(rl.GetFrameTime())
	scale := min(float32(1), min(float32(width-48)/840, float32(height-48)/660))
	x, y := (float32(width)-840*scale)/2, (float32(height)-660*scale)/2
	rect := func(left, top, w, h float32) rl.Rectangle {
		return rl.NewRectangle(x+left*scale, y+top*scale, w*scale, h*scale)
	}
	gold, teal := rl.NewColor(211, 181, 119, 255), rl.NewColor(83, 197, 188, 255)
	muted := rl.NewColor(143, 157, 167, 255)
	alpha := float32(min(1, b.elapsed/.4))
	text := func(label string, top float32, size int32, tint rl.Color) {
		fontSize := max(int32(1), int32(float32(size)*scale))
		for fontSize > 1 && primitives.MeasureText(label, fontSize) > int32(712*scale) {
			fontSize--
		}
		primitives.DrawText(label, (width-primitives.MeasureText(label, fontSize))/2, int32(y+top*scale), fontSize, rl.Fade(tint, alpha))
	}
	line := func(ax, ay, bx, by float32, tint rl.Color) {
		rl.DrawLineEx(rl.NewVector2(x+ax*scale, y+ay*scale), rl.NewVector2(x+bx*scale, y+by*scale), max(float32(1), scale), tint)
	}

	rl.DrawRectangleGradientV(0, 0, width, height, rl.NewColor(7, 13, 20, 255), rl.NewColor(16, 22, 27, 255))

	for row := int32(0); row < height/64+1; row++ {
		for col := int32(-1); col < width/96+1; col++ {
			offset := (row % 2) * 48
			rl.DrawRectangleLines(col*96+offset, row*64, 96, 64, rl.NewColor(24, 32, 39, 100))
		}
	}
	for i := 0; i < 28; i++ {
		px := float32(math.Mod(float64(i*137+31), float64(width)))
		py := float32(math.Mod(float64(i*89)+b.elapsed*(8+float64(i%5)), float64(height)))
		glow := float32(.15 + .12*math.Sin(b.elapsed+float64(i)))
		rl.DrawCircleV(rl.NewVector2(px, float32(height)-py), 2, rl.Fade(teal, glow))
	}
	rl.DrawRectangleRec(rect(10, 10, 840, 660), rl.Fade(rl.Black, .4))
	rl.DrawRectangleRec(rect(0, 0, 840, 660), rl.NewColor(15, 22, 29, 250))
	rl.DrawRectangleLinesEx(rect(0, 0, 840, 660), 1, rl.Fade(gold, .55))
	rl.DrawRectangleLinesEx(rect(9, 9, 822, 642), 1, rl.Fade(gold, .15))
	for _, corner := range [][2]float32{{0, 0}, {840, 0}, {0, 660}, {840, 660}} {
		dx, dy := float32(1), float32(1)
		if corner[0] > 0 {
			dx = -1
		}
		if corner[1] > 0 {
			dy = -1
		}
		line(corner[0], corner[1], corner[0]+28*dx, corner[1], gold)
		line(corner[0], corner[1], corner[0], corner[1]+28*dy, gold)
	}
	text("The dungeon is waking", 66, 46, gold)
	pulse := float32(.5 + .15*math.Sin(b.elapsed*1.8))
	rl.DrawCircleGradient(rl.NewVector2(x+420*scale, y+172*scale), 62*scale, rl.Fade(teal, .12), rl.Fade(teal, 0))
	rl.DrawRectangleLinesEx(rect(394, 136, 52, 70), 2*scale, rl.Fade(gold, pulse))
	rl.DrawRectangleRec(rect(405, 148, 30, 58), rl.NewColor(4, 10, 15, 255))
	line(386, 211, 454, 211, gold)
	line(390, 219, 450, 219, rl.Fade(gold, .4))
	line(270, 174, 350, 174, rl.Fade(gold, .25))
	line(490, 174, 570, 174, rl.Fade(gold, .25))
	text("Something stirs beneath these ancient halls.", 248, 23, rl.RayWhite)

	text("Explore the halls and grow stronger.", 280, 23, muted)
	text("Find a way to the next floor.", 333, 23, gold)

	line(64, 383, 776, 383, rl.Fade(gold, .2))
	controls := [][2]string{{"WASD", "Move"}, {"MOUSE", "Aim"}, {"LEFT CLICK", "Shoot"}, {"SHIFT", "Boost"}, {"E", "Interact"}, {"M", "Map"}, {"P", "Stats"}}
	for i, control := range controls {
		row, col, count := i/3, i%3, 3
		if i >= 3 {
			row, col, count = 1, i-3, 4
		}
		cellWidth := float32(712 / count)
		cx := x + (64+(float32(col)+.5)*cellWidth)*scale
		top := float32(407 + row*65)
		keySize, labelSize := max(int32(1), int32(17*scale)), max(int32(1), int32(18*scale))
		keyWidth := float32(primitives.MeasureText(control[0], keySize)) + 20*scale
		rl.DrawRectangleRounded(rl.NewRectangle(cx-keyWidth/2, y+top*scale, keyWidth, 27*scale), .2, 4, rl.NewColor(31, 43, 51, 255))
		primitives.DrawText(control[0], int32(cx)-primitives.MeasureText(control[0], keySize)/2, int32(y+(top+5)*scale), keySize, teal)
		primitives.DrawText(control[1], int32(cx)-primitives.MeasureText(control[1], labelSize)/2, int32(y+(top+31)*scale), labelSize, muted)
	}
	button := rect(270, 552, 300, 56)
	hovered := rl.CheckCollisionPointRec(rl.GetMousePosition(), button)
	fill := rl.NewColor(33, 51, 53, 255)
	if hovered {
		fill = rl.NewColor(47, 73, 71, 255)
	}
	rl.DrawRectangleRec(button, fill)
	rl.DrawRectangleLinesEx(button, max(float32(1), scale), gold)
	text("Enter the dungeon", 567, 26, rl.RayWhite)

	footer := "Enter or Esc to continue"

	text(footer, 623, 16, muted)
	if b.elapsed > .2 && rl.IsWindowFocused() && (rl.IsKeyPressed(rl.KeyEnter) || hovered && rl.IsMouseButtonPressed(rl.MouseLeftButton)) {
		rendering.PlayMenuClickSound()
		return ActionEnterDungeon
	}
	return ActionNone
}
