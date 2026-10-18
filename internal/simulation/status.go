package simulation

import (
	"math"

	"2d-dungeon/internal/parameters"
)

type StatusEffects struct {
	Poison       Poison
	ShedCooldown float64
	ShedWard     float64
	Sheds        int
}

type Poison struct {
	Damage    float64
	Remaining float64
	Duration  float64
	NextTick  float64
	Dealt     float64
}

func (p Poison) Active() bool { return p.Remaining > 0 }

func (s *StatusEffects) ApplyPoison(damage, seconds float64) {
	if damage <= 0 || seconds <= 0 || math.IsNaN(damage) || math.IsInf(damage, 0) {
		return
	}
	if s.ShedWard > 0 {
		return
	}
	p := &s.Poison
	if !p.Active() {
		p.Damage, p.NextTick = 0, parameters.PoisonTickSeconds
	}
	p.Damage = max(p.Damage, damage)
	p.Remaining = max(p.Remaining, seconds)
	p.Duration = p.Remaining
}

func PoisonStrength(attacker Stats, segments int) float64 {
	return attacker.Damage * parameters.PoisonDamageRatio * (1 + parameters.PoisonSegmentBonus*float64(segments))
}

func (s *StatusEffects) advance(stats *Stats, protected bool) {
	s.ShedWard = max(0, s.ShedWard-parameters.StepSeconds)
	s.ShedCooldown = max(0, s.ShedCooldown-parameters.StepSeconds)
	p := &s.Poison
	if !p.Active() {
		return
	}
	if rank := stats.Perks[ShedskinPerk]; rank > 0 && s.ShedCooldown == 0 && p.Duration-p.Remaining >= parameters.ShedskinDelaySeconds-1e-9 {
		*p = Poison{Dealt: p.Dealt}
		s.ShedWard = parameters.ShedskinWardSeconds[rank-1]
		s.ShedCooldown = PerkBonus(ShedskinPerk, rank)
		s.Sheds++
		return
	}
	p.Remaining = max(0, p.Remaining-parameters.StepSeconds)
	p.NextTick -= parameters.StepSeconds
	if p.NextTick <= 1e-9 {
		p.NextTick += parameters.PoisonTickSeconds
		if !protected && stats.HP > 1 {
			before := stats.HP
			stats.HP = max(1, stats.HP-p.Damage)
			p.Dealt += before - stats.HP
		}
	}
	if !p.Active() {
		*p = Poison{Dealt: p.Dealt}
	}
}

func (e eater) status() *StatusEffects {
	if e.player != nil {
		return &e.player.Status
	}
	return &e.ai.Status
}

func (w *World) advanceStatusEffects() {
	if p := w.player; !p.Dead {
		p.Status.advance(&p.Stats, eater{player: p}.protected())
	}
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick == 0 {
			ai.Status.advance(&ai.Stats, eater{ai: ai}.protected())
		}
	}
}
