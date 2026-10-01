package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"raylib-2d-sandbox/internal/simulation"
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

func (d Direction) Simulation() simulation.Direction {
	return simulation.Direction{Shoot: d.Shoot, X: d.X, Y: d.Y, Interact: d.Interact, Boost: d.Boost, Dash: d.Dash, AimX: d.AimX, AimY: d.AimY}
}
