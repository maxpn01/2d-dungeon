package rendering

import (
	"image/color"
	"math"
	"strings"

	"2d-dungeon/internal/appearance"
	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func MassColor(source string, playerColors map[string]color.RGBA) color.RGBA {
	if source == "" {
		return EnergyColor
	}
	if tint, ok := playerColors[source]; ok {
		return tint
	}
	if strings.HasPrefix(source, "ai-") {
		return rl.Green
	}
	return rl.SkyBlue
}

func actorRenderSize(size, hp float64) float32 {
	return float32(simulation.BodySize(size, hp))
}

func actorBounds(x, y, size, hp float64) rl.Rectangle {
	side := actorRenderSize(size, hp)
	return rl.NewRectangle(float32(x+size/2)-side/2, float32(y+size/2)-side/2, side, side)
}

func phaseProgress(remaining, duration float64) float32 {
	return float32(1 - max(0, min(1, remaining/duration)))
}

func DrawActor(x, y, size, hp float64, c simulation.CombatState, tint color.RGBA, characters ...appearance.Character) rl.Rectangle {
	character := appearance.Square
	if len(characters) > 0 {
		character = characters[0]
	}
	bounds := actorBounds(x, y, size, hp)
	center := rl.NewVector2(float32(x+size/2), float32(y+size/2))
	dir := rl.NewVector2(float32(c.Direction.X), float32(c.Direction.Y))
	if c.Boosting {
		for i := 3; i > 0; i-- {
			trail := bounds
			trail.X -= dir.X * float32(i) * 10
			trail.Y -= dir.Y * float32(i) * 10
			primitives.DrawCharacterShape(character, trail, rl.Fade(tint, .18/float32(i)))
		}
	}
	if c.Phase == simulation.DashRushing {
		for i := 3; i > 0; i-- {
			trail := bounds
			trail.X -= dir.X * float32(i) * 13
			trail.Y -= dir.Y * float32(i) * 13
			primitives.DrawCharacterShape(character, trail, rl.Fade(tint, 0.19/float32(i)))
		}
	}
	body := actorBody(x, y, size, hp, c, character)
	shadow := body
	shadow.Y += 3
	primitives.DrawCharacterShape(character, shadow, rl.Fade(rl.Black, 0.35))
	primitives.DrawCharacterShape(character, body, tint)

	drawCombat(center, bounds.Width, c)
	if character == appearance.Triangle {
		drawSentinelSword(center, bounds.Width, c)
	}
	return bounds
}

func actorBody(x, y, size, hp float64, c simulation.CombatState, character appearance.Character) rl.Rectangle {
	bounds := actorBounds(x, y, size, hp)
	center := rl.NewVector2(float32(x+size/2), float32(y+size/2))
	dir := rl.NewVector2(float32(c.Direction.X), float32(c.Direction.Y))
	stretch, offset := float32(0), float32(0)
	switch c.Phase {
	case simulation.DashWinding:
		p := phaseProgress(c.Remaining, parameters.DashWindup)
		stretch, offset = -0.18*p, -4*p
	case simulation.DashRushing:
		burst := parameters.DashBurst
		if character == appearance.Triangle {
			burst = parameters.SentinelDashBurst
		}
		stretch = 0.3 * (1 - 0.4*phaseProgress(c.Remaining, burst))
	case simulation.DashBumped:
		p := phaseProgress(c.Remaining, parameters.BumpStun)
		stretch = -0.28 * (1 - p)
		offset = float32(math.Sin(float64(p)*math.Pi*4)) * 4 * (1 - p)
	}
	width := bounds.Width * (1 + stretch*float32(math.Abs(float64(dir.X))) - stretch*0.5*float32(math.Abs(float64(dir.Y))))
	height := bounds.Height * (1 + stretch*float32(math.Abs(float64(dir.Y))) - stretch*0.5*float32(math.Abs(float64(dir.X))))
	return rl.NewRectangle(center.X-width/2+dir.X*offset, center.Y-height/2+dir.Y*offset, width, height)
}

func drawCombat(center rl.Vector2, size float32, c simulation.CombatState) {
	if c.StunRemaining > 0 {
		for i := 0; i < 3; i++ {
			angle := rl.GetTime()*5 + float64(i)*2*math.Pi/3
			star := rl.NewVector2(center.X+float32(math.Cos(angle))*size*.55, center.Y-size*.7+float32(math.Sin(angle))*4)
			rl.DrawCircleV(star, 3, rl.Gold)
		}
	}

	dir := rl.NewVector2(float32(c.Direction.X), float32(c.Direction.Y))
	point := func(distance, sideways float32) rl.Vector2 {
		return rl.NewVector2(center.X+dir.X*distance-dir.Y*sideways, center.Y+dir.Y*distance+dir.X*sideways)
	}
	switch c.Phase {
	case simulation.DashWinding:
		p := phaseProgress(c.Remaining, parameters.DashWindup)
		start, tip := point(size/2+7, 0), point(size/2+48, 0)
		rl.DrawLineEx(start, tip, 2, rl.Fade(rl.Orange, 0.3))
		rl.DrawLineEx(start, point(size/2+7+41*p, 0), 3, rl.Orange)
		rl.DrawLineEx(tip, point(size/2+37, 7), 3, rl.Orange)
		rl.DrawLineEx(tip, point(size/2+37, -7), 3, rl.Orange)
		rl.DrawRing(center, size*0.72+3, size*0.72+5, -90, -90+360*p, 40, rl.Orange)
	case simulation.DashRushing:
		for _, side := range []float32{-1, 1} {
			rl.DrawLineEx(point(-size/2-30, side*size*0.6), point(size/2, side*size*0.6), 2, rl.Fade(rl.RayWhite, 0.65))
		}
	case simulation.DashBumped:
		duration := parameters.BumpStun
		p := phaseProgress(c.Remaining, duration)
		impact := point(size/2+parameters.BumpKnockback, 0)
		for i := 0; i < 7; i++ {
			angle := float64(i)*math.Pi/3 + float64(p)*0.4
			dx, dy := float32(math.Cos(angle)), float32(math.Sin(angle))
			a, b := 5+12*p, 10+22*p
			rl.DrawLineEx(rl.NewVector2(impact.X+dx*a, impact.Y+dy*a), rl.NewVector2(impact.X+dx*b, impact.Y+dy*b), 2, rl.Fade(rl.RayWhite, 1-p))
		}
	case simulation.DashRecovering:
		p := phaseProgress(c.Remaining, parameters.DashRecovery)
		rl.DrawRing(center, size*0.72+3, size*0.72+4, -90+360*p, 270, 40, rl.Fade(rl.Gray, 0.65))
	}
}

func drawStomp(center rl.Vector2, size float32, c simulation.CombatState, kind simulation.MobType) {
	stomp := simulation.MobStomp(kind)
	if c.StompRemaining > 0 {
		direction := rl.NewVector2(float32(c.StompDirection.X), float32(c.StompDirection.Y))
		elapsed := parameters.HeavyStompDuration - c.StompRemaining
		if elapsed < parameters.HeavyStompWindup {
			progress := float32(elapsed / parameters.HeavyStompWindup)
			tip := rl.NewVector2(center.X+direction.X*float32(stomp.Reach), center.Y+direction.Y*float32(stomp.Reach))
			rl.DrawLineEx(center, tip, float32(stomp.WaveSize), rl.Fade(rl.Orange, .08+.12*progress))
			for _, side := range []float32{-1, 1} {
				offset := side * float32(stomp.WaveSize) / 2
				start := rl.NewVector2(center.X-direction.Y*offset, center.Y+direction.X*offset)
				end := rl.NewVector2(tip.X-direction.Y*offset, tip.Y+direction.X*offset)
				rl.DrawLineEx(start, end, 1, rl.Fade(rl.Orange, .4+.4*progress))
			}
			rl.DrawRing(center, size/2+3, size/2+6, -90, -90+360*progress, 32, rl.Gold)
		} else {
			age := elapsed - parameters.HeavyStompWindup
			fade := float32(1 - age/(parameters.HeavyStompDuration-parameters.HeavyStompWindup))
			radius := size/2 + float32(age)*95
			rl.DrawCircleGradient(center, size*1.4, rl.Fade(rl.Gold, fade*.55), rl.Fade(rl.Orange, 0))
			rl.DrawRing(center, radius, radius+5, 0, 360, 40, rl.Fade(rl.Orange, fade))
			rl.DrawRing(center, radius*.7, radius*.7+2, 0, 360, 40, rl.Fade(rl.Gold, fade))
			drawStompDebris(center, direction, age, fade)
		}
	}
}
