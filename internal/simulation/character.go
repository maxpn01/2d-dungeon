package simulation

import (
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

type Character string

const (
	Square   Character = "square"
	Triangle Character = "triangle"
)

func CharacterEnabled(c Character) bool {
	return c == Square
}

func NormalizeCharacter(c Character) Character {
	if CharacterEnabled(c) {
		return c
	}
	return Square
}

func (w *World) SetCharacter(id string, character Character) {
	if p := w.players[id]; p != nil {
		p.Character = NormalizeCharacter(character)
		p.Combat = CombatState{}
	}
}

func AdvancePlayer(p *Player, input Direction, dt float64) {
	defer clampPlayerBodies(p)
	if p.Dead || !input.Valid() || dt <= 0 {
		return
	}
	oldX, oldY := p.X, p.Y
	advanceEntranceGate(p, input, dt)
	defer constrainEntranceGate(p, oldX, oldY)
	defer constrainWingDoor(p, oldX, oldY)
	p.Stats.GainMP(p.Stats.MPRecoveryPerSecond() * dt)
	p.Stats.HP = min(p.Stats.MaxHP(), p.Stats.HP+p.Stats.HPRecoveryPerSecond()*dt)
	p.ShotCooldown = max(-dt, p.ShotCooldown-dt)
	if p.Combat.StunRemaining > 0 {
		p.Combat.StunRemaining = max(0, p.Combat.StunRemaining-dt)
		return
	}
	if p.Combat.StompRemaining > 0 && !p.Stats.Mob {
		return
	}
	if p.HasWings {
		advanceWingPlayer(p, input, dt)
		return
	}
	if p.Stats.Mob && MobMelee(p.Stats.MobType).Duration > 0 {
		AdvanceDashPlayer(p, input, dt)
		return
	}
	advanceBoostPlayer(p, input, dt)
}

func clampPlayerBodies(p *Player) {
	ClampSurfaceBody(&p.X, &p.Y, PlayerSize, p.Stats.HP)
}

func ClampSurfaceBody(x, y *float64, size, hp float64) {
	if *y < MapMin {
		return // the dungeon is governed by its own passages
	}
	if (*x < MapMin || *x+size > MapMax || *y+size > MapMax) && worldmap.Default().CanFly(physics.Square(*x, *y, size)) {
		return
	}
	radius := min(BodySize(size, hp)/2, (MapMax-MapMin)/2)
	body := physics.Rect{X: *x + size/2 - radius, Y: *y + size/2 - radius, W: radius * 2, H: radius * 2}
	body = body.ClampedTo(physics.Rect{X: MapMin, Y: MapMin, W: MapMax - MapMin, H: MapMax - MapMin})
	center := body.Center()
	*x, *y = center.X-size/2, center.Y-size/2
}

func BodySize(size, hp float64) float64 {
	return size
}
