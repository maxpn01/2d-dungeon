package game

import (
	"image/color"
	"math"
	"time"

	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type feedbackActor struct {
	x, y, size, hp, xp float64
	level              int
	tint               color.RGBA
	ai, dead           bool
}

type feedbackEffect struct {
	started    time.Time
	x, y, size float32
	tint       color.RGBA
	kind       string
}

type combatFeedback struct {
	actors  map[string]feedbackActor
	drops   map[string]bool
	effects []feedbackEffect
}

func (f *combatFeedback) observe(now time.Time, actors map[string]feedbackActor, fruits []protocol.Fruit) {
	drops := make(map[string]bool)
	playerColors := make(map[string]color.RGBA)
	for id, actor := range actors {
		if !actor.ai {
			playerColors[id] = actor.tint
		}
	}
	emit := func(x, y, size float64, tint color.RGBA, kind string) {
		f.effects = append(f.effects, feedbackEffect{now, float32(x), float32(y), float32(size), tint, kind})
	}
	if f.actors != nil {
		for id, actor := range actors {
			old, existed := f.actors[id]
			if existed && !old.dead && !actor.dead && (actor.hp > old.hp || actor.level > old.level || actor.xp > old.xp) {
				emit(actor.x+actor.size/2, actor.y+actor.size/2, float64(actorRenderSize(actor.size, actor.hp)), actor.tint, "eat")
			}
		}
		for id, old := range f.actors {
			if _, exists := actors[id]; !exists && old.ai {
				emit(old.x+old.size/2, old.y+old.size/2, float64(actorRenderSize(old.size, old.hp)), old.tint, "prey")
			}
		}
	}
	for _, fruit := range fruits {
		if fruit.Source == "" {
			continue
		}
		drops[fruit.ID] = true
		if f.drops != nil && !f.drops[fruit.ID] {
			emit(fruit.X+fruit.Size/2, fruit.Y+fruit.Size/2, fruit.Size, massColor(fruit.Source, playerColors), "mass")
		}
	}
	f.actors, f.drops = actors, drops
	if len(f.effects) > 128 {
		f.effects = f.effects[len(f.effects)-128:]
	}
}

func (f *combatFeedback) snapshot(now time.Time, s simulation.Snapshot) {
	actors := make(map[string]feedbackActor)
	for _, p := range s.Players {
		actors[p.ID] = feedbackActor{p.X, p.Y, simulation.PlayerSize, p.Stats.HP, p.Stats.XP, p.Stats.Level, characterColor(p.Character), false, p.Dead}
	}
	for _, a := range s.AIs {
		actors[a.ID] = feedbackActor{a.X, a.Y, simulation.AISize, a.Stats.HP, a.Stats.XP, a.Stats.Level, rl.Green, true, false}
	}
	fruits := make([]protocol.Fruit, 0, len(s.Fruits))
	for _, fruit := range s.Fruits {
		fruits = append(fruits, protocol.Fruit{ID: fruit.ID, X: fruit.X, Y: fruit.Y, Size: fruit.Size, Source: fruit.Source})
	}
	f.observe(now, actors, fruits)
}

func (f *combatFeedback) multiplayer(now time.Time, players []protocol.Player, ais []protocol.AI, fruits []protocol.Fruit) {
	actors := make(map[string]feedbackActor)
	for _, p := range players {
		actors[p.ID] = feedbackActor{p.X, p.Y, p.Size, p.Stats.HP, p.Stats.XP, p.Stats.Level, characterColor(p.Character), false, p.Dead}
	}
	for _, a := range ais {
		actors[a.ID] = feedbackActor{a.X, a.Y, a.Size, a.Stats.HP, a.Stats.XP, a.Stats.Level, rl.Green, true, false}
	}
	f.observe(now, actors, fruits)
}

func (f *combatFeedback) draw(now time.Time, exploration *mapExploration) {
	active := f.effects[:0]
	for _, e := range f.effects {
		p := float32(max(0, now.Sub(e.started).Seconds()) / 0.35)
		if p >= 1 {
			continue
		}
		active = append(active, e)
		if !exploration.entityVisible(float64(e.x), float64(e.y), 0) {
			continue
		}
		tint := rl.Fade(e.tint, 1-p)
		center := rl.NewVector2(e.x, e.y)
		switch e.kind {
		case "eat":
			radius := e.size*0.7 + 12*p
			rl.DrawRing(center, radius, radius+2*(1-p), 0, 360, 36, tint)
			for i := range 4 {
				angle := float64(i)*math.Pi/2 + math.Pi/4
				radius := (e.size*0.8 + 12) * (1 - p)
				rl.DrawRectangleRec(rl.NewRectangle(e.x+float32(math.Cos(angle))*radius-2, e.y+float32(math.Sin(angle))*radius-2, 4, 4), tint)
			}
		case "prey":
			side := e.size * (1 - p) * (1 - p)
			rl.DrawRectangleRec(rl.NewRectangle(e.x-side/2, e.y-side/2, side, side), tint)
		case "mass":
			for i := range 6 {
				angle := float64(i) * math.Pi / 3
				radius := e.size/2 + 20*p
				side := 4 * (1 - p)
				rl.DrawRectangleRec(rl.NewRectangle(e.x+float32(math.Cos(angle))*radius-side/2, e.y+float32(math.Sin(angle))*radius-side/2, side, side), tint)
			}
		}
	}
	f.effects = active
}
