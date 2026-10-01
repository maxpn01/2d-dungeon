package worldmap

import (
	mapdata "raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func newFullMapView(windowWidth, windowHeight int32, m *Map) radarView {
	const margin float32 = 72
	left, top, right, bottom := mapdata.Default().Bounds()
	width := max(m.size.X, float32(right)) - float32(left)
	height := max(m.size.Y, float32(bottom)) - float32(top)
	scale := min(max(1, float32(windowWidth)-2*margin)/width, max(1, float32(windowHeight)-2*margin)/height)
	return radarView{
		bounds: rl.NewRectangle((float32(windowWidth)-width*scale)/2, (float32(windowHeight)-height*scale)/2, width*scale, height*scale),
		focus:  rl.NewVector2(float32(left)+width/2, float32(top)+height/2),
		scale:  scale,
	}
}
