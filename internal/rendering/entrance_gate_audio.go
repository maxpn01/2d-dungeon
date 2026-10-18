package rendering

import (
	"encoding/binary"
	"math"
)

func entranceGateWave(duration float64) []byte {
	const rate = 44100
	samples := int(duration * rate)
	data := feedbackWaveHeader(samples)
	seed := uint32(57193)
	grind, body := 0.0, 0.0
	for i := range samples {
		t := float64(i) / rate
		seed = seed*1664525 + 1013904223
		noise := float64(seed>>8)/8388608 - 1
		grind += .09 * (noise - grind)
		body += .008 * (noise - body)
		movement := min(1, t/.045) * min(1, max(0, .42-t)/.07)
		texture := .8 + .12*math.Sin(2*math.Pi*19*t) + .08*math.Sin(2*math.Pi*31*t)
		stone := (grind*1.6 + body*3 + .09*math.Sin(2*math.Pi*73*t)) * texture * movement
		magicEnvelope := (1 - math.Exp(-t/.05)) * math.Exp(-t/.17)
		magicPhase := 2 * math.Pi * (330*t + 55*t*t)
		magic := .12 * math.Sin(magicPhase+.15*math.Sin(2*math.Pi*6*t)) * magicEnvelope
		settle := 0.0
		if age := t - .38; age > 0 {
			envelope := (1 - math.Exp(-age/.008)) * math.Exp(-age/.025)
			settle = (.3*math.Sin(2*math.Pi*92*age) + body) * envelope
		}
		fade := min(1, (duration-t)/.08)
		value := int16(math.Tanh(stone+magic+settle) * fade * 26000)
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(value))
	}
	return data
}
