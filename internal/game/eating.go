package game

import (
	"math"
	"time"

	"raylib-2d-sandbox/internal/client"
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"
)

const (
	eatingContactTime = 80 * time.Millisecond
	eatingShrinkTime  = 140 * time.Millisecond
	eatingSceneTime   = eatingContactTime + eatingShrinkTime + 30*time.Millisecond
)

type eatingScene struct {
	started             time.Time
	preyStart, prey     protocol.Player
	killerStart, killer protocol.Player
	killerAI            bool
	killerBot           protocol.AI
}

func newEatingScene(now time.Time, update client.Update, players []protocol.Player, ais []protocol.AI) *eatingScene {
	scene := &eatingScene{started: now, prey: update.Local, preyStart: update.Local}
	for _, player := range players {
		if player.ID == update.PlayerID {
			scene.preyStart = player
		}
	}
	for _, player := range update.Players {
		if player.ID == update.Local.KilledBy {
			scene.killer = player
		}
	}
	for _, ai := range update.AIs {
		if ai.ID == update.Local.KilledBy {
			scene.killer = protocol.Player{ID: ai.ID, X: ai.X, Y: ai.Y, Size: ai.Size, Stats: ai.Stats}
			scene.killerAI = true
			scene.killerBot = ai
		}
	}
	scene.killerStart = scene.killer
	for _, player := range players {
		if player.ID == scene.killer.ID {
			scene.killerStart = player
		}
	}
	for _, ai := range ais {
		if ai.ID == scene.killer.ID {
			scene.killerStart.X, scene.killerStart.Y = ai.X, ai.Y
		}
	}
	if scene.killer.ID != "" {
		dx := scene.killer.X + scene.killer.Size/2 - (scene.prey.X + scene.prey.Size/2)
		dy := scene.killer.Y + scene.killer.Size/2 - (scene.prey.Y + scene.prey.Size/2)
		distance := math.Max(math.Abs(dx), math.Abs(dy))
		contact := (scene.killer.Size + scene.prey.Size) / 2
		if distance >= contact && distance > 0 {
			scale := contact * .95 / distance
			scene.killer.X = scene.prey.X + scene.prey.Size/2 + dx*scale - scene.killer.Size/2
			scene.killer.Y = scene.prey.Y + scene.prey.Size/2 + dy*scale - scene.killer.Size/2
		}
	}
	return scene
}

func (s *eatingScene) finished(now time.Time) bool {
	return s != nil && now.Sub(s.started) >= eatingSceneTime
}

func (s *eatingScene) actors(now time.Time) (protocol.Player, protocol.Player) {
	elapsed := max(time.Duration(0), now.Sub(s.started))
	contact := min(1, float64(elapsed)/float64(eatingContactTime))
	prey, killer := s.prey, s.killer
	prey.Stats = s.preyStart.Stats
	prey.Combat = simulation.CombatState{}
	prey.X = s.preyStart.X + (prey.X-s.preyStart.X)*contact
	prey.Y = s.preyStart.Y + (prey.Y-s.preyStart.Y)*contact
	killer.X = s.killerStart.X + (killer.X-s.killerStart.X)*contact
	killer.Y = s.killerStart.Y + (killer.Y-s.killerStart.Y)*contact
	shrink := max(0, min(1, float64(elapsed-eatingContactTime)/float64(eatingShrinkTime)))
	shrink = shrink * shrink * (3 - 2*shrink)
	centerX, centerY := prey.X+prey.Size/2, prey.Y+prey.Size/2
	if killer.ID != "" {
		centerX += (killer.X + killer.Size/2 - centerX) * shrink
		centerY += (killer.Y + killer.Size/2 - centerY) * shrink
	}
	prey.Size *= 1 - shrink
	prey.X, prey.Y = centerX-prey.Size/2, centerY-prey.Size/2
	return prey, killer
}

func (s *eatingScene) apply(now time.Time, players []protocol.Player, ais []protocol.AI) ([]protocol.Player, []protocol.AI) {
	prey, killer := s.actors(now)
	visible := players[:0]
	for _, player := range players {
		if player.ID != prey.ID && player.ID != killer.ID {
			visible = append(visible, player)
		}
	}
	if killer.ID != "" {
		if s.killerAI {
			found := false
			for i := range ais {
				if ais[i].ID == killer.ID {
					ais[i].X, ais[i].Y = killer.X, killer.Y
					found = true
				}
			}
			if !found {
				bot := s.killerBot
				bot.X, bot.Y = killer.X, killer.Y
				ais = append(ais, bot)
			}
		} else {
			visible = append(visible, killer)
		}
	}
	if prey.Size > 0 {
		visible = append(visible, prey)
	}
	return visible, ais
}
