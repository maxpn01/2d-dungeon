package protocol

import "raylib-2d-sandbox/internal/simulation"

type Message struct {
	DungeonLayout     string                   `json:"dungeon_layout,omitempty"`
	WingDoorOpen      float64                  `json:"wing_door_open"`
	NorthWingDoorOpen float64                  `json:"north_wing_door_open"`
	WinnerID          string                   `json:"winner_id,omitempty"`
	OpenSecrets       []string                 `json:"open_secrets"`
	Orb               simulation.OrbState      `json:"orb"`
	Statues           []simulation.StatueState `json:"statues"`
	Projectiles       []simulation.Projectile  `json:"projectiles"`
	SafeZones         []simulation.SafeZone    `json:"safe_zones"`
	Boss              simulation.Boss          `json:"boss"`
	AIs               *[]AI                    `json:"ais,omitempty"`
	Fruits            *[]Fruit                 `json:"fruits,omitempty"`
	Type              string                   `json:"type"`
	PlayerID          string                   `json:"player_id,omitempty"`
	Tick              uint64                   `json:"tick"`
	Players           *[]Player                `json:"players,omitempty"`
	Message           string                   `json:"message,omitempty"`
	Name              string                   `json:"name,omitempty"`
}
