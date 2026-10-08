package simulation

import (
	"fmt"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/worldmap"
)

func (w *World) advanceHeavyStomps() {
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.Type != DungeonHeavy || ai.respawnTick != 0 || ai.Stats.HP <= 0 || ai.Combat.StompRemaining <= 0 {
			continue
		}
		before := ai.Combat.StompRemaining
		ai.Combat.StompRemaining = max(0, before-StepSeconds)
		impact := parameters.HeavyStompDuration - parameters.HeavyStompWindup
		if before <= impact || ai.Combat.StompRemaining > impact {
			continue
		}
		if len(w.projectiles) >= parameters.MaxProjectiles {
			continue
		}
		size := parameters.HeavyStompWaveSize
		x, y := ai.X+(AISize-size)/2, ai.Y+(AISize-size)/2
		if !worldmap.Default().WithSecrets(w.openSecrets).CanOccupy(ProjectileWallBody(Projectile{Kind: "shockwave", X: x, Y: y, Size: size})) {
			continue
		}
		w.nextProjectile++
		w.projectiles = append(w.projectiles, Projectile{
			Kind: "shockwave", ID: fmt.Sprintf("stomp-%d", w.nextProjectile), Owner: ai.ID,
			X: x, Y: y, originX: x, originY: y, MobOwned: true,
			Direction: ai.Combat.StompDirection, Size: size,
			Remaining: parameters.HeavyStompWaveLifetime, Damage: ai.Stats.Damage * parameters.HeavyStompDamageMultiplier,
		})
	}
}
