package rendering

import (
	"encoding/binary"
	"math"
)

func eatingWave(duration float64, mana bool) []byte {
	const rate = 44100
	samples := int(duration * rate)
	data := feedbackWaveHeader(samples)
	pitch := 740.0
	if mana {
		pitch = 930
	}
	seed := uint32(72631)
	air := 0.0
	for i := range samples {
		t := float64(i) / rate
		popEnvelope := (1 - math.Exp(-t/.008)) * math.Exp(-t/.026)
		popPhase := 2 * math.Pi * pitch * (t + .18*(t-.018*(1-math.Exp(-t/.018))))
		value := .85 * math.Sin(popPhase) * popEnvelope
		shimmerEnvelope := (1 - math.Exp(-t/.025)) * math.Exp(-t/.055)
		shimmerPhase := 2*math.Pi*pitch*(t+.12*(t-.04*(1-math.Exp(-t/.04)))) + .22*math.Sin(2*math.Pi*17*t)
		value += .16 * (math.Sin(shimmerPhase) + .45*math.Sin(shimmerPhase*1.006)) * shimmerEnvelope
		seed = seed*1664525 + 1013904223
		noise := float64(seed>>8)/8388608 - 1
		air += .12 * (noise - air)
		value += .22 * air * shimmerEnvelope
		value *= min(1, (duration-t)/.025)
		sample := int16(value * 28000)
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(sample))
	}
	return data
}
