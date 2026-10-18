package game

import (
	"2d-dungeon/internal/rendering"
	"2d-dungeon/internal/worldmap"
)

func (w *World) updateDiscoveries(dt float32) {
	if w.Dead() || w.StatsOpen() || w.PerksOpen() {
		return
	}
	center := w.Center()
	room, inside := worldmap.Default().RoomAt(float64(center.X), float64(center.Y))
	if inside && room.ID == worldmap.SecretLibrary && !w.libraryDiscovered {
		w.libraryDiscovered = true
		rendering.PlayLibraryDiscoverySound()
	}
	if inside && room.ID == worldmap.WingChamber && !w.wingRoomDiscovered {
		w.wingRoomDiscovered = true
		rendering.PlayLibraryDiscoverySound()
	}
	portal, exists := worldmap.Default().Portal(worldmap.DescentPortal)
	discovered := inside && room.ID == worldmap.TeleportSanctum && exists && w.gameMap.Exploration.Visible([2]int{portal.At.X, portal.At.Y})
	w.teleporterDiscovery.Update(dt, discovered, w.player.HasOrb || w.snapshot.Orb.Consumed)
}
