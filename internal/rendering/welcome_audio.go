package rendering

import (
	"encoding/binary"
	"math"
)

func PlayDungeonWelcomeSound() { playFeedbackSound("welcome") }

func dungeonWelcomeWave(duration float64) []byte {
	const rate = 44100
	samples := int(duration * rate)
	data := feedbackWaveHeader(samples)
	notes := [...]struct {
		frequency, start float64
	}{
		{293.66, 0},
		{440, .32},
		{523.25, .64},
		{659.25, 1.02},
		{587.33, 1.4},
	}
	for i := range samples {
		t := float64(i) / rate
		sample := 0.0
		for _, note := range notes {
			for echo, gain := range [...]float64{1, .22, .08} {
				age := t - note.start - float64(echo)*.19
				if age < 0 {
					continue
				}
				phase := 2 * math.Pi * note.frequency * age
				envelope := min(1, age/.045) * math.Exp(-age*2.6)
				tone := math.Sin(phase) + .12*math.Sin(phase*2.002)*math.Exp(-age*3)
				sample += tone * envelope * gain * .3
			}
		}
		fade := min(1, (duration-t)/.5)
		value := int16(math.Tanh(sample) * fade * 18000)
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(value))
	}
	return data
}
