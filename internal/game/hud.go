package game

import (
	"fmt"

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

func (h *HUD) draw() {
	w := h.world
	stats := w.player.stats
	progress := fmt.Sprintf("XP %.2f / %.2f", stats.exp, calcExpForNextLvl(stats.lvl))
	if stats.lvl == stats.maxLvl {
		progress = "MAX LEVEL"
	}
	drawStatusPanel(
		fmt.Sprintf("Single player | %s | Fruit: %d | AIs: %d", protocol.NormalizePlayerName(w.player.name), len(w.fruitSpawner.fruits), len(w.ais)),
		fmt.Sprintf("Level %d | %s | HP %.2f | Speed %.2f", stats.lvl, progress, stats.hp, stats.speed),
		"WASD to move | Shift to sprint | M for map | Eat fruit and weaker AIs to grow | Esc for menu",
		h.window.width,
	)
}

func drawStatusPanel(title, stats, controls string, windowWidth int32) {
	width := windowWidth - 268
	rl.DrawRectangle(0, 0, width, 122, rl.Fade(rl.Black, .8))
	for i, line := range []string{title, stats, controls} {
		size := int32(20)
		if i == 0 {
			size = 22
		}
		for size > 10 && rl.MeasureText(line, size) > width-48 {
			size--
		}
		tint := rl.RayWhite
		if i == 2 {
			tint = rl.LightGray
		}
		rl.DrawText(line, 24, int32(20+i*32), size, tint)
	}
}
