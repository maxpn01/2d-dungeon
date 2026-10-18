package worldmap

import (
	"image/color"
	"math"

	"2d-dungeon/internal/parameters"
	mapdata "2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	grassTint    = color.RGBA{54, 94, 66, 255}
	mossTint     = color.RGBA{30, 58, 48, 255}
	dirtTint     = color.RGBA{120, 92, 64, 255}
	caveTint     = color.RGBA{78, 72, 86, 255}
	ruinFloor    = color.RGBA{86, 90, 84, 255}
	mudTint      = color.RGBA{50, 46, 38, 254}
	rockSide     = color.RGBA{82, 78, 96, 254}
	rockTop      = color.RGBA{104, 102, 116, 254}
	rockMoss     = color.RGBA{48, 76, 60, 254}
	barkTint     = color.RGBA{78, 56, 42, 255}
	pineDark     = color.RGBA{32, 96, 84, 255}
	pineLight    = color.RGBA{62, 136, 100, 255}
	leafTint     = color.RGBA{70, 132, 80, 255}
	leafLight    = color.RGBA{108, 162, 90, 255}
	bushTint     = color.RGBA{58, 124, 72, 255}
	stoneTint    = color.RGBA{116, 118, 112, 254}
	stoneTop     = color.RGBA{132, 134, 126, 254}
	plankTint    = color.RGBA{118, 86, 56, 254}
	runeGlow     = color.RGBA{140, 255, 206, 255}
	fruitGlow    = color.RGBA{255, 150, 84, 255}
	mushroomTeal = color.RGBA{90, 232, 214, 255}
	mushroomRose = color.RGBA{214, 136, 255, 255}
	crystalTint  = color.RGBA{176, 128, 255, 255}
	portalGlow   = color.RGBA{120, 220, 255, 255}
)

func cellHash(x, y int) uint32 {
	h := uint32(x*73856093) ^ uint32(y*19349663) ^ 0x9e3779b9
	h ^= h >> 15
	h *= 2246822519
	h ^= h >> 13
	h *= 3266489917
	h ^= h >> 16
	return h
}

func unit(h uint32, k uint32) float32 {
	return float32(cellHash(int(h), int(k))%10000) / 10000
}

type forestGrid struct {
	world *mapdata.Map
	area  mapdata.CellRect
}

func (g forestGrid) at(x, y int) mapdata.Terrain {
	if !g.area.Contains(mapdata.At(x, y)) {
		return mapdata.Tree
	}
	if g.world.Cell(x, y) == 0 {
		return mapdata.Rock
	}
	return g.world.Terrain(x, y)
}

func treeTerrain(t mapdata.Terrain) bool { return t == mapdata.Tree || t == mapdata.FruitTree }

func (g forestGrid) bridge(x, y int) bool {
	if g.at(x, y) != mapdata.DirtPath {
		return false
	}
	water := func(dx, dy int) bool {
		for i := 1; i <= 4; i++ {
			switch g.at(x+dx*i, y+dy*i) {
			case mapdata.Water:
				return true
			case mapdata.DirtPath:
				continue
			}
			return false
		}
		return false
	}
	return water(-1, 0) && water(1, 0) || water(0, -1) && water(0, 1)
}

func (g forestGrid) wet(x, y int) bool { return g.at(x, y) == mapdata.Water || g.bridge(x, y) }

func (g forestGrid) groundTint(x, y int) color.RGBA {
	h := cellHash(x, y)
	jitter := .92 + float32(h%16)/100
	switch t := g.at(x, y); {
	case treeTerrain(t):
		return shade(mossTint, jitter)
	case t == mapdata.DirtPath:
		if g.bridge(x, y) {
			return mudTint
		}
		return shade(dirtTint, jitter)
	case t == mapdata.CaveFloor:
		return shade(caveTint, jitter)
	case t == mapdata.Ruin:
		return ruinFloor
	case t == mapdata.Water:
		return mudTint
	case t == mapdata.Rock:
		return rockSide
	default:
		return shade(grassTint, jitter)
	}
}

