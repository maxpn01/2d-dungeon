package worldmap

import (
	"fmt"
	"raylib-2d-sandbox/internal/physics"
	"time"
)

type FeedingPatch struct {
	ID, RoomID       string
	At               Cell
	Radius, Capacity int
	Respawn          time.Duration
	Style            FeedingStyle
}

type Den struct {
	ID, RoomID string
	At         Cell
	Patrol     CellRect
	Level      int
	HP         float64
	Respawn    time.Duration
}

type GateBounds struct{ Left, Right, Y int }
type PortalPoint struct {
	ID, RoomID, Destination string
	At                      Cell
	Floor                   Floor
}

func (p PortalPoint) Position() physics.Vector { return p.At.World() }

type Map struct {
	layout         Layout
	floors         map[[2]int]int
	cover          map[[2]int]bool
	passages       []CellRect
	pools          []Pool
	patches        []FeedingPatch
	dens           []Den
	gates          map[string]GateBounds
	portals        []PortalPoint
	collisionWorld physics.World
}

var defaultMap = mustBuild(DefaultLayout())

func Default() *Map { return defaultMap }
func mustBuild(layout Layout) *Map {
	m, err := New(layout)
	if err != nil {
		panic(err)
	}
	return m
}

func New(layout Layout) (*Map, error) {
	layout = cloneLayout(layout)
	if err := validateLayout(layout); err != nil {
		return nil, err
	}
	m := &Map{layout: layout, cover: make(map[[2]int]bool), gates: make(map[string]GateBounds)}
	for _, passage := range layout.Passages {
		area, err := layout.passageArea(passage)
		if err != nil {
			return nil, err
		}
		m.passages = append(m.passages, area)
	}
	for _, room := range layout.Rooms {
		for _, pool := range room.Pools {
			pool.Area = room.WorldArea(pool.Area)
			m.pools = append(m.pools, pool)
		}
		for _, area := range room.FeedingAreas {
			at := room.WorldCell(area.At)
			if area.Food.Capacity > 0 {
				m.patches = append(m.patches, FeedingPatch{ID: area.ID, RoomID: room.ID, At: at, Radius: area.Radius, Capacity: area.Food.Capacity, Respawn: area.Food.Respawn, Style: area.Style})
			}
			if area.Guard != nil {
				guard := area.Guard
				m.dens = append(m.dens, Den{ID: area.ID, RoomID: room.ID, At: at, Patrol: Around(at, guard.PatrolRadius), Level: guard.Level, HP: guard.HP, Respawn: guard.Respawn})
			}
		}
		for _, gate := range room.Gates {
			at := room.WorldCell(gate.At)
			m.gates[gate.ID] = GateBounds{Left: at.X * Tile, Right: (at.X + gate.Width) * Tile, Y: at.Y * Tile}
		}
		for _, portal := range room.Portals {
			m.portals = append(m.portals, PortalPoint{ID: portal.ID, RoomID: room.ID, Destination: portal.Destination, At: room.WorldCell(portal.At), Floor: room.Floor})
		}
	}
	m.floors = m.build()
	for _, room := range layout.Rooms {
		for _, prop := range room.Props {
			area := room.WorldArea(prop.Area)
			for y := area.Y; y < area.Bottom(); y++ {
				for x := area.X; x < area.Right(); x++ {
					if m.Cell(x, y) != 0 {
						m.cover[[2]int{x, y}] = true
					}
				}
			}
		}
	}
	m.collisionWorld = physics.World{CanOccupy: m.CanOccupy, MaxStep: Tile / 5}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Map) CanOccupy(body physics.Rect) bool  { return m.Walkable(body.X, body.Y, body.W, body.H) }
