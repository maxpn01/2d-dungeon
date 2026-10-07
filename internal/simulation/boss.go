package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

const (
	BossStartingHP  = parameters.BossStartingHP
	BossMaxHP       = parameters.BossMaxHP
	BossMaxTail     = parameters.BossMaxTail
	BossPartHP      = parameters.BossPartHP
	BossSize        = parameters.BossSize
	BossDashDamage  = parameters.BossDashDamage
	BossCountdown   = parameters.BossCountdown
	BossAggroRadius = parameters.BossAggroRadius
	PortalSeconds   = parameters.PortalSeconds
)

type Point struct{ X, Y float64 }

type Boss struct {
	Name            string
	Level           int
	XP              float64
	Phase           string // waiting, gathering, fighting, defeated
	X, Y, HP, MaxHP float64
	Countdown       float64
	SpawnTicks      int // ticks until the boss appears; zero while it is present
	SpawnDuration   int
	Defeats         int    // keeps the descent unlocked across encounters
	Behavior        string // hunting, winding, rushing, recovering, feeding, patrolling, fleeing
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
	b := Boss{Name: "OROCHI, KING OF SERPENTS", Level: 10, Phase: "waiting", Behavior: "feeding", X: parameters.BossSpawnMinX, Y: parameters.BossSpawnMinY, HP: BossStartingHP, MaxHP: BossStartingHP, SpawnDuration: parameters.BossFirstSpawnSeconds * TickRate}
	b.SpawnTicks = b.SpawnDuration
	b.growTail()
	return b
}

func (w *World) placeBoss() {
	b := &w.boss
	bestDistance := -1.0
	for attempts := 0; attempts < 100; attempts++ {
		x := parameters.BossSpawnMinX + rand.Float64()*(parameters.BossSpawnMaxX-parameters.BossSpawnMinX)
		y := parameters.BossSpawnMinY + rand.Float64()*(parameters.BossSpawnMaxY-parameters.BossSpawnMinY)
		nearest := math.Inf(1)
		for _, p := range w.players {
			if !p.Dead {
				nearest = min(nearest, math.Hypot(p.X-x, p.Y-y))
			}
		}
		if nearest > bestDistance {
			bestDistance = nearest
			b.X, b.Y = x, y
		}
		if nearest >= 120 {
			break
		}
	}
	for i := range b.Tail {
		row, col := i/18, i%18
		if row%2 == 1 {
			col = 17 - col
		}
		b.Tail[i] = Point{b.X - float64(col+1)*30, b.Y + float64(row)*30}
	}
}

func bossArenaBounds() physics.Rect {
	room, ok := worldmap.Default().RoomWithRole(worldmap.BossArena)
	if !ok {
		return physics.Rect{}
	}
	return room.Bounds()
}

func InsideArena(x, y, size float64) bool {
	return bossArenaBounds().Contains(physics.Square(x, y, size))
}

func ConstrainArena(p *Player, oldX, oldY float64) {
	if !p.GatesClosed {
		return
	}
	if p.InArena {
		body := physics.Square(p.X, p.Y, PlayerSize).ClampedTo(bossArenaBounds())
		p.X, p.Y = body.X, body.Y
	} else if bossArenaBounds().Overlaps(physics.Square(p.X, p.Y, PlayerSize)) {
		p.X, p.Y = oldX, oldY
	}
}

func (b *Boss) growTail() {
	count := min(BossMaxTail, max(0, int((b.HP-BossStartingHP)/BossPartHP)))
	for len(b.Tail) < count {
		p := Point{b.X, b.Y}
		if len(b.Tail) > 0 {
			p = b.Tail[len(b.Tail)-1]
		}
		b.Tail = append(b.Tail, p)
	}
	b.MaxHP = max(b.MaxHP, b.HP)
}

func (w *World) bossTargets() []*playerState {
	var players []*playerState
	for _, p := range w.players {
		if !p.Dead && p.Y >= MapMin && !w.bossRefuge(p.X, p.Y, PlayerSize, p.Stats.HP) && math.Hypot(p.X-w.boss.X, p.Y-w.boss.Y) <= BossAggroRadius {
			players = append(players, p)
		}
	}
	sort.Slice(players, func(i, j int) bool { return players[i].ID < players[j].ID })
	return players
}

func (w *World) bossAIs() []*aiState {
	var ais []*aiState
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick == 0 && ai.Y >= MapMin && !w.bossRefuge(ai.X, ai.Y, AISize, ai.Stats.HP) && math.Hypot(ai.X-w.boss.X, ai.Y-w.boss.Y) <= BossAggroRadius {
			ais = append(ais, ai)
		}
	}
	sort.Slice(ais, func(i, j int) bool { return ais[i].ID < ais[j].ID })
	return ais
}

