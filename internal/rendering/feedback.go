package rendering

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"2d-dungeon/internal/appearance"
	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"
	"2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type feedbackActor struct {
	shots              uint64
	x, y, size, hp, xp float64
	level              int
	tint               color.RGBA
	ai, dead, hidden   bool
	escaped            bool
	stone              bool
	recovery           float64
	maxHP              float64
	stompRemaining     float64
	swordRemaining     float64
	mobType            simulation.MobType
	behavior           simulation.Behavior
	charging           bool
}

type feedbackEffect struct {
	started    time.Time
	x, y, size float32
	tint       color.RGBA
	kind       string
	amount     float64
	owner      string
	sounded    bool
}

type Feedback struct {
	libraryEvents map[string]simulation.LibraryState
	observed      time.Time
	secrets       map[string]bool
	wallHits      map[string]int
	walls         []wallBreak
	spawns        map[string]time.Time
	actors        map[string]feedbackActor
	drops         map[string]bool
	fruits        map[string]simulation.Fruit
	engaged       map[string]bool
	alerts        map[string]time.Time
	shakeAt       time.Time
	shakeStrength float64
	shakeSeconds  float64
	effects       []feedbackEffect
	hits          map[string]time.Time
	hurtAt        time.Time
	manaAt        time.Time
	lowHP         bool
	localID       string
	celebrated    bool
	vaultGateOpen float64
	entranceGate  simulation.EntranceGateState
	vaultSeals    [4]bool
	vaultFire     [worldmap.VaultFireTrapCount]uint64
}

