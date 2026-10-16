package worldmap

import (
	"image/color"
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"
	mapdata "raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const radarWorldSpan float32 = parameters.RadarWorldSpan

type radarView struct {
	Exploration *Exploration
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
	if !v.Exploration.ActorVisible(float64(center.X), float64(center.Y), 0) {
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
	v.Exploration = &m.Exploration
	rl.BeginScissorMode(int32(v.bounds.X), int32(v.bounds.Y), int32(v.bounds.Width), int32(v.bounds.Height))
	rl.DrawRectangleRec(v.bounds, color.RGBA{12, 16, 23, 255})

	v.rect(0, 0, m.size.X, m.size.Y, rl.Gray)
	field := color.RGBA{30, 39, 43, 255}
	v.rect(m.edgePosX.start, m.edgePosY.start, m.edgePosX.end-m.edgePosX.start, m.edgePosY.end-m.edgePosY.start, field)
	v.rect(float32(int(m.sizeInCells.X)/2-2)*m.cellSize.X, float32(int(m.sizeInCells.Y)/2-2)*m.cellSize.Y, 5*m.cellSize.X, 5*m.cellSize.Y, rl.Gray)
	v.rect(m.size.X/2-120, 0, 240, m.cellSize.Y, field)

	left := max(mapdata.MinCellX, int(math.Floor(float64((v.focus.X-v.bounds.Width/(2*v.scale))/mapdata.Tile))))
	right := min(mapdata.MaxCellX, int(math.Ceil(float64((v.focus.X+v.bounds.Width/(2*v.scale))/mapdata.Tile))))
	top := max(mapdata.North/mapdata.Tile, int(math.Floor(float64((v.focus.Y-v.bounds.Height/(2*v.scale))/mapdata.Tile))))
	bottom := min(mapdata.MaxCellY, int(math.Ceil(float64((v.focus.Y+v.bounds.Height/(2*v.scale))/mapdata.Tile))))
	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			if m.world.CellFloor(x, y) != m.floor || y >= 0 && y < mapdata.SurfaceCells && x >= 0 && x < mapdata.SurfaceCells {
				continue
			}
			if m.world.Cover(x, y) && m.world.Cell(x, y) != 0 {
				v.rect(float32(x*mapdata.Tile), float32(y*mapdata.Tile), mapdata.Tile, mapdata.Tile, color.RGBA{102, 91, 150, 255})
			} else if m.world.Cell(x, y) != 0 {
				tint := paleStone
				if m.floor == mapdata.SecondFloor {
					tint = forestGrass
				}
				v.rect(float32(x*mapdata.Tile), float32(y*mapdata.Tile), mapdata.Tile, mapdata.Tile, tint)
			} else if mapdata.Default().Pool(x, y) {
				v.rect(float32(x*mapdata.Tile), float32(y*mapdata.Tile), mapdata.Tile, mapdata.Tile, color.RGBA{23, 58, 71, 255})
			}
		}
	}
}

