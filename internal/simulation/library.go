package simulation

import (
	"math"
	"math/bits"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/physics"
	"dungeon-raylib-go/internal/worldmap"
)

const LibraryTomeCount = 3
const LibraryStrikeSpeed = 520.0
const LibraryDrag = .65
const LibraryTomeRadius = 16.0
const LibraryWardRadius = 35.0
const LibraryRecallSeconds = .8

type LibraryState struct {
	StartedTick   uint64
	Owner         string
	Bound         uint8
	Completed     bool
	CompletedTick uint64
	Positions     [LibraryTomeCount]physics.Vector
	Velocities    [LibraryTomeCount]physics.Vector
	HitTicks      [LibraryTomeCount]uint64
	BounceTicks   [LibraryTomeCount]uint64
	BoundTicks    [LibraryTomeCount]uint64
	RecallTick    uint64
	RecallFrom    [LibraryTomeCount]physics.Vector
	NudgeTick     uint64
}

func IsLibraryBook(id string) bool {
	return id == worldmap.SecretInscription
}
func LibraryRoom(id string) string {

	return worldmap.SecretLibrary
}
func LibraryArena(book StatueState) physics.Rect {
	room, _ := worldmap.Default().Room(LibraryRoom(book.ID))
	bounds := room.Bounds()
	return physics.Rect{X: bounds.X + 52, Y: bounds.Y + 52, W: bounds.W - 104, H: bounds.H - 104}
}
func LibraryWard(book StatueState, index int) physics.Vector {
	bounds := LibraryArena(book)
	var at physics.Vector
	switch index {
	case 0:
		at = physics.Vector{X: bounds.X + 42, Y: bounds.Y + 40}
	case 1:
		at = physics.Vector{X: bounds.X + bounds.W - 42, Y: bounds.Y + bounds.H - 40}
	default:
		at = physics.Vector{X: bounds.X + 42, Y: bounds.Y + bounds.H - 65}
	}

	return at
}
func LibraryHome(book StatueState, index int) physics.Vector {
	bounds := LibraryArena(book)
	offsets := [LibraryTomeCount]physics.Vector{{X: -70, Y: -100}, {X: 70, Y: 70}, {X: -105, Y: 115}}
	offset := offsets[index]

	return physics.Vector{X: bounds.X + bounds.W/2 + offset.X, Y: bounds.Y + bounds.H/2 + offset.Y}
}
func LibraryShelves(book StatueState) []physics.Rect {
	room, _ := worldmap.Default().Room(LibraryRoom(book.ID))
	var shelves []physics.Rect
	for _, prop := range room.Props {
		if prop.Kind == worldmap.BookshelfProp {
			shelves = append(shelves, room.WorldArea(prop.Area).Bounds())
		}
	}
	return shelves
}
func LibraryRecalling(book StatueState, tick uint64) bool {
	return book.Library.RecallTick != 0 && tick >= book.Library.RecallTick && float64(tick-book.Library.RecallTick)/parameters.TickRate < LibraryRecallSeconds
}
func LibrarySolved(book StatueState) int { return bits.OnesCount8(book.Library.Bound) }

func (w *World) startLibrary(book *StatueState, p *playerState) bool {
	if book.Library.Completed {
		return true
	}
	if book.Library.StartedTick != 0 {
		if math.Hypot(book.X-p.X-parameters.PlayerSize/2, book.Y-p.Y-parameters.PlayerSize/2) >= 55 {
			return false
		}
		if book.Library.Owner != p.ID || LibraryRecalling(*book, w.tick) {
			return true
		}
		book.Library.RecallTick = w.tick
		book.Library.RecallFrom = book.Library.Positions
		book.Library.Velocities = [LibraryTomeCount]physics.Vector{}
		return true
	}
	book.Library = LibraryState{StartedTick: w.tick, Owner: p.ID}
	for i := 0; i < LibraryTomeCount; i++ {
		book.Library.Positions[i] = LibraryHome(*book, i)
	}
	return true
}

