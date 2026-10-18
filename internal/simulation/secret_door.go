package simulation

import (
	"math"
	"slices"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/physics"
	"dungeon-raylib-go/internal/worldmap"
)

func NearWingDoor(x, y float64) bool {
	for _, door := range worldmap.Default().SecretDoors() {
		if door.Interactable && nearSecretDoor(x, y, door) {
			return true
		}
	}
	return false
}

func nearSecretDoor(x, y float64, door worldmap.SecretDoor) bool {
	bounds := door.Area.Bounds()
	cx, cy := x+parameters.PlayerSize/2, y+parameters.PlayerSize/2
	return math.Hypot(cx-max(bounds.X, min(bounds.X+bounds.W, cx)), cy-max(bounds.Y, min(bounds.Y+bounds.H, cy))) <= 65
}

func (w *World) interactSecret(p *playerState) bool {
	if p.Dead || p.Flight.Flying {
		return false
	}
	for _, door := range worldmap.Default().SecretDoors() {
		if !door.Interactable || slices.Contains(w.openSecrets, door.ID) || !nearSecretDoor(p.X, p.Y, door) {
			continue
		}
		w.openSecrets = append(w.openSecrets, door.ID)
		p.OpenSecrets = w.openSecrets
		return true
	}
	return false
}

func (w *World) advanceWingDoor() {
	if slices.Contains(w.openSecrets, worldmap.WingDoor) {
		w.wingDoorOpen = min(1, w.wingDoorOpen+parameters.StepSeconds/1.1)
	}

}

func constrainWingDoor(p *Player, oldX, oldY float64) {
	if p.Flight.Flying {
		return
	}
	for _, door := range worldmap.Default().SecretDoors() {
		bounds := door.Area.Bounds()
		if door.Interactable {
			border := door.SanctuaryBorder()
			right, bottom := max(bounds.X+bounds.W, border.X+border.W), max(bounds.Y+bounds.H, border.Y+border.H)
			bounds.X, bounds.Y = min(bounds.X, border.X), min(bounds.Y, border.Y)
			bounds.W, bounds.H = right-bounds.X, bottom-bounds.Y
		}
		if door.Interactable && p.WingDoorOpen < 1 && physics.Square(p.X, p.Y, parameters.PlayerSize).Overlaps(bounds) {
			p.X, p.Y = oldX, oldY
			return
		}
	}
}

func (w *World) revealFlownSecrets(p *playerState) {
	if !p.Flight.Flying {
		return
	}
	body := physics.Square(p.X, p.Y, parameters.PlayerSize)
	for _, door := range worldmap.Default().SecretDoors() {
		if slices.Contains(w.openSecrets, door.ID) {
			continue
		}
		room, _ := worldmap.Default().Room(door.RoomID)
		if body.Overlaps(door.Area.Bounds()) || body.Overlaps(room.Bounds()) {
			w.openSecrets = append(w.openSecrets, door.ID)
			p.OpenSecrets = w.openSecrets
		}
	}
}
