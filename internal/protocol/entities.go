package protocol

import "raylib-2d-sandbox/internal/simulation"

type PlayerStats = simulation.Stats

type Player struct {
	OpenSecrets    []string                     `json:"open_secrets"`
	HasOrb         bool                         `json:"has_orb"`
	Escaped        bool                         `json:"escaped"`
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
