package worldmap

import (
	"fmt"
	"math"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"
	mapdata "2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type libraryRenderFrame struct {
	previous                  []simulation.StatueState
	previousTick, currentTick uint64
	alpha, clock              float64
	player                    simulation.Player
	aimBook                   string
	aimTome                   int
	aimBinds                  bool
}

var libraryColors = [simulation.LibraryTomeCount]rl.Color{rl.NewColor(105, 211, 234, 255), rl.NewColor(239, 187, 95, 255), rl.NewColor(195, 141, 240, 255)}
var libraryNames = [simulation.LibraryTomeCount]string{"MOON", "SUN", "STAR"}

func (m *Map) PrepareLibraryFrame(previous []simulation.StatueState, previousTick, tick uint64, alpha float64, player simulation.Player) {
	alpha = max(0, min(1, alpha))
	clock := float64(tick)
	if previousTick != 0 && tick >= previousTick {
		clock = float64(previousTick) + float64(tick-previousTick)*alpha
	}
	m.libraryFrame = libraryRenderFrame{previous: previous, previousTick: previousTick, currentTick: tick, alpha: alpha, clock: clock, player: player, aimTome: -1}
}

func (m *Map) libraryPose(book simulation.StatueState, tome int) physics.Vector {
	at := book.Library.Positions[tome]
	for _, before := range m.libraryFrame.previous {
		if before.ID != book.ID || before.Library.StartedTick != book.Library.StartedTick {
			continue
		}
		if m.libraryFrame.currentTick < m.libraryFrame.previousTick {
			break
		}
		from := before.Library.Positions[tome]
		at.X = from.X + (at.X-from.X)*m.libraryFrame.alpha
		at.Y = from.Y + (at.Y-from.Y)*m.libraryFrame.alpha
		break
	}
	if book.Library.Bound&(1<<tome) != 0 && book.Library.BoundTicks[tome] != 0 {
		t := max(0, min(1, (m.libraryFrame.clock-float64(book.Library.BoundTicks[tome]))/parameters.TickRate/.55))
		t = t * t * (3 - 2*t)
		ward := simulation.LibraryWard(book, tome)
		at.X += (ward.X - at.X) * t
		at.Y += (ward.Y - at.Y) * t
	}
	return at
}

func librarySymbol(center rl.Vector2, kind int, radius float32, tint rl.Color) {
	switch kind {
	case 0:
		for i := 0; i < 20; i++ {
			a, b := float64(i)*math.Pi/20, float64(i+1)*math.Pi/20
			for _, width := range []float32{.9, .3} {
				from := rl.NewVector2(center.X-float32(math.Sin(a))*radius*width, center.Y-float32(math.Cos(a))*radius)
				to := rl.NewVector2(center.X-float32(math.Sin(b))*radius*width, center.Y-float32(math.Cos(b))*radius)
				rl.DrawLineEx(from, to, 2, tint)
			}
		}

	case 1:
		rl.DrawCircleLinesV(center, radius*.52, tint)
		for i := 0; i < 8; i++ {
			a := float64(i) * math.Pi / 4
			from := rl.NewVector2(center.X+float32(math.Cos(a))*radius*.75, center.Y+float32(math.Sin(a))*radius*.75)
			to := rl.NewVector2(center.X+float32(math.Cos(a))*radius, center.Y+float32(math.Sin(a))*radius)
			rl.DrawLineEx(from, to, 2, tint)
		}
	case 2:
		for i := 0; i < 10; i++ {
			a, b := float64(i)*math.Pi/5-math.Pi/2, float64(i+1)*math.Pi/5-math.Pi/2
			ra, rb := radius, radius*.43
			if i%2 == 1 {
				ra, rb = rb, ra
			}
			rl.DrawLineEx(rl.NewVector2(center.X+float32(math.Cos(a))*ra, center.Y+float32(math.Sin(a))*ra), rl.NewVector2(center.X+float32(math.Cos(b))*rb, center.Y+float32(math.Sin(b))*rb), 2, tint)
		}
	}
}

