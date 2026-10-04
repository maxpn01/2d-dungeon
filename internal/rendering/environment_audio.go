package rendering

import (
	"encoding/binary"
	"math"
)

func environmentWave(kind string, duration float64) []byte {
	const rate = 44100
	samples := int(duration * rate)
	data := feedbackWaveHeader(samples)
	seed := uint32(31937)
	filtered := float64(0)
	for i := 0; i < samples; i++ {
		t := float64(i) / rate
		p := t / duration
		seed = seed*1664525 + 1013904223
		noise := float64(seed>>8)/8388608 - 1
		filtered += .08 * (noise - filtered)
		envelope := math.Min(1, t/.04) * math.Min(1, (duration-t)/.22)
		sample := filtered*2.5 + noise*.13 + math.Sin(t*math.Pi*2*55)*.2
		if kind == "awaken" {
			envelope *= .2 + .8*p
			sample = filtered*2 + noise*.2*math.Pow(math.Sin(t*83), 8) + math.Sin(t*math.Pi*2*(85-25*p))*.25
			if p > .72 {
				sample += noise * .65 * math.Exp(-(p-.72)*8)
			}
		}
		value := int16(math.Tanh(sample) * envelope * 25000)
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(value))
	}
	return data
}
