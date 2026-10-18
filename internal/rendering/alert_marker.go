package rendering

import (
	"math"
	"time"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/rendering/primitives"
	"dungeon-raylib-go/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var idleBehaviors = map[simulation.Behavior]bool{
	simulation.BehaviorPatrolling:    true,
	simulation.BehaviorWatching:      true,
	simulation.BehaviorFeeding:       true,
	simulation.BehaviorReturning:     true,
	simulation.BehaviorSearching:     true,
	simulation.BehaviorInvestigating: true,
	simulation.BehaviorStoneGuardian: true,
	simulation.BehaviorAwakening:     true,
}

func (f *Feedback) observeAlerts(now time.Time, actors map[string]feedbackActor) {
	if f.engaged == nil {
		f.engaged = make(map[string]bool)
		f.alerts = make(map[string]time.Time)
	}
	engaged := make(map[string]bool, len(actors))
	for id, actor := range actors {
		if !actor.ai {
			continue
		}
		was, known := f.engaged[id]
		engagedNow := was
		if actor.behavior != simulation.BehaviorDodging {
			engagedNow = !idleBehaviors[actor.behavior]
		}
		engaged[id] = engagedNow
		if known && engagedNow && !was && !actor.dead {
			f.alerts[id] = now
		}
	}
	for id, at := range f.alerts {
		if _, exists := actors[id]; !exists || now.Sub(at).Seconds() > parameters.AlertMarkerSeconds {
			delete(f.alerts, id)
		}
	}
	f.engaged = engaged
}

func (f *Feedback) DrawAlert(id string, bounds rl.Rectangle) {
	if f == nil {
		return
	}
	at, ok := f.alerts[id]
	if !ok {
		return
	}
	age := time.Since(at).Seconds()
	if age >= parameters.AlertMarkerSeconds {
		return
	}
	pop := float32(1 + .7*math.Max(0, 1-age/.14))
	fade := float32(math.Min(1, (parameters.AlertMarkerSeconds-age)/.25))
	center := rl.NewVector2(bounds.X+bounds.Width/2, bounds.Y-44)
	rl.DrawCircleV(center, 11*pop, rl.Fade(rl.NewColor(30, 18, 10, 255), .85*fade))
	rl.DrawCircleLinesV(center, 11*pop, rl.Fade(rl.Orange, fade))
	size := int32(22 * pop)
	primitives.DrawText("!", int32(center.X)-primitives.MeasureText("!", size)/2, int32(center.Y)-size/2, size, rl.Fade(rl.Gold, fade))
}
