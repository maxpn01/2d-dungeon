package game

import (
	"image/color"
	"math"

	"raylib-2d-sandbox/internal/terrain"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const radarWorldSpan float32 = 3000

type radarView struct {
	exploration *mapExploration
	bounds      rl.Rectangle
	focus       rl.Vector2
	scale       float32
}

func newRadarView(windowWidth int32, focus rl.Vector2) radarView {
	const size float32 = 220
	return radarView{
		bounds: rl.NewRectangle(float32(windowWidth)-size, 0, size, size),
		focus:  focus,
		scale:  size / radarWorldSpan,
	}
}

func (v radarView) point(world rl.Vector2) rl.Vector2 {
	return rl.NewVector2(
		v.bounds.X+v.bounds.Width/2+(world.X-v.focus.X)*v.scale,
		v.bounds.Y+v.bounds.Height/2+(world.Y-v.focus.Y)*v.scale,
	)
}

func (v radarView) rect(x, y, width, height float32, tint color.RGBA) {
	p := v.point(rl.NewVector2(x, y))
	rl.DrawRectangleRec(rl.NewRectangle(p.X, p.Y, width*v.scale, height*v.scale), tint)
}

func (v radarView) marker(center rl.Vector2, radius float32, tint color.RGBA) {
	if !v.exploration.visible[explorationCell(center)] {
		return
	}
	p := v.point(center)
	if p.X+radius < v.bounds.X || p.X-radius > v.bounds.X+v.bounds.Width ||
		p.Y+radius < v.bounds.Y || p.Y-radius > v.bounds.Y+v.bounds.Height {
		return
	}
	rl.DrawCircleV(p, radius+1, rl.Black)
	rl.DrawCircleV(p, radius, tint)
}

func (v *radarView) begin(m *Map) {
	v.exploration = &m.exploration
	rl.BeginScissorMode(int32(v.bounds.X), int32(v.bounds.Y), int32(v.bounds.Width), int32(v.bounds.Height))
	rl.DrawRectangleRec(v.bounds, color.RGBA{12, 16, 23, 255})

	v.rect(0, 0, m.size.X, m.size.Y, rl.Gray)
	field := color.RGBA{30, 39, 43, 255}
	v.rect(m.edgePosX.start, m.edgePosY.start, m.edgePosX.end-m.edgePosX.start, m.edgePosY.end-m.edgePosY.start, field)
	v.rect(float32(int(m.sizeInCells.X)/2-2)*m.cellSize.X, float32(int(m.sizeInCells.Y)/2-2)*m.cellSize.Y, 5*m.cellSize.X, 5*m.cellSize.Y, rl.Gray)
	v.rect(880, 0, 240, m.cellSize.Y, field)

	left := max(0, int(math.Floor(float64((v.focus.X-v.bounds.Width/(2*v.scale))/terrain.Tile))))
	right := min(50, int(math.Ceil(float64((v.focus.X+v.bounds.Width/(2*v.scale))/terrain.Tile))))
	top := max(terrain.North/terrain.Tile, int(math.Floor(float64((v.focus.Y-v.bounds.Height/(2*v.scale))/terrain.Tile))))
	bottom := min(0, int(math.Ceil(float64((v.focus.Y+v.bounds.Height/(2*v.scale))/terrain.Tile))))
	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			if terrain.Cell(x, y) != 0 {
				v.rect(float32(x*terrain.Tile), float32(y*terrain.Tile), terrain.Tile, terrain.Tile, paleStone)
			} else if terrain.Pool(x, y) {
				v.rect(float32(x*terrain.Tile), float32(y*terrain.Tile), terrain.Tile, terrain.Tile, color.RGBA{23, 58, 71, 255})
			}
		}
	}
}

