package simulation

import (
	"math"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

const SentinelAwakeningDuration = 3.0
const VaultTrapWarningTicks = TickRate / 2
const VaultTrapFireTicks = TickRate
const VaultTrapCooldownTicks = 5 * TickRate

type VaultFire struct {
	TriggeredTick uint64 `json:"triggered_tick"`
}

func (w *World) awakenVaultSentinels() {
	geometry := worldmap.Default().WithSecrets(w.openSecrets)
	for _, p := range w.players {
		if p.Dead {
			continue
		}
		for i := range w.ais {
			ai := &w.ais[i]
			if ai.den.RoomID != worldmap.OrbVault || !ai.Stone || !w.aiCanSeePlayer(ai, p, 260) {
				continue
			}
			for j := range w.ais {
				guard := &w.ais[j]
				if guard.den.RoomID == worldmap.OrbVault && guard.Stone && geometry.Cell(guard.den.At.X, guard.den.At.Y) != 0 {
					guard.Stone, guard.Awakening, guard.Behavior = false, SentinelAwakeningDuration, "awakening"
				}
			}
			return
		}
	}
}

func (w *World) advanceVaultTraps(actors []eater) {
	geometry := worldmap.Default().WithSecrets(w.openSecrets)
	for i, trap := range geometry.VaultTraps() {
		if geometry.Cell(trap.Tile.X, trap.Tile.Y) == 0 {
			continue
		}
		state := &w.vaultFire[i]
		if state.TriggeredTick == 0 || w.tick-state.TriggeredTick >= VaultTrapCooldownTicks {
			for _, actor := range actors {
				if actor.player == nil || actor.dead() || actor.player.Flight.Flying {
					continue
				}
				start, end := position{actor.start.x + actor.size/2, actor.start.y + actor.size/2}, position{actor.end.x + actor.size/2, actor.end.y + actor.size/2}
				tile := position{float64(trap.Tile.X * worldmap.Tile), float64(trap.Tile.Y * worldmap.Tile)}
				if _, hit := eatingContact(start, end, 1, tile, tile, worldmap.Tile); hit {
					state.TriggeredTick = w.tick
					w.vaultFireHits[i] = make(map[string]bool)
					break
				}
			}
		}
		if state.TriggeredTick == 0 {
			continue
		}
		age := w.tick - state.TriggeredTick
		if age < VaultTrapWarningTicks || age >= VaultTrapWarningTicks+VaultTrapFireTicks {
			continue
		}
		for _, actor := range actors {
			if actor.player == nil || actor.dead() || actor.protectedAt(1, w.safeZones) || w.vaultFireHits[i][actor.id] {
				continue
			}
			delta := physics.Vector{X: actor.end.x - actor.start.x, Y: actor.end.y - actor.start.y}
			world := physics.World{CanOccupy: func(body physics.Rect) bool { return !body.Overlaps(trap.Lane) }}
			if _, blocked := world.MovementLimit(physics.Square(actor.start.x, actor.start.y, actor.size), delta); !blocked {
				continue
			}
			w.vaultFireHits[i][actor.id] = true
			hp := actor.stats.HP
			actor.stats.HP = math.Max(0, hp-actor.stats.MaxHP()*.3)
			if actor.stats.HP == 0 {
				w.dropKillLoot(actor, hp, actor.ownerStats().MP)
				actor.die("vault fire", w.tick)
			}
		}
	}
}
