package worldmap

import (
	"image/color"
	"math"

	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ForestActor struct {
	ID            string
	Label         string
	Center        rl.Vector2
	Width, Depth  float32
	Height, Lift  float32
	Tint          color.RGBA
	Flash, Spawn  float32
	Wings, Flying bool
	NoShadow      bool
}

type actorPose struct {
	ForestActor
	lift float32
}

var wingTint = color.RGBA{192, 220, 244, 255}

func (f *forestRenderer) poses(actors []ForestActor) []actorPose {
	clock := rl.GetTime()
	poses := make([]actorPose, 0, len(actors))
	for _, a := range actors {
		pose := actorPose{ForestActor: a, lift: a.Lift}
		if a.Flying {
			pose.lift = 18 + float32(math.Sin(clock*5))*3
		}
		poses = append(poses, pose)
	}
	return poses
}

func (f *forestRenderer) drawActorShadows(poses []actorPose) {
	for _, p := range poses {
		if p.NoShadow {
			continue
		}
		height := min(p.lift, 40)
		size := (p.Width + p.Depth) * (1.6 - height/80) * p.Spawn
		alpha := uint8(235 - height*3)
		offset := rl.Vector2Scale(rl.Vector2Normalize(rl.NewVector2(forestSunDir.X, forestSunDir.Z)), 6+height*.4)
		f.drawFlatShadow(v3(p.Center.X+offset.X, 1.2, p.Center.Y+offset.Y), size, color.RGBA{0, 4, 10, alpha})
	}
}

func (f *forestRenderer) drawFlatShadow(at rl.Vector3, size float32, tint color.RGBA) {
	rl.BeginBlendMode(rl.BlendAlpha)
	f.drawFlatGlow(at, size, tint)
	rl.EndBlendMode()
}

func (f *forestRenderer) drawActors(poses []actorPose) {
	for _, p := range poses {
		scale := p.Spawn * p.Spawn * (3 - 2*p.Spawn)
		if scale <= .01 {
			continue
		}
		width, depth, height := p.Width*scale, p.Depth*scale, p.Height*scale
		base := v3(p.Center.X, p.lift, p.Center.Y)
		tint := mixColor(p.Tint, rl.White, p.Flash*.85)
		f.drawCube(base, width, height, depth, tint)
		f.drawCube(v3(base.X, base.Y+height, base.Z), width*.84, height*.08, depth*.84, mixColor(tint, rl.White, .18))
		if p.Wings {
			f.drawWings(p, base, width, height)
		}
	}
}

func (f *forestRenderer) drawCube(base rl.Vector3, width, height, depth float32, tint color.RGBA) {
	f.actorMaterial.GetMap(rl.MapDiffuse).Color = tint
	transform := rl.MatrixMultiply(rl.MatrixScale(width, height, depth), rl.MatrixTranslate(base.X, base.Y, base.Z))
	rl.DrawMesh(f.cube, f.actorMaterial, transform)
}

func (f *forestRenderer) drawWings(p actorPose, base rl.Vector3, width, height float32) {
	clock := rl.GetTime()
	spread, flap, back := float32(16), float32(-.55), float32(8)
	if p.Flying {
		spread, back = 40, 2
		flap = float32(math.Sin(clock*11))*.6 + .2
	}
	rl.DisableBackfaceCulling()
	rl.SetTexture(rl.GetTextureIdDefault())
	rl.Begin(rl.Triangles)
	for _, s := range []float32{-1, 1} {
		root := v3(base.X+s*width*.42, base.Y+height*.72, base.Z+2)
		place := func(out, up, z float32) rl.Vector3 {
			sin, cos := float32(math.Sin(float64(flap))), float32(math.Cos(float64(flap)))
			return v3(root.X+s*(out*cos-up*sin), root.Y+out*sin+up*cos, root.Z+z)
		}
		for i := range 5 {
			k := float32(i)
			tip := place(spread+10-k*3, 13-k*6, back*.4+k*3)
			end := place(10+k*1.5, -6, back+5-k)
			alpha := uint8((.9 - k*.1) * 255)
			shadeTint := mixColor(wingTint, rl.White, .3-k*.06)
			rl.Color4ub(shadeTint.R, shadeTint.G, shadeTint.B, alpha)
			rl.Vertex3f(root.X, root.Y, root.Z)
			rl.Vertex3f(tip.X, tip.Y, tip.Z)
			rl.Color4ub(wingTint.R, wingTint.G, wingTint.B, alpha)
			rl.Vertex3f(end.X, end.Y, end.Z)
		}
	}
	rl.End()
	rl.DrawRenderBatchActive()
	rl.EnableBackfaceCulling()
	if p.Flying {
		rl.BeginBlendMode(rl.BlendAdditive)
		f.drawFlatGlow(v3(p.Center.X, 1.4, p.Center.Y), 130, color.RGBA{wingTint.R, wingTint.G, wingTint.B, 60})
		rl.EndBlendMode()
	}
}

func (f *forestRenderer) drawActorLabels(poses []actorPose, width, height int32) {
	for _, p := range poses {
		if p.Label == "" {
			continue
		}
		top := v3(p.Center.X, p.lift+p.Height*1.1+16, p.Center.Y)
		at := rl.GetWorldToScreenEx(top, f.camera, width, height)
		size := primitives.MeasureText(p.Label, 18)
		primitives.DrawTextAt(p.Label, rl.NewVector2(at.X-float32(size)/2, at.Y-18), 18, rl.RayWhite)
	}
}

func (f *forestRenderer) pickActor(camera rl.Camera3D, screen rl.Vector2, width, height int32) (rl.Vector2, bool) {
	ray := rl.GetScreenToWorldRayEx(screen, camera, width, height)
	best, found := float32(math.Inf(1)), false
	var target rl.Vector2
	for _, p := range f.lastPoses {
		if p.ID == simulation.PlayerID || p.Height <= 0 {
			continue
		}
		box := rl.NewBoundingBox(v3(p.Center.X-p.Width/2, p.lift, p.Center.Y-p.Depth/2), v3(p.Center.X+p.Width/2, p.lift+p.Height, p.Center.Y+p.Depth/2))
		if hit := rl.GetRayCollisionBox(ray, box); hit.Hit && hit.Distance < best {
			best, target, found = hit.Distance, p.Center, true
		}
	}
	return target, found
}