func (m *Map) drawHauntedLibrary(snapshot simulation.Snapshot, local simulation.Player) {
	if m.floor != mapdata.FirstFloor {
		return
	}
	seconds := m.libraryFrame.clock / parameters.TickRate
	for _, book := range snapshot.Statues {
		if !simulation.IsLibraryBook(book.ID) || m.world.SecretHidden(book.X, book.Y) {
			continue
		}
		room, ok := m.world.Room(simulation.LibraryRoom(book.ID))
		if !ok {
			continue
		}
		drawLibraryShelves(m.world, room)
		drawLibraryStacks(room)
		arena := simulation.LibraryArena(book)
		rail := rl.NewRectangle(float32(arena.X), float32(arena.Y), float32(arena.W), float32(arena.H))
		rl.DrawRectangleLinesEx(rail, 1, rl.NewColor(81, 69, 92, 255))
		rl.DrawRectangleLinesEx(rl.NewRectangle(rail.X-8, rail.Y-8, rail.Width+16, rail.Height+16), 1, rl.NewColor(47, 42, 55, 255))
		for i := 0; i < simulation.LibraryTomeCount; i++ {
			goal := simulation.LibraryWard(book, i)
			at := rl.NewVector2(float32(goal.X), float32(goal.Y))
			tint := libraryColors[i]
			lit := book.Library.Bound&(1<<i) != 0
			strength := float32(.12)
			if lit {
				strength = .3
			}
			rl.DrawCircleGradient(at, 65, rl.Fade(tint, strength), rl.Blank)
			rl.DrawPoly(at, 8, 37, 22.5, rl.NewColor(29, 28, 38, 255))
			rl.DrawPolyLines(at, 8, 37, 22.5, rl.Fade(tint, .75))
			rl.DrawCircleLinesV(at, 28, rl.Fade(tint, .35))
			librarySymbol(at, i, 17, rl.Fade(tint, .8))
			label := libraryNames[i]
			width := primitives.MeasureText(label, 12)
			primitives.DrawTextAt(label, rl.NewVector2(at.X-float32(width)/2, at.Y+43), 12, rl.Fade(tint, .8))
			if lit {
				for p := 0; p < 7; p++ {
					angle := seconds*.6 + float64(p)*math.Pi*2/7
					rl.DrawCircleV(rl.NewVector2(at.X+float32(math.Cos(angle))*31, at.Y+float32(math.Sin(angle))*31), 1.5, tint)
				}
			}
		}
		drawLibraryLectern(book, seconds)
		if book.Library.StartedTick != 0 {
			m.drawLibraryAim(book)
			for tome := 0; tome < simulation.LibraryTomeCount; tome++ {
				at := m.libraryPose(book, tome)
				if !m.Exploration.EntityVisible(at.X, at.Y, 0) {
					continue
				}
				drawLibraryTome(book, tome, at, m.libraryFrame.clock)
			}
		}
		if book.Library.CompletedTick != 0 {
			t := (m.libraryFrame.clock - float64(book.Library.CompletedTick)) / parameters.TickRate
			if t >= 0 && t < 3 {
				center := rl.NewVector2(float32(book.X), float32(book.Y))
				rl.DrawRing(center, float32(40+t*95), float32(43+t*95), 0, 360, 72, rl.Fade(rl.Gold, float32(1-t/3)))
				for spark := 0; spark < 30; spark++ {
					angle := float64(spark) * 2.399
					distance := (35 + float64(spark%7)*14) * t
					p := rl.NewVector2(center.X+float32(math.Cos(angle)*distance), center.Y+float32(math.Sin(angle)*distance-t*24))
					rl.DrawRectanglePro(rl.NewRectangle(p.X, p.Y, 3, 7), rl.NewVector2(1.5, 3.5), float32(angle*180/math.Pi), rl.Fade(libraryColors[spark%3], float32(1-t/3)))
				}
			}
		}
		closeEnough := book.Library.StartedTick == 0 || math.Hypot(book.X-local.X-parameters.PlayerSize/2, book.Y-local.Y-parameters.PlayerSize/2) < 55
		if !local.Dead && book.Near(local.X, local.Y, snapshot.OpenSecrets) && closeEnough {
			label := "[E] WAKE THE ENCHANTED TOMES"
			if book.Library.StartedTick != 0 {
				label = "[E] RECALL UNBOUND TOMES"
			}
			if book.Library.Completed {
				label = "THE LIBRARY REMEMBERS YOU"
			}
			w := float32(primitives.MeasureText(label, 14))
			pos := rl.NewVector2(float32(book.X)-w/2, float32(book.Y)+53)
			rl.DrawRectangleRec(rl.NewRectangle(pos.X-6, pos.Y-4, w+12, 23), rl.NewColor(18, 16, 25, 225))
			primitives.DrawTextAt(label, pos, 14, rl.NewColor(186, 172, 213, 255))
		}
	}
}

