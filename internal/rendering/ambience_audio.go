package rendering

import (
	"encoding/binary"
	"math"
	"runtime"

	"dungeon-raylib-go/internal/parameters"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var floorAmbience struct {
	music  rl.Music
	data   []byte
	pin    runtime.Pinner
	ready  bool
	volume float32
}

func initFloorAmbience() {
	floorAmbience.data = floorAmbienceWave()
	floorAmbience.pin.Pin(&floorAmbience.data[0])
	floorAmbience.music = rl.LoadMusicStreamFromMemory(".wav", floorAmbience.data, int32(len(floorAmbience.data)))
	floorAmbience.ready = rl.IsMusicValid(floorAmbience.music)
	if floorAmbience.ready {
		floorAmbience.music.Looping = true
		rl.SetMusicVolume(floorAmbience.music, 0)
	}
}

func closeFloorAmbience() {
	if floorAmbience.ready {
		rl.UnloadMusicStream(floorAmbience.music)
	}
	floorAmbience.pin.Unpin()
	floorAmbience.ready, floorAmbience.volume, floorAmbience.data = false, 0, nil
}

func UpdateFloorAmbience(active bool, dt float32) {
	if !floorAmbience.ready {
		return
	}
	target := float32(0)
	if active && FeedbackAudioEnabled() {
		target = parameters.FloorAmbienceVolume
	}
	step := min(max(dt, 0), .25) * parameters.FloorAmbienceVolume
	floorAmbience.volume += min(step/2, max(-step/1.5, target-floorAmbience.volume))
	if math.Abs(float64(target-floorAmbience.volume)) < .001 {
		floorAmbience.volume = target
	}
	if floorAmbience.volume == 0 {
		if rl.IsMusicStreamPlaying(floorAmbience.music) {
			rl.StopMusicStream(floorAmbience.music)
		}
		return
	}
	rl.SetMusicVolume(floorAmbience.music, floorAmbience.volume)
	if !rl.IsMusicStreamPlaying(floorAmbience.music) {
		rl.PlayMusicStream(floorAmbience.music)
	}
	rl.UpdateMusicStream(floorAmbience.music)
}

type whisperResonator struct {
	a, b, gain, previous, older float64
}

func newWhisperResonator(frequency, bandwidth, rate float64) whisperResonator {
	radius := math.Exp(-math.Pi * bandwidth / rate)
	return whisperResonator{a: 2 * radius * math.Cos(2*math.Pi*frequency/rate), b: -radius * radius, gain: 1 - radius}
}

func (r *whisperResonator) sample(noise float64) float64 {
	value := r.gain*noise + r.a*r.previous + r.b*r.older
	r.older, r.previous = r.previous, value
	return value
}

func floorAmbienceWave() []byte {
	const rate, seconds, overlap = 22050, 24, 1
	frames, blend := rate*seconds, rate*overlap
	samples := make([][2]float64, frames+blend)
	seed := uint32(91357)
	air := 0.0
	voices := [...]whisperResonator{
		newWhisperResonator(530, 140, rate),
		newWhisperResonator(1250, 200, rate),
		newWhisperResonator(2450, 320, rate),
	}
	for i := range samples {
		t := float64(i) / rate
		seed = seed*1664525 + 1013904223
		noise := float64(seed>>8)/8388608 - 1
		air += .018 * (noise - air)
		drift := .2 * math.Sin(2*math.Pi*t/12)
		drone := .16*math.Sin(2*math.Pi*(1762.0/seconds)*t+drift) + .08*math.Sin(2*math.Pi*110*t-drift)
		drone += .045 * math.Sin(2*math.Pi*(3529.0/seconds)*t)
		drone *= .7 + .3*math.Sin(2*math.Pi*t/seconds)
		breath := math.Pow(max(0, math.Sin(2*math.Pi*(t-3)/8)), 2)
		syllables := .35 + .65*math.Pow(.5+.5*math.Sin(t*8+.6*math.Sin(t*2.3)), 2)
		voice := (voices[0].sample(noise)*.5 + voices[1].sample(noise)*1.4 + voices[2].sample(noise)*1.1) * breath * syllables
		voice += .055 * math.Sin(2*math.Pi*196*t+.03*math.Sin(t*4)) * breath * syllables
		bell := 0.0
		for note, frequency := range [...]float64{293.66, 415.3, 440} {
			age := t - float64(note)*7 - 2
			if age >= 0 {
				bell += .09 * math.Sin(2*math.Pi*frequency*age) * min(1, age/.3) * math.Exp(-age*.65)
			}
		}
		pan := .65 * math.Sin(2*math.Pi*t/16)
		for channel := range 2 {
			side := float64(channel*2 - 1)
			samples[i][channel] = math.Tanh(drone*parameters.AmbienceDroneGain + air*parameters.AmbienceAirGain + bell*parameters.AmbienceBellGain + voice*parameters.AmbienceWhisperGain*(1+side*pan))
		}
	}
	data := feedbackWaveHeader(frames * 2)
	binary.LittleEndian.PutUint16(data[22:], 2)
	binary.LittleEndian.PutUint32(data[24:], rate)
	binary.LittleEndian.PutUint32(data[28:], rate*4)
	binary.LittleEndian.PutUint16(data[32:], 4)
	for i := range frames {
		for channel := range 2 {
			sample := samples[i][channel]
			if i < blend {
				mix := float64(i) / float64(blend)
				mix = mix * mix * (3 - 2*mix)
				sample = sample*mix + samples[frames+i][channel]*(1-mix)
			}
			binary.LittleEndian.PutUint16(data[44+(i*2+channel)*2:], uint16(int16(sample*22000)))
		}
	}
	return data
}
