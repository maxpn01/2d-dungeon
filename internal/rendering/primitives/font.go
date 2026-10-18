package primitives

import (
	"image/color"
	"math"

	"2d-dungeon/assets"
	"2d-dungeon/internal/display"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var textFonts = make(map[int32]rl.Font)

func fontForSize(size int32) rl.Font {
	if font, ok := textFonts[size]; ok {
		return font
	}
	data := assets.PixelOperator
	if size >= 24 {
		data = assets.PixelOperatorBold
	}
	font := rl.LoadFontFromMemory(".ttf", data, size, nil)
	rl.SetTextureFilter(font.Texture, rl.FilterPoint)
	textFonts[size] = font
	return font
}

func UnloadTextFonts() {
	for size, font := range textFonts {
		rl.UnloadFont(font)
		delete(textFonts, size)
	}
}

func DrawText(text string, x, y, size int32, tint color.RGBA) {
	DrawTextAt(text, rl.NewVector2(float32(x), float32(y)), size, tint)
}

func DrawTextAt(text string, position rl.Vector2, size int32, tint color.RGBA) {
	rl.DrawTextEx(fontForSize(size), text, position, float32(size), 0, tint)
}

func MeasureText(text string, size int32) int32 {
	width := rl.MeasureTextEx(fontForSize(size), text, float32(size), 0).X
	return int32(math.Ceil(float64(width)))
}

func DrawCenteredText(text string, y, size int32, tint color.RGBA) {
	DrawText(text, (int32(display.Width())-MeasureText(text, size))/2, y, size, tint)
}
