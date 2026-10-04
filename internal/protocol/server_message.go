package protocol

import "raylib-2d-sandbox/internal/simulation"

type Message struct {
	SecretWallHits    map[string]int           `json:"secret_wall_hits,omitempty"`
	RoundID           uint64                   `json:"round_id"`
	MatchStarted      bool                     `json:"match_started"`
	BriefingActive    bool                     `json:"briefing_active"`
	StartCountdown    int                      `json:"start_countdown"`
	DungeonLayout     string                   `json:"dungeon_layout,omitempty"`
	WingDoorOpen      float64                  `json:"wing_door_open"`
	NorthWingDoorOpen float64                  `json:"north_wing_door_open"`
	FinishedTick      uint64                   `json:"finished_tick"`
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
