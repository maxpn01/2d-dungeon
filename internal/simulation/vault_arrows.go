package simulation

import (
	"fmt"
	"math"
	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

const VaultArrowDurationTicks = 10 * TickRate
const VaultArrowIntervalTicks = TickRate * 2 / 3
const VaultArrowSpeed = 720.0

func ProjectileSpeed(shot Projectile) float64 {
	if shot.Kind == "arrow" {
		return VaultArrowSpeed
	}
	if shot.Kind == "shockwave" {
		return parameters.HeavyStompWaveSpeed
	}
	return parameters.ShotSpeed
}

func (w *World) advanceVaultArrows(actors []eater) {
	if w.vaultArrowStarted == 0 || w.tick < w.vaultArrowStarted {
		return
	}
	age := w.tick - w.vaultArrowStarted
	if age >= VaultArrowDurationTicks || age < VaultArrowIntervalTicks || age%VaultArrowIntervalTicks != 0 {
		return
	}
	geometry := worldmap.Default().WithSecrets(w.openSecrets)
	ports := geometry.VaultArrowPorts()
	for _, actor := range actors {
		if actor.player == nil || actor.cell != nil || actor.dead() || !geometry.RoomContains(worldmap.OrbVault, physics.Square(actor.end.x, actor.end.y, actor.size)) {
			continue
		}
		for side := 0; side < 2; side++ {
			nearest := -1
			best := math.Inf(1)
			for i := side * len(ports) / 2; i < (side+1)*len(ports)/2; i++ {
				if distance := math.Abs(ports[i].Y - actor.end.y - actor.size/2); distance < best {
					nearest, best = i, distance
				}
			}
			if nearest < 0 || len(w.projectiles) >= parameters.MaxProjectiles {
				continue
			}
			at := ports[nearest]
			dx, dy := actor.end.x+actor.size/2-at.X, actor.end.y+actor.size/2-at.Y
			length := math.Hypot(dx, dy)
			if length < 1 {
				continue
			}
			w.nextProjectile++
			w.projectiles = append(w.projectiles, Projectile{Kind: "arrow", ID: fmt.Sprintf("arrow-%d", w.nextProjectile), Owner: "vault arrows", X: at.X - parameters.ShotSize/2, Y: at.Y - parameters.ShotSize/2, Direction: Direction{X: dx / length, Y: dy / length}, Size: parameters.ShotSize, Remaining: parameters.ShotLifetime, Damage: 8})
		}
	}
}
