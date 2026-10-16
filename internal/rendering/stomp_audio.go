package rendering

import (
	"encoding/binary"
	"math"
)

func heavyStompWave(duration float64) []byte {
	const rate = 44100
	samples := int(duration * rate)
	data := feedbackWaveHeader(samples)
	seed := uint32(8317)
	gravel := 0.0
	for i := range samples {
		t := float64(i) / rate
		seed = seed*1664525 + 1013904223
		noise := float64(seed>>8)/8388608 - 1
		gravel += .12 * (noise - gravel)
		phase := 2 * math.Pi * (78*t - 28*t*t/(2*duration))
		body := math.Sin(phase)*math.Exp(-t*11) + .22*math.Sin(phase*1.7)*math.Exp(-t*18)
		crack := (gravel*.9 + noise*.12) * math.Exp(-t*35)
		ring := .08 * math.Sin(2*math.Pi*146.83*t) * math.Exp(-t*8)
		envelope := min(1, t/.003) * min(1, (duration-t)/.08)
		value := int16(math.Tanh(body+crack+ring) * envelope * 22000)
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(value))
	}
	return data
}