func (w *World) interactLibrary(p *playerState, input Direction) bool {
	for i := range w.statues {
		book := &w.statues[i]
		if !IsLibraryBook(book.ID) || book.Library.StartedTick == 0 || book.Library.Completed || book.Library.Owner != p.ID {
			continue
		}
		if !worldmap.Default().WithSecrets(w.openSecrets).RoomContains(LibraryRoom(book.ID), physics.Square(p.X, p.Y, parameters.PlayerSize)) {
			continue
		}
		nearest, distance := -1, 78.0
		for tome, at := range book.Library.Positions {
			if book.Library.Bound&(1<<tome) != 0 {
				continue
			}
			d := math.Hypot(at.X-p.X-parameters.PlayerSize/2, at.Y-p.Y-parameters.PlayerSize/2)
			if d < distance {
				half := float64(parameters.PlayerSize) / 2
				geometry := worldmap.Default().WithSecrets(w.openSecrets)
				if _, blocked := geometry.MovementLimit(p.X+half-.5, p.Y+half-.5, at.X-p.X-half, at.Y-p.Y-half, 1, 1); blocked {
					continue
				}
				nearest, distance = tome, d
			}
		}
		if nearest < 0 || math.Hypot(book.X-p.X-parameters.PlayerSize/2, book.Y-p.Y-parameters.PlayerSize/2) < 55 {
			continue
		}
		if LibraryRecalling(*book, w.tick) || book.Library.NudgeTick != 0 && w.tick-book.Library.NudgeTick < parameters.TickRate/4 {
			return true
		}
		dx, dy := input.AimX, input.AimY
		if math.Hypot(dx, dy) < .01 {
			dx, dy = p.Combat.Facing.X, p.Combat.Facing.Y
		}
		if math.Hypot(dx, dy) < .01 {
			at := book.Library.Positions[nearest]
			dx, dy = at.X-p.X-parameters.PlayerSize/2, at.Y-p.Y-parameters.PlayerSize/2
		}
		length := math.Hypot(dx, dy)
		if length > .01 {
			w.strikeLibraryTome(book, nearest, physics.Vector{X: dx / length, Y: dy / length})
			book.Library.NudgeTick = w.tick
		}
		return true
	}
	return false
}

func (w *World) strikeLibraryTome(book *StatueState, tome int, direction physics.Vector) {
	book.Library.Velocities[tome] = physics.Vector{X: direction.X * LibraryStrikeSpeed, Y: direction.Y * LibraryStrikeSpeed}
	book.Library.HitTicks[tome] = w.tick
}

func (w *World) advanceLibraries() {
	for i := range w.statues {
		book := &w.statues[i]
		if !IsLibraryBook(book.ID) || book.Library.StartedTick == 0 || book.Library.Completed {
			continue
		}
		owner := w.playerWithID(book.Library.Owner)
		if owner == nil || owner.Dead {
			book.Library = LibraryState{}
			continue
		}
		arena := LibraryArena(*book)
		shelves := LibraryShelves(*book)
		for tome := 0; tome < LibraryTomeCount; tome++ {
			if book.Library.Bound&(1<<tome) != 0 {
				continue
			}
			if LibraryRecalling(*book, w.tick) {
				t := float64(w.tick-book.Library.RecallTick) / parameters.TickRate / LibraryRecallSeconds
				t = t * t * (3 - 2*t)
				from, to := book.Library.RecallFrom[tome], LibraryHome(*book, tome)
				book.Library.Positions[tome] = physics.Vector{X: from.X + (to.X-from.X)*t, Y: from.Y + (to.Y-from.Y)*t}
				continue
			}
			if book.Library.RecallTick != 0 && w.tick-book.Library.RecallTick == uint64(LibraryRecallSeconds*parameters.TickRate) {
				book.Library.Positions[tome] = LibraryHome(*book, tome)
			}
			wards := [LibraryTomeCount]physics.Vector{}
			for ward := range wards {
				wards[ward] = LibraryWard(*book, ward)
			}
			at, velocity, bounced, bound := LibraryTomeStep(arena, wards, shelves, tome, book.Library.Positions[tome], book.Library.Velocities[tome])
			if bound {
				book.Library.Bound |= 1 << tome
				book.Library.BoundTicks[tome] = w.tick
			}
			book.Library.Positions[tome], book.Library.Velocities[tome] = at, velocity
			if bounced {
				book.Library.BounceTicks[tome] = w.tick
			}

		}
		if LibrarySolved(*book) == LibraryTomeCount {
			book.Library.Completed = true
			book.Library.CompletedTick = w.tick
			book.ClaimedBy = owner.ID
			owner.Stats.SpectralShots = true
		}
	}
}

func (w *World) hitLibraryTome(shot Projectile, start, end position, wallFraction float64) bool {
	if shot.MobOwned {
		return false
	}
	for i := range w.statues {
		book := &w.statues[i]
		if !IsLibraryBook(book.ID) || book.Library.Completed || book.Library.Owner != shot.Owner || book.Library.StartedTick == 0 || LibraryRecalling(*book, w.tick) {
			continue
		}
		nearest, first := -1, wallFraction
		for tome, at := range book.Library.Positions {
			if book.Library.Bound&(1<<tome) != 0 {
				continue
			}
			radius := LibraryTomeRadius
			if t, hit := eatingContact(start, end, shot.Size, position{at.X - radius, at.Y - radius}, position{at.X - radius, at.Y - radius}, radius*2); hit && t < first {
				nearest, first = tome, t
			}
		}
		if nearest >= 0 {
			w.strikeLibraryTome(book, nearest, physics.Vector{X: shot.Direction.X, Y: shot.Direction.Y})
			return true
		}
	}
	return false
}

