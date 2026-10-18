package rendering

import (
	"image/color"
	"math"
	"time"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/rendering/primitives"
	maprender "2d-dungeon/internal/rendering/worldmap"
	"2d-dungeon/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	hydraTint      = rl.NewColor(62, 98, 108, 255)
	hydraBellyTint = rl.NewColor(118, 146, 120, 255)
	hydraRidgeTint = rl.NewColor(150, 92, 178, 255)
	hydraHeadTint  = rl.NewColor(74, 116, 120, 255)
	hydraEyeTint   = rl.NewColor(255, 176, 64, 255)
	hydraRageTint  = rl.NewColor(255, 72, 56, 255)
	hydraWoundTint = rl.NewColor(150, 34, 58, 255)
	hydraVenomTint = rl.NewColor(150, 236, 120, 255)
)

func hydraVisible(ai simulation.AI, exploration Visibility) bool {
	center := simulation.HydraCenter(ai)
	radius := simulation.HydraBodyRadius(ai.Hydra) + parameters.HydraLungeReach*ai.Hydra.Scale
	return exploration.EntityVisible(center.X-radius, center.Y-radius, radius*2)
}

func hydraHeadLift(ai simulation.AI, head simulation.HydraHead) float32 {
	scale := float32(ai.Hydra.Scale)
	if ai.Hydra.Phase == simulation.HydraSleeping && !ai.Hydra.Engaged {
		return 5 * scale
	}
	lift := 46 * scale
	switch head.LungePhase {
	case simulation.DashWinding:
		p := float32(1 - head.LungeRemaining/parameters.HydraLungeWindup)
		lift += 22 * scale * p
	case simulation.DashRushing:
		lift = 14 * scale
	}
	if head.SpitRemaining > 0 && head.SpitDelay <= 0 {
		p := float32(1 - head.SpitRemaining/parameters.HydraSpitWindup)
		lift += 16 * scale * p
	}
	return lift
}

