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

func (e *Exploration) fogOpacity(x, y float64) float32 {
	if e.world == nil || e.state == nil {
		return 1
	}
	for _, area := range e.hiddenAreas {
		if x >= area.X && x < area.X+area.W && y >= area.Y && y < area.Y+area.H {
			return 1
		}
	}
	gx, gy := x/mapdata.Tile-.5, y/mapdata.Tile-.5
	cx, cy := int(math.Floor(gx)), int(math.Floor(gy))
	fx, fy := gx-float64(cx), gy-float64(cy)
	visible := func(dx, dy int) float64 {
		return e.fogLight[[2]int{cx + dx, cy + dy}]
	}
	top := visible(0, 0)*(1-fx) + visible(1, 0)*fx
	bottom := visible(0, 1)*(1-fx) + visible(1, 1)*fx
	light := top*(1-fy) + bottom*fy
	cell := mapdata.WorldCell(x, y)
	if e.state.WallVisible(cell) {
		light = max(light, e.fogLight[cell])
	}
	distance := math.Hypot(x-float64(e.focus.X), y-float64(e.focus.Y))
	fade := max(0, min(1, (parameters.ExplorationRadius-distance)/110))
	fade = fade * fade * (3 - 2*fade)
	return float32(1 - light*fade)
}

type fogRenderer struct {
	texture              rl.Texture2D
	shader               rl.Shader
	pixels               []color.RGBA
	width, height        int
	origin, scale, clock int32
}

var fogLayer fogRenderer

const fogFragment = `#version 330
in vec2 fragTexCoord;
in vec4 fragColor;
uniform sampler2D texture0;
uniform vec2 worldOrigin;
uniform vec2 worldScale;
uniform float clock;
out vec4 finalColor;
float hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453);
}
float noise(vec2 p) {
    vec2 i = floor(p), f = fract(p);
    f = f*f*(3.0-2.0*f);
    return mix(mix(hash(i), hash(i+vec2(1,0)), f.x),
               mix(hash(i+vec2(0,1)), hash(i+vec2(1,1)), f.x), f.y);
}
void main() {
    vec2 world = worldOrigin + fragTexCoord * worldScale;
    vec2 p = world * 0.007;
    float mist = noise(p + vec2(clock*0.035, -clock*0.022))*0.6
               + noise(p*2.1 + vec2(-clock*0.026, clock*0.018))*0.3
               + noise(p*4.3)*0.1;
    float mask = texture(texture0, fragTexCoord).r;
    float edge = mask*(1.0-mask);
    float alpha = clamp(mask + (mist-0.5)*edge*0.65, 0.0, 1.0);
    vec3 grey = mix(vec3(0.065,0.070,0.088), vec3(0.125,0.129,0.153), mist);
    float violet = noise(p*0.55 + clock*0.012);
    grey += vec3(0.012,0.002,0.019)*violet;
    finalColor = vec4(grey, alpha);
}`

func (e *Exploration) advanceFog() {
	if e.state == nil {
		return
	}
	reset := e.fogLight == nil || math.Hypot(float64(e.focus.X-e.fogFocus.X), float64(e.focus.Y-e.fogFocus.Y)) > parameters.ExplorationRadius
	if reset {
		e.fogLight = make(map[[2]int]float64)
	}
	e.fogFocus = e.focus
	origin := explorationCell(e.focus)
	blend := 1 - math.Exp(-float64(min(rl.GetFrameTime(), .05))*14)
	for y := origin[1] - 12; y <= origin[1]+12; y++ {
		for x := origin[0] - 12; x <= origin[0]+12; x++ {
			cell := [2]int{x, y}
			target := 0.0
			if e.Visible(cell) {
				target = 1
			}
			light := e.fogLight[cell]
			if reset || target < light {
				light = target
			} else {
				light += (target - light) * blend
			}
			e.fogLight[cell] = light
		}
	}
	for cell := range e.fogLight {
		if absInt(cell[0]-origin[0]) > 12 || absInt(cell[1]-origin[1]) > 12 {
			delete(e.fogLight, cell)
		}
	}
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func (m *Map) DrawFog(ctx primitives.Context) {
	m.Exploration.advanceFog()
	f := &fogLayer
	width, height := (int(ctx.WindowWidth)+5)/6, (int(ctx.WindowHeight)+5)/6
	if width <= 0 || height <= 0 {
		return
	}
	if f.shader.ID == 0 {
		f.shader = rl.LoadShaderFromMemory("", fogFragment)
		f.origin = rl.GetShaderLocation(f.shader, "worldOrigin")
		f.scale = rl.GetShaderLocation(f.shader, "worldScale")
		f.clock = rl.GetShaderLocation(f.shader, "clock")
	}
	if width != f.width || height != f.height {
		if f.texture.ID != 0 {
			rl.UnloadTexture(f.texture)
		}
		img := rl.GenImageColor(width, height, rl.White)
		f.texture = rl.LoadTextureFromImage(img)
		rl.UnloadImage(img)
		rl.SetTextureFilter(f.texture, rl.FilterBilinear)
		f.width, f.height = width, height
		f.pixels = make([]color.RGBA, width*height)
	}
	origin := rl.GetScreenToWorld2D(rl.NewVector2(0, 0), ctx.Camera)
	end := rl.GetScreenToWorld2D(rl.NewVector2(float32(ctx.WindowWidth), float32(ctx.WindowHeight)), ctx.Camera)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			p := rl.GetScreenToWorld2D(rl.NewVector2((float32(x)+.5)*float32(ctx.WindowWidth)/float32(width), (float32(y)+.5)*float32(ctx.WindowHeight)/float32(height)), ctx.Camera)
			opacity := m.Exploration.fogOpacity(float64(p.X), float64(p.Y))
			v := uint8(opacity * 255)
			f.pixels[y*width+x] = color.RGBA{R: v, G: v, B: v, A: 255}
		}
	}
	rl.UpdateTexture(f.texture, f.pixels)
	rl.SetShaderValue(f.shader, f.origin, []float32{origin.X, origin.Y}, rl.ShaderUniformVec2)
	rl.SetShaderValue(f.shader, f.scale, []float32{end.X - origin.X, end.Y - origin.Y}, rl.ShaderUniformVec2)
	rl.SetShaderValue(f.shader, f.clock, []float32{float32(rl.GetTime())}, rl.ShaderUniformFloat)
	rl.BeginShaderMode(f.shader)
	rl.DrawTexturePro(f.texture, rl.NewRectangle(0, 0, float32(width), float32(height)), rl.NewRectangle(0, 0, float32(ctx.WindowWidth), float32(ctx.WindowHeight)), rl.NewVector2(0, 0), 0, rl.White)
	rl.EndShaderMode()
}

func (m *Map) DrawVisibleVaultGate(orb simulation.OrbState, local simulation.Player) {
	gate, ok := m.world.Gate(mapdata.VaultGate)
	if ok && m.Exploration.EntityVisible(float64(gate.Left+gate.Right)/2, float64(gate.Y+mapdata.Tile), 0) {
		m.drawVaultGate(orb, local)
	}
}
