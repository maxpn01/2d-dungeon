package worldmap

import (
	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/physics"
)

type Cell struct{ X, Y int }
type CellRect struct{ X, Y, Width, Height int }

func At(x, y int) Cell { return Cell{X: x, Y: y} }
func Cells(x, y, width, height int) CellRect {
	return CellRect{X: x, Y: y, Width: width, Height: height}
}
func Around(center Cell, radius int) CellRect {
	return Cells(center.X-radius, center.Y-radius, radius*2+1, radius*2+1)
}
func (c Cell) Add(other Cell) Cell { return At(c.X+other.X, c.Y+other.Y) }
func (c Cell) World() physics.Vector {
	return physics.Vector{X: float64(c.X * parameters.Tile), Y: float64(c.Y * parameters.Tile)}
}
func (c Cell) Center() physics.Vector {
	return c.World().Add(physics.Vector{X: parameters.Tile / 2, Y: parameters.Tile / 2})
}
func (c Cell) key() [2]int      { return [2]int{c.X, c.Y} }
func (r CellRect) Origin() Cell { return At(r.X, r.Y) }
func (r CellRect) Right() int   { return r.X + r.Width }
func (r CellRect) Bottom() int  { return r.Y + r.Height }
func (r CellRect) Contains(cell Cell) bool {
	return cell.X >= r.X && cell.X < r.Right() && cell.Y >= r.Y && cell.Y < r.Bottom()
}
func (r CellRect) ContainsArea(area CellRect) bool {
	return area.Width > 0 && area.Height > 0 && area.X >= r.X && area.Y >= r.Y && area.X < r.Right() && area.Y < r.Bottom() && area.Width <= r.Right()-area.X && area.Height <= r.Bottom()-area.Y
}
func (r CellRect) Shift(offset Cell) CellRect {
	return Cells(r.X+offset.X, r.Y+offset.Y, r.Width, r.Height)
}
func (r CellRect) Bounds() physics.Rect {
	return physics.Rect{X: float64(r.X * parameters.Tile), Y: float64(r.Y * parameters.Tile), W: float64(r.Width * parameters.Tile), H: float64(r.Height * parameters.Tile)}
}
func (r Room) Bounds() physics.Rect              { return r.Area.Bounds() }
func (r Room) WorldCell(local Cell) Cell         { return r.Area.Origin().Add(local) }
func (r Room) WorldArea(local CellRect) CellRect { return local.Shift(r.Area.Origin()) }

func (r CellRect) ContainsWorld(x, y float64) bool {
	bounds := r.Bounds()
	return x >= bounds.X && x < bounds.X+bounds.W && y >= bounds.Y && y < bounds.Y+bounds.H
}