func (g forestGrid) cornerTint(x, y int) color.RGBA {
	var r, gg, b float32
	for _, c := range [][2]int{{x - 1, y - 1}, {x, y - 1}, {x - 1, y}, {x, y}} {
		t := g.groundTint(c[0], c[1])
		r, gg, b = r+float32(t.R), gg+float32(t.G), b+float32(t.B)
	}
	return color.RGBA{uint8(r / 4), uint8(gg / 4), uint8(b / 4), 255}
}

func (g forestGrid) rockCorner(x, y int) float32 {
	h := cellHash(x*3+1, y*5+2)
	open := 0
	for _, c := range [][2]int{{x - 1, y - 1}, {x, y - 1}, {x - 1, y}, {x, y}} {
		if g.at(c[0], c[1]) != mapdata.Rock {
			open++
		}
	}
	if open > 0 {
		return 46 + float32(h%34) - float32(open)*6
	}
	depth := float32(0)
	for r := 1; r <= 3; r++ {
		if g.at(x-r, y) == mapdata.Rock && g.at(x+r-1, y) == mapdata.Rock && g.at(x, y-r) == mapdata.Rock && g.at(x, y+r-1) == mapdata.Rock {
			depth++
		}
	}
	return 70 + depth*22 + float32(h%40)
}

func (f *forestRenderer) build(world *mapdata.Map) {
	f.built = true
	f.shader = rl.LoadShaderFromMemory(forestVertex, forestFragment)
	f.locs = map[string]int32{}
	for _, name := range []string{"mode", "clock", "sunDir", "sunColor", "ambient", "lantern", "cameraPos", "playerScreen", "playerDepthZ", "hazeColor", "fogArea", "texture1"} {
		f.locs[name] = rl.GetShaderLocation(f.shader, name)
	}
	f.material = rl.LoadMaterialDefault()
	f.material.Shader = f.shader
	f.actorMaterial = rl.LoadMaterialDefault()
	f.actorMaterial.Shader = f.shader
	var cube meshBuilder
	cube.box(v3(-.5, 0, -.5), v3(.5, 1, .5), color.RGBA{255, 255, 255, 254}, color.RGBA{255, 255, 255, 254})
	f.cube = cube.upload()
	img := rl.GenImageGradientRadial(64, 64, 0, rl.White, rl.Blank)
	f.glowTex = rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)
	rl.SetTextureFilter(f.glowTex, rl.FilterBilinear)
	room, _ := world.Room(mapdata.Forest)
	grid := forestGrid{world: world, area: room.Area}
	left, top := room.Area.X-forestBorder, room.Area.Y-forestBorder
	right, bottom := room.Area.Right()+forestBorder, room.Area.Bottom()+forestBorder
	for cy := top; cy < bottom; cy += forestChunk {
		for cx := left; cx < right; cx += forestChunk {
			var ground, water, solid, shadow meshBuilder
			for y := cy; y < min(cy+forestChunk, bottom); y++ {
				for x := cx; x < min(cx+forestChunk, right); x++ {
					f.buildCell(grid, x, y, &ground, &water, &solid, &shadow)
				}
			}
			chunk := forestChunkMeshes{
				min: rl.NewVector2(float32(cx*parameters.Tile), float32(cy*parameters.Tile)),
				max: rl.NewVector2(float32((cx+forestChunk)*parameters.Tile), float32((cy+forestChunk)*parameters.Tile)),
			}
			if chunk.hasGround = !ground.empty(); chunk.hasGround {
				chunk.ground = ground.upload()
			}
			if chunk.hasWater = !water.empty(); chunk.hasWater {
				chunk.water = water.upload()
			}
			if chunk.hasSolid = !solid.empty(); chunk.hasSolid {
				chunk.solid = solid.upload()
			}
			if chunk.hasShadow = !shadow.empty(); chunk.hasShadow {
				chunk.shadow = shadow.upload()
			}
			f.chunks = append(f.chunks, chunk)
		}
	}
	for y := room.Area.Y; y < room.Area.Bottom(); y++ {
		for x := room.Area.X; x < room.Area.Right(); x++ {
			if grid.at(x, y) != mapdata.Grass || cellHash(x, y)%61 != 0 {
				continue
			}
			open := true
			for dy := -4; dy <= 4 && open; dy++ {
				for dx := -4; dx <= 4; dx++ {
					if t := grid.at(x+dx, y+dy); treeTerrain(t) || t == mapdata.Rock || t == mapdata.CaveFloor {
						open = false
						break
					}
				}
			}
			if open {
				f.shafts = append(f.shafts, v3(float32(x*parameters.Tile+parameters.Tile/2), 0, float32(y*parameters.Tile+parameters.Tile/2)))
			}
		}
	}
	var props meshBuilder
	for _, prop := range room.Props {
		f.buildProp(room.WorldArea(prop.Area), prop.Kind, &props)
	}
	for _, portal := range world.Portals() {
		if portal.Floor == mapdata.SecondFloor {
			position := portal.Position()
			f.buildPortalRing(rl.NewVector2(float32(position.X), float32(position.Y)), &props)
		}
	}
	if !props.empty() {
		bounds := room.Area.Bounds()
		f.chunks = append(f.chunks, forestChunkMeshes{min: rl.NewVector2(float32(bounds.X), float32(bounds.Y)), max: rl.NewVector2(float32(bounds.X+bounds.W), float32(bounds.Y+bounds.H)), solid: props.upload(), hasSolid: true})
	}
}