func drawLibraryTome(book simulation.StatueState, tome int, at physics.Vector, clock float64) {
	tint := libraryColors[tome]
	bound := book.Library.Bound&(1<<tome) != 0
	age := (clock - float64(book.Library.StartedTick)) / parameters.TickRate
	bob := math.Sin(age*2+float64(tome)*2) * 2.2
	speed := math.Hypot(book.Library.Velocities[tome].X, book.Library.Velocities[tome].Y)
	center := rl.NewVector2(float32(at.X), float32(at.Y))
	if speed > 15 && !bound {
		vx, vy := book.Library.Velocities[tome].X/speed, book.Library.Velocities[tome].Y/speed
		for i := 1; i <= 7; i++ {
			p := rl.NewVector2(center.X-float32(vx)*float32(i*9), center.Y-float32(vy)*float32(i*9))
			rl.DrawCircleGradient(p, 13, rl.Fade(tint, .2*(1-float32(i)/8)), rl.Blank)
		}
	}
	rl.DrawEllipse(int32(center.X), int32(center.Y+13), 17, 7, rl.Fade(rl.Black, .45))
	center.Y += float32(bob)
	flash := 0.0
	if book.Library.HitTicks[tome] != 0 {
		flash = max(0, min(1, 1-(clock-float64(book.Library.HitTicks[tome]))/parameters.TickRate/.25))
	}
	rl.DrawCircleGradient(center, 40+float32(flash)*18, rl.Fade(tint, .12+float32(flash)*.2), rl.Blank)
	rl.DrawRectangleRec(rl.NewRectangle(center.X-17, center.Y-21, 34, 42), rl.NewColor(26, 22, 31, 255))
	rl.DrawRectangleRec(rl.NewRectangle(center.X-14, center.Y-20, 29, 37), rl.NewColor(tint.R/3, tint.G/3, tint.B/3, 255))
	rl.DrawRectangleLinesEx(rl.NewRectangle(center.X-14, center.Y-20, 29, 37), 1, tint)
	rl.DrawRectangleRec(rl.NewRectangle(center.X-17, center.Y-20, 5, 39), rl.NewColor(58, 44, 50, 255))
	for line := 0; line < 3; line++ {
		rl.DrawLineEx(rl.NewVector2(center.X-11, center.Y+18+float32(line)), rl.NewVector2(center.X+14, center.Y+18+float32(line)), 1, rl.NewColor(185, 175, 156, 255))
	}
	librarySymbol(rl.NewVector2(center.X+1, center.Y-2), tome, 9, tint)
	rl.DrawLineEx(rl.NewVector2(center.X+9, center.Y+16), rl.NewVector2(center.X+9, center.Y+28), 3, tint)
	if bound {
		t := (clock - float64(book.Library.BoundTicks[tome])) / parameters.TickRate
		if t >= 0 && t < .7 {
			rl.DrawRing(center, float32(20+t*55), float32(22+t*55), 0, 360, 48, rl.Fade(tint, float32(1-t/.7)))
		}
	}
	bounce := (clock - float64(book.Library.BounceTicks[tome])) / parameters.TickRate
	if book.Library.BounceTicks[tome] != 0 && bounce >= 0 && bounce < .3 {
		rl.DrawCircleLinesV(center, float32(18+bounce*35), rl.Fade(tint, float32(1-bounce/.3)))
	}
}

