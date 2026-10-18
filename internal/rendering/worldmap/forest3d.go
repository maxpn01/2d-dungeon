package worldmap

import (
	"image/color"
	"math"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/rendering/primitives"
	mapdata "2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	forestChunk     = 16
	forestBorder    = 10
	forestPitch     = 56.0
	forestFovy      = 32.0
	forestViewScale = 1.15
	forestFogCells  = 72
	overlayWidth    = 2560
	overlayHeight   = 2048
)

var forestSunDir = rl.Vector3Normalize(v3(.62, -.58, .34))

type forestChunkMeshes struct {
	min, max             rl.Vector2
	ground, water, solid rl.Mesh
	shadow               rl.Mesh
	hasWater, hasShadow  bool
	hasGround, hasSolid  bool
}

type forestGlow struct {
	at     rl.Vector3
	size   float32
	tint   color.RGBA
	phase  float32
	pulses bool
	flat   bool
}

type forestRenderer struct {
	built         bool
	chunks        []forestChunkMeshes
	glows         []forestGlow
	shafts        []rl.Vector3
	shader        rl.Shader
	material      rl.Material
	locs          map[string]int32
	overlay       rl.RenderTexture2D
	glowTex       rl.Texture2D
	fogTex        rl.Texture2D
	fogPixels     []color.RGBA
	fogOrigin     [2]int
	camera        rl.Camera3D
	overlayMin    rl.Vector2
	overlayMax    rl.Vector2
	cube          rl.Mesh
	actorMaterial rl.Material
}

var forest3D forestRenderer

func (m *Map) InForest() bool { return m.floor == mapdata.SecondFloor }

func ForestCamera(camera rl.Camera2D, width, height int32) rl.Camera3D {
	zoom := max(camera.Zoom, .1)
	tx := camera.Target.X - (camera.Offset.X-float32(width)/2)/zoom
	ty := camera.Target.Y - (camera.Offset.Y-float32(height)/2)/zoom
	distance := float32(height) / (2 * float32(math.Tan(forestFovy*math.Pi/360)) * zoom) * forestViewScale
	pitch := forestPitch * math.Pi / 180
	return rl.Camera3D{
		Position:   v3(tx, distance*float32(math.Sin(pitch)), ty+distance*float32(math.Cos(pitch))),
		Target:     v3(tx, 0, ty),
		Up:         v3(0, 1, 0),
		Fovy:       forestFovy,
		Projection: rl.CameraPerspective,
	}
}

func groundPoint(camera rl.Camera3D, screen rl.Vector2, width, height int32) (rl.Vector2, bool) {
	ray := rl.GetScreenToWorldRayEx(screen, camera, width, height)
	if ray.Direction.Y > -1e-4 {
		return rl.Vector2{}, false
	}
	t := -ray.Position.Y / ray.Direction.Y
	return rl.NewVector2(ray.Position.X+ray.Direction.X*t, ray.Position.Z+ray.Direction.Z*t), true
}

func (m *Map) ScreenToWorld(camera rl.Camera2D, width, height int32, screen rl.Vector2) rl.Vector2 {
	if m.InForest() {
		if p, ok := groundPoint(ForestCamera(camera, width, height), screen, width, height); ok {
			return p
		}
	}
	return rl.GetScreenToWorld2D(screen, camera)
}

func (m *Map) DrawForest3D(ctx primitives.Context, focus rl.Vector2, actors []ForestActor, drawFlat func(rl.Camera2D)) {
	f := &forest3D
	if !f.built {
		f.build(m.world)
	}
	m.Exploration.advanceFog()
	f.camera = ForestCamera(ctx.Camera, ctx.WindowWidth, ctx.WindowHeight)
	poses := f.poses(actors)
	f.drawOverlay(ctx, focus, drawFlat)
	f.updateFog(m, focus)
	rl.ClearBackground(rl.NewColor(10, 14, 24, 255))
	rl.SetClipPlanes(20, 12000)
	rl.BeginMode3D(f.camera)
	f.setUniforms(ctx, focus)
	visible := f.visibleChunks()
	identity := rl.MatrixIdentity()
	f.setMode(4)
	for _, c := range visible {
		if c.hasGround {
			rl.DrawMesh(c.ground, f.material, identity)
		}
	}
	f.setMode(2)
	for _, c := range visible {
		if c.hasWater {
			rl.DrawMesh(c.water, f.material, identity)
		}
	}
	rl.DisableDepthMask()
	f.setMode(1)
	for _, c := range visible {
		if c.hasShadow {
			rl.DrawMesh(c.shadow, f.material, identity)
		}
	}
	f.drawActorShadows(poses)
	f.drawOverlayQuad()
	rl.EnableDepthMask()
	f.setMode(0)
	for _, c := range visible {
		if c.hasSolid {
			rl.DrawMesh(c.solid, f.material, identity)
		}
	}
	f.drawActors(poses)
	f.drawLights(focus)
	rl.EndMode3D()
	rl.SetClipPlanes(.01, 1000)
	drawForestVignette(ctx.WindowWidth, ctx.WindowHeight)
	f.drawActorLabels(poses, ctx.WindowWidth, ctx.WindowHeight)
}

