package rendering

import (
	"encoding/binary"
	"math"
	"time"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (f *Feedback) observeShootingAudio(actors map[string]feedbackActor) {
	local := actors[f.localID]
	if local.dead {
		return
	}
	geometry := worldmap.Default()
	floor := geometry.BodyFloor(local.x, local.y, local.size)
	nearest := parameters.ShotSoundRadius
	mobID := ""
	for id, actor := range actors {
		old, known := f.actors[id]
		if !known || actor.dead || actor.shots <= old.shots {
			continue
		}
		if id == f.localID {
			playShootingSound(false, 1, actor.shots)
			continue
		}
		if !actor.ai || geometry.BodyFloor(actor.x, actor.y, actor.size) != floor {
			continue
		}
		distance := math.Hypot(actor.x-local.x, actor.y-local.y)
		if distance < nearest || distance == nearest && mobID != "" && id < mobID {
			nearest, mobID = distance, id
		}
	}
	if mobID != "" {
		variation := actors[mobID].shots
		for _, char := range mobID {
			variation = variation*31 + uint64(char)
		}
		playShootingSound(true, 1-nearest/parameters.ShotSoundRadius, variation)
	}
}

func playShootingSound(mob bool, gain float64, variation uint64) {
	if !FeedbackAudioEnabled() {
		return
	}
	kind := "player-shot"
	volume := parameters.PlayerShotSoundVolume
	interval := parameters.PlayerShotSoundInterval
	pitch := float32(1)
	if mob {
		kind = "mob-shot"
		volume = parameters.MobShotSoundVolume
		interval = parameters.MobShotSoundInterval
		pitch = .88 + float32(variation%9)*.012
	}
	sound, ok := combatAudio.sounds[kind]
	if !ok || time.Since(combatAudio.last[kind]).Seconds() < interval {
		return
	}
	combatAudio.last[kind] = time.Now()
	rl.SetSoundVolume(sound, volume*float32(gain))
	rl.SetSoundPitch(sound, pitch)
	rl.PlaySound(sound)
}

func shootingWave(duration float64) []byte {
	const rate = 44100
	samples := int(duration * rate)
	data := feedbackWaveHeader(samples)
	seed := uint32(43871)
	air := 0.0
	for i := range samples {
		t := float64(i) / rate
		progress := t / duration
		envelope := (1 - math.Exp(-t/.012)) * math.Exp(-t/.035) * (1 - progress)
		phase := 2 * math.Pi * (520*t + 210*.025*(1-math.Exp(-t/.025)))
		seed = seed*1664525 + 1013904223
		noise := float64(seed>>8)/8388608 - 1
		air += .16 * (noise - air)
		value := (.7*math.Sin(phase+.18*math.Sin(phase*.5)) + air*.85) * envelope
		value *= min(1, (duration-t)/.025)
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(int16(value*28000)))
	}
	return data
}
