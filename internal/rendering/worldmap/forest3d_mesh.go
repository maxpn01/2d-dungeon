package worldmap

import (
	"image/color"
	"math"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type meshBuilder struct {
	positions []float32
	normals   []float32
	colors    []uint8
}

func v3(x, y, z float32) rl.Vector3 { return rl.NewVector3(x, y, z) }

func (b *meshBuilder) vertex(p, n rl.Vector3, c color.RGBA) {
	b.positions = append(b.positions, p.X, p.Y, p.Z)
	b.normals = append(b.normals, n.X, n.Y, n.Z)
	b.colors = append(b.colors, c.R, c.G, c.B, c.A)
}

func (b *meshBuilder) tri(p1, p2, p3, outward rl.Vector3, c color.RGBA) {
	n := rl.Vector3CrossProduct(rl.Vector3Subtract(p2, p1), rl.Vector3Subtract(p3, p1))
	if rl.Vector3Length(n) < 1e-6 {
		return
	}
	if rl.Vector3DotProduct(n, outward) < 0 {
		p2, p3 = p3, p2
		n = rl.Vector3Negate(n)
	}
	n = rl.Vector3Normalize(n)
	b.vertex(p1, n, c)
	b.vertex(p2, n, c)
	b.vertex(p3, n, c)
}

func (b *meshBuilder) smoothQuad(p [4]rl.Vector3, c [4]color.RGBA) {
	n := v3(0, 1, 0)
	for _, i := range []int{0, 1, 2, 0, 2, 3} {
		b.vertex(p[i], n, c[i])
	}
}

func (b *meshBuilder) quad(p1, p2, p3, p4, outward rl.Vector3, c color.RGBA) {
	b.tri(p1, p2, p3, outward, c)
	b.tri(p1, p3, p4, outward, c)
}

func shade(c color.RGBA, f float32) color.RGBA {
	scale := func(v uint8) uint8 { return uint8(min(255, max(0, float32(v)*f))) }
	return color.RGBA{scale(c.R), scale(c.G), scale(c.B), c.A}
}

func mixColor(a, b color.RGBA, t float32) color.RGBA {
	lerp := func(x, y uint8) uint8 { return uint8(float32(x) + (float32(y)-float32(x))*t) }
	return color.RGBA{lerp(a.R, b.R), lerp(a.G, b.G), lerp(a.B, b.B), lerp(a.A, b.A)}
}

func glowing(c color.RGBA, strength float32) color.RGBA {
	c.A = uint8(255 - min(1, strength)*200)
	return c
}

func (b *meshBuilder) prism(base rl.Vector3, bottomRadius, topRadius, height float32, sides int, turn float32, side, top color.RGBA, seed uint32) {
	tip := rl.Vector3Add(base, v3(0, height, 0))
	center := rl.Vector3Add(base, v3(0, height/2, 0))
	ring := func(radius, y float32, i int) rl.Vector3 {
		angle := float64(turn) + float64(i)*2*math.Pi/float64(sides)
		return rl.Vector3Add(base, v3(float32(math.Cos(angle))*radius, y, float32(math.Sin(angle))*radius))
	}
	for i := range sides {
		tone := .9 + float32((seed>>uint(i%16))&3)*.05
		a, d := ring(bottomRadius, 0, i), ring(bottomRadius, 0, i+1)
		face := rl.Vector3Subtract(rl.Vector3Scale(rl.Vector3Add(a, d), .5), center)
		face.Y = 0
		if topRadius <= .01 {
			b.tri(a, d, tip, face, shade(side, tone))
			continue
		}
		bb, cc := ring(topRadius, height, i), ring(topRadius, height, i+1)
		b.quad(a, d, cc, bb, face, shade(side, tone))
		b.tri(tip, bb, cc, v3(0, 1, 0), top)
	}
}

func (b *meshBuilder) cylinder(from, to rl.Vector3, radius float32, sides int, side, cap color.RGBA) {
	axis := rl.Vector3Normalize(rl.Vector3Subtract(to, from))
	helper := v3(0, 1, 0)
	if math.Abs(float64(rl.Vector3DotProduct(axis, helper))) > .9 {
		helper = v3(1, 0, 0)
	}
	u := rl.Vector3Normalize(rl.Vector3CrossProduct(axis, helper))
	w := rl.Vector3CrossProduct(axis, u)
	point := func(origin rl.Vector3, i int) rl.Vector3 {
		angle := float64(i) * 2 * math.Pi / float64(sides)
		return rl.Vector3Add(origin, rl.Vector3Add(rl.Vector3Scale(u, float32(math.Cos(angle))*radius), rl.Vector3Scale(w, float32(math.Sin(angle))*radius)))
	}
	for i := range sides {
		a, d := point(from, i), point(from, i+1)
		bb, cc := point(to, i), point(to, i+1)
		face := rl.Vector3Subtract(rl.Vector3Scale(rl.Vector3Add(a, d), .5), from)
		b.quad(a, d, cc, bb, face, shade(side, .88+float32(i%3)*.06))
		b.tri(from, a, d, rl.Vector3Negate(axis), cap)
		b.tri(to, bb, cc, axis, cap)
	}
}

var blobDirections = func() []rl.Vector3 {
	t := float32((1 + math.Sqrt(5)) / 2)
	points := []rl.Vector3{{X: -1, Y: t}, {X: 1, Y: t}, {X: -1, Y: -t}, {X: 1, Y: -t}, {Y: -1, Z: t}, {Y: 1, Z: t}, {Y: -1, Z: -t}, {Y: 1, Z: -t}, {X: t, Z: -1}, {X: t, Z: 1}, {X: -t, Z: -1}, {X: -t, Z: 1}}
	for i := range points {
		points[i] = rl.Vector3Normalize(points[i])
	}
	return points
}()

var blobFaces = [][3]int{{0, 11, 5}, {0, 5, 1}, {0, 1, 7}, {0, 7, 10}, {0, 10, 11}, {1, 5, 9}, {5, 11, 4}, {11, 10, 2}, {10, 7, 6}, {7, 1, 8}, {3, 9, 4}, {3, 4, 2}, {3, 2, 6}, {3, 6, 8}, {3, 8, 9}, {4, 9, 5}, {2, 4, 11}, {6, 2, 10}, {8, 6, 7}, {9, 8, 1}}

func (b *meshBuilder) blob(center rl.Vector3, radius, height float32, c color.RGBA, seed uint32) {
	points := make([]rl.Vector3, len(blobDirections))
	for i, d := range blobDirections {
		jitter := 1 + (float32((seed*uint32(i+7)*2654435761)>>24)/255-.5)*.28
		points[i] = rl.Vector3Add(center, v3(d.X*radius*jitter, d.Y*height*jitter, d.Z*radius*jitter))
	}
	for i, f := range blobFaces {
		p1, p2, p3 := points[f[0]], points[f[1]], points[f[2]]
		mid := rl.Vector3Scale(rl.Vector3Add(rl.Vector3Add(p1, p2), p3), 1.0/3)
		tone := float32(.92) + float32((seed>>uint(i%20))&3)*.04
		up := rl.Vector3Normalize(rl.Vector3Subtract(mid, center)).Y
		b.tri(p1, p2, p3, rl.Vector3Subtract(mid, center), shade(c, tone+up*.12))
	}
}

func (b *meshBuilder) box(min, max rl.Vector3, side, top color.RGBA) {
	x0, y0, z0, x1, y1, z1 := min.X, min.Y, min.Z, max.X, max.Y, max.Z
	b.quad(v3(x0, y1, z0), v3(x1, y1, z0), v3(x1, y1, z1), v3(x0, y1, z1), v3(0, 1, 0), top)
	b.quad(v3(x0, y0, z1), v3(x1, y0, z1), v3(x1, y1, z1), v3(x0, y1, z1), v3(0, 0, 1), side)
	b.quad(v3(x0, y0, z0), v3(x1, y0, z0), v3(x1, y1, z0), v3(x0, y1, z0), v3(0, 0, -1), shade(side, .8))
	b.quad(v3(x0, y0, z0), v3(x0, y0, z1), v3(x0, y1, z1), v3(x0, y1, z0), v3(-1, 0, 0), shade(side, .9))
	b.quad(v3(x1, y0, z0), v3(x1, y0, z1), v3(x1, y1, z1), v3(x1, y1, z0), v3(1, 0, 0), shade(side, .85))
}

func (b *meshBuilder) disc(center rl.Vector3, rx, rz float32, sides int, c color.RGBA) {
	for i := range sides {
		a0 := float64(i) * 2 * math.Pi / float64(sides)
		a1 := float64(i+1) * 2 * math.Pi / float64(sides)
		p1 := rl.Vector3Add(center, v3(float32(math.Cos(a0))*rx, 0, float32(math.Sin(a0))*rz))
		p2 := rl.Vector3Add(center, v3(float32(math.Cos(a1))*rx, 0, float32(math.Sin(a1))*rz))
		b.tri(center, p1, p2, v3(0, 1, 0), c)
	}
}

func (b *meshBuilder) empty() bool { return len(b.positions) == 0 }

func (b *meshBuilder) upload() rl.Mesh {
	mesh := rl.Mesh{
		VertexCount:   int32(len(b.positions) / 3),
		TriangleCount: int32(len(b.positions) / 9),
		Vertices:      unsafe.SliceData(b.positions),
		Normals:       unsafe.SliceData(b.normals),
		Colors:        unsafe.SliceData(b.colors),
	}
	rl.UploadMesh(&mesh, false)
	mesh.Vertices, mesh.Normals, mesh.Colors = nil, nil, nil
	return mesh
}