func (f *forestRenderer) buildCell(g forestGrid, x, y int, ground, water, solid, shadow *meshBuilder) {
	const tile = float32(parameters.Tile)
	t := g.at(x, y)
	h := cellHash(x, y)
	x0, z0 := float32(x)*tile, float32(y)*tile
	x1, z1 := x0+tile, z0+tile
	cx, cz := x0+tile/2, z0+tile/2
	if g.wet(x, y) {
		water.smoothQuad([4]rl.Vector3{v3(x0, -8, z0), v3(x0, -8, z1), v3(x1, -8, z1), v3(x1, -8, z0)}, [4]color.RGBA{rl.White, rl.White, rl.White, rl.White})
		for _, s := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			if g.wet(x+s[0], y+s[1]) || g.at(x+s[0], y+s[1]) == mapdata.Rock {
				continue
			}
			ex0, ez0, ex1, ez1 := x0, z0, x1, z0
			switch s {
			case [2]int{0, 1}:
				ez0, ez1 = z1, z1
			case [2]int{-1, 0}:
				ex1, ez1 = x0, z1
			case [2]int{1, 0}:
				ex0, ex1, ez1 = x1, x1, z1
			}
			solid.quad(v3(ex0, 0, ez0), v3(ex1, 0, ez1), v3(ex1, -12, ez1), v3(ex0, -12, ez0), v3(-float32(s[0]), 0, -float32(s[1])), mudTint)
		}
		if t == mapdata.Water && h%9 == 0 {
			pad := v3(cx+unit(h, 1)*16-8, -7, cz+unit(h, 2)*16-8)
			solid.disc(pad, 9, 8, 7, color.RGBA{58, 120, 70, 254})
			if h%3 == 0 {
				solid.blob(v3(pad.X+2, -4, pad.Z-1), 2.6, 2.2, glowing(color.RGBA{255, 176, 220, 255}, 1), h)
				f.glows = append(f.glows, forestGlow{at: v3(pad.X, -6, pad.Z), size: 60, flat: true, tint: color.RGBA{255, 160, 220, 140}, phase: unit(h, 3) * 6, pulses: true})
			}
		}
		if g.bridge(x, y) {
			for i := range 4 {
				pz := z0 + float32(i)*10 + 1
				solid.box(v3(x0-1, -2, pz), v3(x1+1, 3, pz+8), shade(plankTint, .9+float32(i%2)*.12), shade(plankTint, 1.1))
			}
			if !g.bridge(x-1, y) {
				solid.box(v3(x0+1, -6, z0+2), v3(x0+6, 16, z0+7), barkTint, barkTint)
				solid.box(v3(x0+1, -6, z1-7), v3(x0+6, 16, z1-2), barkTint, barkTint)
			}
		}
		return
	}
	if t == mapdata.Rock {
		f.buildRock(g, x, y, solid)
		return
	}
	corner := func(dx, dy int) color.RGBA { return g.cornerTint(x+dx, y+dy) }
	ground.smoothQuad([4]rl.Vector3{v3(x0, 0, z0), v3(x0, 0, z1), v3(x1, 0, z1), v3(x1, 0, z0)}, [4]color.RGBA{corner(0, 0), corner(0, 1), corner(1, 1), corner(1, 0)})
	switch {
	case treeTerrain(t):
		f.buildTree(g, x, y, t == mapdata.FruitTree, solid, shadow)
	case t == mapdata.Bush:
		f.buildBush(cx, cz, h, solid)
	case t == mapdata.Ruin:
		height := 22 + unit(h, 1)*44
		solid.box(v3(x0+2, 0, z0+2), v3(x1-2, height, z1-2), shade(stoneTint, .9+unit(h, 2)*.2), stoneTop)
		if h%2 == 0 {
			solid.box(v3(x0+6, height, z0+5), v3(x1-9, height+8+unit(h, 3)*10, z1-6), stoneTint, rockMoss)
		}
		if h%3 == 0 {
			solid.box(v3(x1, 0, z1-4), v3(x1+9, 7, z1+5), stoneTint, stoneTop)
		}
	case t == mapdata.Menhir:
		height := 78 + unit(h, 1)*22
		solid.prism(v3(cx, 0, cz), 21, 12, height, 4, math.Pi/4, color.RGBA{112, 118, 134, 254}, color.RGBA{132, 138, 150, 254}, h)
		solid.blob(v3(cx-5, height-6, cz-2), 9, 4, rockMoss, h)
		for i := range 3 {
			at := float32(24 + i*16)
			radius := (21 + (12-21)*at/height) * float32(math.Sqrt2) / 2
			solid.box(v3(cx-2.5, at, cz+radius-1), v3(cx+2.5, at+9, cz+radius+1.5), glowing(runeGlow, 1), glowing(runeGlow, 1))
		}
		f.glows = append(f.glows, forestGlow{at: v3(cx, 50, cz+10), size: 110, tint: color.RGBA{140, 255, 206, 150}, phase: unit(h, 2) * 6, pulses: true})
	case t == mapdata.CaveFloor:
		switch {
		case h%17 == 0:
			solid.blob(v3(cx+(unit(h, 1)-.5)*20, 3, cz+(unit(h, 2)-.5)*20), 6+unit(h, 3)*5, 5, shade(rockSide, 1.2), h)
		case h%23 == 0:
			solid.disc(v3(cx, .5, cz), 14, 10, 8, color.RGBA{56, 92, 70, 254})
		case h%29 == 0:
			at := v3(cx+(unit(h, 1)-.5)*18, 0, cz+(unit(h, 2)-.5)*18)
			solid.prism(at, 1.8, 1.4, 6, 5, 0, color.RGBA{214, 206, 190, 254}, color.RGBA{214, 206, 190, 254}, h)
			solid.prism(v3(at.X, 6, at.Z), 5, 0, 4, 7, 0, glowing(crystalTint, .9), glowing(crystalTint, .9), h)
			f.glows = append(f.glows, forestGlow{at: v3(at.X, 1, at.Z), size: 90, tint: color.RGBA{176, 128, 255, 120}, phase: unit(h, 4) * 6, pulses: true, flat: true})
		}
	case t == mapdata.DirtPath:
		if h%5 == 0 {
			px, pz := x0+6+unit(h, 1)*26, z0+6+unit(h, 2)*26
			solid.box(v3(px, 0, pz), v3(px+4, 2.5, pz+3), color.RGBA{136, 128, 116, 254}, color.RGBA{156, 148, 134, 254})
		}
	default:
		f.buildGrass(g, x, y, solid)
	}
}

