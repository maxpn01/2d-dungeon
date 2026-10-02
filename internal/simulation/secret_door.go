package simulation

import (
	"math"
	"slices"

	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
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
	cx, cy := x+PlayerSize/2, y+PlayerSize/2
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
		w.wingDoorOpen = min(1, w.wingDoorOpen+StepSeconds/1.1)
	}
	if slices.Contains(w.openSecrets, worldmap.NorthID(worldmap.WingDoor)) {
		w.northWingDoorOpen = min(1, w.northWingDoorOpen+StepSeconds/1.1)
	}
}

func WingDoorProgress(south, north float64, id string) float64 {
	if id == worldmap.NorthID(worldmap.WingDoor) {
		return north
	}
	return south
}

func constrainWingDoor(p *Player, oldX, oldY float64) {
	if p.Flight.Flying {
		return
	}
	for _, door := range worldmap.Default().SecretDoors() {
		bounds := door.Area.Bounds()
		if door.Area.Bottom() == 0 {
			bounds.H += worldmap.SurfaceMin
		}
		if door.Area.Right() == 0 {
			bounds.W += worldmap.SurfaceMin
		}
		if door.Interactable && WingDoorProgress(p.WingDoorOpen, p.NorthWingDoorOpen, door.ID) < 1 && physics.Square(p.X, p.Y, PlayerSize).Overlaps(bounds) {
			p.X, p.Y = oldX, oldY
			return
		}
	}
}

func (w *World) revealFlownSecrets(p *playerState) {
	if !p.Flight.Flying {
		return
	}
	body := physics.Square(p.X, p.Y, PlayerSize)
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