func (w *World) bossRefuge(x, y, size, hp float64) bool {
	return InSafeZone(x, y, size, hp, w.safeZones)
}

func (w *World) openArena() {
	for _, p := range w.players {
		p.GatesClosed = false
		p.InArena = false
	}
}

func (w *World) updateBoss(previous map[string]position) {
	b := &w.boss
	if b.Phase == "waiting" || b.Phase == "defeated" {
		if b.SpawnTicks > 0 {
			b.SpawnTicks--
			if b.SpawnTicks > 0 {
				return
			}
		}
		duration := b.SpawnDuration
		defeats := b.Defeats
		*b = newBoss()
		b.SpawnDuration = duration
		b.SpawnTicks = 0
		b.Defeats = defeats
		initialTail := min(BossMaxTail, int((BossMaxHP-BossStartingHP)/BossPartHP), parameters.BossFirstUnits-1+parameters.BossUnitsPerRespawn*defeats)
		b.HP = BossStartingHP + float64(initialTail)*BossPartHP
		b.MaxHP = b.HP
		b.growTail()
		b.Phase = "fighting"
		b.Countdown = 0
		w.bossHitUntil = make(map[string]uint64)
		w.placeBoss()
		b.tactic("hunting", bossHuntTicks)
		for _, p := range w.bossTargets() {
			w.protectedUntil[p.ID] = w.tick + TickRate
		}
	}
	players := w.bossTargets()
	ais := w.bossAIs()
	if len(players) == 0 && len(ais) == 0 {
		w.updateBossLife()
		return
	}
	old := position{b.X, b.Y}
	var prey []bossTarget
	preyDistance := math.Inf(1)
	threatDistance := math.Inf(1)
	var threat bossTarget
	for _, p := range players {
		distance := math.Hypot(p.X-b.X, p.Y-b.Y)
		if p.Stats.HP < b.HP {
			prey = append(prey, bossTarget{p.ID, p.X, p.Y})
			preyDistance = min(preyDistance, distance)
		} else if p.Stats.HP > b.HP && distance < threatDistance {
			threat, threatDistance = bossTarget{p.ID, p.X, p.Y}, distance
		}
	}
	for _, ai := range ais {
		distance := math.Hypot(ai.X-b.X, ai.Y-b.Y)
		if ai.Stats.HP < b.HP {
			prey = append(prey, bossTarget{ai.ID, ai.X, ai.Y})
			preyDistance = min(preyDistance, distance)
		} else if ai.Stats.HP > b.HP && distance < threatDistance {
			threat, threatDistance = bossTarget{ai.ID, ai.X, ai.Y}, distance
		}
		for i := range ai.Split {
			cell := &ai.Split[i]
			if !cell.Active {
				continue
			}
			cellID := fmt.Sprintf("%s/split/%d", ai.ID, i)
			cellDistance := math.Hypot(cell.X-b.X, cell.Y-b.Y)
			if cell.Stats.HP < b.HP {
				prey = append(prey, bossTarget{cellID, cell.X, cell.Y})
				preyDistance = min(preyDistance, cellDistance)
			} else if cell.Stats.HP > b.HP && cellDistance < threatDistance {
				threat, threatDistance = bossTarget{cellID, cell.X, cell.Y}, cellDistance
			}
		}
	}
	idle := false
	if threatDistance <= AIDetectionRadius {
		w.fleeBoss(threat)
		b.followTail()
	} else if len(prey) > 0 && preyDistance <= w.bossFoodDistance() {
		w.moveBoss(prey, previous)
		b.followTail()
	} else {
		w.updateBossLife()
		idle = true
	}
	if !idle {
		w.feedBoss(old)
	}
}

func bossContact(start, end position, size, hp float64, old position, b *Boss, oldTail []Point) (float64, position, position, int, bool) {
	bodySize := BodySize(size, hp)
	offset := (size - bodySize) / 2
	visualStart := position{start.x + offset, start.y + offset}
	visualEnd := position{end.x + offset, end.y + offset}
	contact, hit := eatingContact(visualStart, visualEnd, bodySize, old, position{b.X, b.Y}, BossSize)
	hitStart, hitEnd := old, position{b.X, b.Y}
	partIndex := -1
	for i, segment := range b.Tail {
		after := position{segment.X, segment.Y}
		before := after
		if i < len(oldTail) {
			before = position{oldTail[i].X, oldTail[i].Y}
		}
		at, touches := eatingContact(visualStart, visualEnd, bodySize, before, after, BossSize)
		if touches && (!hit || at < contact) {
			contact, hit, hitStart, hitEnd, partIndex = at, true, before, after, i
		}
	}
	return contact, hitStart, hitEnd, partIndex, hit
}