func (v radarView) fog() {
	left := max(0, int(math.Floor(float64(v.focus.X-v.bounds.Width/(2*v.scale))/terrain.Tile)))
	right := min(50, int(math.Ceil(float64(v.focus.X+v.bounds.Width/(2*v.scale))/terrain.Tile)))
	top := max(terrain.North/terrain.Tile, int(math.Floor(float64(v.focus.Y-v.bounds.Height/(2*v.scale))/terrain.Tile)))
	bottom := min(50, int(math.Ceil(float64(v.focus.Y+v.bounds.Height/(2*v.scale))/terrain.Tile)))
	state := func(x, y int) int {
		cell := [2]int{x, y}
		if v.exploration.visible[cell] {
			return 0
		}
		if v.exploration.explored[cell] {
			return 1
		}
		return 2
	}
	type region struct{ left, top, right, bottom, state int }
	var regions []region
	previous := make(map[[3]int]int)
	for y := top; y < bottom; y++ {
		current := make(map[[3]int]int)
		for x := left; x < right; {
			start, shade := x, state(x, y)
			for x < right && state(x, y) == shade {
				x++
			}
			if shade == 0 {
				continue
			}
			key := [3]int{start, x, shade}
			if index, ok := previous[key]; ok {
				regions[index].bottom = y + 1
				current[key] = index
			} else {
				current[key] = len(regions)
				regions = append(regions, region{start, y, x, y + 1, shade})
			}
		}
		previous = current
	}
	for _, r := range regions {
		tint := color.RGBA{12, 16, 23, 255}
		if r.state == 1 {
			tint = color.RGBA{0, 0, 0, 100}
		}
		p := v.point(rl.NewVector2(float32(r.left*terrain.Tile), float32(r.top*terrain.Tile)))
		q := v.point(rl.NewVector2(float32(r.right*terrain.Tile), float32(r.bottom*terrain.Tile)))
		x0, y0 := float32(math.Round(float64(p.X))), float32(math.Round(float64(p.Y)))
		x1, y1 := float32(math.Round(float64(q.X))), float32(math.Round(float64(q.Y)))
		rl.DrawRectangleRec(rl.NewRectangle(x0, y0, x1-x0, y1-y0), tint)
	}
}

func (v radarView) end(player rl.Vector2) {
	circle := rl.NewVector2(terrain.DescentCircleX, terrain.DescentCircleY)
	if v.exploration.explored[explorationCell(circle)] {
		p := v.point(circle)
		rl.DrawCircleV(p, 9, rl.Black)
		rl.DrawCircleLinesV(p, 8, runeColor)
		rl.DrawCircleLinesV(p, 5, runeColor)
	}

	v.marker(player, 4, rl.Red)
	rl.EndScissorMode()
	rl.DrawRectangleLinesEx(v.bounds, 1, paleStone)
}

func entityCenter(pos, size rl.Vector2) rl.Vector2 {
	return rl.NewVector2(pos.X+size.X/2, pos.Y+size.Y/2)
}

func (w *World) drawMinimap(windowWidth int32) {
	v := newRadarView(windowWidth, entityCenter(w.player.pos, w.player.size))
	w.drawRadar(v)
}

func (w *World) drawRadar(v radarView) {
	v.begin(w.gameMap)
	v.fog()
	for _, fruit := range w.fruitSpawner.fruits {
		v.marker(entityCenter(fruit.pos, fruit.size), 2, rl.Yellow)
	}
	for _, ai := range w.ais {
		if !ai.isDeadFlag {
			v.marker(entityCenter(ai.pos, ai.size), 3, rl.Green)
		}
	}
	v.end(entityCenter(w.player.pos, w.player.size))
}

func (m *multiplayer) drawMinimap(windowWidth int32) {
	m.drawRadar(windowWidth, 0)
}

func (m *multiplayer) drawRadar(windowWidth, windowHeight int32) {
	if m.view.Local.Dead {
		return
	}
	for _, local := range m.rendered {
		if local.ID != m.view.PlayerID {
			continue
		}
		center := func(x, y, size float64) rl.Vector2 {
			return rl.NewVector2(float32(x+size/2), float32(y+size/2))
		}
		v := newRadarView(windowWidth, center(local.X, local.Y, local.Size))
		if windowHeight > 0 {
			v = newFullMapView(windowWidth, windowHeight, m.gameMap)
		}
		v.begin(m.gameMap)
		v.fog()
		for _, fruit := range m.view.Fruits {
			v.marker(center(fruit.X, fruit.Y, fruit.Size), 2, rl.Yellow)
		}
		for _, ai := range m.renderedAIs {
			v.marker(center(ai.X, ai.Y, ai.Size), 3, rl.Green)
		}
		for _, player := range m.rendered {
			if player.ID != local.ID && !player.Dead {
				v.marker(center(player.X, player.Y, player.Size), 3, rl.SkyBlue)
			}
		}
		v.end(center(local.X, local.Y, local.Size))
		return
	}
}
