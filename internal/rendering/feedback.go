package rendering

import (
	"image/color"
	"math"
	"time"

	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type feedbackActor struct {
	x, y, size, hp, xp float64
	level              int
	tint               color.RGBA
	ai, dead, hidden   bool
	escaped            bool
}

type feedbackEffect struct {
	started    time.Time
	x, y, size float32
	tint       color.RGBA
	kind       string
}

type Feedback struct {
	spawns  map[string]time.Time
	actors  map[string]feedbackActor
	drops   map[string]bool
	effects []feedbackEffect
}

func (f *Feedback) observe(now time.Time, actors map[string]feedbackActor, fruits []protocol.Fruit) {
	if f.spawns == nil {
		f.spawns = make(map[string]time.Time)
	}
	for id, actor := range actors {
		old, existed := f.actors[id]
		if !actor.ai && !actor.dead && (!existed || old.dead) {
			f.spawns[id] = now
		}
	}
	for id, started := range f.spawns {
		actor, exists := actors[id]
		if !exists || actor.dead || now.Sub(started) >= materializeDuration {
			delete(f.spawns, id)
		}
	}
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
			if actor.hidden {
				continue
			}
			old, existed := f.actors[id]
			if actor.escaped && !actor.dead && (!existed || !old.escaped) {
				emit(actor.x+actor.size/2, actor.y+actor.size/2, actor.size, rl.Gold, "completion")
			}
			if existed && !old.dead && !actor.dead && actor.level > old.level && !actor.ai {
				emit(actor.x+actor.size/2, actor.y+actor.size/2, actor.size, primitives.RuneColor, "level")
			}
			if existed && !old.dead && !actor.dead && actor.hp < old.hp {
				emit(actor.x+actor.size/2, actor.y+actor.size/2, actor.size, rl.RayWhite, "mass")
			}
			if existed && !old.dead && !actor.dead && (actor.hp > old.hp || actor.level > old.level || actor.xp > old.xp) {
				emit(actor.x+actor.size/2, actor.y+actor.size/2, float64(actorRenderSize(actor.size, actor.hp)), actor.tint, "eat")
			}
		}
		for id, old := range f.actors {
			if actor, exists := actors[id]; !old.dead && !old.hidden && (!exists || actor.dead && !actor.hidden) {
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
			emit(fruit.X+fruit.Size/2, fruit.Y+fruit.Size/2, fruit.Size, MassColor(fruit.Source, playerColors), "mass")
		}
	}
	f.actors, f.drops = actors, drops
	if len(f.effects) > 128 {
		f.effects = f.effects[len(f.effects)-128:]
	}
}

func (f *Feedback) Observe(now time.Time, snapshot protocol.Snapshot, localID string) {
	actors := make(map[string]feedbackActor)
	for _, p := range snapshot.Players {
		actors[p.ID] = feedbackActor{p.X, p.Y, p.Size, p.Stats.HP, p.Stats.XP, p.Stats.Level, primitives.CharacterColor(p.Character), false, p.Dead, false, p.Escaped}
	}
	for _, a := range snapshot.AIs {
		actors[a.ID] = feedbackActor{a.X, a.Y, a.Size, a.Stats.HP, a.Stats.XP, a.Stats.Level, rl.Green, true, false, false, false}
	}
	f.observe(now, actors, snapshot.Fruits)
}