func (w *World) detachBossPart(index int, oldTail *[]Point, drop bool) float64 {
	b := &w.boss
	segment := b.Tail[index]
	b.Tail = append(b.Tail[:index], b.Tail[index+1:]...)
	if index < len(*oldTail) {
		*oldTail = append((*oldTail)[:index], (*oldTail)[index+1:]...)
	}
	amount := min(BossPartHP, b.HP)
	b.HP -= amount
	if !drop {
		return amount
	}
	drops := 0
	for _, fruit := range w.fruits {
		if fruit.Source != "" {
			drops++
		}
	}
	if drops < MaxMassDrops {
		w.nextFruit++
		x, y := worldmap.Default().Move(segment.X, segment.Y, (BossSize-FruitSize)/2, (BossSize-FruitSize)/2, FruitSize, FruitSize)
		w.fruits = append(w.fruits, Fruit{ID: fmt.Sprintf("m%d", w.nextFruit), Source: "boss", HP: amount, X: x, Y: y, Size: FruitSize, AvailableTick: w.tick + HitProtectionTicks, ExpiresTick: w.tick + parameters.BossDropLifetimeTicks})
	}
	return amount
}

func (w *World) defeatBoss(players []*playerState, rewardPlayers bool) {
	w.boss.scheduleRespawn()
	w.openArena()
	if !rewardPlayers {
		return
	}
	for _, survivor := range players {
		if !survivor.Dead {
			survivor.Stats.GainXP(20)
		}
	}
}

func (b *Boss) scheduleRespawn() {
	b.Phase = "defeated"
	b.Defeats++
	b.SpawnDuration += parameters.BossRespawnIncreaseSeconds * TickRate
	b.SpawnTicks = b.SpawnDuration
	b.Accelerating = false
	b.Tail = nil
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
		if hit && f.Y >= MapMin && w.tick >= f.AvailableTick && f.Source != "boss" && (b.HP < BossMaxHP || b.Level < MaxLevel) {
			hp := f.rewardHP()
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
		if p.Dead || p.Flight.Flying {
			p.TeleportCharge = 0
			continue
		}
		if p.portalCooldown > 0 {
			p.portalCooldown--
			p.TeleportCharge = 0
			continue
		}
		if p.TeleportCharge <= 0 {
			continue
		}
		portal, distance, found := nearestPortal(p.X+PlayerSize/2, p.Y+PlayerSize/2)
		if !found {
			p.TeleportCharge = 0
			p.portalCooldown = 0
			continue
		}
		if portal.ID == worldmap.DescentPortal && !w.orb.Consumed && !p.HasOrb {
			p.TeleportCharge = 0
			continue
		}
		destination, _ := worldmap.Default().Portal(portal.Destination)
		landing := destination.Position()
		toX, toY := landing.X, landing.Y
		if distance > 60 {
			p.TeleportCharge = 0
			continue
		}
		p.TeleportCharge += StepSeconds
		if p.TeleportCharge >= PortalSeconds {
			p.X, p.Y = toX-PlayerSize/2, toY-PlayerSize/2
			if portal.ID == worldmap.DescentPortal {
				w.orb.Consumed = true
				w.releaseOrb(w.orb.Carrier)
				p.Escaped = true
				if w.winnerID == "" {
					w.winnerID = p.ID
					w.finishedTick = w.tick
				}
			}
			if p.Split.Count() > 0 {
				p.Stats.HP += p.Split.Mass()
				p.Split = nil
			}
			p.Combat = CombatState{}
			p.TeleportCharge = 0
			p.portalCooldown = 2 * TickRate
			p.commands = nil
			p.InputSeq = p.lastSequence
			p.direction = Direction{}
		}
	}
}

func nearestPortal(x, y float64) (worldmap.PortalPoint, float64, bool) {
	world := worldmap.Default()
	floor := world.FloorAt(x, y)
	nearest := worldmap.PortalPoint{}
	distance := math.Inf(1)
	for _, portal := range world.Portals() {
		if portal.Floor != floor {
			continue
		}
		position := portal.Position()
		candidate := math.Hypot(x-position.X, y-position.Y)
		if candidate < distance {
			nearest, distance = portal, candidate
		}
	}
	return nearest, distance, nearest.ID != ""
}