func hydraForestActors(scene Scene) []maprender.ForestActor {
	var actors []maprender.ForestActor
	for _, ai := range scene.RenderedAIs {
		if ai.Type != simulation.CaveHydra || !hydraVisible(ai, &scene.Map.Exploration) {
			continue
		}
		h := ai.Hydra
		scale := float32(h.Scale)
		flash, spawn := scene.Feedback.HitFlash(ai.ID), scene.Feedback.SpawnProgress(ai.ID)
		center := simulation.HydraCenter(ai)
		body := rl.NewVector2(float32(center.X), float32(center.Y))
		asleep := h.Phase == simulation.HydraSleeping && !h.Engaged
		breathe := float32(math.Sin(rl.GetTime()*1.4)) * 2 * scale
		if !asleep {
			breathe = 0
		}
		cube := func(at rl.Vector2, width, depth, height, lift float32, tint color.RGBA, shadow bool) {
			actors = append(actors, maprender.ForestActor{
				ID: ai.ID, Center: at, Width: width, Depth: depth, Height: height, Lift: lift,
				Tint: poisonTint(tint, ai.Status.Poison), Flash: flash, Spawn: spawn, NoShadow: !shadow,
			})
		}
		for i := len(ai.Segments) - 1; i >= 0; i-- {
			part := simulation.HydraTailBody(h, i, len(ai.Segments), ai.Segments[i])
			at := rl.NewVector2(float32(part.X+part.W/2), float32(part.Y+part.H/2))
			size := float32(part.W)
			cube(at, size, size, size*.45, 0, hydraTint, true)
			cube(at, size*.3, size*.3, size*.25, size*.45, hydraRidgeTint, false)
		}
		size := float32(parameters.HydraBodySize) * scale
		height := 30*scale + breathe
		facing := rl.NewVector2(float32(h.Facing.X), float32(h.Facing.Y))
		side := rl.NewVector2(-facing.Y, facing.X)
		for _, leg := range [][2]float32{{.32, .38}, {.32, -.38}, {-.3, .4}, {-.3, -.4}} {
			at := rl.NewVector2(body.X+facing.X*size*leg[0]+side.X*size*leg[1], body.Y+facing.Y*size*leg[0]+side.Y*size*leg[1])
			cube(at, size*.22, size*.22, height*.5, 0, shade(hydraTint, .8), false)
		}
		cube(body, size, size, height, height*.25, hydraTint, true)
		cube(rl.NewVector2(body.X+facing.X*size*.12, body.Y+facing.Y*size*.12), size*.7, size*.7, height*.2, height*.12, hydraBellyTint, false)
		for i := range 4 {
			k := float32(i)/3 - .5
			at := rl.NewVector2(body.X-facing.X*size*k*.8, body.Y-facing.Y*size*k*.8)
			plate := size * (.22 - .04*float32(math.Abs(float64(k))))
			cube(at, plate, plate, 12*scale, height*1.25, hydraRidgeTint, false)
		}
		for _, head := range h.Heads {
			hx, hy := float32(head.X), float32(head.Y)
			dx, dy := hx-body.X, hy-body.Y
			length := float32(math.Hypot(float64(dx), float64(dy)))
			dir := rl.NewVector2(dx/max(1, length), dy/max(1, length))
			root := rl.NewVector2(body.X+dir.X*size*.32, body.Y+dir.Y*size*.32)
			headLift := hydraHeadLift(ai, head)
			grown := float32(1)
			if head.Severed {
				grown = float32(head.Growing)
				headLift = 18 * scale
			}
			links := 7
			if head.Severed && head.Growing == 0 {
				links = 3
			}
			end := rl.NewVector2(hx, hy)
			for j := 1; j <= links; j++ {
				t := float32(j) / float32(links+1)
				if head.Severed && head.Growing == 0 {
					t *= .55
				}
				at := rl.NewVector2(root.X+(end.X-root.X)*t, root.Y+(end.Y-root.Y)*t)
				arc := float32(math.Sin(float64(t)*math.Pi)) * 14 * scale
				lift := height*.7 + (headLift-height*.7)*t + arc
				thick := (22 - 8*t) * scale
				tint := hydraTint
				if j%2 == 0 {
					tint = shade(hydraTint, 1.12)
				}
				cube(at, thick, thick, thick*.8, lift, tint, j == links)
			}
			if head.Severed && head.Growing == 0 {
				tip := rl.NewVector2(root.X+(end.X-root.X)*.55*float32(links)/float32(links+1), root.Y+(end.Y-root.Y)*.55*float32(links)/float32(links+1))
				cube(tip, 12*scale, 12*scale, 6*scale, height*.7+5*scale, hydraWoundTint, false)
				continue
			}
			headSize := float32(parameters.HydraHeadSize) * scale * grown
			tint := hydraHeadTint
			if head.Severed {
				tint = mixRGBA(hydraWoundTint, hydraHeadTint, grown)
			}
			cube(end, headSize, headSize, headSize*.55, headLift, tint, true)
			snout := rl.NewVector2(end.X+dir.X*headSize*.55, end.Y+dir.Y*headSize*.55)
			jaw := float32(0)
			if head.LungePhase == simulation.DashRushing || head.SpitRemaining > 0 && head.SpitDelay <= 0 && head.SpitRemaining < parameters.HydraSpitAimLock {
				jaw = 6 * scale
			}
			cube(snout, headSize*.6, headSize*.6, headSize*.3, headLift+jaw, shade(tint, 1.1), false)
			cube(snout, headSize*.5, headSize*.5, headSize*.18, headLift-jaw*.4, hydraBellyTint, false)
			crest := rl.NewVector2(end.X-dir.X*headSize*.35, end.Y-dir.Y*headSize*.35)
			cube(crest, headSize*.25, headSize*.25, 9*scale*grown, headLift+headSize*.55, hydraRidgeTint, false)
			eye := hydraEyeTint
			if h.Engaged {
				eye = hydraRageTint
			}
			if asleep {
				eye = shade(hydraHeadTint, .7)
			}
			normal := rl.NewVector2(-dir.Y, dir.X)
			for _, s := range []float32{-1, 1} {
				at := rl.NewVector2(end.X+dir.X*headSize*.1+normal.X*s*headSize*.32, end.Y+dir.Y*headSize*.1+normal.Y*s*headSize*.32)
				cube(at, 5*scale, 5*scale, 3*scale, headLift+headSize*.5, eye, false)
			}
			if head.SpitRemaining > 0 && head.SpitDelay <= 0 {
				p := float32(1 - head.SpitRemaining/parameters.HydraSpitWindup)
				glob := (4 + 8*p) * scale
				mouth := rl.NewVector2(snout.X+dir.X*headSize*.35, snout.Y+dir.Y*headSize*.35)
				cube(mouth, glob, glob, glob, headLift+2, hydraVenomTint, false)
			}
		}
		actors = append(actors, poisonForestBubbles(ai.ID, body, size, height, ai.Status.Poison)...)
	}
	return actors
}

