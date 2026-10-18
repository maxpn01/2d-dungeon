package simulation

import (
	"fmt"
	"math"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
	"2d-dungeon/internal/worldmap"
)

const VaultArrowDurationTicks = 10 * TickRate
const VaultArrowIntervalTicks = TickRate / 2
const VaultArrowSpeed = 1080.0

func ProjectileSpeed(shot Projectile) float64 {
	if shot.Kind == "arrow" {
		return VaultArrowSpeed
	}
	if shot.Kind == "shockwave" {
		return parameters.HeavyStompWaveSpeed
	}
	if shot.MobOwned {
		return parameters.ShotSpeed * parameters.MobShotSpeedMultiplier
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
	geometry := worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen)
	ports := geometry.VaultArrowPorts()
	for _, actor := range actors {
		if actor.player == nil || actor.dead() || !geometry.RoomContains(worldmap.OrbVault, physics.Square(actor.end.x, actor.end.y, actor.size)) {
			continue
		}
		center := physics.Vector{X: actor.end.x + actor.size/2, Y: actor.end.y + actor.size/2}
		velocity := physics.Vector{X: (actor.end.x - actor.start.x) / StepSeconds, Y: (actor.end.y - actor.start.y) / StepSeconds}
		for side := 0; side < 2; side++ {
			nearest := -1
			best := math.Inf(1)
			var aim physics.Vector
			for i := side * len(ports) / 2; i < (side+1)*len(ports)/2; i++ {
				predicted := vaultArrowIntercept(ports[i], center, velocity)
				if room, ok := geometry.Room(worldmap.OrbVault); ok {
					body := physics.Square(predicted.X-actor.size/2, predicted.Y-actor.size/2, actor.size).ClampedTo(room.Bounds())
					predicted = body.Center()
				}
				if distance := math.Abs(ports[i].Y - predicted.Y); distance < best {
					nearest, best, aim = i, distance, predicted
				}
			}
			if nearest < 0 || len(w.projectiles) >= parameters.MaxProjectiles {
				continue
			}
			at := ports[nearest]
			dx, dy := aim.X-at.X, aim.Y-at.Y
			length := math.Hypot(dx, dy)
			if length < 1 {
				continue
			}
			w.nextProjectile++
			w.projectiles = append(w.projectiles, Projectile{Kind: "arrow", ID: fmt.Sprintf("arrow-%d", w.nextProjectile), Owner: "vault arrows", X: at.X - parameters.ShotSize/2, Y: at.Y - parameters.ShotSize/2, Direction: Direction{X: dx / length, Y: dy / length}, Size: parameters.ShotSize, Remaining: parameters.ShotLifetime, Damage: 8})
		}
	}
}

func vaultArrowIntercept(origin, target, velocity physics.Vector) physics.Vector {
	offset := target.Sub(origin)
	a := velocity.LengthSquared() - VaultArrowSpeed*VaultArrowSpeed
	b := 2 * (offset.X*velocity.X + offset.Y*velocity.Y)
	c := offset.LengthSquared()
	seconds := math.Inf(1)
	if math.Abs(a) < 1e-6 {
		if b < -1e-6 {
			seconds = -c / b
		}
	} else if discriminant := b*b - 4*a*c; discriminant >= 0 {
		root := math.Sqrt(discriminant)
		for _, candidate := range []float64{(-b - root) / (2 * a), (-b + root) / (2 * a)} {
			if candidate >= 0 && candidate < seconds {
				seconds = candidate
			}
		}
	}
	if math.IsInf(seconds, 1) {
		seconds = offset.Length() / VaultArrowSpeed
	}
	return target.Add(velocity.Scale(min(seconds, parameters.ShotLifetime)))
}
