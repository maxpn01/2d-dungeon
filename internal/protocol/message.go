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
)

type Direction struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type ClientMessage struct {
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
	case TypeEcho, "": // Preserve the original echo client format.
		if message.Direction != nil {
			return ClientMessage{}, errors.New("echo cannot contain direction")
		}
	default:
		return ClientMessage{}, errors.New("unsupported message type")
	}
	return *message, nil
}

type Player struct {
	ID   string  `json:"id"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Size float64 `json:"size"`
}

type Message struct {
	Type     string    `json:"type"`
	PlayerID string    `json:"player_id,omitempty"`
	Tick     uint64    `json:"tick"`
	Players  *[]Player `json:"players,omitempty"`
	Message  string    `json:"message,omitempty"`
}