func mixRGBA(a, b color.RGBA, t float32) color.RGBA {
	return color.RGBA{
		R: uint8(float32(a.R) + (float32(b.R)-float32(a.R))*t),
		G: uint8(float32(a.G) + (float32(b.G)-float32(a.G))*t),
		B: uint8(float32(a.B) + (float32(b.B)-float32(a.B))*t),
		A: 255,
	}
}

func shade(c color.RGBA, f float32) color.RGBA {
	clamp := func(v float32) uint8 { return uint8(max(0, min(255, v))) }
	return color.RGBA{clamp(float32(c.R) * f), clamp(float32(c.G) * f), clamp(float32(c.B) * f), c.A}
}

func drawHydra(ai simulation.AI, exploration Visibility, feedback *Feedback) {
	if !hydraVisible(ai, exploration) {
		return
	}
	h := ai.Hydra
	center := simulation.HydraCenter(ai)
	body := rl.NewVector2(float32(center.X), float32(center.Y))
	radius := float32(simulation.HydraBodyRadius(h))
	feedback.DrawAlert(ai.ID, rl.NewRectangle(body.X-radius, body.Y-radius, radius*2, radius*2))
	scale := float32(h.Scale)
	for _, head := range h.Heads {
		at := rl.NewVector2(float32(head.X), float32(head.Y))
		if head.Severed {
			drawHydraStump(head, at, scale)
			continue
		}
		drawHydraLunge(head, body, scale)
		drawHydraSpitWindup(head, at, scale)
	}
}

func DrawHydraSleep(ais []simulation.AI, exploration Visibility, project func(x, y float32) rl.Vector2) {
	for _, ai := range ais {
		if ai.Type == simulation.CaveHydra && ai.Hydra.Phase == simulation.HydraSleeping && !ai.Hydra.Engaged && hydraVisible(ai, exploration) {
			center := simulation.HydraCenter(ai)
			drawHydraSleep(project(float32(center.X), float32(center.Y)))
		}
	}
}

func drawHydraSleep(body rl.Vector2) {
	clock := rl.GetTime()
	for i := range 3 {
		phase := float32(math.Mod(clock*.45+float64(i)/3, 1))
		size := int32(18 + 12*phase)
		at := rl.NewVector2(body.X+20+phase*34+float32(math.Sin(clock*2+float64(i)))*6, body.Y-phase*60)
		fade := float32(math.Sin(float64(phase) * math.Pi))
		primitives.DrawTextAt("z", rl.NewVector2(at.X+1, at.Y+1), size, rl.Fade(rl.Black, .5*fade))
		primitives.DrawTextAt("z", at, size, rl.Fade(rl.NewColor(190, 210, 255, 255), fade))
	}
}

