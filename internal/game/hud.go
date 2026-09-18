package game

import (
	"fmt"
	"image/color"

	"raylib-2d-sandbox/internal/protocol"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type HUD struct {
	world  *World
	window *Window
}

func NewHUD(world *World, window *Window) *HUD {
	return &HUD{world: world, window: window}
}

type playerHUD struct {
	name                  string
	avatar                color.RGBA
	level, maxLevel       int
	hp, maxHP, xp, nextXP float32
	speed                 float32
	fruit, ais, players   int
}

func (h *HUD) draw() {
	w := h.world
	s := w.player.stats
	drawPlayerHUD(playerHUD{
		name: w.player.name, avatar: w.player.color,
		level: s.lvl, maxLevel: s.maxLvl,
		hp: s.hp, maxHP: s.maxHP, xp: float32(s.exp), nextXP: calcExpForNextLvl(s.lvl),
		speed: s.speed, fruit: len(w.fruitSpawner.fruits), ais: len(w.ais),
	}, h.window.width)
}

func drawPlayerHUD(p playerHUD, windowWidth int32) {
	const x, y int32 = 0, 0
	width := min(int32(320), windowWidth-220)
	if width <= 0 {
		return
	}
	background := rl.Fade(rl.Black, .8)
	rl.DrawRectangle(x, y, width, 202, background)

	rl.DrawRectangle(x+12, y+10, 36, 36, color.RGBA{48, 48, 48, 255})
	rl.DrawRectangle(x+18, y+16, 24, 24, p.avatar)
	name := protocol.NormalizePlayerName(p.name)
	fontSize := int32(20)
	for fontSize > 10 && measureText(name, fontSize) > width-138 {
		fontSize--
	}
	drawText(name, x+58, y+28-fontSize/2, fontSize, rl.RayWhite)
	level := fmt.Sprintf("LV %d", p.level)
	drawText(level, x+width-12-measureText(level, 20), y+18, 20, rl.RayWhite)

	barWidth := width - 24
	drawHUDMeter(x+12, y+58, barWidth, "HP", fmt.Sprintf("%.2f / %.0f", p.hp, p.maxHP), p.hp, p.maxHP, color.RGBA{230, 0, 18, 255})
	xpLabel := fmt.Sprintf("%.2f / %.2f", p.xp, p.nextXP)
	xp, nextXP := p.xp, p.nextXP
	if p.level >= p.maxLevel {
		xpLabel, xp, nextXP = "MAX LEVEL", 1, 1
	}
	drawHUDMeter(x+12, y+100, barWidth, "EXP", xpLabel, xp, nextXP, rl.Yellow)
	muted := color.RGBA{160, 170, 180, 255}
	cx, cy := x+19, y+157
	rl.DrawCircleLines(cx, cy, 7, muted)
	rl.DrawLine(cx, cy, cx+4, cy-4, rl.RayWhite)
	drawText(fmt.Sprintf("SPEED %.2f", p.speed), x+34, y+148, 18, muted)

	drawText("SHIFT  Dash toward mouse (-1 HP)", x+12, y+178, 16, muted)

	counts := fmt.Sprintf("Fruit: %d    AIs: %d", p.fruit, p.ais)
	if p.players > 0 {
		counts += fmt.Sprintf("    Players: %d", p.players)
	}
	countsX := x + width
	countsWidth := measureText(counts, 18) + 24
	if countsX+countsWidth <= windowWidth-220 {
		rl.DrawRectangle(countsX, y, countsWidth, 38, background)
		drawText(counts, countsX+12, y+10, 18, rl.RayWhite)
	}
}

func drawHUDMeter(x, y, width int32, label, value string, current, maximum float32, tint color.RGBA) {
	drawText(label, x, y, 18, rl.RayWhite)
	drawText(value, x+width-measureText(value, 18), y, 18, rl.RayWhite)
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
