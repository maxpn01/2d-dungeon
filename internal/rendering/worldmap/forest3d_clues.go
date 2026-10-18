package worldmap

import (
	"image/color"
	"math"

	"2d-dungeon/internal/parameters"
	mapdata "2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	skinTint  = color.RGBA{236, 230, 206, 254}
	skinScale = color.RGBA{196, 190, 160, 254}
	trackTint = color.RGBA{40, 34, 30, 254}
	boneTint  = color.RGBA{226, 218, 196, 254}
)

func (b *meshBuilder) flatQuad(center, along rl.Vector2, length, width, y float32, c color.RGBA) {
	side := rl.NewVector2(-along.Y, along.X)
	corner := func(l, w float32) rl.Vector3 {
		return v3(center.X+along.X*l+side.X*w, y, center.Y+along.Y*l+side.Y*w)
	}
	b.quad(corner(-length/2, -width/2), corner(length/2, -width/2), corner(length/2, width/2), corner(-length/2, width/2), v3(0, 1, 0), c)
}

func hydraLair(world *mapdata.Map) (rl.Vector2, bool) {
	for _, den := range world.Dens() {
		if den.Type == mapdata.CaveHydra {
			at := den.At.Center()
			return rl.NewVector2(float32(at.X), float32(at.Y)), true
		}
	}
	return rl.Vector2{}, false
}

func (f *forestRenderer) buildClue(world *mapdata.Map, area mapdata.CellRect, kind mapdata.PropKind, solid *meshBuilder) {
	bounds := area.Bounds()
	center := rl.NewVector2(float32(bounds.X+bounds.W/2), float32(bounds.Y+bounds.H/2))
	h := cellHash(area.X, area.Y)
	toward := rl.NewVector2(0, -1)
	if lair, ok := hydraLair(world); ok {
		size := float64(parameters.ActorSize)
		x, y := float64(center.X)-size/2, float64(center.Y)-size/2
		if nx, ny, found := world.NextStep(x, y, float64(lair.X)-size/2, float64(lair.Y)-size/2, size); found && math.Hypot(nx-x, ny-y) > 1 {
			toward = rl.Vector2Normalize(rl.NewVector2(float32(nx-x), float32(ny-y)))
		} else {
			toward = rl.Vector2Normalize(rl.Vector2Subtract(lair, center))
		}
	}
	switch kind {
	case mapdata.TracksProp:
		side := rl.NewVector2(-toward.Y, toward.X)
		for _, s := range []float32{-1, 1} {
			stride := s * 10
			foot := rl.NewVector2(center.X+side.X*s*15+toward.X*stride, center.Y+side.Y*s*15+toward.Y*stride)
			solid.disc(v3(foot.X, .5, foot.Y), 9.5, 9.5, 7, trackTint)
			for toe := -1; toe <= 1; toe++ {
				angle := float64(toe) * .45
				dir := rl.NewVector2(toward.X*float32(math.Cos(angle))-toward.Y*float32(math.Sin(angle)), toward.Y*float32(math.Cos(angle))+toward.X*float32(math.Sin(angle)))
				tip := rl.NewVector2(foot.X+dir.X*16, foot.Y+dir.Y*16)
				solid.flatQuad(tip, dir, 17, 3.6, .6, trackTint)
			}
		}
		scrape := rl.NewVector2(center.X-toward.X*14, center.Y-toward.Y*14)
		solid.flatQuad(scrape, toward, 26, 3, .45, shade(trackTint, 1.4))
	case mapdata.ShedSkinProp:
		side := rl.NewVector2(-toward.Y, toward.X)
		count := 14
		previous := rl.Vector2{}
		for i := range count {
			t := float32(i)/float32(count-1) - .5
			wave := float32(math.Sin(float64(t)*6+float64(unit(h, 1))*6)) * 14
			at := rl.NewVector2(center.X+toward.X*t*150+side.X*wave, center.Y+toward.Y*t*150+side.Y*wave)
			if i > 0 {
				dir := rl.Vector2Normalize(rl.Vector2Subtract(at, previous))
				mid := rl.Vector2Lerp(at, previous, .5)
				width := 24 * (1 - .5*float32(math.Abs(float64(t))))
				tint := skinTint
				if i%2 == 0 {
					tint = skinScale
				}
				solid.flatQuad(mid, dir, rl.Vector2Distance(at, previous)+2, width, .7+float32(i%3)*.15, tint)
				if i%3 == 0 {
					solid.box(v3(mid.X-2, 0, mid.Y-2), v3(mid.X+2, 3, mid.Y+2), skinScale, skinTint)
				}
			}
			previous = at
		}
		head := rl.NewVector2(center.X+toward.X*84, center.Y+toward.Y*84)
		solid.blob(v3(head.X, 2, head.Y), 13, 6, skinTint, h)
	}
}

func (f *forestRenderer) buildLair(world *mapdata.Map, solid *meshBuilder) {
	lair, ok := hydraLair(world)
	if !ok {
		return
	}
	for i := range 26 {
		h := cellHash(i*31+7, i*17+3)
		angle := float64(unit(h, 1)) * 2 * math.Pi
		distance := 70 + unit(h, 2)*130
		at := rl.NewVector2(lair.X+float32(math.Cos(angle))*distance, lair.Y+float32(math.Sin(angle))*distance)
		cell := mapdata.WorldCell(float64(at.X), float64(at.Y))
		if world.Terrain(cell[0], cell[1]) != mapdata.CaveFloor {
			continue
		}
		dir := rl.NewVector2(float32(math.Cos(float64(unit(h, 3))*6)), float32(math.Sin(float64(unit(h, 3))*6)))
		switch h % 4 {
		case 0:
			solid.blob(v3(at.X, 4, at.Y), 6, 5, boneTint, h)
			solid.box(v3(at.X-3, 2, at.Y+3), v3(at.X-1, 4, at.Y+5), trackTint, trackTint)
			solid.box(v3(at.X+1, 2, at.Y+3), v3(at.X+3, 4, at.Y+5), trackTint, trackTint)
		default:
			length := 14 + unit(h, 4)*14
			from := v3(at.X-dir.X*length/2, 2, at.Y-dir.Y*length/2)
			to := v3(at.X+dir.X*length/2, 2, at.Y+dir.Y*length/2)
			solid.cylinder(from, to, 1.8, 5, boneTint, boneTint)
			solid.blob(from, 2.8, 2.4, boneTint, h)
			solid.blob(to, 2.8, 2.4, boneTint, h>>2)
		}
	}
}