func (m *Map) ForestScreenPoint(ctx primitives.Context, x, y, lift float32) rl.Vector2 {
	return rl.GetWorldToScreenEx(v3(x, lift, y), forest3D.camera, ctx.WindowWidth, ctx.WindowHeight)
}

func (f *forestRenderer) setUniforms(ctx primitives.Context, focus rl.Vector2) {
	set := func(name string, values []float32, kind rl.ShaderUniformDataType) {
		rl.SetShaderValue(f.shader, f.locs[name], values, kind)
	}
	player := rl.GetWorldToScreenEx(v3(focus.X, 0, focus.Y), f.camera, ctx.WindowWidth, ctx.WindowHeight)
	set("clock", []float32{float32(rl.GetTime())}, rl.ShaderUniformFloat)
	set("sunDir", []float32{forestSunDir.X, forestSunDir.Y, forestSunDir.Z}, rl.ShaderUniformVec3)
	set("sunColor", []float32{1.05, .8, .58}, rl.ShaderUniformVec3)
	set("ambient", []float32{.5, .62, .74}, rl.ShaderUniformVec3)
	set("lantern", []float32{focus.X, 320, focus.Y}, rl.ShaderUniformVec3)
	set("cameraPos", []float32{f.camera.Position.X, f.camera.Position.Y, f.camera.Position.Z}, rl.ShaderUniformVec3)
	set("playerScreen", []float32{player.X, float32(ctx.WindowHeight) - player.Y}, rl.ShaderUniformVec2)
	set("playerDepthZ", []float32{focus.Y}, rl.ShaderUniformFloat)
	set("hazeColor", []float32{.16, .2, .3}, rl.ShaderUniformVec3)
	size := float32(forestFogCells * parameters.Tile)
	set("fogArea", []float32{float32(f.fogOrigin[0] * parameters.Tile), float32(f.fogOrigin[1] * parameters.Tile), size, size}, rl.ShaderUniformVec4)
}

func (f *forestRenderer) setMode(mode int32) {
	rl.SetShaderValue(f.shader, f.locs["mode"], []float32{float32(mode)}, rl.ShaderUniformFloat)
}

func (f *forestRenderer) visibleChunks() []forestChunkMeshes {
	minX, minY, maxX, maxY := f.overlayMin.X-240, f.overlayMin.Y-200, f.overlayMax.X+240, f.overlayMax.Y+400
	result := make([]forestChunkMeshes, 0, 32)
	for _, c := range f.chunks {
		if c.max.X < minX || c.min.X > maxX || c.max.Y < minY || c.min.Y > maxY {
			continue
		}
		result = append(result, c)
	}
	return result
}

func (f *forestRenderer) drawOverlay(ctx primitives.Context, focus rl.Vector2, drawFlat func(rl.Camera2D)) {
	if f.overlay.ID == 0 {
		f.overlay = rl.LoadRenderTexture(overlayWidth, overlayHeight)
		rl.SetTextureFilter(f.overlay.Texture, rl.FilterBilinear)
	}
	minX, minY := float32(math.Inf(1)), float32(math.Inf(1))
	maxX, maxY := float32(math.Inf(-1)), float32(math.Inf(-1))
	w, h := float32(ctx.WindowWidth), float32(ctx.WindowHeight)
	for _, s := range []rl.Vector2{{X: 0, Y: 0}, {X: w, Y: 0}, {X: 0, Y: h}, {X: w, Y: h}, {X: w / 2, Y: 0}} {
		p, ok := groundPoint(f.camera, s, ctx.WindowWidth, ctx.WindowHeight)
		if !ok {
			p = rl.NewVector2(focus.X+(s.X/w-.5)*4000, focus.Y-3000)
		}
		minX, minY, maxX, maxY = min(minX, p.X), min(minY, p.Y), max(maxX, p.X), max(maxY, p.Y)
	}
	minX, maxX = max(minX, focus.X-3200), min(maxX, focus.X+3200)
	minY, maxY = max(minY, focus.Y-3600), min(maxY, focus.Y+1600)
	minX, minY, maxX, maxY = minX-60, minY-60, maxX+60, maxY+60
	f.overlayMin, f.overlayMax = rl.NewVector2(minX, minY), rl.NewVector2(maxX, maxY)
	scale := min(overlayWidth/(maxX-minX), overlayHeight/(maxY-minY), 1.5)
	f.overlayMax = rl.NewVector2(minX+overlayWidth/scale, minY+overlayHeight/scale)
	camera := rl.NewCamera2D(rl.Vector2{}, rl.NewVector2(minX, minY), 0, scale)
	rl.BeginTextureMode(f.overlay)
	rl.ClearBackground(rl.Blank)
	rl.SetBlendFactorsSeparate(0x0302, 0x0303, 1, 0x0303, 0x8006, 0x8006)
	rl.BeginBlendMode(rl.BlendCustomSeparate)
	drawFlat(camera)
	rl.EndBlendMode()
	rl.EndTextureMode()
}