func drawLibraryLectern(book simulation.StatueState, seconds float64) {
	x, y := float32(book.X), float32(book.Y)
	rug := rl.NewRectangle(x-78, y-64, 156, 150)
	rl.DrawRectangleRec(rug, rl.NewColor(43, 34, 52, 255))
	rl.DrawRectangleLinesEx(rug, 2, rl.NewColor(106, 81, 89, 255))
	rl.DrawRectangleLinesEx(rl.NewRectangle(rug.X+7, rug.Y+7, rug.Width-14, rug.Height-14), 1, rl.NewColor(76, 60, 82, 255))
	for _, side := range []float32{-1, 1} {
		at := rl.NewVector2(x+side*58, y+17)
		rl.DrawCircleGradient(at, 64, rl.Fade(rl.NewColor(223, 158, 95, 255), .12), rl.Blank)
		rl.DrawEllipse(int32(at.X), int32(at.Y+6), 8, 4, rl.NewColor(89, 68, 44, 255))
		rl.DrawRectangleRec(rl.NewRectangle(at.X-3, at.Y-8, 6, 14), rl.NewColor(196, 181, 153, 255))
		rl.DrawCircleV(rl.NewVector2(at.X, at.Y-10), 2.5+float32(math.Sin(seconds*6+float64(side)))*.35, rl.Gold)
	}
	rl.DrawEllipse(int32(x), int32(y+20), 27, 10, rl.Fade(rl.Black, .5))
	rl.DrawRectangleRec(rl.NewRectangle(x-15, y-2, 30, 23), rl.NewColor(59, 45, 42, 255))
	rl.DrawRectangleRec(rl.NewRectangle(x-25, y-8, 50, 8), rl.NewColor(113, 80, 57, 255))
	bob := float32(math.Sin(seconds*1.5)) * 2
	rl.DrawRectangleRec(rl.NewRectangle(x-23, y-33+bob, 46, 26), rl.NewColor(69, 53, 94, 255))
	rl.DrawRectangleRec(rl.NewRectangle(x-20, y-31+bob, 19, 21), rl.NewColor(197, 191, 177, 255))
	rl.DrawRectangleRec(rl.NewRectangle(x+1, y-31+bob, 19, 21), rl.NewColor(184, 179, 171, 255))
	for line := 0; line < 4; line++ {
		for _, side := range []float32{-1, 1} {
			rl.DrawLineEx(rl.NewVector2(x+side*5, y-27+bob+float32(line*4)), rl.NewVector2(x+side*16, y-27+bob+float32(line*4)), 1, rl.NewColor(112, 94, 122, 255))
		}
	}
	if book.Library.StartedTick != 0 && !book.Library.Completed {
		ghost := rl.NewVector2(x, y-66+float32(math.Sin(seconds*1.8))*3)
		rl.DrawCircleGradient(ghost, 29, rl.Fade(rl.NewColor(175, 155, 225, 255), .18), rl.Blank)
		rl.DrawCircleV(ghost, 10, rl.Fade(rl.NewColor(162, 148, 214, 255), .5))
		rl.DrawTriangle(rl.NewVector2(ghost.X-9, ghost.Y+3), rl.NewVector2(ghost.X, ghost.Y+28), rl.NewVector2(ghost.X+9, ghost.Y+3), rl.Fade(rl.NewColor(162, 148, 214, 255), .3))
	}
}

