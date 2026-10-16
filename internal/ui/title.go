package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawPixelTitle(title string, centerX, y, pixelSize int32, tint color.RGBA) {
	var glyphs = map[rune][7]uint8{
		'2': {14, 17, 1, 2, 4, 8, 31},
		'D': {30, 17, 17, 17, 17, 17, 30},
		'S': {15, 16, 16, 14, 1, 1, 30},
		'N': {17, 25, 25, 21, 19, 19, 17},
		'B': {30, 17, 17, 30, 17, 17, 30},
		'X': {17, 17, 10, 4, 10, 17, 17},
		'G': {14, 17, 16, 23, 17, 17, 14},
		'A': {14, 17, 17, 31, 17, 17, 17},
		'M': {17, 27, 21, 21, 17, 17, 17},
		'E': {31, 16, 16, 30, 16, 16, 31},
		'O': {14, 17, 17, 17, 17, 17, 14},
		'V': {17, 17, 17, 17, 17, 10, 4},
		'R': {30, 17, 17, 30, 20, 18, 17},
	}

	width := int32(len(title)*6 - 1)

	x := centerX - width*pixelSize/2

	for _, letter := range title {
		for row, bits := range glyphs[letter] {
			for col := range 5 {
				if bits&(1<<uint(4-col)) != 0 {
					rl.DrawRectangle(x+int32(col)*pixelSize, y+int32(row)*pixelSize, pixelSize, pixelSize, tint)
				}
			}
		}
		x += 6 * pixelSize
	}
}