func (f *forestRenderer) drawOverlayQuad() {
	rl.BeginShaderMode(f.shader)
	f.setMode(3)
	rl.SetShaderValueTexture(f.shader, f.locs["texture1"], f.fogTex)
	rl.BeginBlendMode(rl.BlendAlphaPremultiply)
	rl.SetTexture(f.overlay.Texture.ID)
	rl.Begin(rl.Quads)
	rl.Color4ub(255, 255, 255, 255)
	y := float32(.6)
	rl.TexCoord2f(0, 1)
	rl.Vertex3f(f.overlayMin.X, y, f.overlayMin.Y)
	rl.TexCoord2f(0, 0)
	rl.Vertex3f(f.overlayMin.X, y, f.overlayMax.Y)
	rl.TexCoord2f(1, 0)
	rl.Vertex3f(f.overlayMax.X, y, f.overlayMax.Y)
	rl.TexCoord2f(1, 1)
	rl.Vertex3f(f.overlayMax.X, y, f.overlayMin.Y)
	rl.End()
	rl.SetTexture(0)
	rl.EndBlendMode()
	rl.EndShaderMode()
}

func (f *forestRenderer) updateFog(m *Map, focus rl.Vector2) {
	if f.fogTex.ID == 0 {
		img := rl.GenImageColor(forestFogCells, forestFogCells, rl.White)
		f.fogTex = rl.LoadTextureFromImage(img)
		rl.UnloadImage(img)
		rl.SetTextureFilter(f.fogTex, rl.FilterBilinear)
		f.fogPixels = make([]color.RGBA, forestFogCells*forestFogCells)
		f.material.GetMap(rl.MapSpecular).Texture = f.fogTex
	}
	center := mapdata.WorldCell(float64(focus.X), float64(focus.Y))
	f.fogOrigin = [2]int{center[0] - forestFogCells/2, center[1] - forestFogCells/2}
	for y := range forestFogCells {
		for x := range forestFogCells {
			cell := [2]int{f.fogOrigin[0] + x, f.fogOrigin[1] + y}
			opacity := m.Exploration.fogOpacity(float64(cell[0]*parameters.Tile+parameters.Tile/2), float64(cell[1]*parameters.Tile+parameters.Tile/2))
			if m.Exploration.Explored(cell) {
				opacity = min(opacity, .42)
			} else {
				opacity = min(opacity, .8)
			}
			v := uint8(opacity * 255)
			f.fogPixels[y*forestFogCells+x] = color.RGBA{v, v, v, 255}
		}
	}
	rl.UpdateTexture(f.fogTex, f.fogPixels)
}

func (f *forestRenderer) fogAt(x, z float32) float32 {
	cx := int(math.Floor(float64(x)/parameters.Tile)) - f.fogOrigin[0]
	cy := int(math.Floor(float64(z)/parameters.Tile)) - f.fogOrigin[1]
	if cx < 0 || cy < 0 || cx >= forestFogCells || cy >= forestFogCells {
		return 1
	}
	return float32(f.fogPixels[cy*forestFogCells+cx].R) / 255
}

