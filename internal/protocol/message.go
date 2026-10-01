package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"raylib-2d-sandbox/internal/simulation"
)

const (
	TypeWelcome    = "welcome"
	TypeSnapshot   = "snapshot"
	TypeEcho       = "echo"
	TypeInput      = "input"
	TypeChat       = "chat"
	TypePerkChoice = "perk_choice"
	TypeCloseOffer = "close_offer"
	TypeRespawn    = "respawn"
)

type Direction struct {
	Shoot    bool    `json:"shoot,omitempty"`
	Interact bool    `json:"interact,omitempty"`
	Boost    bool    `json:"boost,omitempty"`
	AimX     float64 `json:"aim_x,omitempty"`
	AimY     float64 `json:"aim_y,omitempty"`
	Dash     bool    `json:"dash,omitempty"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

type ClientMessage struct {
	Perk      *PerkSelection `json:"perk,omitempty"`
	Sequence  uint64         `json:"sequence,omitempty"`
	Type      string         `json:"type"`
	Direction *Direction     `json:"direction,omitempty"`
	Message   string         `json:"message,omitempty"`
}

type PerkSelection struct {
	StatueID string `json:"statue_id"`
	Cycle    uint64 `json:"cycle"`
	Card     int    `json:"card"`
}

func DecodeClientMessage(data []byte) (ClientMessage, error) {
	var message *ClientMessage
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&message); err != nil {
		return ClientMessage{}, err
	}
	if message == nil {
		return ClientMessage{}, errors.New("message must be an object")
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return ClientMessage{}, errors.New("expected one JSON object")
	}
	switch message.Type {
	case TypePerkChoice:
		if message.Perk == nil || message.Direction != nil || message.Message != "" || message.Sequence != 0 || message.Perk.StatueID == "" || len(message.Perk.StatueID) > 96 || message.Perk.Cycle == 0 || message.Perk.Card < 0 || message.Perk.Card > 2 {
			return ClientMessage{}, errors.New("invalid perk selection")
		}
	case TypeCloseOffer, TypeRespawn:
		if message.Perk != nil || message.Direction != nil || message.Message != "" || message.Sequence != 0 {
			return ClientMessage{}, errors.New("invalid control message")
		}
	case TypeInput:
		if message.Direction == nil || message.Message != "" || message.Perk != nil {
			return ClientMessage{}, errors.New("input requires a direction")
		}
		d := simulation.Direction{X: message.Direction.X, Y: message.Direction.Y, AimX: message.Direction.AimX, AimY: message.Direction.AimY}
		if !d.Valid() {
			return ClientMessage{}, errors.New("direction components must be between -1 and 1")
		}
	case TypeChat:
		if message.Direction != nil || message.Sequence != 0 || message.Perk != nil {
			return ClientMessage{}, errors.New("chat cannot contain direction")
		}
		if NormalizeChatMessage(message.Message) == "" {
			return ClientMessage{}, errors.New("chat requires a message")
		}
	case TypeEcho, "":
		if message.Direction != nil || message.Sequence != 0 || message.Perk != nil {
			return ClientMessage{}, errors.New("echo cannot contain direction")
		}
	default:
		return ClientMessage{}, errors.New("unsupported message type")
	}
	return *message, nil
}

type PlayerStats = simulation.Stats

type Player struct {
	ActiveStatue   string                       `json:"active_statue,omitempty"`
	EntranceGate   simulation.EntranceGateState `json:"entrance_gate"`
	ShotCooldown   float64                      `json:"shot_cooldown,omitempty"`
	EjectCooldown  float64                      `json:"eject_cooldown,omitempty"`
	Character      simulation.Character         `json:"character"`
	Split          simulation.SplitCells        `json:"split"`
	GatesClosed    bool                         `json:"gates_closed,omitempty"`
	InArena        bool                         `json:"in_arena,omitempty"`
	TeleportCharge float64                      `json:"teleport_charge,omitempty"`
	Combat         simulation.CombatState       `json:"combat"`
	Name           string                       `json:"name"`
	Dead           bool                         `json:"dead"`
	KilledBy       string                       `json:"killed_by,omitempty"`
	Stats          PlayerStats                  `json:"stats"`
	InputSeq       uint64                       `json:"input_seq"`
	ID             string                       `json:"id"`
	X              float64                      `json:"x"`
	Y              float64                      `json:"y"`
	Size           float64                      `json:"size"`
}

type Fruit struct {
	Source string  `json:"source,omitempty"`
	HP     float64 `json:"hp,omitempty"`
	MP     float64 `json:"mp,omitempty"`
	ID     string  `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Size   float64 `json:"size"`
}

type AI struct {
	Combat    simulation.CombatState `json:"combat"`
	Behavior  string                 `json:"behavior"`
	Character simulation.Character   `json:"character"`
	Split     simulation.SplitCells  `json:"split"`
	ID        string                 `json:"id"`
	X         float64                `json:"x"`
	Y         float64                `json:"y"`
	Size      float64                `json:"size"`
	Stats     PlayerStats            `json:"stats"`
}

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
