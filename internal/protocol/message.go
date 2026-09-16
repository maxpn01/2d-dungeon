package protocol

const (
	TypeWelcome  = "welcome"
	TypeSnapshot = "snapshot"
	TypeEcho     = "echo"
)

type Player struct {
	ID string `json:"id"`
}

type Message struct {
	Type     string    `json:"type"`
	PlayerID string    `json:"player_id,omitempty"`
	Players  *[]Player `json:"players,omitempty"`
	Message  string    `json:"message,omitempty"`
}
