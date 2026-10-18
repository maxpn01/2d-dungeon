package worldmap

import (
	"image/color"
	"math"
	"sort"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/rendering/primitives"
	mapdata "dungeon-raylib-go/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var EmberTimeOffset float64

const emberGroundRows = 3

var (
	emberLava     = color.RGBA{255, 112, 34, 255}
	emberHotLava  = color.RGBA{255, 214, 120, 255}
	emberBasalt   = color.RGBA{30, 20, 24, 255}
	emberCrust    = color.RGBA{78, 44, 40, 255}
	emberSmoke    = color.RGBA{38, 28, 32, 255}
	emberSkyTop   = color.RGBA{22, 6, 14, 255}
	emberSkyLow   = color.RGBA{132, 40, 24, 255}
	emberFarRidge = color.RGBA{70, 22, 26, 255}
	emberMidRidge = color.RGBA{34, 12, 16, 255}
)

func emberTime() float64 { return rl.GetTime() + EmberTimeOffset }

func (m *Map) InEmberHollow() bool { return m.floor == mapdata.ThirdFloor }

func emberRoom() (mapdata.Room, bool) { return mapdata.Default().Room(mapdata.EmberHollow) }

func EmberCameraTarget(focus rl.Vector2, width, height int32) (rl.Vector2, bool) {
	world := mapdata.Default()
	room, ok := emberRoom()
	if !ok || world.FloorAt(float64(focus.X), float64(focus.Y)) != mapdata.ThirdFloor {
		return focus, false
	}
	b := room.Bounds()
	w, h := float32(width), float32(height)
	x := float32(b.X + b.W/2)
	if float32(b.W) > w {
		x = min(max(focus.X, float32(b.X)+w/2), float32(b.X+b.W)-w/2)
	}
	y := min(max(focus.Y, float32(b.Y)+h/2-80), float32(b.Y+b.H)-.4*h)
	return rl.NewVector2(x, y), true
}

func emberHash(i int, salt float64) float64 {
	v := math.Sin(float64(i)*127.1+salt*311.7) * 43758.5453
	return v - math.Floor(v)
}

func coloredTriangle(a, b, c rl.Vector2, tint color.RGBA) {
	if (b.X-a.X)*(c.Y-a.Y)-(b.Y-a.Y)*(c.X-a.X) > 0 {
		b, c = c, b
	}
	rl.DrawTriangle(a, b, c, tint)
}

type volcano struct {
	cx, base, height, half, crater float32
}

func (v volcano) surface(x float32) float32 {
	d := float32(math.Abs(float64(x - v.cx)))
	if d >= v.half {
		return v.base
	}
	if d < v.crater {
		return v.base - v.height + v.height*.05*(1-(d/v.crater)*(d/v.crater))
	}
	slope := float64((d - v.crater) / (v.half - v.crater))
	rise := math.Pow(1-slope, 1.6)
	noise := math.Sin(float64(x)*.045)*.012 + math.Sin(float64(x)*.13)*.006
	return v.base - v.height*float32(rise+noise*(1-slope))
}

func (v volcano) peak() rl.Vector2 { return rl.NewVector2(v.cx, v.base-v.height*.97) }

func ridgeHeight(x, seed, amplitude float32) float32 {
	fx := float64(x)
	s := float64(seed)
	n := math.Sin(fx*.006+s)*.5 + math.Sin(fx*.017+s*2.1)*.3 + math.Sin(fx*.041+s*3.7)*.2
	return amplitude * float32(.6+.4*n)
}

func (m *Map) DrawEmberBackdrop(ctx primitives.Context) {
	room, ok := emberRoom()
	if !ok {
		return
	}
	t := emberTime()
	w, h := float32(ctx.WindowWidth), float32(ctx.WindowHeight)
	b := room.Bounds()
	ground := rl.GetWorldToScreen2D(rl.NewVector2(float32(b.X), float32(b.Y+b.H)-emberGroundRows*parameters.Tile), ctx.Camera)
	horizon := ground.Y + 6
	shift := ctx.Camera.Target.X - float32(b.X+b.W/2)
	rl.DrawRectangleGradientV(0, 0, int32(w), int32(horizon), emberSkyTop, emberSkyLow)
	rl.DrawRectangle(0, int32(horizon), int32(w), int32(h-horizon), emberMidRidge)
	drawEmberClouds(w, horizon, t)
	for x := float32(-4); x < w+4; x += 4 {
		top := horizon - 40 - ridgeHeight(x+shift*.08, 1.3, 120)
		rl.DrawRectangleGradientV(int32(x), int32(top), 4, int32(horizon-top)+2, emberFarRidge, color.RGBA{96, 30, 26, 255})
	}
	v := volcano{cx: w*.56 - shift*.18, base: horizon + 8, height: min(h*.5, 540), crater: 52}
	v.half = v.height * 1.35
	phase := math.Mod(t, 7)
	eruption := float32(0)
	if phase < 3 {
		eruption = float32(math.Min(1, phase*5) * math.Pow(1-phase/3, 1.5))
	}
	drawVolcanoPlume(v, t, eruption)
	peak := v.peak()
	glow := float32(.45+.12*math.Sin(t*2.3)) + .5*eruption
	rl.DrawCircleGradient(peak, 170+90*eruption, rl.Fade(emberLava, .32*glow), rl.Blank)
	for x := v.cx - v.half; x < v.cx+v.half; x += 3 {
		top := v.surface(x)
		near := 1 - float32(math.Abs(float64(x-v.cx)))/v.half
		lit := color.RGBA{uint8(52 + 36*near), uint8(18 + 8*near), 20, 255}
		rl.DrawRectangleGradientV(int32(x), int32(top), 3, int32(v.base-top)+2, lit, color.RGBA{24, 8, 12, 255})
		rl.DrawRectangle(int32(x), int32(top), 3, 2, rl.Fade(color.RGBA{150, 58, 36, 255}, near*near))
	}
	rl.DrawEllipse(int32(v.cx), int32(peak.Y+4), v.crater*.9, 7, rl.Fade(emberHotLava, .7+.3*glow))
	drawLavaStreams(v, t)
	drawLavaFountain(v, eruption)
	drawLavaBombs(v, phase)
	for x := float32(-4); x < w+4; x += 4 {
		top := horizon - ridgeHeight(x+shift*.4, 4.2, 110)
		rl.DrawRectangle(int32(x), int32(top), 4, int32(horizon-top)+4, emberMidRidge)
	}
	for i := range 9 {
		x := float32(emberHash(i, 2))*w*1.2 - shift*.4 - w*.1
		x = float32(math.Mod(float64(x)+float64(w)*1.2, float64(w)*1.2)) - w*.1
		flicker := float32(.5 + .5*math.Sin(t*3+float64(i)))
		rl.DrawCircleGradient(rl.NewVector2(x, horizon), 50, rl.Fade(emberLava, .22*flicker), rl.Blank)
	}
	rl.DrawRectangleGradientV(0, int32(horizon-90), int32(w), 92, rl.Blank, rl.Fade(color.RGBA{255, 96, 40, 255}, .28))
}

func drawEmberClouds(w, horizon float32, t float64) {
	for i := range 7 {
		speed := 6 + emberHash(i, 5)*10
		span := float64(w) + 600
		x := float32(math.Mod(emberHash(i, 1)*span+t*speed, span)) - 300
		y := horizon * float32(.12+emberHash(i, 3)*.45)
		rx, ry := float32(180+emberHash(i, 4)*220), float32(18+emberHash(i, 6)*20)
		rl.DrawEllipse(int32(x), int32(y), rx, ry, rl.Fade(color.RGBA{52, 16, 22, 255}, .55))
		rl.DrawEllipse(int32(x), int32(y+ry*.45), rx*.8, ry*.35, rl.Fade(color.RGBA{170, 60, 34, 255}, .25))
	}
}

func drawVolcanoPlume(v volcano, t float64, eruption float32) {
	peak := v.peak()
	type puff struct {
		at     rl.Vector2
		radius float32
		age    float32
	}
	puffs := make([]puff, 0, 26)
	for i := range 26 {
		age := float32(math.Mod(t*.06+float64(i)/26, 1))
		drift := float32(math.Sin(float64(i)*1.7+t*.3)) * 34 * age
		at := rl.NewVector2(peak.X+age*age*300+drift, peak.Y-10-age*v.height*.95)
		puffs = append(puffs, puff{at, 20 + age*120, age})
	}
	sort.Slice(puffs, func(i, j int) bool { return puffs[i].age > puffs[j].age })
	for _, p := range puffs {
		fade := (1 - p.age) * min(1, p.age*10)
		rl.DrawCircleV(p.at, p.radius, rl.Fade(emberSmoke, .62*fade))
		heat := (1 - p.age) * (1 - p.age) * (1 - p.age)
		rl.DrawCircleV(rl.NewVector2(p.at.X, p.at.Y+p.radius*.35), p.radius*.7, rl.Fade(emberLava, (.22+.35*eruption)*heat))
	}
}

func drawLavaStreams(v volcano, t float64) {
	for i, side := range []float32{-1, 1, -1, 1} {
		length := float32(.55 + .2*emberHash(i, 8))
		start := v.crater * (.5 + .3*float32(i%2))
		var prev rl.Vector2
		for step := 0; step <= 40; step++ {
			s := float32(step) / 40 * length
			x := v.cx + side*(start+s*(v.half-start)*(.35+.15*float32(i/2))) + float32(math.Sin(float64(s)*11+float64(i)*2))*12*s
			at := rl.NewVector2(x, v.surface(x)+4+s*v.height*(.18+.08*float32(i/2)))
			if step > 0 {
				width := 5 * (1 - s*.7)
				pulse := float32(.65 + .35*math.Sin(float64(s)*24-t*3.2+float64(i)))
				rl.DrawLineEx(prev, at, width*3, rl.Fade(emberLava, .18*pulse))
				rl.DrawLineEx(prev, at, width, rl.Fade(emberLava, .55+.4*pulse))
				if pulse > .93 {
					rl.DrawCircleV(at, width*.6, emberHotLava)
				}
			}
			prev = at
		}
	}
}

func drawLavaFountain(v volcano, eruption float32) {
	if eruption <= 0 {
		return
	}
	peak := v.peak()
	height := 260 * eruption
	rl.DrawRectangleGradientV(int32(peak.X-16*eruption), int32(peak.Y-height), int32(32*eruption)+1, int32(height), rl.Blank, rl.Fade(emberHotLava, .85))
	rl.DrawCircleGradient(peak, 60+50*eruption, rl.Fade(emberHotLava, .5*eruption), rl.Blank)
}

func drawLavaBombs(v volcano, phase float64) {
	peak := v.peak()
	const gravity = 430.0
	for i := range 18 {
		launch := float64(i) * .07
		vx := (emberHash(i, 11) - .5) * 300
		vy := -(360 + emberHash(i, 12)*340)
		for trail := range 5 {
			tt := phase - launch - float64(trail)*.035
			if tt <= 0 {
				continue
			}
			at := rl.NewVector2(peak.X+float32(vx*tt), peak.Y+float32(vy*tt+.5*gravity*tt*tt))
			if at.Y > v.surface(at.X) {
				break
			}
			radius := (4 - float32(trail)*.7) * float32(.7+emberHash(i, 13)*.6)
			tint := emberHotLava
			if trail > 0 {
				tint = emberLava
			}
			rl.DrawCircleV(at, radius, rl.Fade(tint, 1-float32(trail)*.18))
			if trail == 0 {
				rl.DrawCircleGradient(at, radius*4, rl.Fade(emberLava, .35), rl.Blank)
			}
		}
	}
}

func (m *Map) emberSolid(area mapdata.CellRect, x, y int) bool {
	if y < area.Y {
		return false
	}
	if x < area.X || x >= area.Right() || y >= area.Bottom() {
		return true
	}
	return m.world.Cell(x, y) == 0 || m.world.TerrainBlocks(x, y)
}

func (m *Map) DrawEmberHollow(ctx primitives.Context) {
	room, ok := emberRoom()
	if !ok {
		return
	}
	t := emberTime()
	area := room.Area
	zoom := max(ctx.Camera.Zoom, .1)
	left := int(math.Floor(float64(ctx.Camera.Target.X-ctx.Camera.Offset.X/zoom)/parameters.Tile)) - 1
	right := int(math.Ceil(float64(ctx.Camera.Target.X+(float32(ctx.WindowWidth)-ctx.Camera.Offset.X)/zoom)/parameters.Tile)) + 1
	top := int(math.Floor(float64(ctx.Camera.Target.Y-ctx.Camera.Offset.Y/zoom)/parameters.Tile)) - 1
	bottom := int(math.Ceil(float64(ctx.Camera.Target.Y+(float32(ctx.WindowHeight)-ctx.Camera.Offset.Y)/zoom)/parameters.Tile)) + 1
	for y := max(top, area.Y); y < bottom; y++ {
		for x := left; x < right; x++ {
			if !m.emberSolid(area, x, y) {
				continue
			}
			m.drawEmberCell(area, x, y, t)
		}
	}
	for y := max(top, area.Y); y < bottom; y++ {
		for x := left; x < right; x++ {
			if m.emberSolid(area, x, y) && !m.emberSolid(area, x, y+1) {
				drawLedgeUnderside(x, y, t)
			}
		}
	}
}

func (m *Map) drawEmberCell(area mapdata.CellRect, x, y int, t float64) {
	px, py := float32(x*parameters.Tile), float32(y*parameters.Tile)
	hash := uint32(x*73856093) ^ uint32((y+400)*19349663)
	tone := uint8(hash % 9)
	surface := !m.emberSolid(area, x, y-1)
	depth := 0
	for depth < 4 && m.emberSolid(area, x, y-depth-1) {
		depth++
	}
	shade := uint8(depth * 3)
	base := color.RGBA{emberBasalt.R + tone - shade, emberBasalt.G + tone/2 - shade/2, emberBasalt.B + tone/2 - shade/2, 255}
	rl.DrawRectangleV(rl.NewVector2(px, py), rl.NewVector2(parameters.Tile, parameters.Tile), base)
	for col := float32(6 + hash%9); col < parameters.Tile; col += 13 {
		rl.DrawRectangleV(rl.NewVector2(px+col, py), rl.NewVector2(2, parameters.Tile), color.RGBA{18, 12, 16, 120})
	}
	if surface {
		rl.DrawRectangleV(rl.NewVector2(px, py), rl.NewVector2(parameters.Tile, 7), emberCrust)
		rl.DrawRectangleV(rl.NewVector2(px, py), rl.NewVector2(parameters.Tile, 2), color.RGBA{132, 70, 54, 255})
		if hash%3 == 0 {
			glow := float32(.5 + .4*math.Sin(t*2+float64(hash%37)))
			start := px + float32(hash%20)
			rl.DrawLineEx(rl.NewVector2(start, py+7), rl.NewVector2(start+8, py+14), 2, rl.Fade(emberLava, glow))
			rl.DrawCircleGradient(rl.NewVector2(start+4, py+9), 12, rl.Fade(emberLava, .2*glow), rl.Blank)
		}
		return
	}
	if depth >= 2 && hash%5 == 0 {
		glow := float32(.45 + .35*math.Sin(t*1.6+float64(hash%41)))
		a := rl.NewVector2(px+6+float32(hash%12), py+8+float32((hash>>4)%10))
		b := rl.NewVector2(a.X+14, a.Y+10+float32((hash>>8)%8))
		c := rl.NewVector2(b.X+8, b.Y-6)
		rl.DrawLineEx(a, b, 2, rl.Fade(emberLava, glow))
		rl.DrawLineEx(b, c, 2, rl.Fade(emberLava, glow*.8))
		rl.DrawCircleGradient(b, 14, rl.Fade(emberLava, .22*glow), rl.Blank)
	}
}

func drawLedgeUnderside(x, y int, t float64) {
	px, py := float32(x*parameters.Tile), float32((y+1)*parameters.Tile)
	hash := uint32(x*2654435761) ^ uint32(y*40503)
	for i := range 3 {
		w := float32(8 + (hash>>uint(i*3))%7)
		cx := px + 6 + float32(i)*13 + float32((hash>>uint(i*5))%5)
		depth := float32(6 + (hash>>uint(i*4))%12)
		coloredTriangle(rl.NewVector2(cx-w/2, py-1), rl.NewVector2(cx+w/2, py-1), rl.NewVector2(cx, py+depth), color.RGBA{26, 16, 20, 255})
	}
	if hash%3 != 0 {
		return
	}
	drip := px + 10 + float32(hash%20)
	period := 1.6 + float64(hash%7)*.2
	age := math.Mod(t+float64(hash%13)*.37, period) / period
	rl.DrawCircleGradient(rl.NewVector2(drip, py+4), 10, rl.Fade(emberLava, .35), rl.Blank)
	if age < .45 {
		rl.DrawCircleV(rl.NewVector2(drip, py+3), float32(1+age*5), emberLava)
		return
	}
	fall := float32((age - .45) / .55)
	rl.DrawCircleV(rl.NewVector2(drip, py+4+fall*fall*220), 2.2, rl.Fade(emberHotLava, 1-fall))
}

func (m *Map) DrawEmberOverlay(ctx primitives.Context) {
	t := emberTime()
	w, h := float32(ctx.WindowWidth), float32(ctx.WindowHeight)
	for i := range 70 {
		speed := 30 + emberHash(i, 21)*60
		span := float64(h) + 60
		y := float32(float64(h) + 30 - math.Mod(emberHash(i, 22)*span+t*speed, span))
		x := float32(emberHash(i, 23))*w + float32(math.Sin(t*.8+float64(i)))*24
		flicker := float32(.5 + .5*math.Sin(t*6+float64(i)*1.3))
		size := float32(1 + emberHash(i, 24)*2)
		rl.DrawCircleV(rl.NewVector2(x, y), size, rl.Fade(emberHotLava, .35+.5*flicker))
	}
	rl.DrawRectangleGradientV(0, int32(h*.72), int32(w), int32(h*.28)+1, rl.Blank, rl.Fade(emberLava, .14))
	rl.DrawRectangleGradientH(0, 0, int32(w*.12), int32(h), rl.Fade(rl.Black, .45), rl.Blank)
	rl.DrawRectangleGradientH(int32(w*.88), 0, int32(w*.12)+1, int32(h), rl.Blank, rl.Fade(rl.Black, .45))
}