func (f *Feedback) Draw(now time.Time, exploration Visibility) {
	active := f.effects[:0]
	for _, e := range f.effects {
		duration := .35
		if e.kind == "level" {
			duration = 1.6
		}
		if e.kind == "completion" {
			duration = 4.5
		}
		p := float32(max(0, now.Sub(e.started).Seconds()) / duration)
		if p >= 1 {
			continue
		}
		active = append(active, e)
		if !exploration.EntityVisible(float64(e.x), float64(e.y), 0) {
			continue
		}
		tint := rl.Fade(e.tint, 1-p)
		center := rl.NewVector2(e.x, e.y)
		switch e.kind {
		case "completion":
			glow := float32(math.Sin(float64(p) * math.Pi))
			rl.DrawCircleGradient(center, 150, rl.Fade(e.tint, .32*glow), rl.Blank)
			radius := 40 + 100*p
			rl.DrawRing(center, radius, radius+2, 0, 360, 64, rl.Fade(e.tint, .55*(1-p)))
			rl.DrawCircleLinesV(center, 65+20*p, rl.Fade(primitives.RuneColor, .5*glow))
			for i := 0; i < 8; i++ {
				angle := float64(i)*math.Pi/4 + float64(p)*.25
				inner, outer := float32(66+12*p), float32(76+12*p)
				a := rl.NewVector2(center.X+float32(math.Cos(angle))*inner, center.Y+float32(math.Sin(angle))*inner)
				b := rl.NewVector2(center.X+float32(math.Cos(angle))*outer, center.Y+float32(math.Sin(angle))*outer)
				rl.DrawLineEx(a, b, 2, rl.Fade(e.tint, .7*glow))
			}
		case "level":
			label := "LEVEL UP / +5 MAX HP & MP"
			primitives.DrawText(label, int32(e.x)-primitives.MeasureText(label, 16)/2, int32(e.y-50-30*p), 16, tint)
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

const materializeDuration = 850 * time.Millisecond

func (f *Feedback) DrawPlayer(id string, x, y, size, hp float64, combat simulation.CombatState, tint color.RGBA, character simulation.Character) rl.Rectangle {
	if f == nil {
		return DrawActor(x, y, size, hp, combat, tint, character)
	}
	started, spawning := f.spawns[id]
	if !spawning {
		return DrawActor(x, y, size, hp, combat, tint, character)
	}
	p := float32(time.Since(started).Seconds() / materializeDuration.Seconds())
	if p >= 1 {
		return DrawActor(x, y, size, hp, combat, tint, character)
	}
	p = max(0, p)
	bounds := actorBounds(x, y, size, hp)
	center := rl.NewVector2(float32(x+size/2), float32(y+size/2))
	assembly := min(1, p/0.75)
	eased := assembly * assembly * (3 - 2*assembly)
	side := bounds.Width * (0.15 + 0.85*eased)
	body := rl.NewRectangle(center.X-side/2, center.Y-side/2, side, side)
	shadow := body
	shadow.Y += 3
	primitives.DrawCharacterShape(character, shadow, rl.Fade(rl.Black, 0.35*eased))
	primitives.DrawCharacterShape(character, body, rl.Fade(tint, eased))
	glow := float32(math.Sin(float64(p) * math.Pi))
	primitives.DrawCharacterShape(character, body, rl.Fade(rl.RayWhite, 0.65*glow*(1-eased)))
	radius := bounds.Width*0.7 + 32*(1-eased)
	rl.DrawRing(center, radius, radius+2, 0, 360, 48, rl.Fade(tint, 0.7*glow))
	for i := range 12 {
		angle := float64(i)*math.Pi/6 + float64(p)*0.8
		distance := bounds.Width*0.35 + (36+float32(i%3)*8)*(1-eased)
		particle := 2 + 3*glow
		px := center.X + float32(math.Cos(angle))*distance
		py := center.Y + float32(math.Sin(angle))*distance
		rl.DrawRectangleRec(rl.NewRectangle(px-particle/2, py-particle/2, particle, particle), rl.Fade(rl.RayWhite, glow))
	}
	if p > 0.65 {
		pulse := (p - 0.65) / 0.35
		radius := bounds.Width*0.6 + 24*pulse
		rl.DrawRing(center, radius, radius+2*(1-pulse), 0, 360, 48, rl.Fade(tint, 1-pulse))
	}
	drawCombat(center, bounds.Width, combat)
	return bounds
}

func (f *Feedback) Spawn(id string, now time.Time) {
	if f.spawns == nil {
		f.spawns = make(map[string]time.Time)
	}
	f.spawns[id] = now
}
