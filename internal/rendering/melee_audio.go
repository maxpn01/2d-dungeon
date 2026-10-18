package rendering

import (
	"encoding/binary"
	"math"
)

func meleeSwingWave(duration float64, heavy bool) []byte {
	const rate = 44100
	samples := int(duration * rate)
	data := feedbackWaveHeader(samples)
	seed := uint32(19427)
	filtered := 0.0
	frequency, cutoff := 420.0, .35
	if heavy {
		frequency, cutoff = 150, .12
	}
	for i := range samples {
		t := float64(i) / rate
		progress := t / duration
		seed = seed*1664525 + 1013904223
		noise := float64(seed>>8)/8388608 - 1
		filtered += cutoff * (noise - filtered)
		envelope := math.Pow(math.Sin(math.Pi*progress), 2) * math.Exp(-progress*2)
		phase := 2 * math.Pi * frequency * (t - .35*t*t/duration)
		swish := filtered*2.5 + .15*math.Sin(phase)
		if heavy {
			swish += .4 * math.Sin(phase*.5) * math.Exp(-progress*3)
		}
		value := int16(math.Tanh(swish) * envelope * 28000)
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(value))
	}
	return data
}