func drawLibraryShelves(world *mapdata.Map, room mapdata.Room) {
	for side := 0; side < 4; side++ {
		count := room.Area.Width - 2
		if side >= 2 {
			count = room.Area.Height - 2
		}
		for i := 1; i < count; i += 2 {
			cell := room.WorldCell(mapdata.At(i, 0))
			outside := mapdata.At(cell.X, cell.Y-1)
			vertical := side >= 2
			switch side {
			case 1:
				cell = room.WorldCell(mapdata.At(i, room.Area.Height-1))
				outside = mapdata.At(cell.X, cell.Y+1)
			case 2:
				cell = room.WorldCell(mapdata.At(0, i))
				outside = mapdata.At(cell.X-1, cell.Y)
			case 3:
				cell = room.WorldCell(mapdata.At(room.Area.Width-1, i))
				outside = mapdata.At(cell.X+1, cell.Y)
			}
			next := outside
			if vertical {
				next.Y++
			} else {
				next.X++
			}
			if world.Cell(outside.X, outside.Y) != 0 || world.Cell(next.X, next.Y) != 0 {
				continue
			}
			rect := rl.NewRectangle(float32(cell.X*parameters.Tile)+5, float32(cell.Y*parameters.Tile)+5, 70, 30)
			if vertical {
				rect.Width, rect.Height = 30, 70
			}
			drawLibraryShelf(rect, vertical, i+side*7)
		}
	}
}
func drawLibraryShelf(rect rl.Rectangle, vertical bool, seed int) {
	rl.DrawRectangleRec(rl.NewRectangle(rect.X+3, rect.Y+4, rect.Width+3, rect.Height+4), rl.Fade(rl.Black, .4))
	rl.DrawRectangleRec(rect, rl.NewColor(57, 43, 39, 255))
	rl.DrawRectangleLinesEx(rect, 2, rl.NewColor(106, 77, 54, 255))
	point := func(along, depth float32) rl.Vector2 {
		if vertical {
			return rl.NewVector2(rect.X+depth, rect.Y+along)
		}
		return rl.NewVector2(rect.X+along, rect.Y+depth)
	}
	colors := []rl.Color{rl.NewColor(98, 67, 72, 255), rl.NewColor(66, 83, 77, 255), rl.NewColor(79, 72, 103, 255), rl.NewColor(118, 93, 63, 255), rl.NewColor(59, 75, 100, 255)}
	length := rect.Width
	if vertical {
		length = rect.Height
	}
	for row := 0; row < 2; row++ {
		depth := float32(4 + row*13)
		for b := 0; b < int((length-8)/7); b++ {
			height := float32(7 + (b*3+seed)%5)
			at := point(float32(4+b*7), depth+11-height)
			book := rl.NewRectangle(at.X, at.Y, 5, height)
			if vertical {
				book.Width, book.Height = height, 5
			}
			rl.DrawRectangleRec(book, colors[(b+seed+row)%len(colors)])
			rl.DrawLineEx(point(float32(5+b*7), depth+9), point(float32(7+b*7), depth+9), 1, rl.NewColor(174, 144, 90, 255))
		}
		rl.DrawLineEx(point(2, depth+12), point(length-2, depth+12), 2, rl.NewColor(125, 92, 65, 255))
	}
}

func drawLibraryStacks(room mapdata.Room) {
	for i, prop := range room.Props {
		if prop.Kind != mapdata.BookshelfProp {
			continue
		}
		bounds := room.WorldArea(prop.Area).Bounds()
		rect := rl.NewRectangle(float32(bounds.X), float32(bounds.Y), float32(bounds.W), float32(bounds.H))
		rl.DrawRectangleRec(rl.NewRectangle(rect.X+5, rect.Y+7, rect.Width+3, rect.Height+3), rl.Fade(rl.Black, .5))
		rl.DrawRectangleRec(rect, rl.NewColor(63, 46, 42, 255))
		rl.DrawRectangleLinesEx(rect, 2, rl.NewColor(141, 109, 76, 255))
		vertical, length := rect.Height > rect.Width, rect.Width
		if vertical {
			length = rect.Height
		}
		for along := float32(0); along < length; along += 80 {
			segment := rl.NewRectangle(rect.X+along+5, rect.Y+5, min(70, length-along-10), 30)
			if vertical {
				segment = rl.NewRectangle(rect.X+5, rect.Y+along+5, 30, min(70, length-along-10))
			}
			drawLibraryShelf(segment, vertical, i*7+int(along/80))
		}
		rl.DrawRectangleLinesEx(rl.NewRectangle(rect.X-2, rect.Y-2, rect.Width+4, rect.Height+4), 1, rl.NewColor(113, 94, 138, 255))
	}
}