func (v radarView) fog() {
	left := max(mapdata.MinCellX, int(math.Floor(float64(v.focus.X-v.bounds.Width/(2*v.scale))/mapdata.Tile)))
	right := min(mapdata.MaxCellX, int(math.Ceil(float64(v.focus.X+v.bounds.Width/(2*v.scale))/mapdata.Tile)))
	top := max(mapdata.North/mapdata.Tile, int(math.Floor(float64(v.focus.Y-v.bounds.Height/(2*v.scale))/mapdata.Tile)))
	bottom := min(mapdata.MaxCellY, int(math.Ceil(float64(v.focus.Y+v.bounds.Height/(2*v.scale))/mapdata.Tile)))
	state := func(x, y int) int {
		cell := [2]int{x, y}
		if v.Exploration.Visible(cell) {
			return 0
		}
		if v.Exploration.Explored(cell) {
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
		p := v.point(rl.NewVector2(float32(r.left*mapdata.Tile), float32(r.top*mapdata.Tile)))
		q := v.point(rl.NewVector2(float32(r.right*mapdata.Tile), float32(r.bottom*mapdata.Tile)))
		x0, y0 := float32(math.Round(float64(p.X))), float32(math.Round(float64(p.Y)))
		x1, y1 := float32(math.Round(float64(q.X))), float32(math.Round(float64(q.Y)))
		rl.DrawRectangleRec(rl.NewRectangle(x0, y0, x1-x0, y1-y0), tint)
	}
}

func (v radarView) end(player rl.Vector2) {
	for _, portal := range mapdata.Default().Portals() {
		position := portal.Position()
		circle := rl.NewVector2(float32(position.X), float32(position.Y))
		if v.Exploration.Explored(explorationCell(circle)) {
			p := v.point(circle)
			rl.DrawCircleV(p, 9, rl.Black)
			rl.DrawCircleLinesV(p, 8, primitives.RuneColor)
			rl.DrawCircleLinesV(p, 5, primitives.RuneColor)
		}
	}

	v.marker(player, 4, rl.Red)
	rl.EndScissorMode()
	rl.DrawRectangleLinesEx(v.bounds, 1, paleStone)
}

func (v radarView) statueMarkers(statues []simulation.StatueState) {
	for _, s := range statues {
		center := rl.NewVector2(float32(s.X), float32(s.Y))
		if !v.Exploration.Explored(explorationCell(center)) {
			continue
		}
		point := v.point(center)
		tint := primitives.RuneColor
		if s.ClaimedBy != "" {
			tint = rl.Gray
		}
		rl.DrawPoly(point, 4, 6, 45, rl.Black)
		rl.DrawPoly(point, 4, 4, 45, tint)
	}
}

type Marker struct {
	Center rl.Vector2
	Radius float32
	Tint   color.RGBA
}

func (m *Map) DrawRadar(width, height int32, focus rl.Vector2, markers []Marker, snapshot simulation.Snapshot) {
	m.world = mapdata.Default().WithSecrets(snapshot.OpenSecrets)
	m.floor = m.world.FloorAt(float64(focus.X), float64(focus.Y))
	v := newRadarView(width, focus)
	if height > 0 {
		v = newFullMapView(width, height, m)
	}
	v.begin(m)
	v.fog()
	for _, marker := range markers {
		v.marker(marker.Center, marker.Radius, marker.Tint)
	}
	v.statueMarkers(snapshot.Statues)
	if m.floor == mapdata.FirstFloor {
		gate, ok := m.world.Gate(mapdata.VaultGate)
		center := rl.NewVector2(float32(gate.Left+gate.Right)/2, float32(gate.Y))
		if ok && v.Exploration.Explored(explorationCell(center)) {
			point := v.point(center)
			tint := astralGold
			if snapshot.Orb.GateOpen >= .95 {
				tint = astralViolet
			}
			rl.DrawRectangleRec(rl.NewRectangle(point.X-7, point.Y-5, 14, 10), rl.Black)
			rl.DrawRectangleLinesEx(rl.NewRectangle(point.X-6, point.Y-4, 12, 8), 2, tint)
		}
	}
	if !snapshot.Orb.Consumed && snapshot.Orb.Carrier == "" {
		center := rl.NewVector2(float32(snapshot.Orb.X), float32(snapshot.Orb.Y))
		if v.Exploration.Explored(explorationCell(center)) {
			point := v.point(center)
			tint := rl.Orange
			if snapshot.Orb.Unlocked {
				tint = primitives.RuneColor
			}
			rl.DrawCircleV(point, 6, rl.Black)
			rl.DrawCircleV(point, 4, tint)
		}
	} else {
		for _, player := range snapshot.Players {
			if player.ID == snapshot.Orb.Carrier && m.Exploration.PlayerVisible(player.X, player.Y, simulation.PlayerSize) {
				v.marker(rl.NewVector2(float32(player.X+simulation.PlayerSize/2), float32(player.Y+simulation.PlayerSize/2)), 5, primitives.RuneColor)
			}
		}
	}
	v.end(focus)
}
