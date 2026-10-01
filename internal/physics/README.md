# Physics

Small, stateless 2D movement and geometry tools. Coordinates and distances use
`float64` in the caller's units. Rectangles use their top-left position and positive
width and height. Operations return values without changing their inputs.

## Geometry

```go
player := physics.Square(x, y, size)
obstacle := physics.Rect{X: 100, Y: 200, W: 40, H: 80}
overlapping := player.Overlaps(obstacle)
distanceSquared := player.Center().DistanceSquared(obstacle.Center())
inside := room.Contains(player)
player = player.ClampedTo(room)
```

`Vector` provides addition, subtraction, scaling, normalization, length, and
distance, including squared length and distance. Normalizing a zero or invalid
vector returns the zero vector.

`Rect` provides centers, translation, expansion, overlap, containment, distance
between rectangle edges, clamping, and overlap with a regular octagon.
`Circle.Contains` checks whether another circle fits completely inside it.

Touching edges do not count as overlap. Containment includes the boundary.
Rectangle distance is zero for touching or overlapping shapes. A rectangle larger
than its clamp bounds is centered along each oversized axis; clamping does not
resize it.

## Movement through terrain

```go
world := physics.World{
	CanOccupy: func(body physics.Rect) bool {
		return mapAllows(body)
	},
	MaxStep: 8,
}
delta := (physics.Vector{X: inputX, Y: inputY}).Normalized().Scale(speed * dt)
motion := world.MoveAndSlide(player, delta)
if motion.Rejected {
	return
}
player = motion.Bounds
if motion.Blocked() {
	handleWallCollision()
}
```

`World` holds a terrain query and movement settings, rather than registered bodies.
`MoveAndSlide` advances X and then Y in each substep, refines blocked movement to
the obstacle boundary, and allows movement along the unblocked axis.
`BlockedX` and `BlockedY` record whether either axis was blocked at any point.
`Rejected` distinguishes invalid or excessive movement from an actual collision;
rejected movement leaves the rectangle unchanged.

`MaxStep` limits displacement per axis per substep. Zero, negative, or nonfinite
values use the default of 8. Steps are also limited to half the body's smaller
dimension. Requests requiring more than 65,536 substeps are rejected.

`world.MovementLimit(player, delta)` checks a straight path without sliding. It returns
`(fraction, blocked)`, where the fraction is in `[0, 1]` and identifies the last clear
position before a detected obstacle. A clear path returns `(1, false)`. An invalid
request, missing terrain query, or blocked starting position returns `(0, true)`.

Terrain movement and movement-limit checks sample the occupancy query; they are not exact
continuous collision detection for arbitrary geometry. Configure `MaxStep` for
the map's smallest relevant features. Movement can leave an occupied starting
position if a substep reaches a clear position, but does not automatically resolve
existing penetration.

## Contact between moving bodies

```go
fraction, hit := player.CollisionFraction(playerDelta, enemy, enemyDelta)
```

`Rect.CollisionFraction` computes continuous contact between two translating rectangles using
their relative movement. Fractions use the same `[0, 1]` interval for both bodies.
Existing overlap returns a hit at zero. Merely touching, grazing a corner, or
first touching at exactly the end of the interval does not count as overlap.
Invalid inputs return `(0, false)`.

Map layout, pathfinding, gate state, damage, ability timing, and movement-speed
rules belong to worldmap or simulation. Rendering and interface hit checks belong
to the game package. Physics has no dependency on those packages or raylib.
