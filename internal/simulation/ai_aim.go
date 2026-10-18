package simulation

import "2d-dungeon/internal/parameters"

func (w *World) aiInput(ai *aiState) Direction {
	ai.shotTargetID = ""
	input := w.aiIntent(ai)
	ai.Aiming = false
	if !input.Shoot {
		if !ai.aimReady || w.tick-ai.aimLastSeen >= parameters.AIAimResetTicks {
			ai.aimTargetID = ""
			ai.aimReady = false
		}
		return input
	}
	if ai.aimTargetID != ai.shotTargetID || w.tick-ai.aimLastSeen >= parameters.AIAimResetTicks {
		ai.aimTargetID = ai.shotTargetID
		ai.aimSince = w.tick
		ai.aimReady = false
	}
	ai.aimLastSeen = w.tick
	if !ai.aimReady {
		ai.aimReady = w.tick-ai.aimSince >= parameters.AIAimDelayTicks
	}
	if !ai.aimReady {
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
