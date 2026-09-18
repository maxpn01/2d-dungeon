package simulation

import (
	"fmt"
	"math"
	"raylib-2d-sandbox/internal/terrain"
	"sort"
)

const (
	BossStartingHP = 30.0
	BossMaxHP      = 1000.0
	BossMaxTail    = 100
	BossSize       = 36.0
	BossDashDamage = 8.0
	BossCountdown  = 3.0
	PortalSeconds  = 0.8
	ArenaLeft      = 12 * terrain.Tile
	ArenaRight     = 38 * terrain.Tile
	ArenaTop       = -134 * terrain.Tile
	ArenaBottom    = -118 * terrain.Tile
)

type Point struct{ X, Y float64 }

type Boss struct {
	Name            string
	Level           int
	XP              float64
	Phase           string // waiting, gathering, fighting, defeated
	X, Y, HP, MaxHP float64
	Countdown       float64
	Behavior        string // hunting, winding, rushing, recovering, feeding
	AttackTarget    Point
	Enraged         bool
	Chain           int
	Accelerating    bool
	Tail            []Point
	speed           float64
	tacticTicks     int
	attackCount     int
	targetID        string
	heading         float64
	patrol          int
	staggerTicks    int
}

func newBoss() Boss {
	b := Boss{Name: "OROCHI, KING OF SERPENTS", Level: 10, Phase: "waiting", Behavior: "feeding", X: 980, Y: -5220, HP: BossStartingHP, MaxHP: BossStartingHP}
	b.growTail()
	for i := range b.Tail {
		b.Tail[i] = Point{b.X + float64(i+1)*30, b.Y}
	}
	return b
}

func InsideArena(x, y, size float64) bool {
	return x >= ArenaLeft && y >= ArenaTop && x+size <= ArenaRight && y+size <= ArenaBottom
}

// Shared by authoritative movement and client prediction; arena state belongs to
// each world/player, never the global terrain map.
func ConstrainArena(p *Player, oldX, oldY float64) {
	if !p.GatesClosed {
		return
	}
	if p.InArena {
		p.X = max(ArenaLeft, min(ArenaRight-PlayerSize, p.X))
		p.Y = max(ArenaTop, min(ArenaBottom-PlayerSize, p.Y))
	} else if p.X+PlayerSize > ArenaLeft && p.X < ArenaRight && p.Y+PlayerSize > ArenaTop && p.Y < ArenaBottom {
		p.X, p.Y = oldX, oldY
	}
}

func (b *Boss) growTail() {
	count := min(BossMaxTail, int(b.HP/10))
	for len(b.Tail) < count {
		p := Point{b.X, b.Y}
		if len(b.Tail) > 0 {
			p = b.Tail[len(b.Tail)-1]
		}
		b.Tail = append(b.Tail, p)
	}
	b.MaxHP = max(b.MaxHP, b.HP)
}

func (w *World) arenaPlayers() []*playerState {
	var players []*playerState
	for _, p := range w.players {
		if !p.Dead && InsideArena(p.X, p.Y, PlayerSize) {
			players = append(players, p)
		}
	}
	sort.Slice(players, func(i, j int) bool { return players[i].ID < players[j].ID })
	return players
}

func (w *World) openArena() {
	for _, p := range w.players {
		p.GatesClosed = false
		p.InArena = false
	}
}

