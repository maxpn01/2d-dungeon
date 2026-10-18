package simulation

import (
	"slices"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
	"2d-dungeon/internal/worldmap"
)

func (w *World) EnableTesting() {
	p := w.player
	p.testing, p.god = true, true
	p.Stats = NewStatsAtLevel(10)
	w.refillTestingPlayer(p)
}

func (w *World) TestingEnabled() bool { return w.player.testing }
func (w *World) GodMode() bool        { return w.player.testing && w.player.god }

func (w *World) refillTestingPlayer(p *playerState) {
	p.Stats.HP = p.Stats.MaxHP()
	p.Stats.MP = p.Stats.MaxMP()
	p.Dead, p.KilledBy, p.progressionReset = false, "", false
}

func (w *World) ToggleTestingGod() {
	p := w.player
	if !p.testing {
		return
	}
	p.god = !p.god
	if p.god {
		w.refillTestingPlayer(p)
	}
}

func (w *World) CycleTestingLevel() {
	p := w.player
	if !p.testing {
		return
	}
	levels := []int{1, 5, 10, 20, parameters.MaxLevel}
	index := slices.Index(levels, p.Stats.Level)
	p.Stats = NewStatsAtLevel(levels[(index+1)%len(levels)])
	w.refillTestingPlayer(p)
}

func (w *World) ToggleTestingWings() {
	p := w.player
	if !p.testing {
		return
	}
	if p.HasWings && !w.geometry().Walkable(p.X, p.Y, parameters.PlayerSize, parameters.PlayerSize) {
		return
	}
	p.HasWings = !p.HasWings
	p.Flight = FlightState{}
	p.Combat = CombatState{Facing: p.Combat.Facing}
}

func (w *World) TestingWarp(roomID string) bool {
	p := w.player
	if !p.testing {
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
		w.releaseOrb(p.ID)
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
		w.orb.Carrier, p.HasOrb = p.ID, true
		w.orb.FrenzyAt = w.tick + OrbFrenzyDelayTicks
	}
	x, y := worldmap.CellPosition(at.X, at.Y, parameters.PlayerSize)
	geometry := w.geometry()
	if !geometry.CanOccupy(physics.Square(x, y, parameters.PlayerSize)) {
		return false
	}
	p.X, p.Y = x, y
	p.Flight, p.Combat = FlightState{}, CombatState{}
	p.ActiveStatue, p.TeleportCharge, p.portalCooldown = "", 0, 0
	p.direction = Direction{}
	p.velocityX, p.velocityY = 0, 0
	p.VaultGateOpen = w.orb.GateOpen
	p.OpenSecrets = w.openSecrets
	p.FloorCleared, p.FloorClearedAt = false, 0
	w.winnerID, w.finishedTick = "", 0
	w.projectiles = nil
	w.refillTestingPlayer(p)
	w.updateOrb()
	return true
}