func LibraryTomeStep(arena physics.Rect, wards [LibraryTomeCount]physics.Vector, shelves []physics.Rect, tome int, at, velocity physics.Vector) (physics.Vector, physics.Vector, bool, bool) {
	at, velocity, bounced := libraryShelfMotion(at, velocity, shelves)
	if at.X < arena.X {
		at.X = 2*arena.X - at.X
		velocity.X = math.Abs(velocity.X) * .9
		bounced = true
	}
	if at.X > arena.X+arena.W {
		at.X = 2*(arena.X+arena.W) - at.X
		velocity.X = -math.Abs(velocity.X) * .9
		bounced = true
	}
	if at.Y < arena.Y {
		at.Y = 2*arena.Y - at.Y
		velocity.Y = math.Abs(velocity.Y) * .9
		bounced = true
	}
	if at.Y > arena.Y+arena.H {
		at.Y = 2*(arena.Y+arena.H) - at.Y
		velocity.Y = -math.Abs(velocity.Y) * .9
		bounced = true
	}
	for ward, target := range wards {
		dx, dy := at.X-target.X, at.Y-target.Y
		distance := math.Hypot(dx, dy)
		if distance > LibraryWardRadius {
			continue
		}
		if ward == tome {
			return at, physics.Vector{}, bounced, true
		}
		if distance > .001 {
			nx, ny := dx/distance, dy/distance
			inward := velocity.X*nx + velocity.Y*ny
			if inward < 0 {
				velocity.X -= 2 * inward * nx
				velocity.Y -= 2 * inward * ny
				bounced = true
			}
			at.X, at.Y = target.X+nx*(LibraryWardRadius+1), target.Y+ny*(LibraryWardRadius+1)
		}
	}
	damping := math.Exp(-LibraryDrag * parameters.StepSeconds)
	velocity.X *= damping
	velocity.Y *= damping
	if math.Hypot(velocity.X, velocity.Y) < 8 {
		velocity = physics.Vector{}
	}
	return at, velocity, bounced, false
}

func libraryShelfMotion(at, velocity physics.Vector, shelves []physics.Rect) (physics.Vector, physics.Vector, bool) {
	remaining, bounced := parameters.StepSeconds, false
	for contact := 0; contact < 4 && remaining > 1e-9; contact++ {
		delta := physics.Vector{X: velocity.X * remaining, Y: velocity.Y * remaining}
		body := physics.Square(at.X-LibraryTomeRadius, at.Y-LibraryTomeRadius, LibraryTomeRadius*2)
		nearest, first := -1, 1.0
		for i, shelf := range shelves {
			if fraction, hit := body.CollisionFraction(delta, shelf, physics.Vector{}); hit && fraction < first {
				nearest, first = i, fraction
			}
		}
		at.X += delta.X * first
		at.Y += delta.Y * first
		if nearest < 0 {
			break
		}
		bounds := shelves[nearest].Expanded(LibraryTomeRadius)
		if math.Abs(at.X-bounds.X) < .001 && velocity.X > 0 || math.Abs(at.X-bounds.X-bounds.W) < .001 && velocity.X < 0 {
			velocity.X *= -.9
			at.X += math.Copysign(.01, velocity.X)
		}
		if math.Abs(at.Y-bounds.Y) < .001 && velocity.Y > 0 || math.Abs(at.Y-bounds.Y-bounds.H) < .001 && velocity.Y < 0 {
			velocity.Y *= -.9
			at.Y += math.Copysign(.01, velocity.Y)
		}
		remaining *= 1 - first
		bounced = true
	}
	return at, velocity, bounced
}

func LibraryShotPreview(book StatueState, tome int, start, direction physics.Vector) ([]physics.Vector, bool) {
	points := []physics.Vector{start}
	arena := LibraryArena(book)
	shelves := LibraryShelves(book)
	wards := [LibraryTomeCount]physics.Vector{}
	for ward := range wards {
		wards[ward] = LibraryWard(book, ward)
	}
	velocity := physics.Vector{X: direction.X * LibraryStrikeSpeed, Y: direction.Y * LibraryStrikeSpeed}
	at := start
	for step := 0; step < 7*parameters.TickRate; step++ {
		var bounced, bound bool
		at, velocity, bounced, bound = LibraryTomeStep(arena, wards, shelves, tome, at, velocity)
		if step%3 == 0 || bounced || bound {
			points = append(points, at)
		}
		if bound {
			points = append(points, wards[tome])
			return points, true
		}
		if velocity == (physics.Vector{}) {
			break
		}
	}
	points = append(points, at)
	return points, false
}
