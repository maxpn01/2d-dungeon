package rendering

import (
	"encoding/binary"
	"math"
)

const DungeonAwakeningShakeDuration = 3.0

func PlayDungeonAwakeningSound() {
	playFeedbackSound("rumble")
}

func environmentWave(kind string, duration float64) []byte {
	if kind == "gate" {
		return magicalGateWave(duration)
	}
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
		filter := .08
		if kind == "rumble" {
			filter = .012
		}
		filtered += filter * (noise - filtered)
		envelope := math.Min(1, t/.04) * math.Min(1, (duration-t)/.22)
		sample := filtered*2.5 + noise*.13 + math.Sin(t*math.Pi*2*55)*.2
		if kind == "rumble" {
			envelope = math.Min(1, t/.06) * math.Pow(1-p, 2)
			pulse := .8 + .2*math.Sin(t*13)
			sample = (math.Sin(t*math.Pi*2*38)*.55 + math.Sin(t*math.Pi*2*57)*.25 + math.Sin(t*math.Pi*2*76)*.12 + filtered*3) * pulse
		}
		value := int16(math.Tanh(sample) * envelope * 25000)
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(value))
	}
	return data
}

func magicalGateWave(duration float64) []byte {
	const rate = 44100
	samples := int(duration * rate)
	data := feedbackWaveHeader(samples)
	tones := [...]float64{293.66, 440, 587.33, 880}
	for i := range samples {
		t := float64(i) / rate
		sample := 0.0
		for note, frequency := range tones {
			age := t - float64(note)*.12
			if age < 0 {
				continue
			}
			phase := 2*math.Pi*frequency*age + .012*math.Sin(age*7)
			envelope := min(1, age/.09) * math.Exp(-age*2.4)
			tone := math.Sin(phase) + .16*math.Sin(phase*2.001)*math.Exp(-age*3)
			sample += tone * envelope * (.38 / (1 + float64(note)*.3))
		}
		fade := min(1, (duration-t)/.4)
		value := int16(math.Tanh(sample) * fade * 18000)
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(value))
	}
	return data
}