func (m *Map) Layout() Layout                    { return cloneLayout(m.layout) }
func (m *Map) Rooms() []Room                     { return cloneRooms(m.layout.Rooms) }
func (m *Map) Pools() []Pool                     { return append([]Pool(nil), m.pools...) }
func (m *Map) Patches() []FeedingPatch           { return append([]FeedingPatch(nil), m.patches...) }
func (m *Map) Dens() []Den                       { return append([]Den(nil), m.dens...) }
func (m *Map) Portals() []PortalPoint            { return append([]PortalPoint(nil), m.portals...) }
func (m *Map) Gate(id string) (GateBounds, bool) { gate, ok := m.gates[id]; return gate, ok }
func (m *Map) EntranceGate() GateBounds          { return m.gates[EntranceGate] }
func (m *Map) Portal(id string) (PortalPoint, bool) {
	for _, portal := range m.portals {
		if portal.ID == id {
			return portal, true
		}
	}
	return PortalPoint{}, false
}
func (m *Map) Room(id string) (Room, bool) { return m.layout.Room(id) }
func (m *Map) RoomWithRole(role RoomRole) (Room, bool) {
	for _, room := range m.layout.Rooms {
		if room.HasRole(role) {
			return cloneRoom(room), true
		}
	}
	return Room{}, false
}
func (m *Map) RoomAt(x, y float64) (Room, bool) {
	for _, room := range m.layout.Rooms {
		bounds := room.Bounds()
		if x >= bounds.X && x < bounds.X+bounds.W && y >= bounds.Y && y < bounds.Y+bounds.H {
			return cloneRoom(room), true
		}
	}
	return Room{}, false
}
func (m *Map) RoomContains(id string, body physics.Rect) bool {
	for _, room := range m.layout.Rooms {
		if room.ID == id {
			return room.Bounds().Contains(body)
		}
	}
	return false
}
func (m *Map) FloorAt(x, y float64) Floor {
	for _, room := range m.layout.Rooms {
		bounds := room.Bounds()
		if x >= bounds.X && x < bounds.X+bounds.W && y >= bounds.Y && y < bounds.Y+bounds.H {
			return room.Floor
		}
	}
	return FirstFloor
}

func cloneRoom(room Room) Room {
	room.Roles = append([]RoomRole(nil), room.Roles...)
	room.Props = append([]Prop(nil), room.Props...)
	room.Pools = append([]Pool(nil), room.Pools...)
	room.FeedingAreas = append([]FeedingArea(nil), room.FeedingAreas...)
	for i := range room.FeedingAreas {
		if room.FeedingAreas[i].Guard != nil {
			guard := *room.FeedingAreas[i].Guard
			room.FeedingAreas[i].Guard = &guard
		}
	}
	room.Gates = append([]Gate(nil), room.Gates...)
	room.Portals = append([]Portal(nil), room.Portals...)
	return room
}
func cloneRooms(rooms []Room) []Room {
	result := make([]Room, len(rooms))
	for i, room := range rooms {
		result[i] = cloneRoom(room)
	}
	return result
}
func cloneLayout(layout Layout) Layout {
	layout.Rooms = cloneRooms(layout.Rooms)
	layout.Passages = append([]Passage(nil), layout.Passages...)
	layout.Entrances = append([]Location(nil), layout.Entrances...)
	return layout
}

func (l Layout) resolve(location Location) (Cell, error) {
	if location.Room == "" {
		return location.At, nil
	}
	for _, room := range l.Rooms {
		if room.ID == location.Room {
			return room.WorldCell(location.At), nil
		}
	}
	return Cell{}, fmt.Errorf("unknown room %q in location", location.Room)
}
func (l Layout) passageArea(p Passage) (CellRect, error) {
	from, err := l.resolve(p.From)
	if err != nil {
		return CellRect{}, fmt.Errorf("passage %s: %w", p.ID, err)
	}
	to, err := l.resolve(p.To)
	if err != nil {
		return CellRect{}, fmt.Errorf("passage %s: %w", p.ID, err)
	}
	if from.X == to.X {
		return Cells(from.X, min(from.Y, to.Y), p.Width, max(from.Y, to.Y)-min(from.Y, to.Y)+1), nil
	}
	if from.Y == to.Y {
		return Cells(min(from.X, to.X), from.Y, max(from.X, to.X)-min(from.X, to.X)+1, p.Width), nil
	}
	return CellRect{}, fmt.Errorf("passage %s endpoints must align horizontally or vertically", p.ID)
}

func (m *Map) Passage(id string) (CellRect, bool) {
	for i, passage := range m.layout.Passages {
		if passage.ID == id {
			return m.passages[i], true
		}
	}
	return CellRect{}, false
}
