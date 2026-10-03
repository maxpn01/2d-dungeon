package protocol

import "raylib-2d-sandbox/internal/simulation"

type Snapshot struct {
	MatchStarted      bool                `json:"match_started"`
	StartCountdown    int                 `json:"start_countdown"`
	DungeonLayout     string              `json:"dungeon_layout"`
	WingDoorOpen      float64             `json:"wing_door_open"`
	NorthWingDoorOpen float64             `json:"north_wing_door_open"`
	WinnerID          string              `json:"winner_id,omitempty"`
	OpenSecrets       []string            `json:"open_secrets"`
	Orb               simulation.OrbState `json:"orb"`
	Statues           []simulation.StatueState
	Projectiles       []simulation.Projectile
	SafeZones         []simulation.SafeZone
	Boss              simulation.Boss
	AIs               []AI
	Tick              uint64
	Players           []Player
	Fruits            []Fruit
}

func SnapshotFromSimulation(state simulation.Snapshot, names map[string]string) Snapshot {
	snapshot := Snapshot{
		DungeonLayout:     state.DungeonLayout,
		NorthWingDoorOpen: state.NorthWingDoorOpen, WingDoorOpen: state.WingDoorOpen, WinnerID: state.WinnerID, OpenSecrets: state.OpenSecrets, Orb: state.Orb, Statues: state.Statues, Projectiles: state.Projectiles, SafeZones: state.SafeZones,
		Boss: state.Boss, Tick: state.Tick,
		Players: make([]Player, 0, len(state.Players)),
		AIs:     make([]AI, 0, len(state.AIs)), Fruits: make([]Fruit, 0, len(state.Fruits)),
	}
	for _, p := range state.Players {
		snapshot.Players = append(snapshot.Players, Player{
			EntranceID: p.EntranceID, HasWings: p.HasWings, Flight: p.Flight, NorthWingDoorOpen: p.NorthWingDoorOpen, WingDoorOpen: p.WingDoorOpen, OpenSecrets: p.OpenSecrets, HasOrb: p.HasOrb, Escaped: p.Escaped, ActiveStatue: p.ActiveStatue, EntranceGate: p.EntranceGate,
			ShotCooldown: p.ShotCooldown, EjectCooldown: p.EjectCooldown,
			Character: p.Character, Split: p.Split, GatesClosed: p.GatesClosed,
			InArena: p.InArena, TeleportCharge: p.TeleportCharge, Combat: p.Combat,
			Name: names[p.ID], Dead: p.Dead, KilledBy: p.KilledBy,
			Stats: p.Stats, InputSeq: p.InputSeq, ID: p.ID, X: p.X, Y: p.Y, Size: simulation.PlayerSize,
		})
	}
	for _, ai := range state.AIs {
		snapshot.AIs = append(snapshot.AIs, AI{
			Frenzied: ai.Frenzied, Combat: ai.Combat, Behavior: ai.Behavior, Character: ai.Character, Split: ai.Split,
			ID: ai.ID, X: ai.X, Y: ai.Y, Size: simulation.AISize, Stats: ai.Stats,
		})
	}
	for _, f := range state.Fruits {
		snapshot.Fruits = append(snapshot.Fruits, Fruit{Source: f.Source, HP: f.HP, MP: f.MP, ID: f.ID, X: f.X, Y: f.Y, Size: f.Size})
	}
	return snapshot
}

func (s Snapshot) Message() Message {
	return Message{
		MatchStarted: s.MatchStarted, StartCountdown: s.StartCountdown,
		DungeonLayout:     s.DungeonLayout,
		NorthWingDoorOpen: s.NorthWingDoorOpen, WingDoorOpen: s.WingDoorOpen, WinnerID: s.WinnerID, OpenSecrets: s.OpenSecrets, Orb: s.Orb, Statues: s.Statues, Projectiles: s.Projectiles, SafeZones: s.SafeZones, Boss: s.Boss,
		Type: TypeSnapshot, Tick: s.Tick, Players: &s.Players, Fruits: &s.Fruits, AIs: &s.AIs,
	}
}

func (m Message) Snapshot() Snapshot {
	snapshot := Snapshot{MatchStarted: m.MatchStarted, StartCountdown: m.StartCountdown, DungeonLayout: m.DungeonLayout, NorthWingDoorOpen: m.NorthWingDoorOpen, WingDoorOpen: m.WingDoorOpen, WinnerID: m.WinnerID, OpenSecrets: m.OpenSecrets, Orb: m.Orb, Statues: m.Statues, Projectiles: m.Projectiles, SafeZones: m.SafeZones, Boss: m.Boss, Tick: m.Tick}
	if m.Players != nil {
		snapshot.Players = *m.Players
	}
	if m.Fruits != nil {
		snapshot.Fruits = *m.Fruits
	}
	if m.AIs != nil {
		snapshot.AIs = *m.AIs
	}
	return snapshot
}
