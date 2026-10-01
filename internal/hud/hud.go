package hud

import (
	"fmt"
	"image/color"

	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type playerView struct {
	Perks                            [simulation.PerkFamilyCount]int
	Character                        simulation.Character
	Name                             string
	Avatar                           color.RGBA
	Level, MaxLevel                  int
	HP, MaxHP, MP, MaxMP, XP, NextXP float32
	Speed                            float32
	Damage, AttackRate, Boost        float32
	Fruit, AIs, Players              int
}

func drawPlayer(p playerView, windowWidth int32) {
	const x, y int32 = 0, 0
	width := min(int32(320), windowWidth-220)
	if width <= 0 {
		return
	}
	background := rl.Fade(rl.Black, .8)
	owned := []string{}
	for family, rank := range p.Perks {
		if rank > 0 {
			owned = append(owned, fmt.Sprintf("%s %s", simulation.PerkName(simulation.PerkFamily(family)), simulation.RankName(rank)))
		}
	}
	height := int32(272)
	if len(owned) > 0 {
		height = 294 + int32((len(owned)+1)/2)*22
	}
	rl.DrawRectangle(x, y, width, height, background)

	rl.DrawRectangle(x+12, y+10, 36, 36, color.RGBA{48, 48, 48, 255})
	primitives.DrawCharacterShape(p.Character, rl.NewRectangle(float32(x+18), float32(y+16), 24, 24), p.Avatar)
	name := protocol.NormalizePlayerName(p.Name)
	fontSize := int32(20)
	for fontSize > 10 && primitives.MeasureText(name, fontSize) > width-138 {
		fontSize--
	}
	primitives.DrawText(name, x+58, y+28-fontSize/2, fontSize, rl.RayWhite)
	level := fmt.Sprintf("LV %d", p.Level)
	primitives.DrawText(level, x+width-12-primitives.MeasureText(level, 20), y+18, 20, rl.RayWhite)

	barWidth := width - 24
	drawHUDMeter(x+12, y+58, barWidth, "HP", fmt.Sprintf("%.2f / %.0f", p.HP, p.MaxHP), p.HP, p.MaxHP, color.RGBA{230, 0, 18, 255})
	drawHUDMeter(x+12, y+100, barWidth, "MP", fmt.Sprintf("%.2f / %.0f", p.MP, p.MaxMP), p.MP, p.MaxMP, rl.SkyBlue)
	xpLabel := fmt.Sprintf("%.2f / %.2f", p.XP, p.NextXP)
	xp, nextXP := p.XP, p.NextXP
	if p.Level >= p.MaxLevel {
		xpLabel, xp, nextXP = "MAX LEVEL", 1, 1
	}
	drawHUDMeter(x+12, y+142, barWidth, "EXP", xpLabel, xp, nextXP, rl.Yellow)
	muted := color.RGBA{160, 170, 180, 255}
	cx, cy := x+19, y+199
	rl.DrawCircleLines(cx, cy, 7, muted)
	rl.DrawLine(cx, cy, cx+4, cy-4, rl.RayWhite)
	primitives.DrawText(fmt.Sprintf("SPEED %.2f", p.Speed), x+34, y+190, 18, muted)
	primitives.DrawText(fmt.Sprintf("DMG %.2f   RATE %.2f/s", p.Damage, p.AttackRate), x+12, y+218, 18, muted)
	primitives.DrawText(fmt.Sprintf("SHIFT x%.2f", p.Boost), x+12, y+245, 16, muted)
	if len(owned) > 0 {
		primitives.DrawText("BLESSINGS", x+12, y+274, 14, primitives.RuneColor)
		for i, label := range owned {
			primitives.DrawText(label, x+12+int32(i%2)*(width-24)/2, y+298+int32(i/2)*22, 16, rl.LightGray)
		}
	}

	counts := fmt.Sprintf("Food: %d    AIs: %d", p.Fruit, p.AIs)
	if p.Players > 0 {
		counts += fmt.Sprintf("    Players: %d", p.Players)
	}
	countsX := x + width
	countsWidth := primitives.MeasureText(counts, 18) + 24
	if countsX+countsWidth <= windowWidth-220 {
		rl.DrawRectangle(countsX, y, countsWidth, 38, background)
		primitives.DrawText(counts, countsX+12, y+10, 18, rl.RayWhite)
	}
}

func drawHUDMeter(x, y, width int32, label, value string, current, maximum float32, tint color.RGBA) {
	primitives.DrawText(label, x, y, 18, rl.RayWhite)
	primitives.DrawText(value, x+width-primitives.MeasureText(value, 18), y, 18, rl.RayWhite)
	track := rl.NewRectangle(float32(x), float32(y+22), float32(width), 16)
	rl.DrawRectangleRec(track, color.RGBA{40, 40, 40, 255})
	fraction := float32(0)
	if maximum > 0 {
		fraction = max(float32(0), min(float32(1), current/maximum))
	}
	if fraction > 0 {
		track.Width *= fraction
		rl.DrawRectangleRec(track, tint)
	}
}

type Counts struct{ Food, AIs, Players int }

func DrawPlayer(player protocol.Player, counts Counts, width int32) {
	stats := player.Stats
	hp := float32(stats.HP + player.Split.Mass())
	if player.Dead {
		hp = 0
	}
	drawPlayer(playerView{
		Perks: stats.Perks, Character: player.Character, Name: player.Name, Avatar: primitives.CharacterColor(player.Character),
		Level: stats.Level, MaxLevel: simulation.MaxLevel, HP: hp, MaxHP: float32(stats.MaxHP()),
		MP: float32(stats.MP), MaxMP: float32(stats.MaxMP()), XP: float32(stats.XP), NextXP: float32(simulation.XPForNextLevel(stats.Level)),
		Speed:  float32(simulation.PlayerMovementSpeed(player.Character, stats.Speed, stats.HP, player.Split)),
		Damage: float32(stats.Damage), AttackRate: float32(stats.AttackRate), Boost: float32(stats.BoostMultiplier),
		Fruit: counts.Food, AIs: counts.AIs, Players: counts.Players,
	}, width)
}