func (m *Map) drawLibraryAim(book simulation.StatueState) {
	p := m.libraryFrame.player
	if p.Dead || book.Library.Completed || book.Library.Owner != p.ID || simulation.LibraryRecalling(book, m.libraryFrame.currentTick) {
		return
	}
	dx, dy := p.Combat.Facing.X, p.Combat.Facing.Y
	length := math.Hypot(dx, dy)
	if length < .01 {
		return
	}
	dx /= length
	dy /= length
	origin := physics.Vector{X: p.X + parameters.PlayerSize/2, Y: p.Y + parameters.PlayerSize/2}
	selected, distance := -1, 600.0
	for tome := 0; tome < simulation.LibraryTomeCount; tome++ {
		if book.Library.Bound&(1<<tome) != 0 {
			continue
		}
		at := m.libraryPose(book, tome)
		ax, ay := at.X-origin.X, at.Y-origin.Y
		along := ax*dx + ay*dy
		across := math.Abs(ax*dy - ay*dx)
		if along > 0 && along < distance && across < simulation.LibraryTomeRadius+5 {
			size := p.Stats.ShotSize()
			if _, blocked := m.world.MovementLimit(origin.X-size/2, origin.Y-size/2, dx*max(0, along-simulation.LibraryTomeRadius), dy*max(0, along-simulation.LibraryTomeRadius), size, size); blocked {
				continue
			}
			selected, distance = tome, along
		}
	}
	if selected < 0 {
		return
	}
	at := m.libraryPose(book, selected)
	tint := libraryColors[selected]
	center := rl.NewVector2(float32(at.X), float32(at.Y))
	rl.DrawRing(center, 26, 28, 0, 360, 40, tint)
	points, binds := simulation.LibraryShotPreview(book, selected, at, physics.Vector{X: dx, Y: dy})
	m.libraryFrame.aimBook, m.libraryFrame.aimTome, m.libraryFrame.aimBinds = book.ID, selected, binds
	for i := 1; i < len(points); i++ {
		if i%2 == 0 {
			continue
		}
		a, b := points[i-1], points[i]
		fade := float32(.65) * (1 - float32(i)/float32(len(points))*.5)
		rl.DrawLineEx(rl.NewVector2(float32(a.X), float32(a.Y)), rl.NewVector2(float32(b.X), float32(b.Y)), 2, rl.Fade(tint, fade))
	}
	end := points[len(points)-1]
	pos := rl.NewVector2(float32(end.X), float32(end.Y))
	if binds {
		rl.DrawCircleGradient(pos, 50, rl.Fade(tint, .25), rl.Blank)
		rl.DrawCircleLinesV(pos, 39, tint)
	} else {
		rl.DrawPolyLines(pos, 4, 8, 0, rl.Fade(tint, .65))
	}
}

func (m *Map) DrawLibraryHUD(ctx primitives.Context, snapshot simulation.Snapshot, local simulation.Player) {
	if local.Dead {
		return
	}
	for _, book := range snapshot.Statues {
		if !simulation.IsLibraryBook(book.ID) || m.world.SecretHidden(book.X, book.Y) || !m.world.RoomContains(simulation.LibraryRoom(book.ID), physics.Square(local.X, local.Y, parameters.PlayerSize)) {
			continue
		}
		title := "THE LIBRARIAN'S SPELL BILLIARDS"
		hint := "Match each tome to its rune alcove."
		detail := "Bank around the shelves / cast to launch / E nearby to nudge"
		if book.Library.StartedTick == 0 {
			detail = "[E] at the lectern to begin / no timer / no MP needed"
		}
		if book.Library.StartedTick != 0 {
			title = fmt.Sprintf("THE LIBRARIAN'S SPELL BILLIARDS   %d / 3", simulation.LibrarySolved(book))
			hint = "MOON to MOON / SUN to SUN / STAR to STAR"
		}
		if m.libraryFrame.aimBook == book.ID && m.libraryFrame.aimTome >= 0 {
			if m.libraryFrame.aimBinds {
				detail = libraryNames[m.libraryFrame.aimTome] + " / BINDING SHOT LINED UP"
			} else {
				detail = libraryNames[m.libraryFrame.aimTome] + " / aim a direct shot or bounce off a rail"
			}
		}
		if book.Library.Completed {
			title = "SPECTRAL SHOTS UNLOCKED"
			hint = "Every shot casts an echo / MP cost +25%"
			detail = "The librarian nods. A little order, at last."
		}
		if book.Library.StartedTick != 0 && book.Library.Owner != local.ID && !book.Library.Completed {
			detail = "Another explorer is restoring these tomes."
		}
		width := max(primitives.MeasureText(title, 18), max(primitives.MeasureText(hint, 16), primitives.MeasureText(detail, 14))) + 30
		x := (ctx.WindowWidth - width) / 2
		rl.DrawRectangle(x, 18, width, 86, rl.NewColor(23, 19, 30, 235))
		rl.DrawRectangle(x, 18, width, 2, rl.NewColor(160, 138, 193, 255))
		primitives.DrawText(title, x+(width-primitives.MeasureText(title, 18))/2, 30, 18, rl.NewColor(218, 196, 149, 255))
		primitives.DrawText(hint, x+(width-primitives.MeasureText(hint, 16))/2, 56, 16, rl.RayWhite)
		primitives.DrawText(detail, x+(width-primitives.MeasureText(detail, 14))/2, 81, 14, rl.NewColor(180, 174, 193, 255))
	}
}
