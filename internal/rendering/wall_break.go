package rendering

import (
	"math"
	"time"

	"2d-dungeon/internal/physics"
	"2d-dungeon/internal/simulation"
	"2d-dungeon/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type wallBreak struct {
	started time.Time
	area    worldmap.CellRect
	hit     bool
}

func (f *Feedback) observeWalls(now time.Time, snapshot simulation.Snapshot) {
	open := snapshot.OpenSecrets
	current := make(map[string]bool, len(open))
	for _, id := range open {
		current[id] = true
	}
	if f.secrets != nil {
		for _, door := range worldmap.Default().SecretDoors() {
			if !door.Interactable && !current[door.ID] && snapshot.SecretWallHits[door.ID] > f.wallHits[door.ID] {
				f.walls = append(f.walls, wallBreak{started: now, area: door.Area, hit: true})
				center := door.Area.Bounds().Center()
				p := snapshot.Player
				if worldmap.Default().FloorAt(p.X, p.Y) == worldmap.Default().FloorAt(center.X, center.Y) && math.Hypot(p.X-center.X, p.Y-center.Y) < 300 {
					playFeedbackSound("stone")
				}
			}
			if !door.Interactable && current[door.ID] && !f.secrets[door.ID] {
				f.walls = append(f.walls, wallBreak{started: now, area: door.Area})
			}
		}
	}
	f.secrets = current
	f.wallHits = snapshot.SecretWallHits
}

func (f *Feedback) DrawWallDebris(now time.Time, visibility Visibility) {
	if f == nil || len(f.walls) == 0 {
		return
	}
	active := f.walls[:0]
	open := make([]string, 0, len(f.secrets))
	for id := range f.secrets {
		open = append(open, id)
	}
	geometry := worldmap.Default().WithSecrets(open)
	for _, wall := range f.walls {
		age := max(0, now.Sub(wall.started).Seconds())
		if wall.hit {
			if age < .45 {
				active = append(active, wall)
				visible := false
				neighbors := []worldmap.Cell{
					worldmap.At(wall.area.X-1, wall.area.Y), worldmap.At(wall.area.Right(), wall.area.Y),
					worldmap.At(wall.area.X, wall.area.Y-1), worldmap.At(wall.area.X, wall.area.Bottom()),
				}
				for _, cell := range neighbors {
					at := cell.Center()
					visible = visible || visibility.EntityVisible(at.X, at.Y, 0)
				}
				if !visible {
					continue
				}
				for y := wall.area.Y; y < wall.area.Bottom(); y++ {
					for x := wall.area.X; x < wall.area.Right(); x++ {
						center := worldmap.At(x, y).Center()
						p := float32(age / .45)
						at := rl.NewVector2(float32(center.X), float32(center.Y))
						rl.DrawCircleGradient(at, 12+30*p, rl.Fade(rl.NewColor(165, 178, 189, 255), .4*(1-p)), rl.Blank)
						for i := 0; i < 6; i++ {
							angle := float64(i)*math.Pi/3 + float64(x+y)
							rl.DrawCircleV(rl.NewVector2(at.X+float32(math.Cos(angle))*p*28, at.Y+float32(math.Sin(angle))*p*20+p*p*16), 2*(1-p), rl.Fade(rl.NewColor(132, 143, 153, 255), 1-p))
						}
					}
				}
			}
			continue
		}
		if age >= 10 {
			continue
		}
		active = append(active, wall)
		fade := float32(min(1, (10-age)/3))
		for y := wall.area.Y; y < wall.area.Bottom(); y++ {
			for x := wall.area.X; x < wall.area.Right(); x++ {
				center := worldmap.At(x, y).Center()
				if !visibility.EntityVisible(center.X, center.Y, 0) {
					continue
				}
				seed := uint32(x*73856093) ^ uint32(y*19349663)
				random := func() float32 {
					seed = seed*1664525 + 1013904223
					return float32(seed>>8) / 16777216
				}
				for i := 0; i < 14; i++ {
					start := rl.NewVector2(float32(center.X)+(random()-.5)*30, float32(center.Y)+(random()-.5)*28)
					angle := float64(random()) * math.Pi * 2
					distance := 18 + random()*55
					end := rl.NewVector2(start.X+float32(math.Cos(angle))*distance, start.Y+float32(math.Sin(angle))*distance)
					size := 3 + random()*6
					duration := .5 + float64(random())*.45
					height := 12 + random()*26
					rotation := random() * 360
					spin := (random() - .5) * 500
					tone := uint8(48 + random()*35)
					if !geometry.CanOccupy(physics.Square(float64(end.X-size/2), float64(end.Y-size/2), float64(size))) {
						end = rl.NewVector2(float32(center.X)+(random()-.5)*24, float32(center.Y)+(random()-.5)*26)
					}
					progress := float32(min(1, age/duration))
					travel := 1 - (1-progress)*(1-progress)
					ground := rl.NewVector2(start.X+(end.X-start.X)*travel, start.Y+(end.Y-start.Y)*travel)
					lift := 4 * height * progress * (1 - progress)
					if age > duration && age < duration+.22 {
						lift = float32(math.Sin((age-duration)/.22*math.Pi)) * height * .15
					}
					rl.DrawEllipse(int32(ground.X), int32(ground.Y+3), size*.7, size*.35, rl.Fade(rl.Black, .45*fade))
					piece := rl.NewVector2(ground.X, ground.Y-lift)
					rl.DrawPoly(piece, 5, size*.65, rotation+spin*progress, rl.Fade(rl.NewColor(tone, tone+7, tone+14, 255), fade))
					rl.DrawLineEx(rl.NewVector2(piece.X-size*.4, piece.Y-size*.2), rl.NewVector2(piece.X+size*.2, piece.Y-size*.4), 1, rl.Fade(rl.NewColor(105, 117, 124, 255), .7*fade))
				}
				if age < .7 {
					p := float32(age / .7)
					rl.DrawCircleGradient(rl.NewVector2(float32(center.X), float32(center.Y)), 18+p*46, rl.Fade(rl.NewColor(118, 126, 131, 255), .25*(1-p)), rl.Blank)
				}
			}
		}
	}
	f.walls = active
}