func (f *forestRenderer) buildRock(g forestGrid, x, y int, solid *meshBuilder) {
	const tile = float32(parameters.Tile)
	h := cellHash(x, y)
	x0, z0 := float32(x)*tile, float32(y)*tile
	x1, z1 := x0+tile, z0+tile
	a := v3(x0, g.rockCorner(x, y), z0)
	b := v3(x1, g.rockCorner(x+1, y), z0)
	c := v3(x1, g.rockCorner(x+1, y+1), z1)
	d := v3(x0, g.rockCorner(x, y+1), z1)
	tone := func(p1, p2, p3 rl.Vector3) color.RGBA {
		n := rl.Vector3Normalize(rl.Vector3CrossProduct(rl.Vector3Subtract(p2, p1), rl.Vector3Subtract(p3, p1)))
		flat := float32(math.Abs(float64(n.Y)))
		base := mixColor(rockSide, rockTop, flat)
		if flat > .93 && (p1.Y+p2.Y+p3.Y)/3 > 60 && h%2 == 0 {
			base = rockMoss
		}
		return shade(base, .9+unit(h, 1)*.18)
	}
	if h%2 == 0 {
		solid.tri(a, b, c, v3(0, 1, 0), tone(a, b, c))
		solid.tri(a, c, d, v3(0, 1, 0), tone(a, c, d))
	} else {
		solid.tri(a, b, d, v3(0, 1, 0), tone(a, b, d))
		solid.tri(b, c, d, v3(0, 1, 0), tone(b, c, d))
	}
	for _, s := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		neighbor := g.at(x+s[0], y+s[1])
		if neighbor == mapdata.Rock {
			continue
		}
		var p, q rl.Vector3
		switch s {
		case [2]int{0, -1}:
			p, q = a, b
		case [2]int{0, 1}:
			p, q = d, c
		case [2]int{-1, 0}:
			p, q = a, d
		default:
			p, q = b, c
		}
		mid := rl.Vector3Lerp(p, q, .5)
		bulge := v3(float32(s[0])*(3+unit(h, 2)*5), 0, float32(s[1])*(3+unit(h, 2)*5))
		middle := rl.Vector3Add(v3(mid.X, mid.Y*.55, mid.Z), bulge)
		outward := v3(float32(s[0]), 0, float32(s[1]))
		low := func(v rl.Vector3) rl.Vector3 { return v3(v.X, -12, v.Z) }
		side := shade(rockSide, .82+unit(h, 3)*.2)
		solid.tri(low(p), p, middle, outward, side)
		solid.tri(p, q, middle, outward, shade(side, 1.08))
		solid.tri(q, low(q), middle, outward, shade(side, .94))
		solid.tri(low(q), low(p), middle, outward, shade(side, .78))
		if neighbor == mapdata.CaveFloor && h%4 == 0 {
			tint := crystalTint
			if h%8 == 0 {
				tint = color.RGBA{110, 220, 255, 255}
			}
			base := rl.Vector3Add(v3(mid.X, 0, mid.Z), rl.Vector3Scale(outward, 9))
			solid.prism(base, 6, 0, 26+unit(h, 6)*22, 5, unit(h, 7)*3, glowing(tint, 1), glowing(tint, 1), h)
			solid.prism(rl.Vector3Add(base, v3(float32(s[1])*9, 0, float32(s[0])*9)), 4, 0, 16, 5, 1, glowing(tint, .9), glowing(tint, .9), h>>3)
			f.glows = append(f.glows, forestGlow{at: v3(base.X, 1, base.Z), size: 170, tint: color.RGBA{tint.R, tint.G, tint.B, 150}, phase: unit(h, 8) * 6, pulses: true, flat: true})
		}
	}
}

