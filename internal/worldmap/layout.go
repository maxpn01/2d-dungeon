package worldmap

import (
	"slices"
	"time"
)

type RoomRole string

const (
	Entrance        RoomRole = "entrance"
	Private         RoomRole = "private"
	ExplorationRoom RoomRole = "exploration"
	Refuge          RoomRole = "refuge"
	BossArena       RoomRole = "boss-arena"
	PortalSanctum   RoomRole = "portal-sanctum"
)

type Floor int

const (
	FirstFloor  Floor = 1
	SecondFloor Floor = 2
)

type PropKind string

const CoverProp PropKind = "cover"

type FeedingStyle string

const (
	TealGarden   FeedingStyle = "teal"
	VioletGarden FeedingStyle = "violet"
)

type Room struct {
	ID           string
	Name         string
	Area         CellRect
	Floor        Floor
	Roles        []RoomRole
	Props        []Prop
	Pools        []Pool
	FeedingAreas []FeedingArea
	Gates        []Gate
	Portals      []Portal
	Statues      []Statue
}

type Prop struct {
	ID   string
	Kind PropKind
	Area CellRect
}

func Cover(id string, area CellRect) Prop { return Prop{ID: id, Kind: CoverProp, Area: area} }

type Passage struct {
	ID    string
	From  Location
	To    Location
	Width int
}

type Pool struct {
	ID   string
	Area CellRect
}

type FoodSource struct {
	Capacity int
	Respawn  time.Duration
}

type Guard struct {
	PatrolRadius int
	Level        int
	Damage       float64
	AttackRate   float64
	Respawn      time.Duration
}

type FeedingArea struct {
	ID     string
	At     Cell
	Radius int
	Style  FeedingStyle
	Food   FoodSource
	Guard  *Guard
}

type Gate struct {
	ID    string
	At    Cell
	Width int
}

type Portal struct {
	ID          string
	At          Cell
	Destination string
}

type Statue struct {
	ID string
	At Cell
}

type Location struct {
	Room string
	At   Cell
}

type Layout struct {
	Rooms     []Room
	Passages  []Passage
	Entrances []Location
}

func (r Room) HasRole(role RoomRole) bool { return slices.Contains(r.Roles, role) }

func (l Layout) Room(id string) (Room, bool) {
	for _, room := range l.Rooms {
		if room.ID == id {
			return cloneRoom(room), true
		}
	}
	return Room{}, false
}

func (l *Layout) SetRoom(room Room) {
	l.Rooms = setByID(l.Rooms, room, func(r Room) string { return r.ID })
}
func (l *Layout) RemoveRoom(id string) bool {
	return removeByID(&l.Rooms, id, func(r Room) string { return r.ID })
}
func (l *Layout) SetPassage(passage Passage) {
	l.Passages = setByID(l.Passages, passage, func(p Passage) string { return p.ID })
}
func (l *Layout) RemovePassage(id string) bool {
	return removeByID(&l.Passages, id, func(p Passage) string { return p.ID })
}
func (r *Room) SetProp(prop Prop) {
	r.Props = setByID(r.Props, prop, func(p Prop) string { return p.ID })
}
func (r *Room) RemoveProp(id string) bool {
	return removeByID(&r.Props, id, func(p Prop) string { return p.ID })
}
func (r *Room) SetPool(pool Pool) {
	r.Pools = setByID(r.Pools, pool, func(p Pool) string { return p.ID })
}
func (r *Room) RemovePool(id string) bool {
	return removeByID(&r.Pools, id, func(p Pool) string { return p.ID })
}
func (r *Room) SetFeedingArea(area FeedingArea) {
	r.FeedingAreas = setByID(r.FeedingAreas, area, func(a FeedingArea) string { return a.ID })
}
func (r *Room) RemoveFeedingArea(id string) bool {
	return removeByID(&r.FeedingAreas, id, func(a FeedingArea) string { return a.ID })
}
func (r *Room) SetGate(gate Gate) {
	r.Gates = setByID(r.Gates, gate, func(g Gate) string { return g.ID })
}
func (r *Room) RemoveGate(id string) bool {
	return removeByID(&r.Gates, id, func(g Gate) string { return g.ID })
}
func (r *Room) SetPortal(portal Portal) {
	r.Portals = setByID(r.Portals, portal, func(p Portal) string { return p.ID })
}
func (r *Room) RemovePortal(id string) bool {
	return removeByID(&r.Portals, id, func(p Portal) string { return p.ID })
}

func (r *Room) SetStatue(statue Statue) {
	r.Statues = setByID(r.Statues, statue, func(s Statue) string { return s.ID })
}
func (r *Room) RemoveStatue(id string) bool {
	return removeByID(&r.Statues, id, func(s Statue) string { return s.ID })
}

func setByID[T any](items []T, value T, id func(T) string) []T {
	for i := range items {
		if id(items[i]) == id(value) {
			items[i] = value
			return items
		}
	}
	return append(items, value)
}

func removeByID[T any](items *[]T, key string, id func(T) string) bool {
	for i := range *items {
		if id((*items)[i]) == key {
			*items = slices.Delete(*items, i, i+1)
			return true
		}
	}
	return false
}
