package simulation

import "dungeon-raylib-go/internal/parameters"

func (w *World) aiInput(ai *aiState) Direction {
	ai.shotTargetID = ""
	input := w.aiIntent(ai)
	ai.Aiming = false
	if !input.Shoot {
		if !ai.aim.ready || w.tick-ai.aim.lastSeen >= parameters.AIAimResetTicks {
			ai.aim.targetID = ""
			ai.aim.ready = false
		}
		return input
	}
	if ai.aim.targetID != ai.shotTargetID || w.tick-ai.aim.lastSeen >= parameters.AIAimResetTicks {
		ai.aim.targetID = ai.shotTargetID
		ai.aim.since = w.tick
		ai.aim.ready = false
	}
	ai.aim.lastSeen = w.tick
	if !ai.aim.ready {
		ai.aim.ready = w.tick-ai.aim.since >= parameters.AIAimDelayTicks
	}
	if !ai.aim.ready {
		input.Shoot = false
		ai.Aiming = true
		ai.Behavior = BehaviorAiming
	}
	if input.Shoot {
		if target := w.playerWithID(ai.shotTargetID); target != nil && !w.clearMobShot(ai, target) {
			input.Shoot = false
			ai.Behavior = BehaviorWaitingForClearShot
		}
	}
	return input
}