func (f *forestRenderer) buildTree(g forestGrid, x, y int, fruit bool, solid, shadow *meshBuilder) {
	const tile = float32(parameters.Tile)
	h := cellHash(x, y)
	edge := false
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if t := g.at(x+dx, y+dy); !treeTerrain(t) && t != mapdata.Rock {
				edge = true
			}
		}
	}
	if !edge && !fruit && h%5 < 2 {
		return
	}
	cx := float32(x)*tile + tile/2 + (unit(h, 1)-.5)*14
	cz := float32(y)*tile + tile/2 + (unit(h, 2)-.5)*14
	scale := 1.25 + unit(h, 3)*.45
	if !edge {
		scale += .3
	}
	kind := h % 3
	if fruit {
		kind = 1
	}
	var height, radius float32
	switch kind {
	case 1:
		radius, height = 23*scale, 78*scale
		solid.prism(v3(cx, 0, cz), 6*scale, 4.5*scale, 40*scale, 6, unit(h, 4)*3, barkTint, barkTint, h)
		canopy := mixColor(leafTint, leafLight, unit(h, 5)*.6)
		switch h % 41 {
		case 3:
			canopy = color.RGBA{214, 132, 176, 255}
		case 7:
			canopy = color.RGBA{96, 148, 196, 255}
		case 11:
			canopy = color.RGBA{204, 170, 84, 255}
		}
		solid.blob(v3(cx, 54*scale, cz), radius, 21*scale, canopy, h)
		solid.blob(v3(cx+(unit(h, 6)-.5)*18, 70*scale, cz+(unit(h, 7)-.5)*14), radius*.66, 15*scale, shade(canopy, 1.1), h>>4)
		if fruit {
			for i := range 6 {
				angle := float64(i)*math.Pi/3 + float64(unit(h, 8))
				at := v3(cx+float32(math.Cos(angle))*radius*.9, 50*scale+float32(i%3)*6, cz+float32(math.Sin(angle))*radius*.75)
				solid.blob(at, 3.6, 3.6, glowing(fruitGlow, 1), h+uint32(i))
			}
			f.glows = append(f.glows, forestGlow{at: v3(cx, 56*scale, cz+radius*.6), size: 130, tint: color.RGBA{255, 150, 84, 120}, phase: unit(h, 9) * 6, pulses: true})
		}
	default:
		layers := 3
		if kind == 2 {
			layers, scale = 4, scale*1.08
		}
		radius = 20 * scale
		solid.prism(v3(cx, 0, cz), 5*scale, 4*scale, 24*scale, 6, unit(h, 4)*3, barkTint, barkTint, h)
		base := 14 * scale
		for i := range layers {
			r := radius * (1 - float32(i)*.2)
			height := 36 * scale * (1 - float32(i)*.08)
			tint := mixColor(pineDark, pineLight, float32(i)/float32(layers)+unit(h, uint32(10+i))*.2)
			solid.prism(v3(cx, base, cz), r, 0, height, 7, unit(h, uint32(20+i))*3, tint, tint, h>>uint(i))
			base += height * .52
		}
		height = base + 16*scale
	}
	if edge {
		offset := rl.NewVector2(forestSunDir.X, forestSunDir.Z)
		offset = rl.Vector2Scale(rl.Vector2Normalize(offset), height*.38)
		shadow.disc(v3(cx+offset.X, .3, cz+offset.Y), radius*1.15, radius*.85, 9, color.RGBA{0, 0, 0, 105})
	}
}

