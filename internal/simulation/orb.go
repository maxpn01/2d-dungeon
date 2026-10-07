package simulation

import (
	"math"
	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
	"slices"
)

type OrbState struct {
	Consumed        bool    `json:"consumed,omitempty"`
	FrenzyAt        uint64  `json:"frenzy_at,omitempty"`
	X               float64 `json:"x"`
	Y               float64 `json:"y"`
	Unlocked        bool    `json:"unlocked"`
	Carrier         string  `json:"carrier,omitempty"`
	GuardsRemaining int     `json:"guards_remaining"`
}

func (w *World) revealSecret(x, y, dx, dy, size float64) {
	geometry := worldmap.Default().WithSecrets(w.openSecrets)
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
	remaining := 0
	for _, ai := range w.ais {
		if ai.den.RoomID == worldmap.OrbVault && ai.respawnTick == 0 && ai.Stats.HP > 0 {
			remaining++
		}
	}
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
	geometry := worldmap.Default().WithSecrets(w.openSecrets)
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
