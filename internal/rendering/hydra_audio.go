package rendering

import (
	"encoding/binary"
	"math"
)

func hydraRoarWave(duration float64) []byte {
	const rate = 44100
	samples := int(duration * rate)
	data := feedbackWaveHeader(samples)
	seed := uint32(52711)
	rasp := 0.0
	phase := 0.0
	for i := range samples {
		t := float64(i) / rate
		progress := t / duration
		seed = seed*1664525 + 1013904223
		noise := float64(seed>>8)/8388608 - 1
		rasp += .08 * (noise - rasp)
		frequency := 62 + 38*math.Sin(math.Pi*progress) + 6*math.Sin(2*math.Pi*7*t)
		phase += 2 * math.Pi * frequency / rate
		growl := math.Sin(phase) + .5*math.Sin(phase*2.02) + .3*math.Sin(phase*3.1)
		growl *= 1 + .5*math.Sin(2*math.Pi*23*t)
		envelope := min(1, t/.12) * math.Pow(1-progress, 1.4)
		value := int16(math.Tanh((growl*.55+rasp*3.2)*1.4) * envelope * 22000)
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(value))
	}
	return data
}

func severWave(duration float64) []byte {
	const rate = 44100
	samples := int(duration * rate)
	data := feedbackWaveHeader(samples)
	seed := uint32(90121)
	wet := 0.0
	for i := range samples {
		t := float64(i) / rate
		seed = seed*1664525 + 1013904223
		noise := float64(seed>>8)/8388608 - 1
		wet += .25 * (noise - wet)
		chop := noise * math.Exp(-t*60)
		thump := math.Sin(2*math.Pi*(110*t-60*t*t)) * math.Exp(-t*14)
		squelch := wet * math.Exp(-t*9) * (1 + math.Sin(2*math.Pi*31*t))
		envelope := min(1, t/.002) * min(1, (duration-t)/.05)
		value := int16(math.Tanh(chop*.8+thump+squelch*1.6) * envelope * 24000)
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(value))
	}
	return data
}
