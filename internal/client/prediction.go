package client

import (
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"
)

func predictPlayer(player *protocol.Player, input protocol.Direction, zones []simulation.SafeZone) {
	p := simulation.Player{EntranceGate: player.EntranceGate, ShotCooldown: player.ShotCooldown, EjectCooldown: player.EjectCooldown, Character: player.Character, Split: player.Split.Clone(), GatesClosed: player.GatesClosed, InArena: player.InArena, X: player.X, Y: player.Y, Dead: player.Dead, Combat: player.Combat, Stats: player.Stats}
	d := input.Simulation()
	simulation.AdvancePlayer(&p, d, simulation.StepSeconds, zones)
	player.X, player.Y, player.Combat, player.Stats.HP = p.X, p.Y, p.Combat, p.Stats.HP
	player.Stats.MP = p.Stats.MP
	player.Split = p.Split
	if !simulation.InSafeZone(p.X, p.Y, simulation.PlayerSize, p.Stats.HP, zones) {
		simulation.PrepareShot(&p, d)
	}
	p.ShotCooldown = max(0, p.ShotCooldown)
	player.Stats.MP = p.Stats.MP
	player.ShotCooldown = p.ShotCooldown
	player.EjectCooldown = p.EjectCooldown
	player.EntranceGate = p.EntranceGate
}

func (s *Session) PredictedPlayer() (protocol.Player, bool) {
	s.inputMu.Lock()
	defer s.inputMu.Unlock()
	p := s.predicted
	p.Split = p.Split.Clone()
	return p, s.predictionReady
}
