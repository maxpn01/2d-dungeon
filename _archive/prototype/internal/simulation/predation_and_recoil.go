package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
)

const (
	ContactRatio                 = parameters.ContactRatio
	KnockedMass                  = parameters.KnockedMass
	DashRecoiling      DashPhase = "recoil"
	BossHitGrace                 = parameters.BossHitGrace
	BossRecoilDuration           = parameters.BossRecoilDuration
	BossImpactHold               = parameters.BossImpactHold
	BossRecoilDistance           = parameters.BossRecoilDistance
)

func CanContactEat(hp, preyHP float64) bool {
	return hp > 0 && (preyHP == 0 || hp >= preyHP*ContactRatio)
}

func advanceBossRecoil(p *Player, step float64, move func(Direction, float64) physics.Motion) {
	c := &p.Combat

	progress := func(elapsed float64) float64 {
		t := max(0, min(1, (elapsed-BossImpactHold)/(BossRecoilDuration-BossImpactHold)))
		return 1 - (1-t)*(1-t)
	}
	elapsed := BossRecoilDuration - c.Remaining
	fraction := progress(elapsed+step) - progress(elapsed)
	length := math.Hypot(c.RecoilX, c.RecoilY)
	if length > 0 {
		motion := move(Direction{X: c.RecoilX / length, Y: c.RecoilY / length}, length*fraction)
		p.X, p.Y = motion.Bounds.X, motion.Bounds.Y
	}

}
