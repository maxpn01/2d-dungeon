package simulation

import (
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
	"slices"
)

func (w *World) EnableTesting(id string) bool {
	p := w.players[id]
	if p == nil || worldmap.Multiplayer() || len(w.players) != 1 {
		return false
	}
	p.testing, p.god = true, true
	p.Stats = NewStatsAtLevel(10)
	w.refillTestingPlayer(p)
	return true
}

func (w *World) TestingEnabled(id string) bool { p := w.players[id]; return p != nil && p.testing }
func (w *World) GodMode(id string) bool        { p := w.players[id]; return p != nil && p.testing && p.god }

func (w *World) refillTestingPlayer(p *playerState) {
	p.Stats.HP = max(0, p.Stats.MaxHP()-p.Split.Mass())
	p.Stats.MP = p.Stats.MaxMP()
	p.Dead, p.KilledBy, p.progressionReset = false, "", false
}

func (w *World) ToggleTestingGod(id string) {
	p := w.players[id]
	if p == nil || !p.testing {
		return
	}
	p.god = !p.god
	if p.god {
		w.refillTestingPlayer(p)
	}
}

func (w *World) CycleTestingLevel(id string) {
	p := w.players[id]
	if p == nil || !p.testing {
		return
	}
	levels := []int{1, 5, 10, 20, MaxLevel}
	index := slices.Index(levels, p.Stats.Level)
	p.Stats = NewStatsAtLevel(levels[(index+1)%len(levels)])
	p.Split = nil
	w.refillTestingPlayer(p)
}

func (w *World) ToggleTestingWings(id string) {
	p := w.players[id]
	if p == nil || !p.testing {
		return
	}
	if p.HasWings && !worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen).Walkable(p.X, p.Y, PlayerSize, PlayerSize) {
		return
	}
	p.HasWings = !p.HasWings
	p.Flight = FlightState{}
	p.Combat = CombatState{Facing: p.Combat.Facing}
}

func (w *World) TestingWarp(id, roomID string) bool {
	p := w.players[id]
	if p == nil || !p.testing {
		return false
	}
	room, ok := worldmap.Default().Room(roomID)
	if !ok {
		return false
	}
	for _, door := range worldmap.Default().SecretDoors() {
		if door.RoomID == roomID && !slices.Contains(w.openSecrets, door.ID) {
			w.openSecrets = append(w.openSecrets, door.ID)
		}
	}
	if roomID == worldmap.SecretLibrary {
		for i := range w.statues {
			if IsLibraryBook(w.statues[i].ID) && LibraryRoom(w.statues[i].ID) == roomID {
				w.statues[i].Library = LibraryState{}
				w.statues[i].ClaimedBy = ""
			}
		}
		p.Stats.SpectralShots = false
	}
	at := room.WorldCell(worldmap.At(room.Area.Width/2, room.Area.Height/2))
	if roomID == worldmap.OrbVault || roomID == worldmap.VaultCourt {
		if roomID == worldmap.OrbVault {
			at, _ = worldmap.Default().VaultApproach()
		}
		for i := range w.ais {
			ai := &w.ais[i]
			if ai.Type != GateHeavy {
				continue
			}
			if roomID == worldmap.VaultCourt {
				oldID := ai.ID
				*ai = w.newDenAI(ai.den)
				ai.ID = oldID
			} else {
				ai.Stats.HP, ai.respawnTick = 0, w.tick+1
			}
		}
		w.orb.GateActivated = roomID == worldmap.OrbVault
		w.orb.GateOpen = 0
		if w.orb.GateActivated {
			w.orb.GateOpen = 1
		}
		for i := range w.ais {
			ai := &w.ais[i]
			if ai.den.RoomID == worldmap.OrbVault {
				oldID := ai.ID
				*ai = w.newDenAI(ai.den)
				ai.ID = oldID
			}
		}
		remaining := w.fruits[:0]
		for _, fruit := range w.fruits {
			vaultLoot := false
			for _, ai := range w.ais {
				if (ai.den.RoomID == worldmap.OrbVault || ai.Type == GateHeavy) && fruit.Source == ai.ID {
					vaultLoot = true
					break
				}
			}
			if !vaultLoot {
				remaining = append(remaining, fruit)
			}
		}
		w.fruits = remaining
		w.vaultFire = [worldmap.VaultFireTrapCount]VaultFire{}
		w.vaultArrowStarted = 0
		w.vaultFireHits = [worldmap.VaultFireTrapCount]map[string]bool{}
		w.releaseOrb(id)
		w.orb.Unlocked = false
		w.orb.Consumed = false
	}
	if roomID == worldmap.TeleportSanctum && !w.orb.Consumed {
		for i := range w.ais {
			if w.ais[i].den.RoomID == worldmap.OrbVault {
				w.ais[i].Stats.HP = 0
				w.ais[i].respawnTick = w.tick + 1
			}
		}
		w.updateOrb()
		w.orb.Carrier, p.HasOrb = id, true
		w.orb.FrenzyAt = w.tick + OrbFrenzyDelayTicks
	}
	x, y := worldmap.CellPosition(at.X, at.Y, PlayerSize)
	geometry := worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen)
	if !geometry.CanOccupy(physics.Square(x, y, PlayerSize)) {
		return false
	}
	p.X, p.Y = x, y
	p.Flight, p.Combat, p.Split = FlightState{}, CombatState{}, nil
	p.ActiveStatue, p.TeleportCharge, p.portalCooldown = "", 0, 0
	p.direction, p.commands = Direction{}, nil
	p.velocityX, p.velocityY = 0, 0
	p.VaultGateOpen = w.orb.GateOpen
	p.OpenSecrets = w.openSecrets
	p.Escaped, p.GatesClosed, p.InArena = false, false, false
	w.winnerID, w.finishedTick = "", 0
	w.projectiles = nil
	w.refillTestingPlayer(p)
	w.updateOrb()
	return true
}