func (f *forestRenderer) buildBush(cx, cz float32, h uint32, solid *meshBuilder) {
	radius := 13 + unit(h, 1)*5
	solid.blob(v3(cx-4, 9, cz), radius, 11, shade(bushTint, .95+unit(h, 2)*.15), h)
	solid.blob(v3(cx+7, 8, cz+3), radius*.7, 9, shade(bushTint, 1.1), h>>3)
	if h%3 == 0 {
		tint := color.RGBA{150, 200, 255, 255}
		if h%2 == 0 {
			tint = color.RGBA{255, 170, 220, 255}
		}
		for i := range 4 {
			angle := float64(i)*1.7 + float64(unit(h, 3))*6
			at := v3(cx+float32(math.Cos(angle))*radius*.7, 14+float32(i%2)*4, cz+float32(math.Sin(angle))*radius*.6)
			solid.blob(at, 2.2, 2, glowing(tint, 1), h+uint32(i))
		}
		f.glows = append(f.glows, forestGlow{at: v3(cx, 1, cz), size: 80, flat: true, tint: color.RGBA{tint.R, tint.G, tint.B, 110}, phase: unit(h, 4) * 6, pulses: true})
	}
}

func (f *forestRenderer) buildGrass(g forestGrid, x, y int, solid *meshBuilder) {
	const tile = float32(parameters.Tile)
	h := cellHash(x, y)
	x0, z0 := float32(x)*tile, float32(y)*tile
	nearTree := false
	for _, s := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		nearTree = nearTree || treeTerrain(g.at(x+s[0], y+s[1]))
	}
	for i := range int(1 + h%3) {
		px, pz := x0+4+unit(h, uint32(30+i))*32, z0+4+unit(h, uint32(40+i))*32
		height := 7 + unit(h, uint32(50+i))*7
		tint := mixColor(color.RGBA{64, 116, 72, 255}, color.RGBA{96, 150, 88, 255}, unit(h, uint32(60+i)))
		for j := range 3 {
			lean := float32(j-1) * 3
			solid.tri(v3(px+float32(j)*2-2, 0, pz), v3(px+float32(j)*2+.5, 0, pz), v3(px+float32(j)*2-1+lean, height-float32(j%2)*2, pz-1), v3(0, .3, 1), tint)
		}
	}
	if nearTree && h%7 == 0 {
		tint := mushroomTeal
		if h%2 == 0 {
			tint = mushroomRose
		}
		count := int(1 + h%3)
		for i := range count {
			px, pz := x0+8+unit(h, uint32(70+i))*24, z0+8+unit(h, uint32(80+i))*24
			stem := 5 + unit(h, uint32(90+i))*5
			solid.prism(v3(px, 0, pz), 1.8, 1.4, stem, 5, 0, color.RGBA{214, 206, 190, 254}, color.RGBA{214, 206, 190, 254}, h)
			solid.prism(v3(px, stem, pz), 5.5-float32(i), 0, 4, 7, float32(i), glowing(tint, .9), glowing(tint, .9), h>>uint(i))
		}
		f.glows = append(f.glows, forestGlow{at: v3(x0+20, 1, z0+20), size: 110, flat: true, tint: color.RGBA{tint.R, tint.G, tint.B, 130}, phase: unit(h, 99) * 6, pulses: true})
	} else if h%17 == 0 {
		px, pz := x0+8+unit(h, 5)*24, z0+8+unit(h, 6)*24
		tint := color.RGBA{255, 230, 140, 255}
		if h%2 == 0 {
			tint = color.RGBA{180, 190, 255, 255}
		}
		solid.prism(v3(px, 0, pz), .8, .8, 8, 4, 0, color.RGBA{80, 130, 70, 254}, color.RGBA{80, 130, 70, 254}, h)
		solid.blob(v3(px, 9, pz), 2.4, 1.8, glowing(tint, .8), h)
	}
}