func (w *World) updateBoss(previous map[string]position) {
	b := &w.boss
	if b.Phase == "defeated" {
		return
	}
	players := w.arenaPlayers()
	if b.Phase == "waiting" {
		if len(players) == 0 {
			w.updateBossLife()
			return
		}
		b.Phase = "gathering"
		b.Countdown = BossCountdown
	}
	if b.Phase == "gathering" {
		if len(players) == 0 {
			w.resumeBossLife()
			w.updateBossLife()
			return
		}
		b.Countdown = max(0, b.Countdown-StepSeconds)
		if b.Countdown > 0 {
			w.updateBossLife()
			return
		}
		b.Phase = "fighting"
		b.targetID = ""
		b.tactic("hunting", bossHuntTicks)
		for _, p := range w.players {
			p.GatesClosed = true
			p.InArena = !p.Dead && InsideArena(p.X, p.Y, PlayerSize)
			if p.InArena {
				w.protectedUntil[p.ID] = w.tick + TickRate
			}
		}
	}
	// A disconnect is a wipe too; never strand the shared world behind gates.
	alive := players[:0]
	for _, p := range players {
		if p.InArena {
			alive = append(alive, p)
		}
	}
	players = alive
	if len(players) == 0 {
		w.resumeBossLife()
		w.updateBossLife()
		return
	}
	old := position{b.X, b.Y}
	w.moveBoss(players, previous)
	oldTail := append([]Point(nil), b.Tail...)
	b.followTail()
	for _, p := range players {
		start := previous[p.ID]
		end := position{p.X, p.Y}
		contact, hit := eatingContact(start, end, PlayerSize, old, position{b.X, b.Y}, BossSize)
		hitStart, hitEnd := old, position{b.X, b.Y}
		for i, segment := range b.Tail {
			before, after := position{oldTail[i].X, oldTail[i].Y}, position{segment.X, segment.Y}
			at, touches := eatingContact(start, end, PlayerSize, before, after, BossSize)
			if touches && (!hit || at < contact) {
				contact, hit, hitStart, hitEnd = at, true, before, after
			}
		}
		if !hit {
			continue
		}
		if p.Combat.Attack && w.tick >= w.bossHitUntil[p.ID] {
			b.HP = max(0, b.HP-BossDashDamage)
			w.bossHitUntil[p.ID] = w.tick + TickRate/2
			// Stop at the first segment actually struck. Carry that contact with the
			// moving segment to its final tick position so the bodies visibly touch.
			p.X = start.x + (end.x-start.x)*contact + (hitEnd.x-hitStart.x)*(1-contact)
			p.Y = start.y + (end.y-start.y)*contact + (hitEnd.y-hitStart.y)*(1-contact)
			ConstrainArena(&p.Player, start.x, start.y)
			w.bossHitEscape(p)
			// A successful hit releases mass for recovery, without immediately healing
			// the boss that shed it.
			drops := 0
			for _, f := range w.fruits {
				if f.Source != "" {
					drops++
				}
			}
			for i := 0; i < 4; i++ {
				if drops >= MaxMassDrops {
					break
				}
				drops++
				w.nextFruit++
				x, y := terrain.Move(p.X, p.Y, float64(i%2)*25-12, float64(i/2)*25-12, FruitSize, FruitSize)
				w.fruits = append(w.fruits, Fruit{ID: fmt.Sprintf("m%d", w.nextFruit), Source: "boss", HP: 1, X: x, Y: y, Size: FruitSize, AvailableTick: w.tick + HitProtectionTicks, ExpiresTick: w.tick + 15*TickRate})
			}
			if b.HP == 0 {
				b.Phase = "defeated"
				b.Accelerating = false
				b.Tail = nil
				w.openArena()
				for _, survivor := range players {
					if !survivor.Dead {
						survivor.Stats.GainXP(20)
					}
				}
				return
			}
		} else if w.tick >= w.protectedUntil[p.ID] && CanContactEat(b.HP, p.Stats.HP) {
			b.eat(p.Stats.HP*ActorHPFraction, p.Stats.HP)
			(eater{player: p}).die("boss", w.tick)
		}
	}
	w.feedBoss(old)
}

func (b *Boss) followTail() {
	lead := Point{b.X, b.Y}
	for i := range b.Tail {
		p := &b.Tail[i]
		dx, dy := lead.X-p.X, lead.Y-p.Y
		distance := math.Hypot(dx, dy)
		if distance > 30 {
			p.X += dx / distance * (distance - 30)
			p.Y += dy / distance * (distance - 30)
		}
		lead = *p
	}
}

// Use the same XP curve as ordinary creatures; mass is also limited by level.
func (b *Boss) eat(hp, xp float64) {
	stats := Stats{Level: b.Level, XP: b.XP}
	stats.GainXP(xp)
	b.Level, b.XP = stats.Level, stats.XP
	b.HP = min(BossMaxHP, min(MaxHPForLevel(b.Level), b.HP+hp))
	b.MaxHP = max(b.MaxHP, b.HP)
}

func (w *World) feedBoss(old position) {
	b := &w.boss
	remaining := w.fruits[:0]
	for _, f := range w.fruits {
		if f.ExpiresTick != 0 && w.tick >= f.ExpiresTick {
			continue
		}
		_, hit := eatingContact(old, position{b.X, b.Y}, BossSize, position{f.X, f.Y}, position{f.X, f.Y}, f.Size)
		if hit && InsideArena(f.X, f.Y, f.Size) && w.tick >= f.AvailableTick && f.Source != "boss" && (b.HP < BossMaxHP || b.Level < MaxLevel) {
			hp := FruitHP
			if f.Source != "" {
				hp = f.HP
			}
			b.eat(hp, FruitXP*hp)
		} else {
			remaining = append(remaining, f)
		}
	}
	w.fruits = remaining
	b.growTail()
}

func (w *World) updatePortals() {
	for _, p := range w.players {
		if p.Dead {
			p.TeleportCharge = 0
			continue
		}
		if w.boss.Phase != "defeated" {
			p.TeleportCharge = 0
			continue
		}
		x, y := float64(terrain.DescentCircleX), float64(terrain.DescentCircleY)
		toX, toY := float64(terrain.ReturnCircleX), float64(terrain.ReturnCircleY)
		if terrain.FloorTwo(p.X, p.Y) {
			x, y, toX, toY = toX, toY, x, y
		}
		distance := math.Hypot(p.X+PlayerSize/2-x, p.Y+PlayerSize/2-y)
		if p.portalCooldown != 0 {
			if distance > 90 {
				p.portalCooldown = 0
			}
			p.TeleportCharge = 0
			continue
		}
		if distance > 60 {
			p.TeleportCharge = 0
			continue
		}
		p.TeleportCharge += StepSeconds
		if p.TeleportCharge >= PortalSeconds {
			p.X, p.Y = toX-PlayerSize/2, toY-PlayerSize/2
			p.Combat = CombatState{}
			p.TeleportCharge = 0
			p.portalCooldown = 2 * TickRate
			// Retire pre-teleport inputs so lag cannot fling someone off the arrival pad.
			p.commands = nil
			p.InputSeq = p.lastSequence
			p.direction = Direction{}
		}
	}
}