func (f *forestRenderer) drawLights(focus rl.Vector2) {
	rl.DisableDepthMask()
	rl.BeginBlendMode(rl.BlendAdditive)
	clock := float32(rl.GetTime())
	for _, s := range f.shafts {
		if s.X < f.overlayMin.X || s.X > f.overlayMax.X || s.Z < f.overlayMin.Y || s.Z > f.overlayMax.Y {
			continue
		}
		fog := f.fogAt(s.X, s.Z)
		if fog > .6 {
			continue
		}
		strength := (1 - fog/.6) * (.6 + .4*float32(math.Sin(float64(clock*.35+s.X*.01))))
		drawLightShaft(s, strength)
		f.drawFlatGlow(v3(s.X, 1, s.Z), 150, rl.Fade(rl.NewColor(255, 214, 150, 255), .22*strength))
	}
	for _, g := range f.glows {
		if g.at.X < f.overlayMin.X || g.at.X > f.overlayMax.X || g.at.Z < f.overlayMin.Y || g.at.Z > f.overlayMax.Y {
			continue
		}
		strength := 1 - f.fogAt(g.at.X, g.at.Z)*.85
		if g.pulses {
			strength *= .65 + .35*float32(math.Sin(float64(clock*2+g.phase)))
		}
		if g.flat {
			f.drawFlatGlow(g.at, g.size, rl.Fade(g.tint, strength))
			continue
		}
		rl.DrawBillboard(f.camera, f.glowTex, g.at, g.size, rl.Fade(g.tint, strength))
	}
	fireflyCell := float32(150)
	gx0, gy0 := int(math.Floor(float64((focus.X-1300)/fireflyCell))), int(math.Floor(float64((focus.Y-1800)/fireflyCell)))
	for gy := gy0; gy < gy0+int(2600/fireflyCell); gy++ {
		for gx := gx0; gx < gx0+int(2600/fireflyCell); gx++ {
			hash := uint32(gx*73856093) ^ uint32(gy*19349663)
			if hash%4 != 0 {
				continue
			}
			phase := float64(hash%997) * .1
			t := float64(clock)*.35 + phase
			x := float32(gx)*fireflyCell + float32(hash%120) + float32(math.Sin(t*1.3))*38
			z := float32(gy)*fireflyCell + float32((hash>>8)%120) + float32(math.Cos(t*.9))*30
			y := 26 + float32((hash>>16)%50) + float32(math.Sin(t*2.1))*10
			blink := float32(math.Max(0, math.Sin(t*3.1+phase)))
			strength := blink * (1 - f.fogAt(x, z)*.9)
			if strength < .05 {
				continue
			}
			rl.DrawBillboard(f.camera, f.glowTex, v3(x, y, z), 16, rl.Fade(rl.NewColor(214, 255, 140, 255), strength))
			rl.DrawBillboard(f.camera, f.glowTex, v3(x, y, z), 5, rl.Fade(rl.NewColor(250, 255, 220, 255), strength))
		}
	}
	rl.EndBlendMode()
	rl.EnableDepthMask()
}

func drawLightShaft(base rl.Vector3, strength float32) {
	top := rl.Vector3Add(base, rl.Vector3Scale(forestSunDir, -560))
	rl.SetTexture(rl.GetTextureIdDefault())
	rl.Begin(rl.Quads)
	for i, width := range []float32{34, 20, 9} {
		alpha := uint8(float32(14+i*10) * strength)
		mid := rl.Vector3Lerp(base, top, .45)
		for _, segment := range [][2]rl.Vector3{{base, mid}, {mid, top}} {
			lower, upper := alpha, uint8(0)
			if segment[0] == base {
				lower, upper = 0, alpha
			}
			rl.Color4ub(255, 222, 168, lower)
			rl.Vertex3f(segment[0].X-width, segment[0].Y, segment[0].Z)
			rl.Color4ub(255, 222, 168, lower)
			rl.Vertex3f(segment[0].X+width, segment[0].Y, segment[0].Z)
			rl.Color4ub(255, 230, 180, upper)
			rl.Vertex3f(segment[1].X+width*1.4, segment[1].Y, segment[1].Z)
			rl.Color4ub(255, 230, 180, upper)
			rl.Vertex3f(segment[1].X-width*1.4, segment[1].Y, segment[1].Z)
		}
	}
	rl.End()
	rl.SetTexture(0)
}

func (f *forestRenderer) drawFlatGlow(at rl.Vector3, size float32, tint color.RGBA) {
	half := size / 2
	rl.SetTexture(f.glowTex.ID)
	rl.Begin(rl.Quads)
	rl.Color4ub(tint.R, tint.G, tint.B, tint.A)
	rl.TexCoord2f(0, 0)
	rl.Vertex3f(at.X-half, at.Y, at.Z-half)
	rl.TexCoord2f(0, 1)
	rl.Vertex3f(at.X-half, at.Y, at.Z+half)
	rl.TexCoord2f(1, 1)
	rl.Vertex3f(at.X+half, at.Y, at.Z+half)
	rl.TexCoord2f(1, 0)
	rl.Vertex3f(at.X+half, at.Y, at.Z-half)
	rl.End()
	rl.SetTexture(0)
}

func drawForestVignette(width, height int32) {
	edge := rl.NewColor(4, 8, 18, 200)
	band := height / 4
	rl.DrawRectangleGradientV(0, 0, width, band, edge, rl.Blank)
	rl.DrawRectangleGradientV(0, height-band/2, width, band/2, rl.Blank, rl.NewColor(4, 8, 18, 150))
	side := width / 6
	rl.DrawRectangleGradientH(0, 0, side, height, rl.NewColor(4, 8, 18, 150), rl.Blank)
	rl.DrawRectangleGradientH(width-side, 0, side, height, rl.Blank, rl.NewColor(4, 8, 18, 150))
}