func (f *forestRenderer) buildProp(area mapdata.CellRect, kind mapdata.PropKind, solid *meshBuilder) {
	bounds := area.Bounds()
	cx, cz := float32(bounds.X+bounds.W/2), float32(bounds.Y+bounds.H/2)
	h := cellHash(area.X, area.Y)
	switch kind {
	case mapdata.StumpProp:
		radius := float32(min(bounds.W, bounds.H)) / 2 * .9
		for i := range 7 {
			angle := float64(i)*2*math.Pi/7 + .4
			end := v3(cx+float32(math.Cos(angle))*radius*1.35, 0, cz+float32(math.Sin(angle))*radius*1.35)
			solid.cylinder(v3(cx+float32(math.Cos(angle))*radius*.7, 14, cz+float32(math.Sin(angle))*radius*.7), end, 10, 5, barkTint, barkTint)
		}
		height := float32(48)
		solid.prism(v3(cx, 0, cz), radius, radius*.94, height, 16, .2, shade(barkTint, 1.05), color.RGBA{176, 138, 92, 254}, h)
		for i := range 4 {
			r := radius * (.8 - float32(i)*.18)
			tint := color.RGBA{150, 112, 72, 254}
			if i%2 == 1 {
				tint = color.RGBA{180, 142, 96, 254}
			}
			solid.disc(v3(cx, height+.4+float32(i)*.2, cz), r, r, 16, tint)
		}
		solid.blob(v3(cx-radius*.4, height, cz-radius*.3), radius*.3, 6, rockMoss, h)
		for i := range 5 {
			angle := float64(i) * 1.3
			at := v3(cx+float32(math.Cos(angle))*radius*.45, height, cz+float32(math.Sin(angle))*radius*.35)
			solid.prism(at, 1.6, 1.2, 6, 5, 0, color.RGBA{214, 206, 190, 254}, color.RGBA{214, 206, 190, 254}, h)
			solid.prism(v3(at.X, height+6, at.Z), 5, 0, 4, 7, 0, glowing(mushroomTeal, .9), glowing(mushroomTeal, .9), h+uint32(i))
		}
		f.glows = append(f.glows, forestGlow{at: v3(cx, height+2, cz), size: 260, tint: color.RGBA{110, 240, 220, 110}, phase: 1, pulses: true, flat: true})
	case mapdata.LogProp:
		from, to := v3(float32(bounds.X)+6, 13, cz), v3(float32(bounds.X+bounds.W)-6, 13, cz)
		if area.Height > area.Width {
			from, to = v3(cx, 13, float32(bounds.Y)+6), v3(cx, 13, float32(bounds.Y+bounds.H)-6)
		}
		solid.cylinder(from, to, 13, 9, barkTint, color.RGBA{170, 130, 84, 254})
		mid := rl.Vector3Lerp(from, to, .4)
		solid.blob(v3(mid.X, 25, mid.Z), 12, 5, rockMoss, h)
		end := rl.Vector3Lerp(from, to, .8)
		solid.prism(v3(end.X, 24, end.Z), 1.6, 1.2, 5, 5, 0, color.RGBA{214, 206, 190, 254}, color.RGBA{214, 206, 190, 254}, h)
		solid.prism(v3(end.X, 29, end.Z), 5, 0, 4, 7, 0, glowing(mushroomRose, .9), glowing(mushroomRose, .9), h)
		f.glows = append(f.glows, forestGlow{at: v3(end.X, 1, end.Z), size: 110, flat: true, tint: color.RGBA{214, 136, 255, 130}, phase: 2, pulses: true})
	}
}

func (f *forestRenderer) buildPortalRing(center rl.Vector2, solid *meshBuilder) {
	for i := range 10 {
		angle := float64(i) * 2 * math.Pi / 10
		at := v3(center.X+float32(math.Cos(angle))*112, 0, center.Y+float32(math.Sin(angle))*112)
		height := 20 + float32(i%3)*7
		solid.prism(at, 9, 5, height, 5, float32(angle), stoneTint, stoneTop, uint32(i))
		solid.blob(v3(at.X, height+3, at.Z), 3, 3, glowing(portalGlow, 1), uint32(i))
	}
	f.glows = append(f.glows, forestGlow{at: v3(center.X, 1.5, center.Y), size: 340, tint: color.RGBA{120, 220, 255, 110}, phase: 0, pulses: true, flat: true})
}