func (f *Feedback) observe(now time.Time, actors map[string]feedbackActor, fruits []simulation.Fruit) {
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
		f.effects = append(f.effects, feedbackEffect{started: now, x: float32(x), y: float32(y), size: float32(size), tint: tint, kind: kind})
	}
	if f.actors != nil {
		for id, actor := range actors {
			if actor.hidden {
				continue
			}
			old, existed := f.actors[id]
			if actor.ai && existed && !actor.dead {
				local := actors[f.localID]
				geometry := worldmap.Default()
				distance := math.Hypot(actor.x-local.x, actor.y-local.y)
				nearby := distance <= 600
				sameFloor := geometry.BodyFloor(actor.x, actor.y, actor.size) == geometry.BodyFloor(local.x, local.y, local.size)
				if !local.dead && nearby && sameFloor {
					impact := parameters.HeavyStompDuration - parameters.HeavyStompWindup
					if old.stompRemaining > impact && actor.stompRemaining > 0 && actor.stompRemaining <= impact {
						playFeedbackSound("stomp")
						f.shake(now, parameters.StompShakeStrength*(1-distance/600), parameters.ScreenShakeSeconds)
					}
					attack := simulation.MobMelee(actor.mobType)
					strike := attack.Duration - attack.Windup
					if attack.Duration > 0 && old.swordRemaining > strike && actor.swordRemaining <= strike || actor.mobType == simulation.Sentinel && actor.charging && !old.charging {
						cue := "melee-swing"
						if actor.mobType.IsHeavy() {
							cue = "heavy-melee"
						}
						playFeedbackSound(cue)
					}
				}
			}
			if existed && !actor.ai && !old.dead && !actor.dead && worldmap.Default().BodyFloor(old.x, old.y, old.size) != worldmap.Default().BodyFloor(actor.x, actor.y, actor.size) {
				for _, portal := range worldmap.Default().Portals() {
					at := portal.Position()
					if math.Hypot(actor.x+actor.size/2-at.X, actor.y+actor.size/2-at.Y) <= 60 {
						emit(at.X, at.Y, actor.size, primitives.RuneColor, "teleport-arrival")
						break
					}
				}
			}
			if existed && old.stone && !actor.stone {
				local := actors[f.localID]
				if math.Hypot(actor.x-local.x, actor.y-local.y) < 900 {
					playFeedbackSound("awaken")
				}
			}
			if actor.escaped && !actor.dead && (!existed || !old.escaped) {
				emit(actor.x+actor.size/2, actor.y+actor.size/2, actor.size, rl.Gold, "completion")
				if existed && id == f.localID && !f.celebrated {
					f.celebrated = true
					playFeedbackSound("completion")
				}
			}
			if existed && !old.dead && !actor.dead && actor.level > old.level && !actor.ai {
				emit(actor.x+actor.size/2, actor.y+actor.size/2, actor.size, primitives.RuneColor, "level")
			}
			previousHP := old.hp
			if actor.ai && actor.level != old.level && old.maxHP > 0 {
				previousHP = old.hp / old.maxHP * actor.maxHP
			}
			if existed && !old.dead && actor.hp < previousHP-1e-6 {
				if f.hits == nil {
					f.hits = make(map[string]time.Time)
				}
				f.hits[id] = now
				f.effects = append(f.effects, feedbackEffect{
					started: now, x: float32(actor.x + actor.size/2), y: float32(actor.y + actor.size/2), size: float32(actor.size),
					tint: rl.Orange, kind: "hit", amount: previousHP - actor.hp, owner: id,
				})
			}
			if existed && !old.dead && !actor.dead && (actor.hp > old.hp+actor.recovery*(max(0, now.Sub(f.observed).Seconds())+parameters.StepSeconds)+1e-6 || actor.level > old.level || actor.xp > old.xp) {
				emit(actor.x+actor.size/2, actor.y+actor.size/2, float64(actorRenderSize(actor.size, actor.hp)), actor.tint, "eat")
			}
		}
		for id, old := range f.actors {
			if actor, exists := actors[id]; !old.dead && !old.hidden && (!exists || actor.dead && !actor.hidden) {
				kind := "prey"
				if old.ai {
					kind = "defeat"
				} else if id == f.localID {
					f.shake(now, parameters.DeathShakeStrength, parameters.DeathShakeSeconds)
				}
				emit(old.x+old.size/2, old.y+old.size/2, float64(actorRenderSize(old.size, old.hp)), old.tint, kind)
			}
		}
	}
	for id, at := range f.hits {
		if now.Sub(at) > time.Second {
			delete(f.hits, id)
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
	f.actors, f.drops, f.observed = actors, drops, now
	if len(f.effects) > 128 {
		f.effects = f.effects[len(f.effects)-128:]
	}
}

func (f *Feedback) Observe(now time.Time, snapshot simulation.Snapshot) {
	p := snapshot.Player
	localID := p.ID
	f.localID = localID
	f.observeLibraryAudio(snapshot)
	actors := make(map[string]feedbackActor)
	actors[p.ID] = feedbackActor{
		shots: p.ShotsFired,
		x:     p.X, y: p.Y, size: parameters.PlayerSize, hp: p.Stats.HP, xp: p.Stats.XP, level: p.Stats.Level,
		tint: primitives.CharacterColor(p.Character), dead: p.Dead, escaped: p.Escaped,
		recovery: p.Stats.HPRecoveryPerSecond(), maxHP: p.Stats.MaxHP(),
	}
	for _, a := range snapshot.AIs {
		actors[a.ID] = feedbackActor{x: a.X, y: a.Y, size: parameters.AISize, hp: a.Stats.HP, xp: a.Stats.XP, level: a.Stats.Level, tint: mobTint(a.Type), ai: true, maxHP: a.Stats.MaxHP()}
		actor := actors[a.ID]
		actor.shots = a.ShotsFired
		actor.stone = a.Stone
		actor.stompRemaining = a.Combat.StompRemaining
		actor.swordRemaining = a.Combat.SwordRemaining
		actor.mobType = a.Type
		actor.charging = a.Combat.Phase == simulation.DashRushing
		actor.behavior = a.Behavior
		actors[a.ID] = actor
	}
	if !f.observed.IsZero() {
		local := actors[localID]
		geometry := worldmap.Default().WithSecrets(snapshot.OpenSecrets)
		traps := geometry.VaultTraps()
		body := physics.Square(local.x, local.y, local.size)
		for i, state := range snapshot.VaultFire {
			if i >= len(f.vaultFire) || i >= len(traps) || state.TriggeredTick == 0 || snapshot.Tick < state.TriggeredTick {
				continue
			}
			age := snapshot.Tick - state.TriggeredTick
			if age >= simulation.VaultTrapWarningTicks && age < simulation.VaultTrapWarningTicks+simulation.VaultTrapFireTicks && f.vaultFire[i] != state.TriggeredTick {
				f.vaultFire[i] = state.TriggeredTick
				if !local.dead && geometry.BodyFloor(local.x, local.y, local.size) == worldmap.OrbFloor && body.Overlaps(traps[i].Lane.Expanded(300)) {
					playFeedbackSound("fire")
				}
			}
		}
	}
	if !f.observed.IsZero() {
		local := actors[localID]
		if !local.dead && p.EntranceGate.ID == f.entranceGate.ID && p.EntranceGate.Remaining > 0 && f.entranceGate.Remaining == 0 && p.EntranceGate.Open > f.entranceGate.Open {
			playFeedbackSound("entrance-gate")
		}
		if snapshot.Orb.GateOpen > 0 && f.vaultGateOpen == 0 && simulation.NearVaultGate(local.x, local.y, local.size) {
			playFeedbackSound("gate")
		}
		for i, lit := range snapshot.Orb.Seals {
			if lit && !f.vaultSeals[i] && worldmap.Default().RoomContains(worldmap.OrbVault, physics.Square(local.x, local.y, local.size)) {
				playFeedbackSound("seal")
			}
		}
	}
	f.vaultGateOpen, f.vaultSeals = snapshot.Orb.GateOpen, snapshot.Orb.Seals
	f.entranceGate = p.EntranceGate
	f.observeWalls(now, snapshot)
	f.observeShootingAudio(actors)
	f.observeAlerts(now, actors)
	f.observePickups(now, snapshot.Tick, actors[localID], snapshot.Fruits)
	f.observe(now, actors, snapshot.Fruits)
	if at := f.hits[localID]; at == now {
		f.hurtAt = now
		f.shake(now, parameters.HurtShakeStrength, parameters.ScreenShakeSeconds)
		playFeedbackSound("hurt")
	}
	low := !p.Dead && p.Stats.HP <= p.Stats.MaxHP()*.25
	if low && !f.lowHP {
		playFeedbackSound("low")
	}
	f.lowHP = low
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
		if e.kind == "teleport-arrival" {
			duration = .9
		}
		if e.kind == "hit" {
			duration = .65
		}
		if e.kind == "defeat" {
			duration = .8
		}
		if e.kind == "prey" {
			duration = parameters.DeathSequenceSeconds
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
		case "teleport-arrival":
			drawTeleportEnergy(center, center, parameters.PortalSeconds+float64(p)*duration, (1-p)*(1-p), 1-p)
			rl.DrawRing(center, 84+55*p, 87+55*p, 0, 360, 64, rl.Fade(e.tint, (1-p)*.6))
		case "hit":
			if !e.sounded {
				if e.owner != f.localID {
					playFeedbackSound("hit")
				}
				active[len(active)-1].sounded = true
			}
			label := fmt.Sprintf("-%.1f", e.amount)
			primitives.DrawText(label, int32(e.x)-primitives.MeasureText(label, 20)/2, int32(e.y-e.size/2-56-28*p), 20, tint)
			if p < .4 {
				for i := range 6 {
					angle := float64(i) * math.Pi / 3
					radius := e.size/2 + 42*p
					a := rl.NewVector2(e.x+float32(math.Cos(angle))*radius, e.y+float32(math.Sin(angle))*radius)
					b := rl.NewVector2(a.X+float32(math.Cos(angle))*7, a.Y+float32(math.Sin(angle))*7)
					rl.DrawLineEx(a, b, 2, rl.Fade(rl.RayWhite, 1-p/.4))
				}
			}
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
		case "prey", "defeat":
			if e.kind == "defeat" {
				if !e.sounded {
					playFeedbackSound("hit")
					active[len(active)-1].sounded = true
				}
				drawDeathBurst(center, e.size, p, e.tint, uint32(e.x*7)^uint32(e.y*13))
				label := "DEFEATED"
				primitives.DrawText(label, int32(e.x)-primitives.MeasureText(label, 18)/2, int32(e.y-e.size/2-28-22*p), 18, rl.Fade(rl.Gold, 1-p))
			}
			if e.kind == "prey" {
				drawPlayerDeath(center, e.size, p, e.tint, uint32(e.x*7)^uint32(e.y*13))
				break
			}
			side := e.size * (1 - p) * (1 - p)
			rl.DrawRectangleRec(rl.NewRectangle(e.x-side/2, e.y-side/2, side, side), tint)
		case "pop":
			drawPickupPop(center, e.size, p, e.tint)
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

func (f *Feedback) DrawPlayer(id string, x, y, size, hp float64, combat simulation.CombatState, tint color.RGBA, character appearance.Character) rl.Rectangle {
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

func (f *Feedback) observeLibraryAudio(snapshot simulation.Snapshot) {
	if f.libraryEvents == nil {
		f.libraryEvents = make(map[string]simulation.LibraryState)
	}
	local := snapshot.Player
	for _, book := range snapshot.Statues {
		if !simulation.IsLibraryBook(book.ID) {
			continue
		}
		before, known := f.libraryEvents[book.ID]
		current := book.Library
		audible := !local.Dead && worldmap.Default().WithSecrets(snapshot.OpenSecrets).RoomContains(simulation.LibraryRoom(book.ID), physics.Square(local.X, local.Y, parameters.PlayerSize))
		if known && audible && current.StartedTick != 0 && before.StartedTick == current.StartedTick {
			for tome := 0; tome < simulation.LibraryTomeCount; tome++ {
				if current.HitTicks[tome] > before.HitTicks[tome] {
					playFeedbackSound("library-tap")
				}
				if current.BounceTicks[tome] > before.BounceTicks[tome] {
					playFeedbackSound("library-bounce")
				}
				if current.BoundTicks[tome] > before.BoundTicks[tome] {
					playFeedbackSound("library-bind")
				}
			}
			if current.Completed && !before.Completed {
				playFeedbackSound("library-done")
			}
		}
		f.libraryEvents[book.ID] = current
	}
}
