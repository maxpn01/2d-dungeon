package rendering

import (
	"encoding/binary"
	"math"
)

func PlayLibraryDiscoverySound() { playFeedbackSound("library-discovery") }

func libraryDiscoveryWave(duration float64) []byte {
	const rate = 44100
	samples := int(duration * rate)
	data := feedbackWaveHeader(samples)
	seed := uint32(6173)
	air := 0.0
	notes := [...]float64{523.25, 415.3, 293.66}
	for i := range samples {
		t := float64(i) / rate
		seed = seed*1664525 + 1013904223
		noise := float64(seed>>8)/8388608 - 1
		air += .035 * (noise - air)
		swell := math.Sin(math.Pi * t / duration)
		sample := (math.Sin(2*math.Pi*146.83*t)+.4*math.Sin(2*math.Pi*147.6*t))*.12*swell + air*.15*swell
		for note, frequency := range notes {
			for echo, gain := range [...]float64{1, .3, .12} {
				age := t - float64(note)*.4 - float64(echo)*.27
				if age < 0 {
					continue
				}
				phase := 2*math.Pi*frequency*age + .025*math.Sin(age*5)
				envelope := min(1, age/.12) * math.Exp(-age*1.9)
				sample += (math.Sin(phase) + .08*math.Sin(phase*2.003)) * envelope * gain * .25
			}
		}
		fade := min(1, (duration-t)/.7)
		value := int16(math.Tanh(sample) * fade * 18000)
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(value))
	}
	return data
}