func drawHydraStump(head simulation.HydraHead, at rl.Vector2, scale float32) {
	clock := rl.GetTime()
	for i := range 3 {
		phase := float32(math.Mod(clock*1.3+float64(i)/3, 1))
		drop := rl.NewVector2(at.X+float32(i-1)*5*scale, at.Y+phase*18)
		rl.DrawCircleV(drop, 2.5*(1-phase)+1, rl.Fade(hydraWoundTint, 1-phase))
	}
	if head.Growing > 0 {
		p := float32(head.Growing)
		rl.DrawRing(at, 18*scale, 21*scale, -90, -90+360*p, 32, rl.Fade(hydraVenomTint, .85))
		rl.BeginBlendMode(rl.BlendAdditive)
		rl.DrawCircleGradient(at, 34*scale*p, rl.Fade(hydraVenomTint, .45), rl.Fade(hydraVenomTint, 0))
		rl.EndBlendMode()
		return
	}
	left := float32(1 - head.Regrow/parameters.HydraRegrowSeconds)
	rl.DrawRing(at, 16*scale, 18*scale, -90, -90+360*left, 32, rl.Fade(hydraWoundTint, .7))
}

func drawHydraLunge(head simulation.HydraHead, body rl.Vector2, scale float32) {
	if head.LungePhase != simulation.DashWinding && head.LungePhase != simulation.DashRushing {
		return
	}
	dir := rl.NewVector2(float32(head.LungeDirection.X), float32(head.LungeDirection.Y))
	reach := float32(parameters.HydraLungeReach)*scale + float32(parameters.HydraHeadSize)*scale/2
	width := float32(parameters.HydraHeadSize)*scale + 12
	angle := float32(math.Atan2(float64(dir.Y), float64(dir.X)) * 180 / math.Pi)
	if head.LungePhase == simulation.DashRushing {
		rl.DrawRectanglePro(rl.NewRectangle(body.X, body.Y, reach, width), rl.NewVector2(0, width/2), angle, rl.Fade(hydraRageTint, .28))
		return
	}
	p := float32(1 - head.LungeRemaining/parameters.HydraLungeWindup)
	tint := rl.Orange
	if head.LungeRemaining <= parameters.HydraLungeAimLock {
		tint = hydraRageTint
	}
	rl.DrawRectanglePro(rl.NewRectangle(body.X, body.Y, reach, width), rl.NewVector2(0, width/2), angle, rl.Fade(tint, .06+.18*p))
	for _, s := range []float32{-1, 1} {
		edge := rl.NewVector2(body.X-dir.Y*s*width/2, body.Y+dir.X*s*width/2)
		rl.DrawLineEx(edge, rl.NewVector2(edge.X+dir.X*reach, edge.Y+dir.Y*reach), 2, rl.Fade(tint, .35+.55*p))
	}
	tip := rl.NewVector2(body.X+dir.X*reach, body.Y+dir.Y*reach)
	rl.DrawRing(tip, width*.35*p, width*.35*p+3, 0, 360, 24, rl.Fade(tint, .8))
}

func drawHydraSpitWindup(head simulation.HydraHead, at rl.Vector2, scale float32) {
	if head.SpitRemaining <= 0 || head.SpitDelay > 0 {
		return
	}
	p := float32(1 - head.SpitRemaining/parameters.HydraSpitWindup)
	dir := rl.NewVector2(float32(head.SpitDirection.X), float32(head.SpitDirection.Y))
	line := rl.Fade(PoisonColor, .15+.35*p)
	if head.SpitRemaining <= parameters.HydraSpitAimLock {
		line = rl.Fade(hydraVenomTint, .8)
	}
	reach := float32(parameters.HydraSpitRange) * (.3 + .5*p)
	for d := float32(20) * scale; d < reach; d += 22 {
		from := rl.NewVector2(at.X+dir.X*d, at.Y+dir.Y*d)
		to := rl.NewVector2(at.X+dir.X*min(reach, d+11), at.Y+dir.Y*min(reach, d+11))
		rl.DrawLineEx(from, to, 2, line)
	}
}

func hydraHUDActor(ais []simulation.AI) (simulation.AI, bool) {
	for _, ai := range ais {
		if ai.Type == simulation.CaveHydra {
			return ai, true
		}
	}
	return simulation.AI{}, false
}

func HydraInView(ais []simulation.AI, player simulation.Player, exploration Visibility) (simulation.AI, bool) {
	ai, ok := hydraHUDActor(ais)
	if !ok || player.Dead {
		return ai, false
	}
	center := simulation.HydraCenter(ai)
	near := math.Hypot(center.X-player.X, center.Y-player.Y) < 900
	return ai, ai.Hydra.Engaged || near && hydraVisible(ai, exploration)
}

