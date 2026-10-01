package protocol

import "raylib-2d-sandbox/internal/simulation"

type Message struct {
	Statues     []simulation.StatueState `json:"statues"`
	Projectiles []simulation.Projectile  `json:"projectiles"`
	SafeZones   []simulation.SafeZone    `json:"safe_zones"`
	Boss        simulation.Boss          `json:"boss"`
	AIs         *[]AI                    `json:"ais,omitempty"`
	Fruits      *[]Fruit                 `json:"fruits,omitempty"`
	Type        string                   `json:"type"`
	PlayerID    string                   `json:"player_id,omitempty"`
	Tick        uint64                   `json:"tick"`
	Players     *[]Player                `json:"players,omitempty"`
	Message     string                   `json:"message,omitempty"`
	Name        string                   `json:"name,omitempty"`
}
