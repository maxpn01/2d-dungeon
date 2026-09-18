package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"raylib-2d-sandbox/internal/simulation"
)

const (
	TypeWelcome  = "welcome"
	TypeSnapshot = "snapshot"
	TypeEcho     = "echo"
	TypeInput    = "input"
	TypeChat     = "chat"
)

type Direction struct {
	Dash bool    `json:"dash,omitempty"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

type ClientMessage struct {
	Sequence  uint64     `json:"sequence,omitempty"`
	Type      string     `json:"type"`
	Direction *Direction `json:"direction,omitempty"`
	Message   string     `json:"message,omitempty"`
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
	case TypeInput:
		if message.Direction == nil || message.Message != "" {
			return ClientMessage{}, errors.New("input requires a direction")
		}
		d := simulation.Direction{X: message.Direction.X, Y: message.Direction.Y}
		if !d.Valid() {
			return ClientMessage{}, errors.New("direction components must be between -1 and 1")
		}
	case TypeChat:
		if message.Direction != nil || message.Sequence != 0 {
			return ClientMessage{}, errors.New("chat cannot contain direction")
		}
		if NormalizeChatMessage(message.Message) == "" {
			return ClientMessage{}, errors.New("chat requires a message")
		}
	case TypeEcho, "":
		if message.Direction != nil || message.Sequence != 0 {
			return ClientMessage{}, errors.New("echo cannot contain direction")
		}
	default:
		return ClientMessage{}, errors.New("unsupported message type")
	}
	return *message, nil
}

type PlayerStats struct {
	Level int     `json:"level"`
	XP    float64 `json:"xp"`
	HP    float64 `json:"hp"`
	Speed float64 `json:"speed"`
}

type Player struct {
	Combat   simulation.CombatState `json:"combat"`
	Name     string                 `json:"name"`
	Dead     bool                   `json:"dead"`
	KilledBy string                 `json:"killed_by,omitempty"`
	Stats    PlayerStats            `json:"stats"`
	InputSeq uint64                 `json:"input_seq"`
	ID       string                 `json:"id"`
	X        float64                `json:"x"`
	Y        float64                `json:"y"`
	Size     float64                `json:"size"`
}

type Fruit struct {
	Source string  `json:"source,omitempty"`
	HP     float64 `json:"hp,omitempty"`
	ID     string  `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Size   float64 `json:"size"`
}

type AI struct {
	Combat   simulation.CombatState `json:"combat"`
	Behavior string                 `json:"behavior"`
	ID       string                 `json:"id"`
	X        float64                `json:"x"`
	Y        float64                `json:"y"`
	Size     float64                `json:"size"`
	Stats    PlayerStats            `json:"stats"`
}

type Message struct {
	AIs      *[]AI     `json:"ais,omitempty"`
	Fruits   *[]Fruit  `json:"fruits,omitempty"`
	Type     string    `json:"type"`
	PlayerID string    `json:"player_id,omitempty"`
	Tick     uint64    `json:"tick"`
	Players  *[]Player `json:"players,omitempty"`
	Message  string    `json:"message,omitempty"`
	Name     string    `json:"name,omitempty"`
}
