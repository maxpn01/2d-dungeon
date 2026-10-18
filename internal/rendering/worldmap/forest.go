package worldmap

import (
	"math"

	"2d-dungeon/internal/parameters"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var forestGrass = rl.NewColor(77, 128, 48, 255)
var forestLeaf = rl.NewColor(104, 146, 66, 255)

func (m *Map) drawForest(left, top, right, bottom int) {
	const tile = parameters.Tile
	dungeonRect(left*tile, top*tile, (right-left)*tile, (bottom-top)*tile, rl.NewColor(12, 25, 19, 255))
	for y := max(parameters.MinCellY, top); y < min(parameters.MaxCellY, bottom); y++ {
		for x := max(parameters.MinCellX, left); x < min(parameters.MaxCellX, right); x++ {
			if m.world.CellFloor(x, y) != m.floor {
				continue
			}
			px, py := x*tile, y*tile
			hash := uint32(x*73856093) ^ uint32((y+100)*19349663)
			if m.world.Cell(x, y) == 0 {
				near := false
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						near = near || m.world.Cell(x+dx, y+dy) != 0
					}
				}
				if near {
					drawForestTree(px, py, hash)
				}
				continue
			}
			tone := uint8(hash % 5)
			dungeonRect(px, py, tile, tile, rl.NewColor(43+tone, 83+tone, 33+tone/2, 255))
			at := rl.NewVector2(float32(px+12+int(hash%17)), float32(py+12+int(hash>>5)%17))
			for i := 0; i < 7; i++ {
				seed := (hash ^ uint32(i+1)*2654435761) * 2246822519
				gx, gy := px+4+int(seed%32), py+10+int(seed>>5)%27
				height := 5 + int(seed>>10)%4
				rl.DrawLine(int32(gx), int32(gy), int32(gx-3), int32(gy-height+2), rl.Fade(forestGrass, .8))
				rl.DrawLine(int32(gx), int32(gy), int32(gx), int32(gy-height), rl.Fade(forestLeaf, .65))
				rl.DrawLine(int32(gx), int32(gy), int32(gx+3), int32(gy-height+1), rl.Fade(forestGrass, .9))
			}
			if hash%11 == 0 {
				rl.DrawEllipse(int32(at.X), int32(at.Y), 4, 2, rl.NewColor(109, 87, 47, 255))
				rl.DrawLine(int32(at.X-2), int32(at.Y), int32(at.X+2), int32(at.Y), rl.NewColor(156, 128, 64, 255))
			}
			if hash%17 == 0 {
				rl.DrawCircleV(at, 3, rl.NewColor(77, 88, 66, 255))
				rl.DrawCircleV(rl.NewVector2(at.X-1, at.Y-1), 1, rl.NewColor(125, 136, 98, 255))
			}
			if hash%23 == 0 {
				drawForestFlower(at, hash)
			}
			if hash%37 == 0 {
				phase := rl.GetTime()*1.2 + float64(hash%97)
				mote := rl.NewVector2(at.X+float32(math.Sin(phase*.7))*7, at.Y+float32(math.Cos(phase))*5)
				glow := float32(.4 + .3*math.Sin(phase*2))
				rl.DrawCircleGradient(mote, 12, rl.Fade(rl.NewColor(204, 224, 112, 255), glow*.2), rl.Blank)
				rl.DrawCircleV(mote, 1.2, rl.Fade(rl.NewColor(225, 241, 151, 255), glow))
			}
			if m.world.Cell(x, y-1) == 0 {
				dungeonRect(px, py, tile, 8, rl.NewColor(10, 24, 15, 100))
			}
		}
	}
	for _, patch := range m.world.Patches() {
		if m.world.CellFloor(patch.At.X, patch.At.Y) != m.floor || m.world.Cell(patch.At.X, patch.At.Y) == 0 {
			continue
		}
		if patch.At.X+patch.Radius < left || patch.At.X-patch.Radius > right || patch.At.Y+patch.Radius < top || patch.At.Y-patch.Radius > bottom {
			continue
		}
		center := rl.NewVector2(float32(patch.At.X*tile+tile/2), float32(patch.At.Y*tile+tile/2))
		radius := float32(patch.Radius * tile)
		rl.DrawCircleGradient(center, radius, rl.Fade(forestGrass, .3), rl.Blank)
		for i := 0; i < 18; i++ {
			angle := float64(i) * math.Pi / 9
			at := rl.NewVector2(center.X+float32(math.Cos(angle))*radius*.85, center.Y+float32(math.Sin(angle))*radius*.85)
			rl.DrawEllipse(int32(at.X), int32(at.Y), 7, 3, rl.Fade(forestLeaf, .65))
			if i%3 == 0 {
				drawForestFlower(at, uint32(i))
			}
		}
	}
	for _, portal := range m.world.Portals() {
		if portal.Floor != m.floor {
			continue
		}
		position := portal.Position()
		center := rl.NewVector2(float32(position.X), float32(position.Y))
		rl.DrawCircleV(center, 96, rl.NewColor(41, 52, 36, 255))
		rl.DrawRing(center, 92, 97, 0, 360, 64, rl.NewColor(98, 112, 75, 255))
		for i := 0; i < 12; i++ {
			angle := float64(i) * math.Pi / 6
			at := rl.NewVector2(center.X+float32(math.Cos(angle))*101, center.Y+float32(math.Sin(angle))*101)
			rl.DrawLineV(at, rl.NewVector2(at.X-4, at.Y-6), forestGrass)
			rl.DrawLineV(at, rl.NewVector2(at.X, at.Y-9), forestLeaf)
			rl.DrawLineV(at, rl.NewVector2(at.X+4, at.Y-7), forestGrass)
		}
	}
}

func drawForestTree(x, y int, hash uint32) {
	dungeonRect(x, y, parameters.Tile, parameters.Tile, rl.NewColor(17, 32, 21, 255))
	dungeonRect(x+15, y+9, 11, 29, rl.NewColor(72, 56, 35, 255))
	dungeonRect(x+18, y+12, 3, 24, rl.NewColor(109, 83, 48, 255))
	rl.DrawLineEx(rl.NewVector2(float32(x+20), float32(y+29)), rl.NewVector2(float32(x+7), float32(y+37)), 3, rl.NewColor(72, 56, 35, 255))
	rl.DrawLineEx(rl.NewVector2(float32(x+22), float32(y+29)), rl.NewVector2(float32(x+33), float32(y+35)), 3, rl.NewColor(72, 56, 35, 255))
	tone := uint8(hash % 12)
	rl.DrawEllipse(int32(x+20), int32(y+18), 19, 16, rl.NewColor(25+tone, 54+tone, 30, 255))
	rl.DrawEllipse(int32(x+13), int32(y+12), 11, 9, rl.NewColor(44+tone, 78+tone, 37, 255))
	rl.DrawEllipse(int32(x+28), int32(y+16), 10, 10, rl.NewColor(35+tone, 68+tone, 32, 255))
	rl.DrawEllipse(int32(x+13), int32(y+9), 7, 3, rl.Fade(forestLeaf, .45))
}

func drawForestFlower(at rl.Vector2, hash uint32) {
	tint := rl.NewColor(194, 181, 112, 255)
	if hash%2 == 0 {
		tint = rl.NewColor(156, 156, 196, 255)
	}
	for _, offset := range []rl.Vector2{{X: -2}, {X: 2}, {Y: -2}, {Y: 2}} {
		rl.DrawCircleV(rl.Vector2Add(at, offset), 1.6, tint)
	}
	rl.DrawCircleV(at, 1, rl.NewColor(231, 212, 121, 255))
}
