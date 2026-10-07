package rendering

import (
	"encoding/binary"
	"math"
	"time"

	"raylib-2d-sandbox/assets"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const victorySoundVolume float32 = .04

var combatAudio struct {
	sounds       map[string]rl.Sound
	last         map[string]time.Time
	muted, owned bool
}

func InitFeedbackAudio() {
	if combatAudio.sounds != nil {
		return
	}
	if !rl.IsAudioDeviceReady() {
		rl.InitAudioDevice()
		combatAudio.owned = true
	}
	if !rl.IsAudioDeviceReady() {
		return
	}
	combatAudio.sounds = make(map[string]rl.Sound)
	combatAudio.last = make(map[string]time.Time)
	for _, cue := range []struct {
		name                 string
		start, end, duration float64
	}{{"hit", 950, 220, .075}, {"hurt", 180, 65, .18}, {"mana", 240, 170, .16}, {"low", 100, 75, .3}, {"stone", 130, 42, .14}, {"awaken", 0, 0, 3.0}, {"fire", 0, 0, .9}, {"rumble", 0, 0, DungeonAwakeningShakeDuration}, {"completion", 0, 0, 8}} {
		var data []byte
		if cue.name == "completion" {
			data = assets.RoundComplete
		} else {
			data = feedbackWave(cue.start, cue.end, cue.duration)
		}
		if cue.name == "fire" || cue.name == "awaken" || cue.name == "rumble" {
			data = environmentWave(cue.name, cue.duration)
		}
		wave := rl.LoadWaveFromMemory(".wav", data, int32(len(data)))
		if !rl.IsWaveValid(wave) {
			continue
		}
		sound := rl.LoadSoundFromWave(wave)
		rl.UnloadWave(wave)
		if rl.IsSoundValid(sound) {
			volume := float32(.3)
			if cue.name == "completion" {
				volume = victorySoundVolume
			}
			rl.SetSoundVolume(sound, volume)
			combatAudio.sounds[cue.name] = sound
		}
	}
}

func CloseFeedbackAudio() {
	for _, sound := range combatAudio.sounds {
		rl.UnloadSound(sound)
	}
	combatAudio.sounds = nil
	combatAudio.last = nil
	if combatAudio.owned && rl.IsAudioDeviceReady() {
		rl.CloseAudioDevice()
	}
	combatAudio.owned = false
}

func FeedbackAudioEnabled() bool { return !combatAudio.muted && len(combatAudio.sounds) > 0 }
func ToggleFeedbackAudio() {
	combatAudio.muted = !combatAudio.muted
	if combatAudio.muted {
		for _, sound := range combatAudio.sounds {
			rl.StopSound(sound)
		}
	}
}

func playFeedbackSound(kind string) {
	if !FeedbackAudioEnabled() {
		return
	}
	sound, ok := combatAudio.sounds[kind]
	if !ok || time.Since(combatAudio.last[kind]) < 100*time.Millisecond {
		return
	}
	combatAudio.last[kind] = time.Now()
	rl.PlaySound(sound)
}

func feedbackWave(start, end, duration float64) []byte {
	const rate = 44100
	samples := int(duration * rate)
	data := feedbackWaveHeader(samples)
	for i := range samples {
		t := float64(i) / rate
		p := t / duration
		phase := 2 * math.Pi * (start*t + (end-start)*t*t/(2*duration))
		envelope := min(1, t/.004) * (1 - p) * (1 - p)
		sample := int16((math.Sin(phase)*.65 + math.Sin(phase*2.7)*.18) * envelope * 24000)
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(sample))
	}
	return data
}

func feedbackWaveHeader(samples int) []byte {
	const rate = 44100
	data := make([]byte, 44+samples*2)
	copy(data, "RIFF")
	binary.LittleEndian.PutUint32(data[4:], uint32(len(data)-8))
	copy(data[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(data[16:], 16)
	binary.LittleEndian.PutUint16(data[20:], 1)
	binary.LittleEndian.PutUint16(data[22:], 1)
	binary.LittleEndian.PutUint32(data[24:], rate)
	binary.LittleEndian.PutUint32(data[28:], rate*2)
	binary.LittleEndian.PutUint16(data[32:], 2)
	binary.LittleEndian.PutUint16(data[34:], 16)
	copy(data[36:], "data")
	binary.LittleEndian.PutUint32(data[40:], uint32(samples*2))
	return data
}
