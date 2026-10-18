package simulation

import (
	"math"
	"slices"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
	"2d-dungeon/internal/worldmap"
)

type OrbState struct {
	Seals               [4]bool `json:"seals"`
	GateGuardsRemaining int     `json:"gate_guards_remaining"`
	GateUnlocked        bool    `json:"gate_unlocked"`
	GateActivated       bool    `json:"gate_activated"`
	GateOpen            float64 `json:"gate_open"`
	Consumed            bool    `json:"consumed,omitempty"`
	FrenzyAt            uint64  `json:"frenzy_at,omitempty"`
	X                   float64 `json:"x"`
	Y                   float64 `json:"y"`
	Unlocked            bool    `json:"unlocked"`
	Carrier             string  `json:"carrier,omitempty"`
	GuardsRemaining     int     `json:"guards_remaining"`
}

func (w *World) revealSecret(x, y, dx, dy, size float64) {
	geometry := worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen)
	fraction, blocked := geometry.MovementLimit(x, y, dx, dy, size, size)
	if !blocked {
		return
	}
	body := physics.Rect{X: x + dx*min(1, fraction+.001), Y: y + dy*min(1, fraction+.001), W: size, H: size}
	for _, door := range worldmap.Default().SecretDoors() {
		if door.Interactable || slices.Contains(w.openSecrets, door.ID) || !body.Overlaps(door.Area.Bounds()) {
			continue
		}
		w.secretWallHits[door.ID]++
		if w.secretWallHits[door.ID] < parameters.SecretWallHits {
			return
		}
		w.openSecrets = append(w.openSecrets, door.ID)
		return
	}
}

func (w *World) updateOrb() {
	remaining, gateRemaining := 0, 0
	room, _ := worldmap.Default().Room(worldmap.OrbVault)
	for _, ai := range w.ais {
		if ai.Type == GateHeavy && ai.respawnTick == 0 && ai.Stats.HP > 0 {
			gateRemaining++
		}
		if ai.den.RoomID == worldmap.OrbVault {
			for i, encounter := range room.FeedingAreas {
				if encounter.ID == ai.den.ID {
					w.orb.Seals[i] = ai.respawnTick != 0 || ai.Stats.HP <= 0
				}
			}
		}
		if ai.den.RoomID == worldmap.OrbVault && ai.respawnTick == 0 && ai.Stats.HP > 0 {
			remaining++
		}
	}
	w.orb.GateGuardsRemaining = gateRemaining
	w.orb.GateUnlocked = gateRemaining == 0
	w.orb.GuardsRemaining = remaining
	if remaining == 0 {
		w.orb.Unlocked = true
	}
	for _, p := range w.players {
		if p.Dead {
			w.releaseOrb(p.ID)
		}
	}
}

func (w *World) releaseOrb(id string) {
	if w.orb.Carrier == id {
		w.orb.Carrier = ""
		w.orb.FrenzyAt = 0
	}
	if p := w.players[id]; p != nil {
		p.HasOrb = false
		p.TeleportCharge = 0
	}
}

func (w *World) interactObjective(p *playerState) bool {
	if p.Dead || p.Flight.Flying {
		return false
	}
	geometry := worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen)
	x, y := p.X+PlayerSize/2, p.Y+PlayerSize/2
	if !w.orb.Consumed && math.Hypot(x-w.orb.X, y-w.orb.Y) <= 85 && geometry.Visible(x, y, w.orb.X, w.orb.Y) {
		if w.orb.Unlocked && w.orb.Carrier == "" && !p.HasOrb {
			w.orb.Carrier = p.ID
			w.orb.FrenzyAt = w.tick + OrbFrenzyDelayTicks
			p.HasOrb = true
			w.vaultArrowStarted = w.tick
		}
		return true
	}
	portal, distance, found := nearestPortal(x, y)
	if found && distance <= 60 {
		if p.portalCooldown == 0 && (portal.ID != worldmap.DescentPortal || w.orb.Consumed || p.HasOrb) {
			p.TeleportCharge = StepSeconds
		}
		return true
	}
	return false
}
