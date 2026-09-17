package game

import (
	"image/color"
	"math"

	"raylib-2d-sandbox/assets"

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

func unloadTextFonts() {
	for size, font := range textFonts {
		rl.UnloadFont(font)
		delete(textFonts, size)
	}
}

func drawText(text string, x, y, size int32, tint color.RGBA) {
	drawTextAt(text, rl.NewVector2(float32(x), float32(y)), size, tint)
}

func drawTextAt(text string, position rl.Vector2, size int32, tint color.RGBA) {
	rl.DrawTextEx(fontForSize(size), text, position, float32(size), 0, tint)
}

func measureText(text string, size int32) int32 {
	width := rl.MeasureTextEx(fontForSize(size), text, float32(size), 0).X
	return int32(math.Ceil(float64(width)))
}
