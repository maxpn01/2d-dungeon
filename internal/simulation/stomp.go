package simulation

import (
	"fmt"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/worldmap"
)

type MobStompAttack struct {
	Reach, WaveSize, DamageMultiplier, StunDuration float64
}

func MobStomp(t MobType) MobStompAttack {
	if t == GateHeavy {
		return MobStompAttack{Reach: parameters.GateHeavyStompReach, WaveSize: parameters.GateHeavyStompWaveSize, DamageMultiplier: parameters.GateHeavyStompDamageMultiplier, StunDuration: parameters.GateHeavyStompStun}
	}
	return MobStompAttack{Reach: parameters.HeavyStompReach, WaveSize: parameters.HeavyStompWaveSize, DamageMultiplier: parameters.HeavyStompDamageMultiplier, StunDuration: parameters.HeavyStompStun}
}

func (w *World) advanceHeavyStomps() {
	for i := range w.ais {
		ai := &w.ais[i]
		if !ai.Type.IsHeavy() || ai.respawnTick != 0 || ai.Stats.HP <= 0 || ai.Combat.StompRemaining <= 0 {
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
		stomp := MobStomp(ai.Type)
		size := stomp.WaveSize
		x, y := ai.X+(AISize-size)/2, ai.Y+(AISize-size)/2
		if !worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen).CanOccupy(ProjectileWallBody(Projectile{Kind: "shockwave", X: x, Y: y, Size: size})) {
			continue
		}
		w.nextProjectile++
		w.projectiles = append(w.projectiles, Projectile{
			Kind: "shockwave", ID: fmt.Sprintf("stomp-%d", w.nextProjectile), Owner: ai.ID,
			X: x, Y: y, originX: x, originY: y, MobOwned: true,
			Direction: ai.Combat.StompDirection, Size: size,
			Remaining: stomp.Reach / parameters.HeavyStompWaveSpeed, Lifetime: stomp.Reach / parameters.HeavyStompWaveSpeed,
			Damage: ai.Stats.Damage * stomp.DamageMultiplier, StunDuration: stomp.StunDuration,
		})
	}
}