func (f *Feedback) observeHydra(now time.Time, old, actor, local feedbackActor) {
	if actor.dead || local.dead || math.Hypot(actor.x-local.x, actor.y-local.y) > 1400 {
		return
	}
	emit := func(x, y, size float64, tint color.RGBA, kind string) {
		f.effects = append(f.effects, feedbackEffect{started: now, x: float32(x), y: float32(y), size: float32(size), tint: tint, kind: kind})
	}
	h, before := actor.hydra, old.hydra
	if before.Phase == simulation.HydraSleeping && !before.Engaged && (h.Phase != simulation.HydraSleeping || h.Engaged) {
		playFeedbackSound("hydra-roar")
		f.shake(now, parameters.StompShakeStrength, .5)
	}
	for i, head := range h.Heads {
		was := before.Heads[i]
		switch {
		case head.Severed && !was.Severed:
			emit(head.X, head.Y, parameters.HydraHeadSize*h.Scale, hydraWoundTint, "sever")
			playFeedbackSound("sever")
			f.shake(now, parameters.HurtShakeStrength, parameters.ScreenShakeSeconds)
		case !head.Severed && was.Severed:
			emit(head.X, head.Y, parameters.HydraHeadSize*h.Scale, hydraVenomTint, "regrow")
			playFeedbackSound("hydra-roar")
		case head.LungePhase == simulation.DashRushing && was.LungePhase != simulation.DashRushing:
			playFeedbackSound("heavy-melee")
		}
	}
}

func drawSeverBurst(center rl.Vector2, size, p float32) {
	for i := range 10 {
		angle := float64(i)*math.Pi/5 + .3
		distance := size*.4 + 46*p
		drop := rl.NewVector2(center.X+float32(math.Cos(angle))*distance, center.Y+float32(math.Sin(angle))*distance+30*p*p)
		rl.DrawCircleV(drop, 4*(1-p)+1, rl.Fade(hydraWoundTint, 1-p))
	}
	rl.DrawRing(center, size*.5+30*p, size*.5+30*p+3*(1-p), 0, 360, 32, rl.Fade(rl.RayWhite, .8*(1-p)))
	label := "HEAD SEVERED"
	primitives.DrawText(label, int32(center.X)-primitives.MeasureText(label, 18)/2, int32(center.Y-size-30-24*p), 18, rl.Fade(rl.Gold, 1-p))
}

func drawRegrowBloom(center rl.Vector2, size, p float32) {
	rl.BeginBlendMode(rl.BlendAdditive)
	rl.DrawCircleGradient(center, size+40*p, rl.Fade(hydraVenomTint, .5*(1-p)), rl.Fade(hydraVenomTint, 0))
	rl.EndBlendMode()
	label := "HEAD REGROWN"
	primitives.DrawText(label, int32(center.X)-primitives.MeasureText(label, 16)/2, int32(center.Y-size-30-24*p), 16, rl.Fade(hydraVenomTint, 1-p))
}

func drawHydraDefeat(center rl.Vector2, size, p float32) {
	glow := float32(math.Sin(float64(p) * math.Pi))
	rl.BeginBlendMode(rl.BlendAdditive)
	rl.DrawCircleGradient(center, 220, rl.Fade(PoisonColor, .45*glow), rl.Fade(PoisonColor, 0))
	rl.EndBlendMode()
	for i := range 3 {
		radius := 40 + float32(i)*30 + 180*p
		rl.DrawRing(center, radius, radius+3*(1-p), 0, 360, 64, rl.Fade(hydraVenomTint, .7*(1-p)))
	}
	drawDeathBurst(center, size*2, p, hydraTint, uint32(center.X*7)^uint32(center.Y*13))
	label := "THE HYDRA FALLS"
	primitives.DrawText(label, int32(center.X)-primitives.MeasureText(label, 26)/2, int32(center.Y-90-30*p), 26, rl.Fade(rl.Gold, 1-p*p))
}
