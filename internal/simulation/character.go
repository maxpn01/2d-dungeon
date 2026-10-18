package simulation

import (
	"2d-dungeon/internal/appearance"
	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
	"2d-dungeon/internal/worldmap"
)

func (w *World) SetCharacter(character appearance.Character) {
	w.player.Character = appearance.Normalize(character)
	w.player.Combat = CombatState{}
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
	ClampSurfaceBody(&p.X, &p.Y, parameters.PlayerSize, p.Stats.HP)
}

func ClampSurfaceBody(x, y *float64, size, hp float64) {
	if *y < parameters.MapMin {
		return // the dungeon is governed by its own passages
	}
	if (*x < parameters.MapMin || *x+size > parameters.MapMax || *y+size > parameters.MapMax) && worldmap.Default().CanFly(physics.Square(*x, *y, size)) {
		return
	}
	radius := min(BodySize(size, hp)/2, (parameters.MapMax-parameters.MapMin)/2)
	body := physics.Rect{X: *x + size/2 - radius, Y: *y + size/2 - radius, W: radius * 2, H: radius * 2}
	body = body.ClampedTo(physics.Rect{X: parameters.MapMin, Y: parameters.MapMin, W: parameters.MapMax - parameters.MapMin, H: parameters.MapMax - parameters.MapMin})
	center := body.Center()
	*x, *y = center.X-size/2, center.Y-size/2
}

func BodySize(size, hp float64) float64 {
	return size
}
