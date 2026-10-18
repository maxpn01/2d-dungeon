package display

import "dungeon-raylib-go/internal/parameters"

var width, height = int32(parameters.WindowWidth), int32(parameters.WindowHeight)

func Size(w, h int32) {
	width, height = w, h
}

func Width() int32  { return width }
func Height() int32 { return height }
